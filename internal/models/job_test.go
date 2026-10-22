package models

import (
	"encoding/json"
	"sync"
	"testing"
)

// TestJobConcurrentUpdateAndSnapshot exercises the worker/SSE access pattern.
// Run with -race: it fails if any field is read and written unsynchronized.
func TestJobConcurrentUpdateAndSnapshot(t *testing.T) {
	job := &Job{ID: "id", VideoID: "dQw4w9WgXcQ", Status: "pending"}

	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		defer wg.Done()
		for i := 0; i < 1000; i++ {
			job.Update("processing", float64(i)/10, "")
			job.SetFilename("video_1080p.mp4")
		}
		job.Update("ready", 100, "")
	}()
	go func() {
		defer wg.Done()
		for i := 0; i < 1000; i++ {
			snap := job.Snapshot()
			if _, err := json.Marshal(snap); err != nil {
				t.Errorf("marshal: %v", err)
				return
			}
		}
	}()
	wg.Wait()

	snap := job.Snapshot()
	if snap.Status != "ready" || snap.Percentage != 100 || snap.Filename != "video_1080p.mp4" {
		t.Fatalf("unexpected final snapshot: %+v", snap)
	}
	if !snap.IsDone() {
		t.Fatal("ready job should be done")
	}
}

func TestJobSnapshotJSONMatchesJob(t *testing.T) {
	job := &Job{ID: "id", VideoID: "vid", Status: "failed", Error: "boom", FilePath: "/secret/path"}
	got, err := json.Marshal(job.Snapshot())
	if err != nil {
		t.Fatal(err)
	}
	want := `{"id":"id","video_id":"vid","status":"failed","percentage":0,"filename":"","error":"boom"}`
	if string(got) != want {
		t.Fatalf("got %s, want %s", got, want)
	}
}
