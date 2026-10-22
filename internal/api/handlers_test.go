package api

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"ytdl-server/internal/config"
	"ytdl-server/internal/jobs"
)

func newTestRouter(t *testing.T) http.Handler {
	t.Helper()
	cfg := &config.Config{MaxConcurrentJobs: 1, CleanupAfter: time.Hour, DownloadDir: t.TempDir(), TempDir: t.TempDir()}
	return NewRouter(NewHandler(jobs.NewManager(cfg)), t.TempDir())
}

func TestSSEDoesNotOverrideCORSAllowlist(t *testing.T) {
	t.Setenv("ALLOWED_ORIGINS", "https://app.example.com")
	router := newTestRouter(t)

	req := httptest.NewRequest(http.MethodGet, "/api/events/unknown-job", nil)
	req.Header.Set("Origin", "https://app.example.com")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req) // returns after the first tick: job not found

	if got := rec.Header().Get("Access-Control-Allow-Origin"); got != "https://app.example.com" {
		t.Fatalf("Access-Control-Allow-Origin = %q, want the allowlisted origin", got)
	}
}
