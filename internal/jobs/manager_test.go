package jobs

import (
	"context"
	"errors"
	"testing"
	"time"

	"ytdl-server/internal/config"
	"ytdl-server/internal/models"
)

func TestCreateReturnsBusyWhenNoSlot(t *testing.T) {
	m := NewManager(&config.Config{MaxConcurrentJobs: 1, CleanupAfter: time.Hour})
	m.acquireTimeout = 10 * time.Millisecond
	m.queue <- struct{}{} // occupy the only slot

	job, err := m.Create(context.Background(), models.CreateJobRequest{VideoID: "dQw4w9WgXcQ"})
	if !errors.Is(err, ErrServerBusy) || job != nil {
		t.Fatalf("Create = (%v, %v), want (nil, ErrServerBusy)", job, err)
	}
	n := 0
	m.jobs.Range(func(_, _ interface{}) bool { n++; return true })
	if n != 0 {
		t.Fatalf("rejected request must not register a job, got %d", n)
	}
}
