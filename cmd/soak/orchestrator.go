// orchestrator.go — top-level soak coordinator: builds workers, correctness
// probes and memory monitoring; fans tickers over endpoints under a shared
// rate limiter; then reports aggregate results.
package main

import (
	"bufio"
	"context"
	"fmt"
	"log/slog"
	"math/rand"
	"os"
	"runtime"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/bizshuk/yfin/facade"
	"golang.org/x/time/rate"
)

// SoakConfig holds configuration for soak testing
type SoakConfig struct {
	UniverseFile  string
	Endpoints     string
	Fallback      string
	Duration      time.Duration
	Concurrency   int
	QPS           float64
	ProbeInterval time.Duration
	FailureRate   float64
	MemoryCheck   bool
	RunID         string
}

// Orchestrator manages the soak test execution
type Orchestrator struct {
	soakConfig *SoakConfig
	client     *facade.Client
	logger     *slog.Logger
	probes     *CorrectnessProbes
	memMonitor *MemoryMonitor

	// Runtime state
	tickers     []string
	endpoints   []string
	rateLimiter *rate.Limiter
	stats       *Stats

	// Control channels
	stopCh chan struct{}
	wg     sync.WaitGroup
}

// Stats tracks soak test statistics
type Stats struct {
	StartTime         time.Time
	TotalRequests     int64
	SuccessfulReqs    int64
	FailedRequests    int64
	APIRequests       int64
	ScrapeRequests    int64
	FallbackDecisions int64
	RateLimitHits     int64
	RobotsBlocked     int64
	ProbesPassed      int64
	ProbesFailed      int64

	// Per-endpoint stats
	EndpointStats map[string]*EndpointStats
}

// EndpointStats tracks per-endpoint statistics
type EndpointStats struct {
	Requests     int64
	Successes    int64
	Failures     int64
	TotalLatency time.Duration
	mu           sync.RWMutex
}

// NewOrchestrator creates a new soak test orchestrator
func NewOrchestrator(client *facade.Client, logger *slog.Logger, soakCfg *SoakConfig) (*Orchestrator, error) {
	if client == nil {
		return nil, fmt.Errorf("facade client is required")
	}
	if logger == nil {
		return nil, fmt.Errorf("logger is required")
	}
	if err := validateSoakConfig(soakCfg); err != nil {
		return nil, err
	}

	// Load ticker universe
	tickers, err := loadTickerUniverse(soakCfg.UniverseFile)
	if err != nil {
		return nil, fmt.Errorf("failed to load ticker universe: %w", err)
	}

	// Parse endpoints
	endpoints := parseEndpoints(soakCfg.Endpoints)

	// Initialize correctness probes
	probes := NewCorrectnessProbes(client, logger)

	// Initialize memory monitor
	memMonitor := NewMemoryMonitor(logger)

	// Create rate limiter
	rateLimiter := rate.NewLimiter(rate.Limit(soakCfg.QPS), max(1, int(soakCfg.QPS)+1))

	// Initialize stats
	stats := &Stats{
		StartTime:     time.Now(),
		EndpointStats: make(map[string]*EndpointStats),
	}

	// Initialize per-endpoint stats
	for _, endpoint := range endpoints {
		stats.EndpointStats[endpoint] = &EndpointStats{}
	}

	return &Orchestrator{
		soakConfig:  soakCfg,
		client:      client,
		logger:      logger,
		probes:      probes,
		memMonitor:  memMonitor,
		tickers:     tickers,
		endpoints:   endpoints,
		rateLimiter: rateLimiter,
		stats:       stats,
		stopCh:      make(chan struct{}),
	}, nil
}

