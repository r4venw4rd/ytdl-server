package api

import (
	"net/http"
	"net/http/httptest"
	"net/url"
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

func TestGetVideoInfoRejectsInvalidID(t *testing.T) {
	router := newTestRouter(t)
	// None of these may reach YouTube; all must be rejected locally.
	for _, id := range []string{"", "short", "dQw4w9WgXcQx", "dQw4w9WgXc!", "../../etc/pa", "dQw4w9 gXcQ"} {
		req := httptest.NewRequest(http.MethodGet, "/api/info?video_id="+url.QueryEscape(id), nil)
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)
		if rec.Code != http.StatusBadRequest {
			t.Errorf("video_id %q: status %d, want 400", id, rec.Code)
		}
	}
}

func TestVideoIDRegex(t *testing.T) {
	for id, want := range map[string]bool{
		"dQw4w9WgXcQ":  true,
		"a-b_c-d_e-f":  true,
		"dQw4w9WgXc":   false,
		"dQw4w9WgXcQQ": false,
		"dQw4w9WgXc/":  false,
	} {
		if got := videoIDRegex.MatchString(id); got != want {
			t.Errorf("videoIDRegex(%q) = %v, want %v", id, got, want)
		}
	}
}
