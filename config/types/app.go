// app.go — application-level runtime identity: environment name + run ID.
// Capacity: 1 struct (`AppConfig`).
package types

// AppConfig represents application-level configuration
type AppConfig struct {
	Env   string `yaml:"env"`
	RunID string `yaml:"run_id"`
}
