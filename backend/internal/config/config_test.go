package config

import (
	"os"
	"testing"
)

func TestLoad(t *testing.T) {
	// Set required env vars so Load() succeeds without a real .env file.
	t.Setenv("DATABASE_URL", "postgres://user:pass@localhost:5432/testdb")
	t.Setenv("JWT_SECRET", "test-secret-key")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}
	if cfg.Port != "8080" {
		t.Errorf("expected default Port 8080, got %s", cfg.Port)
	}
	if cfg.DatabaseURL == "" {
		t.Error("expected DatabaseURL to be set")
	}
	if cfg.DoclingServiceURL == "" {
		t.Error("expected DoclingServiceURL default")
	}
}

func TestLoadMissingDBURL(t *testing.T) {
	// Ensure DATABASE_URL is empty to trigger validation error.
	os.Unsetenv("DATABASE_URL")
	os.Unsetenv("DATABASE_REPLICA_URL")
	os.Unsetenv("JWT_SECRET")

	_, err := Load()
	if err == nil {
		t.Fatal("expected error when DATABASE_URL is empty")
	}
}
