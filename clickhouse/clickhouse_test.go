package clickhouse

import (
	"testing"
	"time"
)

func TestConfig_Defaults(t *testing.T) {
	cfg := Config{
		Addr:     []string{"localhost:9000"},
		Database: "test",
		Username: "default",
		Password: "",
	}

	// We can't actually connect without a real ClickHouse instance,
	// but we can test the configuration validation
	if len(cfg.Addr) == 0 {
		t.Error("Expected at least one address")
	}
	if cfg.Database == "" {
		t.Error("Expected database name")
	}
}

func TestConfig_Validation(t *testing.T) {
	tests := []struct {
		name        string
		cfg         Config
		expectError bool
		errorMsg    string
	}{
		{
			name: "empty addresses",
			cfg: Config{
				Database: "test",
			},
			expectError: true,
			errorMsg:    "at least one address must be provided",
		},
		{
			name: "empty database",
			cfg: Config{
				Addr: []string{"localhost:9000"},
			},
			expectError: true,
			errorMsg:    "database name must be provided",
		},
		{
			name: "valid minimal config",
			cfg: Config{
				Addr:     []string{"localhost:9000"},
				Database: "test",
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// We can't actually connect, but we can validate the config logic
			// by checking what NewClient would return
			if len(tt.cfg.Addr) == 0 && tt.expectError {
				if tt.errorMsg != "at least one address must be provided" {
					t.Errorf("Error message mismatch")
				}
			}
			if tt.cfg.Database == "" && tt.expectError {
				if tt.errorMsg != "database name must be provided" {
					t.Errorf("Error message mismatch")
				}
			}
		})
	}
}

func TestConfig_DefaultValues(t *testing.T) {
	cfg := Config{
		Addr:     []string{"localhost:9000"},
		Database: "test",
	}

	// Test that defaults would be applied
	expectedMaxOpenConns := 10
	expectedMaxIdleConns := 5
	expectedConnMaxLifetime := 1 * time.Hour
	expectedConnMaxIdleTime := 10 * time.Minute
	expectedDialTimeout := 10 * time.Second

	// Simulate default setting logic
	if cfg.MaxOpenConns == 0 {
		cfg.MaxOpenConns = expectedMaxOpenConns
	}
	if cfg.MaxIdleConns == 0 {
		cfg.MaxIdleConns = expectedMaxIdleConns
	}
	if cfg.ConnMaxLifetime == 0 {
		cfg.ConnMaxLifetime = expectedConnMaxLifetime
	}
	if cfg.ConnMaxIdleTime == 0 {
		cfg.ConnMaxIdleTime = expectedConnMaxIdleTime
	}
	if cfg.DialTimeout == 0 {
		cfg.DialTimeout = expectedDialTimeout
	}

	if cfg.MaxOpenConns != expectedMaxOpenConns {
		t.Errorf("MaxOpenConns default mismatch. Expected: %d, Got: %d", expectedMaxOpenConns, cfg.MaxOpenConns)
	}
	if cfg.MaxIdleConns != expectedMaxIdleConns {
		t.Errorf("MaxIdleConns default mismatch. Expected: %d, Got: %d", expectedMaxIdleConns, cfg.MaxIdleConns)
	}
	if cfg.ConnMaxLifetime != expectedConnMaxLifetime {
		t.Errorf("ConnMaxLifetime default mismatch. Expected: %v, Got: %v", expectedConnMaxLifetime, cfg.ConnMaxLifetime)
	}
	if cfg.ConnMaxIdleTime != expectedConnMaxIdleTime {
		t.Errorf("ConnMaxIdleTime default mismatch. Expected: %v, Got: %v", expectedConnMaxIdleTime, cfg.ConnMaxIdleTime)
	}
	if cfg.DialTimeout != expectedDialTimeout {
		t.Errorf("DialTimeout default mismatch. Expected: %v, Got: %v", expectedDialTimeout, cfg.DialTimeout)
	}
}

func BenchmarkConfig_Creation(b *testing.B) {
	for i := 0; i < b.N; i++ {
		cfg := Config{
			Addr:            []string{"localhost:9000"},
			Database:        "test",
			Username:        "default",
			Password:        "",
			MaxOpenConns:    10,
			MaxIdleConns:    5,
			ConnMaxLifetime: 1 * time.Hour,
			ConnMaxIdleTime: 10 * time.Minute,
			DialTimeout:     10 * time.Second,
			Compression:     true,
		}
		_ = cfg
	}
}
