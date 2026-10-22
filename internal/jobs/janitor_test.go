package jobs

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"ytdl-server/internal/config"
	"ytdl-server/internal/models"
)

func touch(t *testing.T, path string, mtime time.Time) {
	t.Helper()
	if err := os.WriteFile(path, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.Chtimes(path, mtime, mtime); err != nil {
		t.Fatal(err)
	}
}

func exists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func TestRemoveStaleFiles(t *testing.T) {
	dir := t.TempDir()
	now := time.Now()
	old := now.Add(-2 * time.Hour)

	staleJob := filepath.Join(dir, "0b8e6c9e-3f0a-4d8c-9a57-2a1f6a0f3c11.mp4")
	freshJob := filepath.Join(dir, "1b8e6c9e-3f0a-4d8c-9a57-2a1f6a0f3c11.mp4")
	foreign := filepath.Join(dir, "holiday.mp4")
	touch(t, staleJob, old)
	touch(t, freshJob, now)
	touch(t, foreign, old)

	removeStaleFiles(dir, time.Hour, now, isJobOutput)

	if exists(staleJob) {
		t.Error("stale job output should be removed")
	}
	if !exists(freshJob) {
		t.Error("fresh job output must be kept")
	}
	if !exists(foreign) {
		t.Error("files not created by the server must be kept")
	}
}

func TestCleanupJobs(t *testing.T) {
	dir := t.TempDir()
	cfg := &config.Config{MaxConcurrentJobs: 1, CleanupAfter: time.Hour, DownloadDir: dir}
	m := NewManager(cfg)
	now := time.Now()

	add := func(id, status string, created time.Time) string {
		path := filepath.Join(dir, id+".mp4")
		touch(t, path, created)
		m.jobs.Store(id, &models.Job{ID: id, Status: status, CreatedAt: created, FilePath: path})
		return path
	}
	oldReady := add("old-ready", "ready", now.Add(-2*time.Hour))
	oldRunning := add("old-running", "processing", now.Add(-2*time.Hour))
	newReady := add("new-ready", "ready", now)

	m.cleanupJobs(now)

	if _, ok := m.Get("old-ready"); ok || exists(oldReady) {
		t.Error("expired finished job and its file should be removed")
	}
	if _, ok := m.Get("old-running"); !ok || !exists(oldRunning) {
		t.Error("running job must not be cleaned up")
	}
	if _, ok := m.Get("new-ready"); !ok || !exists(newReady) {
		t.Error("recent job must be kept")
	}
}
