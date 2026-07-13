// concurrency.go — global + per-host worker counts that govern Yahoo
// Finance fetch parallelism. Capacity: 1 struct (`ConcurrencyConfig`).
package types

// ConcurrencyConfig represents concurrency configuration
type ConcurrencyConfig struct {
	GlobalWorkers  int `yaml:"global_workers"`
	PerHostWorkers int `yaml:"per_host_workers"`
}
