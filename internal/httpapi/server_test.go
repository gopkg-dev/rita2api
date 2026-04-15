package httpapi

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"rita.ai/internal/app"
	"rita.ai/internal/store"
	"rita.ai/internal/upstream"
)

type fakeGateway struct {
	generateResp upstream.GenerateResponse
	streamEvents []upstream.StreamEvent
}

func (f fakeGateway) Generate(_ context.Context, _ string, _ upstream.GenerateRequest) (upstream.GenerateResponse, error) {
	return f.generateResp, nil
}

func (f fakeGateway) Stream(_ context.Context, _ string, _ string) ([]upstream.StreamEvent, error) {
	return f.streamEvents, nil
}

func TestServerBootstrapCreatesAnonymousCookie(t *testing.T) {
	t.Parallel()

	repo, err := store.OpenSQLite(":memory:")
	if err != nil {
		t.Fatalf("OpenSQLite() error = %v", err)
	}
	t.Cleanup(func() { _ = repo.Close() })

	svc := app.NewService(repo, fakeGateway{}, app.Config{
		VisitorSecret: "secret-key",
	})

	server := NewServer(svc, ServerConfig{})

	req := httptest.NewRequest(http.MethodGet, "/api/v1/bootstrap", nil)
	rec := httptest.NewRecorder()
	server.Handler().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}

	cookies := rec.Result().Cookies()
	if len(cookies) == 0 {
		t.Fatal("cookies = 0, want anonymous session cookie")
	}
}

func TestServerCreatesTaskAndReturnsHistory(t *testing.T) {
	t.Parallel()

	repo, err := store.OpenSQLite(":memory:")
	if err != nil {
		t.Fatalf("OpenSQLite() error = %v", err)
	}
	t.Cleanup(func() { _ = repo.Close() })

	svc := app.NewService(repo, fakeGateway{
		generateResp: upstream.GenerateResponse{
			ParentMessageID: "parent-1",
		},
		streamEvents: []upstream.StreamEvent{
			{ImageURL: "https://img.example/result.png", MessageID: "parent-1-0", Result: "success"},
			{FinishReason: "stop"},
		},
	}, app.Config{
		VisitorSecret: "secret-key",
	})

	server := NewServer(svc, ServerConfig{})

	bootstrapReq := httptest.NewRequest(http.MethodGet, "/api/v1/bootstrap", nil)
	bootstrapRec := httptest.NewRecorder()
	server.Handler().ServeHTTP(bootstrapRec, bootstrapReq)
	sessionCookie := bootstrapRec.Result().Cookies()[0]

	createReq := httptest.NewRequest(http.MethodPost, "/api/v1/generations", strings.NewReader(`{"prompt":"liquid chrome flower"}`))
	createReq.AddCookie(sessionCookie)
	createReq.Header.Set("Content-Type", "application/json")
	createRec := httptest.NewRecorder()
	server.Handler().ServeHTTP(createRec, createReq)

	if createRec.Code != http.StatusAccepted {
		t.Fatalf("create status = %d, want 202", createRec.Code)
	}

	time.Sleep(150 * time.Millisecond)

	historyReq := httptest.NewRequest(http.MethodGet, "/api/v1/history?page=1&limit=10", nil)
	historyReq.AddCookie(sessionCookie)
	historyRec := httptest.NewRecorder()
	server.Handler().ServeHTTP(historyRec, historyReq)

	if historyRec.Code != http.StatusOK {
		t.Fatalf("history status = %d, want 200", historyRec.Code)
	}

	var payload struct {
		Data struct {
			Total int `json:"total"`
		} `json:"data"`
	}
	if err := json.Unmarshal(historyRec.Body.Bytes(), &payload); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}

	if payload.Data.Total != 1 {
		t.Fatalf("history total = %d, want 1", payload.Data.Total)
	}
}
