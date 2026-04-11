package config

import (
	"Online-queue-management-system/libs/logger"
	"context"
	"fmt"
	"os"
	"strconv"
	"time"
)

func MustGet(ctx context.Context, key string) (string, error) {
	log := logger.From(ctx)
	val := os.Getenv(key)
	if val == "" {
		log.Error("env %s is required but not set", key)
		return "", fmt.Errorf("env %s is required but not set", key)
	}
	return val, nil
}

func Get(ctx context.Context, key string, defaultVal string) string {
	log := logger.From(ctx)
	val := os.Getenv(key)
	if val == "" {
		log.Warn("env %s is not set, using default value %s", key, defaultVal)
		return defaultVal
	}
	return val
}

func GetInt(ctx context.Context, key string) (int, error) {
	log := logger.From(ctx)
	val, err := MustGet(ctx, key)
	if err != nil {
		log.Error("error getting env", "key", key, "error", err)
		return 0, err
	}

	i, err := strconv.Atoi(val)
	if err != nil {
		log.Error("env %s must be int, got %s", key, val)
		return 0, err
	}

	return i, nil
}

func GetIntDefault(ctx context.Context, key string, defaultVal int) int {
	log := logger.From(ctx)
	val := os.Getenv(key)
	if val == "" {
		return defaultVal
	}

	i, err := strconv.Atoi(val)
	if err != nil {
		log.Error("env %s must be int, got %s", key, val)
		return defaultVal
	}

	return i
}

func GetDuration(ctx context.Context, key string) (time.Duration, error) {
	log := logger.From(ctx)
	val, err := MustGet(ctx, key)
	if err != nil {
		log.Error("error getting env", "key", key, "error", err)
		return 0, err
	}
	d, err := time.ParseDuration(val)
	if err != nil {
		log.Error("env %s must be duration, got %s", key, val)
		return 0, err
	}

	return d, nil
}

func GetDurationDefault(ctx context.Context, key string, defaultVal time.Duration) time.Duration {
	log := logger.From(ctx)
	val := os.Getenv(key)
	if val == "" {
		return defaultVal
	}

	d, err := time.ParseDuration(val)
	if err != nil {
		log.Error("env %s must be int, got %s", key, val)
		return defaultVal
	}

	return d
}

func GetBool(ctx context.Context, key string, defaultVal bool) (bool, error) {
	log := logger.From(ctx)
	val := os.Getenv(key)
	if val == "" {
		return defaultVal, nil
	}

	b, err := strconv.ParseBool(val)
	if err != nil {
		log.Error("env %s must be bool, got %s", key, val)
		return defaultVal, err
	}

	return b, nil
}
