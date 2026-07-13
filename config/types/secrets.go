// secrets.go — placeholder secret references. The `Ref` field is a
// secret-store handle (env var, vault path, etc.) — actual secret
// values are never written to YAML. Capacity: 1 struct (`SecretConfig`).
package types

// SecretConfig represents secret configuration
type SecretConfig struct {
	Name     string `yaml:"name"`
	Ref      string `yaml:"ref"`
	Required bool   `yaml:"required"`
}