// Run executes the soak test
func (o *Orchestrator) Run(ctx context.Context) error {
	o.logger.Info("Starting soak test",
		"universe_file", o.soakConfig.UniverseFile,
		"ticker_count", len(o.tickers),
		"endpoints", o.endpoints,
		"duration", o.soakConfig.Duration,
		"concurrency", o.soakConfig.Concurrency,
		"qps", o.soakConfig.QPS,
		"run_id", o.soakConfig.RunID,
	)

	// Start memory monitoring
	if o.soakConfig.MemoryCheck {
		o.wg.Add(1)
		go o.memMonitor.Monitor(ctx, &o.wg, o.stopCh)
	}

	// Start correctness probes
	o.wg.Add(1)
	go o.runCorrectnessProbes(ctx)

	// Start metrics collection
	o.wg.Add(1)
	go o.collectMetrics(ctx)

	// Start workers
	o.startWorkers(ctx)

	// Wait for duration or cancellation
	select {
	case <-time.After(o.soakConfig.Duration):
		o.logger.Info("Soak test duration completed")
	case <-ctx.Done():
		o.logger.Info("Soak test canceled")
	}

	// Stop all workers
	close(o.stopCh)
	o.wg.Wait()

	// Print final results
	o.printResults()

	return nil
}

// startWorkers initializes and starts worker goroutines
func (o *Orchestrator) startWorkers(ctx context.Context) {
	for i := 0; i < o.soakConfig.Concurrency; i++ {
		worker := NewWorker(i, o.client, o.rateLimiter, o.logger, o.stats, o.soakConfig.RunID)

		o.wg.Add(1)
		go worker.Run(ctx, &o.wg, o.stopCh, o.tickers, o.endpoints, o.soakConfig)
	}
}

// runCorrectnessProbes runs periodic correctness validation
func (o *Orchestrator) runCorrectnessProbes(ctx context.Context) {
	defer o.wg.Done()

	ticker := time.NewTicker(o.soakConfig.ProbeInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			o.runProbeRound(ctx)
		case <-o.stopCh:
			return
		case <-ctx.Done():
			return
		}
	}
}

// runProbeRound executes a round of correctness probes
func (o *Orchestrator) runProbeRound(ctx context.Context) {
	// Select random sample of tickers for probing
	sampleSize := minInt(10, len(o.tickers))
	sample := make([]string, sampleSize)

	// Random sampling without replacement
	indices := rand.Perm(len(o.tickers))
	for i := 0; i < sampleSize; i++ {
		sample[i] = o.tickers[indices[i]]
	}

	o.logger.Info("Running correctness probe round", "sample", sample)

	for _, ticker := range sample {
		if err := o.probes.ValidateTicker(ctx, ticker); err != nil {
			o.logger.Warn("Correctness probe failed", "ticker", ticker, "error", err)
			atomic.AddInt64(&o.stats.ProbesFailed, 1)
		} else {
			atomic.AddInt64(&o.stats.ProbesPassed, 1)
		}
	}
}

// collectMetrics periodically collects and logs metrics
func (o *Orchestrator) collectMetrics(ctx context.Context) {
	defer o.wg.Done()

	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			o.logCurrentMetrics()
		case <-o.stopCh:
			return
		case <-ctx.Done():
			return
		}
	}
}

// logCurrentMetrics logs current performance metrics
func (o *Orchestrator) logCurrentMetrics() {
	elapsed := time.Since(o.stats.StartTime)
	totalReqs := atomic.LoadInt64(&o.stats.TotalRequests)
	successReqs := atomic.LoadInt64(&o.stats.SuccessfulReqs)
	failedReqs := atomic.LoadInt64(&o.stats.FailedRequests)

	successRate := float64(0)
	if totalReqs > 0 {
		successRate = float64(successReqs) / float64(totalReqs) * 100
	}

	qps := float64(totalReqs) / elapsed.Seconds()

	o.logger.Info("Soak test metrics",
		"elapsed", elapsed,
		"total_requests", totalReqs,
		"successful_requests", successReqs,
		"failed_requests", failedReqs,
		"success_rate", successRate,
		"actual_qps", qps,
		"goroutines", runtime.NumGoroutine(),
	)
}

