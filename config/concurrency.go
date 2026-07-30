// concurrency.go — global + per-host worker counts that govern Yahoo
// Finance fetch parallelism. Capacity: 1 struct (`ConcurrencyConfig`).
package config

const (
	defaultGlobalWorkers  = 64
	defaultPerHostWorkers = 32
)

// ConcurrencyConfig represents concurrency configuration
type ConcurrencyConfig struct {
	GlobalWorkers  int `yaml:"global_workers"`
	PerHostWorkers int `yaml:"per_host_workers"`
}

func applyConcurrencyDefaults(configMap map[string]interface{}, target *ConcurrencyConfig) {
	concurrencyMap, ok := configMap["concurrency"].(map[string]interface{})
	if !ok {
		target.GlobalWorkers = defaultGlobalWorkers
		target.PerHostWorkers = defaultPerHostWorkers
		return
	}
	if _, exists := concurrencyMap["global_workers"]; !exists {
		target.GlobalWorkers = defaultGlobalWorkers
	}
	if _, exists := concurrencyMap["per_host_workers"]; !exists {
		target.PerHostWorkers = defaultPerHostWorkers
	}
}
