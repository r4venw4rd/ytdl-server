package jobs

import (
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"
	"ytdl-server/internal/config"

	"github.com/google/uuid"
)

// StartJanitor periodically removes stale files left on disk: temp stream
// files of crashed/aborted jobs and finished downloads that are no longer
// tracked (e.g. after a restart). Only files older than cfg.CleanupAfter are
// removed, so in-progress jobs are left alone. It runs every cfg.CleanupAfter.
func StartJanitor(cfg *config.Config) {
	ticker := time.NewTicker(cfg.CleanupAfter)

	go func() {
		for now := range ticker.C {
			log.Println("🧹 Janitor: Starting scheduled cleanup...")
			removeStaleFiles(cfg.TempDir, cfg.CleanupAfter, now, func(string) bool { return true })
			removeStaleFiles(cfg.DownloadDir, cfg.CleanupAfter, now, isJobOutput)
			log.Println("✅ Janitor: Cleanup finished.")
		}
	}()
}

// isJobOutput matches the "<uuid>.mp4" files the manager writes, so the
// janitor never deletes unrelated files in a shared download directory.
func isJobOutput(name string) bool {
	id, ok := strings.CutSuffix(name, ".mp4")
	if !ok {
		return false
	}
	_, err := uuid.Parse(id)
	return err == nil
}

func removeStaleFiles(dir string, maxAge time.Duration, now time.Time, match func(name string) bool) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		log.Printf("❌ Janitor Error: Could not read %s: %v", dir, err)
		return
	}
	for _, e := range entries {
		if !e.Type().IsRegular() || !match(e.Name()) {
			continue
		}
		info, err := e.Info()
		if err != nil || now.Sub(info.ModTime()) <= maxAge {
			continue
		}
		path := filepath.Join(dir, e.Name())
		if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
			log.Printf("❌ Janitor Error: Could not remove %s: %v", path, err)
		}
	}
}