// printResults prints comprehensive soak test results
func (o *Orchestrator) printResults() {
	elapsed := time.Since(o.stats.StartTime)
	totalReqs := atomic.LoadInt64(&o.stats.TotalRequests)
	successReqs := atomic.LoadInt64(&o.stats.SuccessfulReqs)
	failedReqs := atomic.LoadInt64(&o.stats.FailedRequests)

	fmt.Printf("\n=== SOAK TEST RESULTS ===\n")
	fmt.Printf("Duration: %v\n", elapsed)
	fmt.Printf("Total Requests: %d\n", totalReqs)
	fmt.Printf("Successful Requests: %d\n", successReqs)
	fmt.Printf("Failed Requests: %d\n", failedReqs)

	if totalReqs > 0 {
		successRate := float64(successReqs) / float64(totalReqs) * 100
		fmt.Printf("Success Rate: %.2f%%\n", successRate)
		fmt.Printf("Actual QPS: %.2f\n", float64(totalReqs)/elapsed.Seconds())
	}

	fmt.Printf("API Requests: %d\n", atomic.LoadInt64(&o.stats.APIRequests))
	fmt.Printf("Scrape Requests: %d\n", atomic.LoadInt64(&o.stats.ScrapeRequests))
	fmt.Printf("Fallback Decisions: %d\n", atomic.LoadInt64(&o.stats.FallbackDecisions))
	fmt.Printf("Rate Limit Hits: %d\n", atomic.LoadInt64(&o.stats.RateLimitHits))
	fmt.Printf("Robots Blocked: %d\n", atomic.LoadInt64(&o.stats.RobotsBlocked))

	probesPassed := atomic.LoadInt64(&o.stats.ProbesPassed)
	probesFailed := atomic.LoadInt64(&o.stats.ProbesFailed)
	fmt.Printf("Correctness Probes Passed: %d\n", probesPassed)
	fmt.Printf("Correctness Probes Failed: %d\n", probesFailed)

	if o.soakConfig.MemoryCheck {
		summary := o.memMonitor.Summary()
		analysis := o.memMonitor.AnalyzeLeaks()

		fmt.Printf("\n=== MEMORY ANALYSIS ===\n")
		fmt.Printf("Initial Memory: %d KB\n", summary.InitialMemory/1024)
		fmt.Printf("Peak Memory: %d KB\n", summary.PeakMemory/1024)
		fmt.Printf("Final Memory: %d KB\n", summary.Final.AllocBytes/1024)
		fmt.Printf("Memory Growth: %d KB\n", summary.MemoryGrowthBytes()/1024)
		fmt.Printf("Initial Goroutines: %d\n", summary.InitialGoroutines)
		fmt.Printf("Peak Goroutines: %d\n", summary.PeakGoroutines)
		fmt.Printf("Final Goroutines: %d\n", summary.Final.Goroutines)
		fmt.Printf("Goroutine Growth: %d\n", summary.GoroutineGrowth())
		fmt.Printf("Leak Analysis: %s\n", analysis.Recommendation)
		for _, detail := range analysis.Details {
			fmt.Printf("- %s\n", detail)
		}
	}

	fmt.Printf("\n=== ENDPOINT BREAKDOWN ===\n")
	for _, endpoint := range o.endpoints {
		stats := o.stats.EndpointStats[endpoint]
		stats.mu.RLock()
		avgLatency := time.Duration(0)
		if stats.Requests > 0 {
			avgLatency = time.Duration(stats.TotalLatency.Nanoseconds() / stats.Requests)
		}
		fmt.Printf("%s: %d requests, %d successes, %d failures, avg latency: %v\n",
			endpoint, stats.Requests, stats.Successes, stats.Failures, avgLatency)
		stats.mu.RUnlock()
	}
}

// loadTickerUniverse loads tickers from a file
func loadTickerUniverse(filename string) ([]string, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to open universe file: %w", err)
	}
	defer file.Close()

	var tickers []string
	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line != "" && !strings.HasPrefix(line, "#") {
			tickers = append(tickers, line)
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("failed to read universe file: %w", err)
	}

	if len(tickers) == 0 {
		return nil, fmt.Errorf("no tickers found in universe file")
	}

	return tickers, nil
}

// parseEndpoints parses comma-separated endpoint list
func parseEndpoints(endpointsStr string) []string {
	var endpoints []string
	for _, endpoint := range strings.Split(endpointsStr, ",") {
		endpoint = strings.TrimSpace(endpoint)
		if endpoint != "" {
			endpoints = append(endpoints, endpoint)
		}
	}
	return endpoints
}

// min returns the minimum of two integers
func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}
