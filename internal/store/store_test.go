package store

import (
	"context"
	"testing"
	"time"
)

func TestSQLiteStoreCreatesSessionAndTracksTasks(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	repo, err := OpenSQLite(":memory:")
	if err != nil {
		t.Fatalf("OpenSQLite() error = %v", err)
	}
	t.Cleanup(func() {
		_ = repo.Close()
	})

	session, err := repo.CreateAnonymousSession(ctx, "session-token", "visitor:sig")
	if err != nil {
		t.Fatalf("CreateAnonymousSession() error = %v", err)
	}

	task, err := repo.CreateTask(ctx, CreateTaskParams{
		SessionID:  session.ID,
		Prompt:     "a chrome flower in the rain",
		Ratio:      "4:5",
		Resolution: "1K",
		ImageNum:   1,
		Status:     TaskStatusQueued,
		IsPublic:   false,
	})
	if err != nil {
		t.Fatalf("CreateTask() error = %v", err)
	}

	if err := repo.UpdateTaskResult(ctx, UpdateTaskResultParams{
		TaskID:          task.ID,
		Status:          TaskStatusSucceeded,
		ParentMessageID: "parent-1",
		MessageID:       "parent-1-0",
		ResultURL:       "https://img.example/result.png",
		FinishedAt:      time.Unix(1_700_000_000, 0).UTC(),
		IsPublic:        true,
	}); err != nil {
		t.Fatalf("UpdateTaskResult() error = %v", err)
	}

	history, err := repo.ListHistory(ctx, session.ID, 1, 20)
	if err != nil {
		t.Fatalf("ListHistory() error = %v", err)
	}

	if history.Total != 1 {
		t.Fatalf("history.Total = %d, want 1", history.Total)
	}

	if got := history.Items[0].ResultURL; got != "https://img.example/result.png" {
		t.Fatalf("history.Items[0].ResultURL = %q, want stored result", got)
	}

	gallery, err := repo.ListGallery(ctx, 1, 20)
	if err != nil {
		t.Fatalf("ListGallery() error = %v", err)
	}

	if gallery.Total != 1 {
		t.Fatalf("gallery.Total = %d, want 1", gallery.Total)
	}
}

func TestSQLiteStoreListsRecoverableTasks(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	repo, err := OpenSQLite(":memory:")
	if err != nil {
		t.Fatalf("OpenSQLite() error = %v", err)
	}
	t.Cleanup(func() {
		_ = repo.Close()
	})

	session, err := repo.CreateAnonymousSession(ctx, "session-token", "visitor:sig")
	if err != nil {
		t.Fatalf("CreateAnonymousSession() error = %v", err)
	}

	queuedTask, err := repo.CreateTask(ctx, CreateTaskParams{
		SessionID:       session.ID,
		Prompt:          "queued prompt",
		Ratio:           "1:1",
		Resolution:      "1K",
		ImageNum:        1,
		Status:          TaskStatusRunning,
		ParentMessageID: "parent-queued",
	})
	if err != nil {
		t.Fatalf("CreateTask() error = %v", err)
	}

	recoverable, err := repo.ListRecoverableTasks(ctx)
	if err != nil {
		t.Fatalf("ListRecoverableTasks() error = %v", err)
	}

	if len(recoverable) != 1 {
		t.Fatalf("recoverable count = %d, want 1", len(recoverable))
	}

	if recoverable[0].ID != queuedTask.ID {
		t.Fatalf("recoverable[0].ID = %q, want %q", recoverable[0].ID, queuedTask.ID)
	}
}
