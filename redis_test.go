package main

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/redis/go-redis/v9/maintnotifications"
)

func TestRedisConnection(t *testing.T) {
	// Get Redis URL from environment or use default
	redisURL := os.Getenv("REDIS_URL")
	if redisURL == "" {
		redisURL = "localhost:6379"
	}
	password := os.Getenv("REDIS_PASSWORD")

	// Create Redis client
	client := redis.NewClient(&redis.Options{
		Addr:     redisURL,
		Password: password,
		DB:       0, // use default DB
		// https://github.com/redis/go-redis/issues/3536
		// Explicitly disable maintenance notifications
		// This prevents the client from sending CLIENT MAINT_NOTIFICATIONS ON
		MaintNotificationsConfig: &maintnotifications.Config{
			Mode: maintnotifications.ModeDisabled,
		},
	})

	ctx := context.Background()

	// Test basic connection
	t.Run("Ping", func(t *testing.T) {
		err := client.Ping(ctx).Err()
		if err != nil {
			t.Fatalf("Failed to ping Redis: %v", err)
		}
	})

	// Test basic operations
	t.Run("SetAndGet", func(t *testing.T) {
		key := "test:key:connection"
		value := "test_value"

		// Set a key
		err := client.Set(ctx, key, value, time.Minute).Err()
		if err != nil {
			t.Fatalf("Failed to set key: %v", err)
		}

		// Get the key
		result, err := client.Get(ctx, key).Result()
		if err != nil {
			t.Fatalf("Failed to get key: %v", err)
		}

		if result != value {
			t.Errorf("Expected %s, got %s", value, result)
		}

		// Clean up
		client.Del(ctx, key)
	})

	// Test that we can execute multiple commands
	t.Run("MultipleCommands", func(t *testing.T) {
		pipe := client.Pipeline()

		// Add multiple commands to pipeline
		pipe.Set(ctx, "test:multi:1", "value1", time.Minute)
		pipe.Set(ctx, "test:multi:2", "value2", time.Minute)
		pipe.Get(ctx, "test:multi:1")
		pipe.Get(ctx, "test:multi:2")

		// Execute pipeline
		cmds, err := pipe.Exec(ctx)
		if err != nil {
			t.Fatalf("Failed to execute pipeline: %v", err)
		}

		// Verify we got expected number of commands
		if len(cmds) != 4 {
			t.Errorf("Expected 4 commands, got %d", len(cmds))
		}

		// Clean up
		client.Del(ctx, "test:multi:1", "test:multi:2")
	})

	// Test Redis info command
	t.Run("Info", func(t *testing.T) {
		info, err := client.Info(ctx).Result()
		if err != nil {
			t.Fatalf("Failed to get Redis info: %v", err)
		}

		if len(info) == 0 {
			t.Error("Redis info should not be empty")
		}

		t.Logf("Redis info length: %d characters", len(info))
	})

	// Close the client
	err := client.Close()
	if err != nil {
		t.Errorf("Failed to close Redis client: %v", err)
	}
}
