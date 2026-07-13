// sessions.go — per-session cookie/crumb rotation knobs. Vestigial in
// current builds (session rotation was removed; see CLAUDE.md), kept
// here for YAML backwards-compatibility only. Capacity: 1 struct
// (`SessionsConfig`).
package types

// SessionsConfig represents session rotation configuration
type SessionsConfig struct {
	N                  int `yaml:"n"`
	EjectAfter         int `yaml:"eject_after"`
	RecreateCooldownMs int `yaml:"recreate_cooldown_ms"`
}
