package app

import (
	"context"
	"strings"
	"sync"
	"testing"
	"time"

	"rita.ai/internal/store"
	"rita.ai/internal/upstream"
)

type fakeGateway struct {
	generateResp    upstream.GenerateResponse
	streamEvents    []upstream.StreamEvent
	generateErr     error
	streamErr       error
	mu              sync.Mutex
	generateHeaders []string
	streamHeaders   []string
}

func (f *fakeGateway) Generate(_ context.Context, visitorHeader string, _ upstream.GenerateRequest) (upstream.GenerateResponse, error) {
	f.mu.Lock()
	f.generateHeaders = append(f.generateHeaders, visitorHeader)
	f.mu.Unlock()
	return f.generateResp, f.generateErr
}

func (f *fakeGateway) Stream(_ context.Context, visitorHeader, _ string) ([]upstream.StreamEvent, error) {
	f.mu.Lock()
	f.streamHeaders = append(f.streamHeaders, visitorHeader)
	f.mu.Unlock()
	return f.streamEvents, f.streamErr
}

func TestServiceEnsureSessionCreatesAndReusesAnonymousSession(t *testing.T) {
	t.Parallel()

	repo, err := store.OpenSQLite(":memory:")
	if err != nil {
		t.Fatalf("OpenSQLite() error = %v", err)
	}
	t.Cleanup(func() { _ = repo.Close() })

	svc := NewService(repo, &fakeGateway{}, Config{
		VisitorSecret: "secret-key",
	})

	ctx := context.Background()
	session, created, err := svc.EnsureSession(ctx, "")
	if err != nil {
		t.Fatalf("EnsureSession() error = %v", err)
	}

	if !created {
		t.Fatal("created = false, want true for first session")
	}

	reused, created, err := svc.EnsureSession(ctx, session.Token)
	if err != nil {
		t.Fatalf("EnsureSession() reuse error = %v", err)
	}

	if created {
		t.Fatal("created = true, want false for existing session")
	}

	if reused.ID != session.ID {
		t.Fatalf("reused.ID = %d, want %d", reused.ID, session.ID)
	}
}

func TestServiceSubmitGenerationPersistsResultAndBroadcastsEvents(t *testing.T) {
	t.Parallel()

	repo, err := store.OpenSQLite(":memory:")
	if err != nil {
		t.Fatalf("OpenSQLite() error = %v", err)
	}
	t.Cleanup(func() { _ = repo.Close() })

	svc := NewService(repo, &fakeGateway{
		generateResp: upstream.GenerateResponse{
			ParentMessageID: "parent-1",
		},
		streamEvents: []upstream.StreamEvent{
			{ID: "AiArt-parent-1"},
			{
				ID:        "AiArt-parent-1",
				ImageURL:  "https://img.example/result.png",
				MessageID: "parent-1-0",
				Result:    "success",
			},
			{ID: "AiArt-parent-1", FinishReason: "stop"},
		},
	}, Config{
		VisitorSecret: "secret-key",
		DefaultRatio:  "1:1",
		DefaultRes:    "1K",
	})

	ctx := context.Background()
	session, _, err := svc.EnsureSession(ctx, "")
	if err != nil {
		t.Fatalf("EnsureSession() error = %v", err)
	}

	task, err := svc.SubmitGeneration(ctx, session.Token, SubmitGenerationInput{
		Prompt: "a chrome rose under moonlight",
	})
	if err != nil {
		t.Fatalf("SubmitGeneration() error = %v", err)
	}

	stream, cancel := svc.SubscribeTask(task.ID)
	defer cancel()

	timeout := time.After(3 * time.Second)
	var gotSucceeded bool
	for !gotSucceeded {
		select {
		case event := <-stream:
			if event.Task.Status == store.TaskStatusSucceeded {
				gotSucceeded = true
			}
		case <-timeout:
			t.Fatal("timed out waiting for succeeded task event")
		}
	}

	saved, err := repo.GetTask(ctx, task.ID)
	if err != nil {
		t.Fatalf("GetTask() error = %v", err)
	}

	if saved.Status != store.TaskStatusSucceeded {
		t.Fatalf("saved.Status = %q, want %q", saved.Status, store.TaskStatusSucceeded)
	}

	if saved.ResultURL != "https://img.example/result.png" {
		t.Fatalf("saved.ResultURL = %q, want final result", saved.ResultURL)
	}
}

