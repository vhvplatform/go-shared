package config

import (
	"os"
	"testing"
)

func TestLoadConfig(t *testing.T) {
	// Test loading config with defaults
	cfg, err := LoadConfig()
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	// Verify defaults are set
	if cfg.Environment != "development" {
		t.Errorf("Expected Environment to be 'development', got '%s'", cfg.Environment)
	}

	if cfg.LogLevel != "info" {
		t.Errorf("Expected LogLevel to be 'info', got '%s'", cfg.LogLevel)
	}

	// Verify new SMS defaults
	if cfg.SMS.Provider != "twilio" {
		t.Errorf("Expected SMS.Provider to be 'twilio', got '%s'", cfg.SMS.Provider)
	}

	if cfg.SMS.AWSRegion != "us-east-1" {
		t.Errorf("Expected SMS.AWSRegion to be 'us-east-1', got '%s'", cfg.SMS.AWSRegion)
	}

	// Verify new Email defaults
	if cfg.Email.Workers != 5 {
		t.Errorf("Expected Email.Workers to be 5, got %d", cfg.Email.Workers)
	}

	// Verify new Server defaults
	if cfg.Server.Port != "8080" {
		t.Errorf("Expected Server.Port to be '8080', got '%s'", cfg.Server.Port)
	}

	if cfg.Server.HTTPPort != "8080" {
		t.Errorf("Expected Server.HTTPPort to be '8080', got '%s'", cfg.Server.HTTPPort)
	}

	if cfg.Server.GRPCPort != "50051" {
		t.Errorf("Expected Server.GRPCPort to be '50051', got '%s'", cfg.Server.GRPCPort)
	}

	if cfg.Server.ShutdownTimeout != 10 {
		t.Errorf("Expected Server.ShutdownTimeout to be 10, got %d", cfg.Server.ShutdownTimeout)
	}

	// Verify new ServiceURLs defaults
	if cfg.ServiceURLs.AuthService != "http://localhost:8081" {
		t.Errorf("Expected ServiceURLs.AuthService to be 'http://localhost:8081', got '%s'", cfg.ServiceURLs.AuthService)
	}

	if cfg.ServiceURLs.UserService != "http://localhost:8082" {
		t.Errorf("Expected ServiceURLs.UserService to be 'http://localhost:8082', got '%s'", cfg.ServiceURLs.UserService)
	}

	// Verify SMTP pool size default
	if cfg.SMTP.PoolSize != 10 {
		t.Errorf("Expected SMTP.PoolSize to be 10, got %d", cfg.SMTP.PoolSize)
	}
}

func TestLoadConfigWithEnvVars(t *testing.T) {
	// Set environment variables
	os.Setenv("SMS_PROVIDER", "aws_sns")
	os.Setenv("EMAIL_WORKERS", "10")
	os.Setenv("SERVER_PORT", "9090")
	os.Setenv("AUTH_SERVICE_URL", "http://auth-service:8081")
	defer func() {
		os.Unsetenv("SMS_PROVIDER")
		os.Unsetenv("EMAIL_WORKERS")
		os.Unsetenv("SERVER_PORT")
		os.Unsetenv("AUTH_SERVICE_URL")
	}()

	cfg, err := LoadConfig()
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	// Verify environment variables override defaults
	if cfg.SMS.Provider != "aws_sns" {
		t.Errorf("Expected SMS.Provider to be 'aws_sns', got '%s'", cfg.SMS.Provider)
	}

	if cfg.Email.Workers != 10 {
		t.Errorf("Expected Email.Workers to be 10, got %d", cfg.Email.Workers)
	}

	if cfg.Server.Port != "9090" {
		t.Errorf("Expected Server.Port to be '9090', got '%s'", cfg.Server.Port)
	}

	if cfg.ServiceURLs.AuthService != "http://auth-service:8081" {
		t.Errorf("Expected ServiceURLs.AuthService to be 'http://auth-service:8081', got '%s'", cfg.ServiceURLs.AuthService)
	}
}

func TestBackwardCompatibility(t *testing.T) {
	// Test that existing fields are still accessible
	cfg, err := LoadConfig()
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	// Verify existing MongoDB config
	if cfg.MongoDB.Database != "saas_framework" {
		t.Errorf("Expected MongoDB.Database to be 'saas_framework', got '%s'", cfg.MongoDB.Database)
	}

	// Verify existing Redis config
	if cfg.Redis.Port != "6379" {
		t.Errorf("Expected Redis.Port to be '6379', got '%s'", cfg.Redis.Port)
	}

	// Verify existing JWT config
	if cfg.JWT.Expiration != 3600 {
		t.Errorf("Expected JWT.Expiration to be 3600, got %d", cfg.JWT.Expiration)
	}

	// Verify existing SMTP config
	if cfg.SMTP.Port != 587 {
		t.Errorf("Expected SMTP.Port to be 587, got %d", cfg.SMTP.Port)
	}
}
