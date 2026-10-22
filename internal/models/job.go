package models

import (
	"sync"
	"time"
)

// Job: Bir indirme işleminin tüm durumunu tutar
//
// ID, VideoID, FilePath and CreatedAt are set once at creation and never
// change afterwards. Status, Percentage, Filename and Error are mutated by
// the worker while HTTP handlers read them, so they must only be accessed
// through the methods below (Update, SetFilename, Snapshot).
type Job struct {
	mu sync.Mutex

	ID         string    `json:"id"`
	VideoID    string    `json:"video_id"`
	Status     string    `json:"status"` // "pending", "processing", "ready", "failed"
	Percentage float64   `json:"percentage"`
	Filename   string    `json:"filename"`
	FilePath   string    `json:"-"`
	Error      string    `json:"error,omitempty"`
	CreatedAt  time.Time `json:"-"`
}

// JobSnapshot is a point-in-time copy of a Job that is safe to read and
// marshal without holding any lock. Its JSON form matches Job's.
type JobSnapshot struct {
	ID         string    `json:"id"`
	VideoID    string    `json:"video_id"`
	Status     string    `json:"status"`
	Percentage float64   `json:"percentage"`
	Filename   string    `json:"filename"`
	FilePath   string    `json:"-"`
	Error      string    `json:"error,omitempty"`
	CreatedAt  time.Time `json:"-"`
}

// Update sets the mutable progress fields atomically.
func (j *Job) Update(status string, pct float64, errStr string) {
	j.mu.Lock()
	defer j.mu.Unlock()
	j.Status = status
	j.Percentage = pct
	j.Error = errStr
}

// SetFilename sets the user-facing download filename.
func (j *Job) SetFilename(name string) {
	j.mu.Lock()
	defer j.mu.Unlock()
	j.Filename = name
}

// Snapshot returns a consistent copy of the job's current state.
func (j *Job) Snapshot() JobSnapshot {
	j.mu.Lock()
	defer j.mu.Unlock()
	return JobSnapshot{
		ID:         j.ID,
		VideoID:    j.VideoID,
		Status:     j.Status,
		Percentage: j.Percentage,
		Filename:   j.Filename,
		FilePath:   j.FilePath,
		Error:      j.Error,
		CreatedAt:  j.CreatedAt,
	}
}

// IsDone reports whether the job reached a terminal state.
func (s JobSnapshot) IsDone() bool {
	return s.Status == "ready" || s.Status == "failed"
}

type CreateJobRequest struct {
	VideoID string `json:"video_id"`
	Quality string `json:"quality"`
}
