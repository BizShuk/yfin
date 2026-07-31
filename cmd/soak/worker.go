// worker.go — rate-limited request dispatch for api-only, scrape-only and the
// bounded auto-fallback policy.
package main

import (
	"context"
	"fmt"
	"log/slog"
	"math/rand"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/bizshuk/yfin/facade"
	"golang.org/x/time/rate"
)

// Worker represents a soak test worker
type Worker struct {
	id          int
	client      *facade.Client
	rateLimiter *rate.Limiter
	logger      *slog.Logger
	stats       *Stats
	rng         *rand.Rand
	runID       string
}

// WorkRequest represents a work item for a worker
type WorkRequest struct {
	Ticker   string
	Endpoint string
	RunID    string
}

// NewWorker creates a new soak test worker
func NewWorker(id int, client *facade.Client, rateLimiter *rate.Limiter, logger *slog.Logger, stats *Stats, runID string) *Worker {
	return &Worker{
		id:          id,
		client:      client,
		rateLimiter: rateLimiter,
		logger:      logger,
		stats:       stats,
		rng:         rand.New(rand.NewSource(time.Now().UnixNano() + int64(id))),
		runID:       runID,
	}
}

// Run starts the worker loop
func (w *Worker) Run(ctx context.Context, wg *sync.WaitGroup, stopCh <-chan struct{}, tickers []string, endpoints []string, config *SoakConfig) {
	defer wg.Done()

	w.logger.Debug("Worker started", "worker_id", w.id)
	defer w.logger.Debug("Worker stopped", "worker_id", w.id)

	for {
		select {
		case <-stopCh:
			return
		case <-ctx.Done():
			return
		default:
			// Wait for rate limiter
			if err := w.rateLimiter.Wait(ctx); err != nil {
				if ctx.Err() != nil {
					return
				}
				w.logger.Warn("Rate limiter error", "error", err)
				continue
			}

			// Generate random work request
			request := w.generateWorkRequest(tickers, endpoints)

			// Execute request
			w.executeRequest(ctx, request, config)
		}
	}
}

// generateWorkRequest creates a random work request
func (w *Worker) generateWorkRequest(tickers []string, endpoints []string) WorkRequest {
	ticker := tickers[w.rng.Intn(len(tickers))]
	endpoint := endpoints[w.rng.Intn(len(endpoints))]
	return WorkRequest{
		Ticker:   ticker,
		Endpoint: endpoint,
		RunID:    w.runID,
	}
}

// executeRequest executes a single work request
func (w *Worker) executeRequest(ctx context.Context, req WorkRequest, config *SoakConfig) {
	startTime := time.Now()

	// Update total requests counter
	atomic.AddInt64(&w.stats.TotalRequests, 1)

	// Get endpoint stats
	endpointStats := w.stats.EndpointStats[req.Endpoint]
	endpointStats.mu.Lock()
	endpointStats.Requests++
	endpointStats.mu.Unlock()

	// Execute based on fallback strategy
	var err error
	var fallbackUsed bool

	switch config.Fallback {
	case "api-only":
		err = w.executeAPIRequest(ctx, req)
		atomic.AddInt64(&w.stats.APIRequests, 1)
	case "scrape-only":
		err = w.executeScrapeRequest(ctx, req)
		atomic.AddInt64(&w.stats.ScrapeRequests, 1)
	case "auto":
		err, fallbackUsed = w.executeAutoFallbackRequest(ctx, req)
		if fallbackUsed {
			atomic.AddInt64(&w.stats.FallbackDecisions, 1)
		}
	default:
		err = fmt.Errorf("unknown fallback strategy: %s", config.Fallback)
	}

	// Record execution time
	duration := time.Since(startTime)
	endpointStats.mu.Lock()
	endpointStats.TotalLatency += duration
	endpointStats.mu.Unlock()

	// Update success/failure counters
	if err != nil {
		atomic.AddInt64(&w.stats.FailedRequests, 1)
		endpointStats.mu.Lock()
		endpointStats.Failures++
		endpointStats.mu.Unlock()

		w.logger.Debug("Request failed",
			"worker_id", w.id,
			"ticker", req.Ticker,
			"endpoint", req.Endpoint,
			"duration", duration,
			"error", err,
		)

		// Check for specific error types
		if isRateLimitError(err) {
			atomic.AddInt64(&w.stats.RateLimitHits, 1)
		}
		if isRobotsError(err) {
			atomic.AddInt64(&w.stats.RobotsBlocked, 1)
		}
	} else {
		atomic.AddInt64(&w.stats.SuccessfulReqs, 1)
		endpointStats.mu.Lock()
		endpointStats.Successes++
		endpointStats.mu.Unlock()

		w.logger.Debug("Request succeeded",
			"worker_id", w.id,
			"ticker", req.Ticker,
			"endpoint", req.Endpoint,
			"duration", duration,
			"fallback_used", fallbackUsed,
		)
	}
}

