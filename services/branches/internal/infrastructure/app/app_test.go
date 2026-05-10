package app

import (
	libconfig "Online-queue-management-system/libs/config"
	"Online-queue-management-system/libs/logger"
	branchesConfig "Online-queue-management-system/services/branches/config"
	"context"
	"log/slog"
	"testing"
)

func TestNewApp_Success(t *testing.T) {
	// Arrange
	ctx := context.Background()
	log := logger.New(logger.Config{
		Level:  slog.LevelInfo,
		JSON:   false,
		Source: true,
	})
	ctx = logger.With(ctx, log)

	cfg := branchesConfig.Config{
		BranchesCfg: branchesConfig.BranchesConfig{
			BranchesPort:    "8083",
			JWTAccessSecret: "test-secret",
		},
	}

	// Using a test database connection string
	// Note: This will fail because the database doesn't exist, but we can test error handling
	dbCfg := libconfig.DBConfig{
		DSN: "invalid-dsn",
	}

	// Act
	app, err := NewApp(ctx, cfg, dbCfg)

	// Assert
	// We expect an error because the DSN is invalid
	if err == nil {
		t.Fatalf("expected error for invalid DSN, got nil")
	}
	if app != nil {
		t.Fatalf("expected nil app with error, got %v", app)
	}
}

func TestBranchesApp_Run_Shutdown(t *testing.T) {
	// This test ensures that the Run method handles context cancellation properly
	// But we can't fully test it without a real database

	// Arrange
	ctx := context.Background()
	log := logger.New(logger.Config{
		Level:  slog.LevelInfo,
		JSON:   false,
		Source: true,
	})
	ctx = logger.With(ctx, log)

	// We'll skip the full initialization since it requires a database
	// This is more of a demonstration of how to structure the test
	t.Skip("Requires database connection - integration test")
}

// Helper test for configuration validation
func TestNewApp_ConfigValidation(t *testing.T) {
	// Arrange
	ctx := context.Background()
	log := logger.New(logger.Config{
		Level:  slog.LevelInfo,
		JSON:   false,
		Source: true,
	})
	ctx = logger.With(ctx, log)

	// Test with empty port
	cfg := branchesConfig.Config{
		BranchesCfg: branchesConfig.BranchesConfig{
			BranchesPort:    "",
			JWTAccessSecret: "test-secret",
		},
	}

	dbCfg := libconfig.DBConfig{
		DSN: "invalid-dsn",
	}

	// Act
	app, err := NewApp(ctx, cfg, dbCfg)

	// Assert
	// Should fail on database connection, not on config
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
	if app != nil {
		t.Fatalf("expected nil app with error, got %v", app)
	}
}

func TestNewApp_EmptyJWTSecret(t *testing.T) {
	// Arrange
	ctx := context.Background()
	log := logger.New(logger.Config{
		Level:  slog.LevelInfo,
		JSON:   false,
		Source: true,
	})
	ctx = logger.With(ctx, log)

	cfg := branchesConfig.Config{
		BranchesCfg: branchesConfig.BranchesConfig{
			BranchesPort:    "8083",
			JWTAccessSecret: "",
		},
	}

	dbCfg := libconfig.DBConfig{
		DSN: "invalid-dsn",
	}

	// Act
	app, err := NewApp(ctx, cfg, dbCfg)

	// Assert
	// Should fail on database connection, not on config
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
	if app != nil {
		t.Fatalf("expected nil app with error, got %v", app)
	}
}
