package config

import (
	"fmt"
	"strings"

	"github.com/spf13/viper"
)

// Config holds all configuration for the application
type Config struct {
	Environment string          `mapstructure:"ENVIRONMENT"`
	LogLevel    string          `mapstructure:"LOG_LEVEL"`
	MongoDB     MongoDBConfig   `mapstructure:",squash"`
	Redis       RedisConfig     `mapstructure:",squash"`
	RabbitMQ    RabbitMQConfig  `mapstructure:",squash"`
	JWT         JWTConfig       `mapstructure:",squash"`
	SMTP        SMTPConfig      `mapstructure:",squash"`
	OAuth       OAuthConfig     `mapstructure:",squash"`
	RateLimit   RateLimitConfig `mapstructure:",squash"`
	CORS        CORSConfig      `mapstructure:",squash"`
}

// MongoDBConfig contains MongoDB configuration
type MongoDBConfig struct {
	URI         string `mapstructure:"MONGODB_URI"`
	Database    string `mapstructure:"MONGODB_DATABASE"`
	MaxPoolSize uint64 `mapstructure:"MONGODB_MAX_POOL_SIZE"`
	MinPoolSize uint64 `mapstructure:"MONGODB_MIN_POOL_SIZE"`
}

// RedisConfig contains Redis configuration
type RedisConfig struct {
	Host     string `mapstructure:"REDIS_HOST"`
	Port     string `mapstructure:"REDIS_PORT"`
	Password string `mapstructure:"REDIS_PASSWORD"`
	DB       int    `mapstructure:"REDIS_DB"`
}

// RabbitMQConfig contains RabbitMQ configuration
type RabbitMQConfig struct {
	URL         string `mapstructure:"RABBITMQ_URL"`
	Exchange    string `mapstructure:"RABBITMQ_EXCHANGE"`
	QueuePrefix string `mapstructure:"RABBITMQ_QUEUE_PREFIX"`
}

// JWTConfig contains JWT configuration
type JWTConfig struct {
	Secret            string `mapstructure:"JWT_SECRET"`
	Expiration        int    `mapstructure:"JWT_EXPIRATION"`
	RefreshExpiration int    `mapstructure:"JWT_REFRESH_EXPIRATION"`
}

// SMTPConfig contains SMTP configuration
type SMTPConfig struct {
	Host      string `mapstructure:"SMTP_HOST"`
	Port      int    `mapstructure:"SMTP_PORT"`
	Username  string `mapstructure:"SMTP_USERNAME"`
	Password  string `mapstructure:"SMTP_PASSWORD"`
	FromEmail string `mapstructure:"SMTP_FROM_EMAIL"`
	FromName  string `mapstructure:"SMTP_FROM_NAME"`
}

// OAuthConfig contains OAuth configuration
type OAuthConfig struct {
	GoogleClientID     string `mapstructure:"GOOGLE_CLIENT_ID"`
	GoogleClientSecret string `mapstructure:"GOOGLE_CLIENT_SECRET"`
	GoogleRedirectURL  string `mapstructure:"GOOGLE_REDIRECT_URL"`
	GitHubClientID     string `mapstructure:"GITHUB_CLIENT_ID"`
	GitHubClientSecret string `mapstructure:"GITHUB_CLIENT_SECRET"`
	GitHubRedirectURL  string `mapstructure:"GITHUB_REDIRECT_URL"`
}

// RateLimitConfig contains rate limiting configuration
type RateLimitConfig struct {
	RequestsPerSecond float64 `mapstructure:"RATE_LIMIT_REQUESTS_PER_SECOND"`
	Burst             int     `mapstructure:"RATE_LIMIT_BURST"`
}

// CORSConfig contains CORS configuration
type CORSConfig struct {
	AllowedOrigins string `mapstructure:"CORS_ALLOWED_ORIGINS"`
	AllowedMethods string `mapstructure:"CORS_ALLOWED_METHODS"`
	AllowedHeaders string `mapstructure:"CORS_ALLOWED_HEADERS"`
}

// LoadConfig loads configuration from environment variables and .env file
func LoadConfig() (*Config, error) {
	return Load()
}

// Load loads configuration from environment variables and .env file
func Load() (*Config, error) {
	v := viper.New()

	// Set default values
	v.SetDefault("ENVIRONMENT", "development")
	v.SetDefault("LOG_LEVEL", "info")
	v.SetDefault("MONGODB_DATABASE", "saas_framework")
	v.SetDefault("MONGODB_MAX_POOL_SIZE", 100)
	v.SetDefault("MONGODB_MIN_POOL_SIZE", 10)
	v.SetDefault("REDIS_PORT", "6379")
	v.SetDefault("REDIS_DB", 0)
	v.SetDefault("RABBITMQ_EXCHANGE", "saas_events")
	v.SetDefault("RABBITMQ_QUEUE_PREFIX", "saas")
	v.SetDefault("JWT_EXPIRATION", 3600)
	v.SetDefault("JWT_REFRESH_EXPIRATION", 604800)
	v.SetDefault("RATE_LIMIT_REQUESTS_PER_SECOND", 100)
	v.SetDefault("RATE_LIMIT_BURST", 200)
	v.SetDefault("SMTP_PORT", 587)

	// Load from .env file if it exists
	v.SetConfigFile(".env")
	v.SetConfigType("env")
	if err := v.ReadInConfig(); err != nil {
		// It's okay if .env file doesn't exist
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("error reading config file: %w", err)
		}
	}

	// Override with environment variables
	v.AutomaticEnv()
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	var config Config
	if err := v.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("error unmarshaling config: %w", err)
	}

	return &config, nil
}

// GetRedisAddr returns the full Redis address
func (c *RedisConfig) GetRedisAddr() string {
	return fmt.Sprintf("%s:%s", c.Host, c.Port)
}
