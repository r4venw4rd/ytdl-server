package config

import (
	"log"
	"os"
	"strconv"
	"time"
)

// Config holds all server settings in correct types
type Config struct {
	Port              string
	MaxConcurrentJobs int
	CleanupAfter      time.Duration
	DownloadDir       string
	TempDir           string
	WebDir            string // directory of the bundled web client served at "/"
}

// Load: The only way to get config in the app
func Load() *Config {
	cfg := &Config{
		Port:              getEnv("PORT", ":8080"),
		MaxConcurrentJobs: getEnvAsInt("MAX_CONCURRENT_JOBS", 3),
		CleanupAfter:      time.Duration(getEnvAsInt("CLEAN_UP_AFTER_MINUTES", 15)) * time.Minute,
		DownloadDir:       getEnv("DOWNLOAD_DIR", "downloads"),
		TempDir:           getEnv("TEMP_DIR", "temp"),
		WebDir:            getEnv("WEB_DIR", "./web"),
	}

	// 🛡️ Post-load Validation
	validate(cfg)

	return cfg
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}

func getEnvAsInt(key string, fallback int) int {
	str := getEnv(key, "")
	if val, err := strconv.Atoi(str); err == nil {
		return val
	}
	return fallback
}

// validate ensures the server won't crash due to misconfiguration
func validate(cfg *Config) {
	if cfg.MaxConcurrentJobs < 1 {
		log.Println("⚠️ Warning: MAX_CONCURRENT_JOBS must be at least 1. Resetting to 3.")
		cfg.MaxConcurrentJobs = 3
	}
	if cfg.CleanupAfter <= 0 {
		log.Println("⚠️ Warning: CLEAN_UP_AFTER_MINUTES must be at least 1. Resetting to 15.")
		cfg.CleanupAfter = 15 * time.Minute
	}
	if _, err := os.Stat(cfg.DownloadDir); os.IsNotExist(err) {
		log.Printf("📂 Notice: Creating missing download directory: %s\n", cfg.DownloadDir)
		os.MkdirAll(cfg.DownloadDir, 0755)
	}
}
