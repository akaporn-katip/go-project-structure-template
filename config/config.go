package config

import (
	"fmt"
	"time"

	"github.com/spf13/viper"
)

type Config struct {
	Server        ServerConfig        `mapstructure:"server"`
	Database      DatabaseConfig      `mapstructure:"database"`
	Observability ObservabilityConfig `mapstructure:"observability"`
	// JWT      JWTConfig
	// Redis    RedisConfig
}

type ServerConfig struct {
	Port string
	Env  string
}

type PostgresConfig struct {
	DSN          string `mapstructure:"dsn"`
	MaxOpenConns int    `mapstructure:"max_open_conns"`
	MaxIdleConns int    `mapstructure:"max_idle_conns"`
}

type MongoConfig struct {
	URI    string `mapstructure:"uri"`
	DBName string `mapstructure:"db_name"`
}

type DatabaseConfig struct {
	Type     string         `mapstructure:"type"`
	Postgres PostgresConfig `mapstructure:"postgres"`
	MongoDB  MongoConfig    `mapstructure:"mongodb"`
}

type ObservabilityConfig struct {
	Enabled         bool          `mapstructure:"enabled"`
	OTLPProtocol    string        `mapstructure:"otlp_protocol"`
	OTLPEndpoint    string        `mapstructure:"otlp_endpoint"`
	ServiceName     string        `mapstructure:"service_name"`
	ServiceVersion  string        `mapstructure:"service_version"`
	Environment     string        `mapstructure:"environment"`
	MetricsInterval time.Duration `mapstructure:"metrics_interval"`
	TraceSampleRate float64       `mapstructure:"trace_sample_rate"`
	EnableTracing   bool          `mapstructure:"enable_tracing"`
	EnableMetrics   bool          `mapstructure:"enable_metrics"`
	EnableLogging   bool          `mapstructure:"enable_logging"`
	LogLevel        string        `mapstructure:"log_level"`
}

func Load() (*Config, error) {
	return LoadWithPath("./config")
}

func LoadWithPath(path string) (*Config, error) {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(path)
	viper.AddConfigPath(".")
	viper.AddConfigPath("../config")
	viper.AddConfigPath("../../config")
	setDefaults()
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		return nil, err
	}

	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, err
	}

	if err := validate(&cfg); err != nil {
		return nil, fmt.Errorf("config validation failed: %w", err)
	}

	return &cfg, nil
}

func setDefaults() {
	viper.SetDefault("observability.enabled", true)
	viper.SetDefault("observability.otlp_endpoint", "localhost:4317")
	viper.SetDefault("observability.service_name", "myapi")
	viper.SetDefault("observability.service_version", "1.0.0")
	viper.SetDefault("observability.environment", "development")
	viper.SetDefault("observability.metrics_interval", 60*time.Second)
	viper.SetDefault("observability.trace_sample_rate", 1.0)
	viper.SetDefault("observability.enable_tracing", true)
	viper.SetDefault("observability.enable_metrics", true)
	viper.SetDefault("observability.enable_logging", true)
	viper.SetDefault("observability.log_level", "info")

	// database
	viper.SetDefault("database.type", "")
	viper.SetDefault("database.postgres.dsn", "")
	viper.SetDefault("database.postgres.max_open_conns", 25)
	viper.SetDefault("database.postgres.max_idle_conns", 5)
	viper.SetDefault("database.mongodb.uri", "")
}

func validate(cfg *Config) error {
	if cfg.Observability.Enabled {
		if cfg.Observability.OTLPEndpoint == "" {
			return fmt.Errorf("observability.otlp_endpoint is required when observability is enabled")
		}
		if cfg.Observability.TraceSampleRate < 0 || cfg.Observability.TraceSampleRate > 1 {
			return fmt.Errorf("observability.trace_sample_rate must be between 0 and 1")
		}
	}

	if cfg.Database.Type == "postgres" {
		if cfg.Database.Postgres.DSN == "" {
			return fmt.Errorf("database.postgres.dsn is required")
		}
	}

	if cfg.Database.Type == "mongodb" {
		if cfg.Database.MongoDB.URI == "" {
			return fmt.Errorf("database.mongodb.uri is required")
		}

		if cfg.Database.MongoDB.DBName == "" {
			return fmt.Errorf("database.mongodb.db_name is required")
		}
	}

	return nil
}

func (c *Config) String() string {
	return fmt.Sprintf(`
		  Observability: {Enabled: %v, Endpoint: %s, Service: %s, Env: %s}
	`,
		c.Observability.Enabled,
		c.Observability.OTLPEndpoint,
		c.Observability.ServiceName,
		c.Observability.Environment,
	)
}
