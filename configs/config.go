package configs

import (
	"strings"

	"github.com/spf13/viper"
)

// Config holds all application configuration.
type Config struct {
	App           AppConfig           `mapstructure:"app"`
	Database      DatabaseConfig      `mapstructure:"database"`
	Redis         RedisConfig         `mapstructure:"redis"`
	Kafka         KafkaConfig         `mapstructure:"kafka"`
	JWT           JWTConfig           `mapstructure:"jwt"`
	RateLimit     RateLimitConfig     `mapstructure:"rate_limit"`
	Observability ObservabilityConfig `mapstructure:"observability"`
	Proxy         ProxyConfig         `mapstructure:"proxy"`
}

type AppConfig struct {
	Name  string `mapstructure:"name"`
	Env   string `mapstructure:"env"`
	Port  int    `mapstructure:"port"`
	Debug bool   `mapstructure:"debug"`
}

type DatabaseConfig struct {
	Host            string `mapstructure:"host"`
	Port            int    `mapstructure:"port"`
	User            string `mapstructure:"user"`
	Password        string `mapstructure:"password"`
	Name            string `mapstructure:"name"`
	SSLMode         string `mapstructure:"sslmode"`
	MaxOpenConns    int    `mapstructure:"max_open_conns"`
	MaxIdleConns    int    `mapstructure:"max_idle_conns"`
	ConnMaxLifetime int    `mapstructure:"conn_max_lifetime"`
}

type RedisConfig struct {
	Host         string `mapstructure:"host"`
	Port         int    `mapstructure:"port"`
	Password     string `mapstructure:"password"`
	DB           int    `mapstructure:"db"`
	PoolSize     int    `mapstructure:"pool_size"`
	DialTimeout  int    `mapstructure:"dial_timeout"`
	ReadTimeout  int    `mapstructure:"read_timeout"`
	WriteTimeout int    `mapstructure:"write_timeout"`
}

type KafkaConfig struct {
	Brokers  []string            `mapstructure:"brokers"`
	GroupID  string              `mapstructure:"group_id"`
	Topics   KafkaTopicsConfig   `mapstructure:"topics"`
	Producer KafkaProducerConfig `mapstructure:"producer"`
	Consumer KafkaConsumerConfig `mapstructure:"consumer"`
}

type KafkaTopicsConfig struct {
	Signin string `mapstructure:"signin"`
	Signup string `mapstructure:"signup"`
	Retry  string `mapstructure:"retry"`
	DLQ    string `mapstructure:"dlq"`
}

type KafkaProducerConfig struct {
	BatchSize    int `mapstructure:"batch_size"`
	BatchTimeout int `mapstructure:"batch_timeout"`
	RequiredAcks int `mapstructure:"required_acks"`
}

type KafkaConsumerConfig struct {
	MinBytes int `mapstructure:"min_bytes"`
	MaxBytes int `mapstructure:"max_bytes"`
	MaxWait  int `mapstructure:"max_wait"`
}

type JWTConfig struct {
	Secret     string `mapstructure:"secret"`
	AccessTTL  int    `mapstructure:"access_ttl"`
	RefreshTTL int    `mapstructure:"refresh_ttl"`
}

type RateLimitConfig struct {
	Requests int `mapstructure:"requests"`
	Window   int `mapstructure:"window"`
}

type ObservabilityConfig struct {
	PrometheusPort int    `mapstructure:"prometheus_port"`
	LogLevel       string `mapstructure:"log_level"`
}

type ProxyConfig struct {
	Sampling string `mapstructure:"sampling"`
	Billing  string `mapstructure:"billing"`
}

// Load reads configuration from file and environment variables.
// Environment variables override file values using the UDIP_ prefix.
// Example: UDIP_DATABASE_HOST overrides database.host
func Load(configPath string) (*Config, error) {
	v := viper.New()

	v.SetConfigFile(configPath)
	v.SetConfigType("yaml")

	// Environment variable overrides — prefix UDIP_
	v.SetEnvPrefix("UDIP")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	if err := v.ReadInConfig(); err != nil {
		return nil, err
	}

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}
