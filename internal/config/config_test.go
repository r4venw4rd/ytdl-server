package config

import (
	"os"
	"testing"
	"time"
)

// setBaseEnv clears all config keys (restored after the test) and points
// DOWNLOAD_DIR at a temp dir, since Load creates it.
func setBaseEnv(t *testing.T) {
	t.Helper()
	for _, k := range []string{"PORT", "MAX_CONCURRENT_JOBS", "CLEAN_UP_AFTER_MINUTES", "CLEAN_UP_MINUTES", "TEMP_DIR", "WEB_DIR"} {
		t.Setenv(k, "") // registers restore of the original value
		os.Unsetenv(k)
	}
	t.Setenv("DOWNLOAD_DIR", t.TempDir())
}

func TestLoadDefaults(t *testing.T) {
	setBaseEnv(t)

	cfg := Load()
	if cfg.Port != ":8080" || cfg.MaxConcurrentJobs != 3 || cfg.CleanupAfter != 15*time.Minute ||
		cfg.TempDir != "temp" || cfg.WebDir != "./web" {
		t.Fatalf("unexpected defaults: %+v", cfg)
	}
}

func TestLoadFromEnv(t *testing.T) {
	setBaseEnv(t)
	t.Setenv("PORT", ":9090")
	t.Setenv("MAX_CONCURRENT_JOBS", "5")
	t.Setenv("CLEAN_UP_AFTER_MINUTES", "30")
	t.Setenv("WEB_DIR", "/srv/web")

	cfg := Load()
	if cfg.Port != ":9090" || cfg.MaxConcurrentJobs != 5 || cfg.CleanupAfter != 30*time.Minute || cfg.WebDir != "/srv/web" {
		t.Fatalf("env not applied: %+v", cfg)
	}
}

func TestCleanupKeyPrecedence(t *testing.T) {
	tests := []struct {
		name          string
		after, legacy string
		want          time.Duration
	}{
		{"new key", "20", "", 20 * time.Minute},
		{"legacy key", "", "45", 45 * time.Minute},
		{"new key wins", "20", "45", 20 * time.Minute},
		{"invalid falls back", "abc", "", 15 * time.Minute},
		{"zero is reset", "0", "", 15 * time.Minute},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			setBaseEnv(t)
			t.Setenv("CLEAN_UP_AFTER_MINUTES", tt.after)
			t.Setenv("CLEAN_UP_MINUTES", tt.legacy)
			if got := Load().CleanupAfter; got != tt.want {
				t.Fatalf("CleanupAfter = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestInvalidMaxConcurrentJobsIsReset(t *testing.T) {
	setBaseEnv(t)
	t.Setenv("MAX_CONCURRENT_JOBS", "0")
	if got := Load().MaxConcurrentJobs; got != 3 {
		t.Fatalf("MaxConcurrentJobs = %d, want 3", got)
	}
}
