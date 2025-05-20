package config

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

// LoadEnv loads environment variables from a .env file into the process.
// Useful for development and local testing.
func LoadEnv() error {
	if err := godotenv.Load(); err != nil {
		return err
	}

	log.Println("✔️ .env file loaded")
	return nil
}

// getString tries to retrieve an environment variable as string,
// falling back to a default value if not present.
func getString(key, fallback string) (string, error) {
	val, ok := os.LookupEnv(key)

	if !ok || strings.TrimSpace(val) == "" {
		if strings.TrimSpace(fallback) == "" {
			return "", fmt.Errorf("missing or empty environment variable '%s' and fallback value is also empty", key)
		}
		return fallback, nil
	}

	return val, nil
}

// mustGetString returns a string from env or exits the app if the variable is invalid or missing.
func mustGetString(key, fallback string) string {
	val, err := getString(key, fallback)
	if err != nil {
		log.Fatalf("❌ %v", err)
	}

	return val
}

// getInt tries to retrieve an environment variable as integer,
// falling back to a default value if not present or invalid.
func getInt(key string, fallback int) (int, error) {
	val, ok := os.LookupEnv(key)

	if !ok || strings.TrimSpace(val) == "" {
		return fallback, nil
	}

	valAsInt, err := strconv.Atoi(val)
	if err != nil {
		return 0, fmt.Errorf("invalid integer value for '%s': %s", key, val)
	}

	return valAsInt, nil
}

// mustGetInt returns an integer from env or exits the app if the variable is invalid or missing.
func mustGetInt(key string, fallback int) int {
	val, err := getInt(key, fallback)
	if err != nil {
		log.Fatalf("❌ %v", err)
	}

	return val
}

// getBool tries to retrieve an environment variable as boolean,
// falling back to a default value if not present or invalid.
func getBool(key string, fallback bool) (bool, error) {
	val, ok := os.LookupEnv(key)

	if !ok || strings.TrimSpace(val) == "" {
		return fallback, nil
	}

	valAsBool, err := strconv.ParseBool(val)
	if err != nil {
		return false, fmt.Errorf("invalid bool value for '%s': %s", key, val)
	}

	return valAsBool, nil
}

// mustGetBool returns a boolean from env or exits the app if the variable is invalid or missing.
func mustGetBool(key string, fallback bool) bool {
	val, err := getBool(key, fallback)
	if err != nil {
		log.Fatalf("❌ %v", err)
	}

	return val
}

// getDuration tries to retrieve an environment variable as time.Duration,
// falling back to a default value if not present or invalid.
// Duration values should be in format accepted by time.ParseDuration (e.g., "1h", "10s").
func getDuration(key string, fallback time.Duration) (time.Duration, error) {
	valStr, ok := os.LookupEnv(key)

	if !ok || strings.TrimSpace(valStr) == "" {
		return fallback, nil
	}

	dur, err := time.ParseDuration(valStr)
	if err != nil {
		return 0, fmt.Errorf("invalid duration for '%s': %v", key, err)
	}

	return dur, nil
}

// mustGetDuration returns a time.Duration from env or exits the app if the variable is invalid or missing.
func mustGetDuration(key string, fallback time.Duration) time.Duration {
	val, err := getDuration(key, fallback)
	if err != nil {
		log.Fatalf("❌ %v", err)
	}

	return val
}

// mustEnvOrPanic retrieves a required environment variable or panics if missing or empty.
func mustEnvOrPanic(key string) string {
	val := os.Getenv(key)

	if strings.TrimSpace(val) == "" {
		log.Fatalf("❌ Required environment variable '%s' is missing or empty", key)
	}

	return val
}