func TestServiceMarksTaskFailedWhenUpstreamErrors(t *testing.T) {
	t.Parallel()

	repo, err := store.OpenSQLite(":memory:")
	if err != nil {
		t.Fatalf("OpenSQLite() error = %v", err)
	}
	t.Cleanup(func() { _ = repo.Close() })

	svc := NewService(repo, &fakeGateway{
		generateErr: context.DeadlineExceeded,
	}, Config{
		VisitorSecret: "secret-key",
	})

	ctx := context.Background()
	session, _, err := svc.EnsureSession(ctx, "")
	if err != nil {
		t.Fatalf("EnsureSession() error = %v", err)
	}

	task, err := svc.SubmitGeneration(ctx, session.Token, SubmitGenerationInput{
		Prompt: "storm glass flower",
	})
	if err != nil {
		t.Fatalf("SubmitGeneration() error = %v", err)
	}

	time.Sleep(150 * time.Millisecond)

	saved, err := repo.GetTask(ctx, task.ID)
	if err != nil {
		t.Fatalf("GetTask() error = %v", err)
	}

	if saved.Status != store.TaskStatusFailed {
		t.Fatalf("saved.Status = %q, want %q", saved.Status, store.TaskStatusFailed)
	}

	if !strings.Contains(saved.ErrorMessage, "deadline") {
		t.Fatalf("saved.ErrorMessage = %q, want upstream error details", saved.ErrorMessage)
	}
}

func TestServiceSubmitGenerationUsesFreshVisitorHeaderPerTask(t *testing.T) {
	t.Parallel()

	repo, err := store.OpenSQLite(":memory:")
	if err != nil {
		t.Fatalf("OpenSQLite() error = %v", err)
	}
	t.Cleanup(func() { _ = repo.Close() })

	gateway := &fakeGateway{
		generateResp: upstream.GenerateResponse{
			ParentMessageID: "parent-1",
		},
		streamEvents: []upstream.StreamEvent{
			{ImageURL: "https://img.example/result.png", MessageID: "parent-1-0", Result: "success"},
		},
	}

	svc := NewService(repo, gateway, Config{
		VisitorSecret: "secret-key",
	})

	ctx := context.Background()
	session, _, err := svc.EnsureSession(ctx, "")
	if err != nil {
		t.Fatalf("EnsureSession() error = %v", err)
	}

	if _, err := svc.SubmitGeneration(ctx, session.Token, SubmitGenerationInput{Prompt: "first task"}); err != nil {
		t.Fatalf("SubmitGeneration(first) error = %v", err)
	}
	if _, err := svc.SubmitGeneration(ctx, session.Token, SubmitGenerationInput{Prompt: "second task"}); err != nil {
		t.Fatalf("SubmitGeneration(second) error = %v", err)
	}

	time.Sleep(200 * time.Millisecond)

	gateway.mu.Lock()
	defer gateway.mu.Unlock()

	if len(gateway.generateHeaders) != 2 {
		t.Fatalf("generate header count = %d, want 2", len(gateway.generateHeaders))
	}

	firstGenerateHeader := gateway.generateHeaders[0]
	secondGenerateHeader := gateway.generateHeaders[1]
	if firstGenerateHeader == secondGenerateHeader {
		t.Fatalf("generate headers matched, want a fresh visitor header per task: %q", firstGenerateHeader)
	}

	if len(gateway.streamHeaders) == 0 {
		t.Fatal("stream headers = 0, want task stream calls to reuse task-scoped visitor header")
	}
}
