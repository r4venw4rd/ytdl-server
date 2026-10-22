package jobs

import (
	"log"
	"os"
	"path/filepath"
	"sync"
	"time"

	"ytdl-server/internal/config"
	"ytdl-server/internal/downloader"
	"ytdl-server/internal/models"

	"github.com/google/uuid"
)

type Manager struct {
	jobs  sync.Map
	queue chan struct{}
	cfg   *config.Config
}

func NewManager(cfg *config.Config) *Manager {
	return &Manager{
		queue: make(chan struct{}, cfg.MaxConcurrentJobs),
		cfg:   cfg,
	}
}

func (m *Manager) Create(req models.CreateJobRequest) *models.Job {
	id := uuid.New().String()
	job := &models.Job{
		ID:        id,
		VideoID:   req.VideoID,
		Status:    "pending",
		CreatedAt: time.Now(),
		FilePath:  filepath.Join(m.cfg.DownloadDir, id+".mp4"),
	}
	m.jobs.Store(id, job)

	go m.runWorker(job, req.Quality)

	return job
}

func (m *Manager) Get(id string) (*models.Job, bool) {
	val, ok := m.jobs.Load(id)
	if !ok {
		return nil, false
	}
	return val.(*models.Job), true
}

func (m *Manager) runWorker(job *models.Job, quality string) {
	// Rate Limiting
	select {
	case m.queue <- struct{}{}:
		defer func() { <-m.queue }()
	case <-time.After(10 * time.Second):
		m.updateStatus(job, "failed", 0, "Server busy")
		return
	}

	m.updateStatus(job, "processing", 0, "")

	err := downloader.Process(job, quality, m.cfg.TempDir, func(pct float64) {
		m.updateStatus(job, "processing", pct, "")
	})

	if err != nil {
		log.Printf("Job %s failed: %v", job.ID, err)
		m.updateStatus(job, "failed", 0, err.Error())
	} else {
		m.updateStatus(job, "ready", 100, "")
	}
}

func (m *Manager) updateStatus(job *models.Job, status string, pct float64, errStr string) {
	job.Update(status, pct, errStr)
}

// StartJanitor periodically drops finished jobs older than cfg.CleanupAfter
// from the job map and deletes their output files. It runs every
// cfg.CleanupAfter in a background goroutine.
func (m *Manager) StartJanitor() {
	go func() {
		ticker := time.NewTicker(m.cfg.CleanupAfter)
		defer ticker.Stop()
		for now := range ticker.C {
			m.cleanupJobs(now)
		}
	}()
}

func (m *Manager) cleanupJobs(now time.Time) {
	m.jobs.Range(func(key, value interface{}) bool {
		job := value.(*models.Job)
		if now.Sub(job.CreatedAt) <= m.cfg.CleanupAfter {
			return true
		}
		// Never pull the file out from under a job that is still running.
		if !job.Snapshot().IsDone() {
			return true
		}
		if err := os.Remove(job.FilePath); err != nil && !os.IsNotExist(err) {
			log.Printf("❌ Janitor: could not remove %s: %v", job.FilePath, err)
		}
		m.jobs.Delete(key)
		log.Println("🧹 Cleaned up job:", job.ID)
		return true
	})
}
