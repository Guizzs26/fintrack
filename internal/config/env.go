package config

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
)

func loadEnv() {
	if err := godotenv.Load(); err != nil {
		log.Fatal("❌ error loading .env file")
	} else {
		log.Println("✔️ .env file loaded")
	}
}

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

func getInt(key string, fallback int) (int, error) {
	val, ok := os.LookupEnv(key)

	if !ok || strings.TrimSpace(val) == "" {
		return fallback, nil
	}

	valAsInt, err := strconv.Atoi(val)
	if err != nil {
		return fallback, fmt.Errorf("invalid integer value for '%s': %s", key, val)
	}

	return valAsInt, nil
}

func mustGetString(key, fallback string) string {
	val, err := getString(key, fallback)
	if err != nil {
		log.Fatalf("❌ %v", err)
	}

	return val
}

func mustGetInt(key string, fallback int) int {
	val, err := getInt(key, fallback)
	if err != nil {
		log.Fatalf("❌ %v", err)
	}

	return val
}