// executeAPIRequest executes an API-only request
func (w *Worker) executeAPIRequest(ctx context.Context, req WorkRequest) error {
	switch req.Endpoint {
	case "quote":
		_, err := w.client.FetchQuote(ctx, req.Ticker, req.RunID)
		return err
	case "daily-bars":
		end := time.Now()
		start := end.AddDate(0, 0, -30) // Last 30 days
		_, err := w.client.FetchDailyBars(ctx, req.Ticker, start, end, true, req.RunID)
		return err
	case "fundamentals":
		_, err := w.client.FetchFundamentalsQuarterly(ctx, req.Ticker, req.RunID)
		return err
	default:
		return fmt.Errorf("API endpoint not supported: %s", req.Endpoint)
	}
}

// executeScrapeRequest executes a scrape-only request
func (w *Worker) executeScrapeRequest(ctx context.Context, req WorkRequest) error {
	switch req.Endpoint {
	case "key-statistics":
		_, err := w.client.ScrapeKeyStatistics(ctx, req.Ticker, req.RunID)
		return err
	case "financials":
		_, err := w.client.ScrapeFinancials(ctx, req.Ticker, req.RunID)
		return err
	case "analysis":
		_, err := w.client.ScrapeAnalysis(ctx, req.Ticker, req.RunID)
		return err
	case "profile":
		_, err := w.client.ScrapeProfileData(ctx, req.Ticker, req.RunID)
		return err
	case "news":
		_, err := w.client.ScrapeNews(ctx, req.Ticker, req.RunID)
		return err
	case "balance-sheet":
		_, err := w.client.ScrapeBalanceSheet(ctx, req.Ticker, req.RunID)
		return err
	case "cash-flow":
		_, err := w.client.ScrapeCashFlow(ctx, req.Ticker, req.RunID)
		return err
	case "analyst-insights":
		_, err := w.client.ScrapeAnalystInsights(ctx, req.Ticker, req.RunID)
		return err
	default:
		return fmt.Errorf("scrape endpoint not supported: %s", req.Endpoint)
	}
}

// executeAutoFallbackRequest executes with automatic fallback
func (w *Worker) executeAutoFallbackRequest(ctx context.Context, req WorkRequest) (error, bool) {
	// First try API if available
	if w.isAPIEndpoint(req.Endpoint) {
		err := w.executeAPIRequest(ctx, req)
		atomic.AddInt64(&w.stats.APIRequests, 1)
		if err == nil {
			return nil, false
		}

		fallbackEndpoint, supported := scrapeFallbackEndpoint(req.Endpoint)
		if supported && shouldFallbackToScrape(err) {
			w.logger.Debug("Falling back to scrape",
				"ticker", req.Ticker,
				"api_endpoint", req.Endpoint,
				"scrape_endpoint", fallbackEndpoint,
				"error", err,
			)

			fallbackRequest := req
			fallbackRequest.Endpoint = fallbackEndpoint
			scrapeErr := w.executeScrapeRequest(ctx, fallbackRequest)
			atomic.AddInt64(&w.stats.ScrapeRequests, 1)
			return scrapeErr, true
		}

		return err, false
	}

	// For scrape-only endpoints, go directly to scraping
	err := w.executeScrapeRequest(ctx, req)
	atomic.AddInt64(&w.stats.ScrapeRequests, 1)
	return err, false
}

// isAPIEndpoint checks if an endpoint has API support
func (w *Worker) isAPIEndpoint(endpoint string) bool {
	_, supported := apiSoakEndpoints[endpoint]
	return supported
}

func scrapeFallbackEndpoint(apiEndpoint string) (string, bool) {
	if apiEndpoint == "fundamentals" {
		return "financials", true
	}
	return "", false
}

// shouldFallbackToScrape determines if an equivalent scrape surface should
// be attempted after a retryable API failure.
func shouldFallbackToScrape(err error) bool {
	if err == nil {
		return false
	}

	errStr := strings.ToLower(err.Error())

	// Fallback on authentication errors (paid features)
	if contains(errStr, "401") || contains(errStr, "authentication") || contains(errStr, "subscription") {
		return true
	}

	// Fallback on rate limiting
	if contains(errStr, "429") || contains(errStr, "rate limit") {
		return true
	}

	// Fallback on server errors
	if contains(errStr, "500") || contains(errStr, "502") || contains(errStr, "503") {
		return true
	}

	return false
}

// isRateLimitError checks if an error is due to rate limiting
func isRateLimitError(err error) bool {
	if err == nil {
		return false
	}
	errStr := err.Error()
	return contains(errStr, "429") || contains(errStr, "rate limit") || contains(errStr, "too many requests")
}

// isRobotsError checks if an error is due to robots.txt blocking
func isRobotsError(err error) bool {
	if err == nil {
		return false
	}
	errStr := err.Error()
	return contains(errStr, "robots") || contains(errStr, "disallowed")
}

// contains checks if a string contains a substring (case-insensitive)
func contains(s, substr string) bool {
	return strings.Contains(strings.ToLower(s), strings.ToLower(substr))
}
