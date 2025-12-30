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
	
	// Infrastructure
	MongoDB     MongoDBConfig   `mapstructure:",squash"`
	Redis       RedisConfig     `mapstructure:",squash"`
	RabbitMQ    RabbitMQConfig  `mapstructure:",squash"`
	
	// Auth & Security
	JWT         JWTConfig       `mapstructure:",squash"`
	OAuth       OAuthConfig     `mapstructure:",squash"`
	
	// Communication
	SMTP        SMTPConfig      `mapstructure:",squash"`
	SMS         SMSConfig       `mapstructure:",squash"`
	
	// Service-specific
	Email       EmailConfig     `mapstructure:",squash"`
	Server      ServerConfig    `mapstructure:",squash"`
	ServiceURLs ServiceURLsConfig `mapstructure:",squash"`
	
	// Cross-cutting concerns
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
	PoolSize  int    `mapstructure:"SMTP_POOL_SIZE"`
}

// SMSConfig contains SMS provider configuration
type SMSConfig struct {
	Provider    string `mapstructure:"SMS_PROVIDER"`
	TwilioSID   string `mapstructure:"TWILIO_SID"`
	TwilioToken string `mapstructure:"TWILIO_TOKEN"`
	TwilioFrom  string `mapstructure:"TWILIO_FROM"`
	AWSSNSARN   string `mapstructure:"AWS_SNS_ARN"`
	AWSRegion   string `mapstructure:"AWS_REGION"`
}

// EmailConfig contains email service configuration
type EmailConfig struct {
	Workers int `mapstructure:"EMAIL_WORKERS"`
}

// ServerConfig contains server configuration for all services
type ServerConfig struct {
	Port            string `mapstructure:"SERVER_PORT"`
	HTTPPort        string `mapstructure:"HTTP_PORT"`
	GRPCPort        string `mapstructure:"GRPC_PORT"`
	ShutdownTimeout int    `mapstructure:"SHUTDOWN_TIMEOUT"`
}

// ServiceURLsConfig contains URLs for all microservices (used by API Gateway)
type ServiceURLsConfig struct {
	AuthService         string `mapstructure:"AUTH_SERVICE_URL"`
	UserService         string `mapstructure:"USER_SERVICE_URL"`
	TenantService       string `mapstructure:"TENANT_SERVICE_URL"`
	NotificationService string `mapstructure:"NOTIFICATION_SERVICE_URL"`
	SystemConfigService string `mapstructure:"SYSTEM_CONFIG_SERVICE_URL"`
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
	v.SetDefault("SMTP_POOL_SIZE", 10)
	
	// SMS defaults
	v.SetDefault("SMS_PROVIDER", "twilio")
	v.SetDefault("AWS_REGION", "us-east-1")
	
	// Email defaults
	v.SetDefault("EMAIL_WORKERS", 5)
	
	// Server defaults
	v.SetDefault("SERVER_PORT", "8080")
	v.SetDefault("HTTP_PORT", "8080")
	v.SetDefault("GRPC_PORT", "50051")
	v.SetDefault("SHUTDOWN_TIMEOUT", 10)
	
	// Service URLs defaults (for local development)
	v.SetDefault("AUTH_SERVICE_URL", "http://localhost:8081")
	v.SetDefault("USER_SERVICE_URL", "http://localhost:8082")
	v.SetDefault("TENANT_SERVICE_URL", "http://localhost:8083")
	v.SetDefault("NOTIFICATION_SERVICE_URL", "http://localhost:8084")
	v.SetDefault("SYSTEM_CONFIG_SERVICE_URL", "http://localhost:8085")

	// Load from .env file if it exists
	v.SetConfigFile(".env")
	v.SetConfigType("env")
	_ = v.ReadInConfig() // Ignore error - it's okay if .env file doesn't exist

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
