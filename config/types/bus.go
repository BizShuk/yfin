// bus.go — message-bus publisher settings (NATS or Kafka), with
// per-publisher retry + circuit-breaker tuning. Capacity: 4 structs
// (`BusConfig`, `PublisherConfig`, `NATSConfig`, `KafkaConfig`).
package types

// BusConfig represents bus configuration
type BusConfig struct {
	Enabled         bool                 `yaml:"enabled"`
	Env             string               `yaml:"env"`
	TopicPrefix     string               `yaml:"topic_prefix"`
	MaxPayloadBytes int64                `yaml:"max_payload_bytes"`
	Publisher       PublisherConfig      `yaml:"publisher"`
	Retry           RetryConfig          `yaml:"retry"`
	CircuitBreaker  CircuitBreakerConfig `yaml:"circuit_breaker"`
}

// PublisherConfig represents publisher configuration
type PublisherConfig struct {
	Backend string      `yaml:"backend"`
	NATS    NATSConfig  `yaml:"nats"`
	Kafka   KafkaConfig `yaml:"kafka"`
}

// NATSConfig represents NATS configuration
type NATSConfig struct {
	URL          string `yaml:"url"`
	SubjectStyle string `yaml:"subject_style"`
	AckWaitMs    int    `yaml:"ack_wait_ms"`
}

// KafkaConfig represents Kafka configuration
type KafkaConfig struct {
	Brokers     []string `yaml:"brokers"`
	Acks        string   `yaml:"acks"`
	Compression string   `yaml:"compression"`
}
