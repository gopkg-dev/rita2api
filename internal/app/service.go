package app

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"sync"
	"time"

	"rita.ai/internal/store"
	"rita.ai/internal/upstream"
)

type Gateway interface {
	Generate(ctx context.Context, visitorHeader string, req upstream.GenerateRequest) (upstream.GenerateResponse, error)
	Stream(ctx context.Context, visitorHeader, parentMessageID string) ([]upstream.StreamEvent, error)
}

type Config struct {
	VisitorSecret   string
	DefaultRatio    string
	DefaultRes      string
	DefaultImageNum int
}

type SubmitGenerationInput struct {
	Prompt     string
	Ratio      string
	Resolution string
	ImageNum   int
}

type TaskEvent struct {
	Type string     `json:"type"`
	Task store.Task `json:"task"`
}

type Service struct {
	repo    *store.Store
	gateway Gateway
	cfg     Config

	mu          sync.RWMutex
	subscribers map[string]map[chan TaskEvent]struct{}
}

func NewService(repo *store.Store, gateway Gateway, cfg Config) *Service {
	if cfg.DefaultRatio == "" {
		cfg.DefaultRatio = "1:1"
	}
	if cfg.DefaultRes == "" {
		cfg.DefaultRes = "1K"
	}
	if cfg.DefaultImageNum == 0 {
		cfg.DefaultImageNum = 1
	}

	return &Service{
		repo:        repo,
		gateway:     gateway,
		cfg:         cfg,
		subscribers: make(map[string]map[chan TaskEvent]struct{}),
	}
}

func (s *Service) EnsureSession(ctx context.Context, token string) (store.AnonymousSession, bool, error) {
	if token != "" {
		session, err := s.repo.GetSessionByToken(ctx, token)
		if err == nil {
			_ = s.repo.TouchSession(ctx, session.ID)
			return session, false, nil
		}
		if !errors.Is(err, store.ErrSessionNotFound) {
			return store.AnonymousSession{}, false, err
		}
	}

	visitorID, err := randomHex(16)
	if err != nil {
		return store.AnonymousSession{}, false, err
	}
	visitorHeader, err := upstream.BuildVisitorIDHeader(visitorID, s.cfg.VisitorSecret)
	if err != nil {
		return store.AnonymousSession{}, false, err
	}

	sessionToken, err := randomHex(24)
	if err != nil {
		return store.AnonymousSession{}, false, err
	}

	session, err := s.repo.CreateAnonymousSession(ctx, sessionToken, visitorHeader)
	if err != nil {
		return store.AnonymousSession{}, false, err
	}

	return session, true, nil
}

func (s *Service) SubmitGeneration(ctx context.Context, sessionToken string, input SubmitGenerationInput) (store.Task, error) {
	session, err := s.repo.GetSessionByToken(ctx, sessionToken)
	if err != nil {
		return store.Task{}, err
	}

	prompt := input.Prompt
	if prompt == "" {
		return store.Task{}, errors.New("prompt is required")
	}

	if input.Ratio == "" {
		input.Ratio = s.cfg.DefaultRatio
	}
	if input.Resolution == "" {
		input.Resolution = s.cfg.DefaultRes
	}
	if input.ImageNum == 0 {
		input.ImageNum = s.cfg.DefaultImageNum
	}

	visitorHeader, err := s.newVisitorHeader()
	if err != nil {
		return store.Task{}, err
	}

	task, err := s.repo.CreateTask(ctx, store.CreateTaskParams{
		SessionID:     session.ID,
		VisitorHeader: visitorHeader,
		Prompt:        prompt,
		Ratio:         input.Ratio,
		Resolution:    input.Resolution,
		ImageNum:      input.ImageNum,
		Status:        store.TaskStatusQueued,
	})
	if err != nil {
		return store.Task{}, err
	}

	s.publish(task.ID, TaskEvent{Type: "queued", Task: task})
	go s.runTask(context.Background(), task)

	return task, nil
}

func (s *Service) SubscribeTask(taskID string) (<-chan TaskEvent, func()) {
	ch := make(chan TaskEvent, 8)

	s.mu.Lock()
	if _, ok := s.subscribers[taskID]; !ok {
		s.subscribers[taskID] = make(map[chan TaskEvent]struct{})
	}
	s.subscribers[taskID][ch] = struct{}{}
	s.mu.Unlock()

	cancel := func() {
		s.mu.Lock()
		defer s.mu.Unlock()
		if subs, ok := s.subscribers[taskID]; ok {
			delete(subs, ch)
			if len(subs) == 0 {
				delete(s.subscribers, taskID)
			}
		}
		close(ch)
	}

	if snapshot, err := s.repo.GetTask(context.Background(), taskID); err == nil {
		ch <- TaskEvent{Type: string(snapshot.Status), Task: snapshot}
	}

	return ch, cancel
}

func (s *Service) GetTask(ctx context.Context, taskID string) (store.Task, error) {
	return s.repo.GetTask(ctx, taskID)
}

func (s *Service) ListHistory(ctx context.Context, sessionToken string, page, limit int) (store.TaskList, error) {
	session, err := s.repo.GetSessionByToken(ctx, sessionToken)
	if err != nil {
		return store.TaskList{}, err
	}
	return s.repo.ListHistory(ctx, session.ID, page, limit)
}

func (s *Service) ListGallery(ctx context.Context, page, limit int) (store.TaskList, error) {
	return s.repo.ListGallery(ctx, page, limit)
}

func (s *Service) RetryTask(ctx context.Context, sessionToken, taskID string) (store.Task, error) {
	existing, err := s.GetTask(ctx, taskID)
	if err != nil {
		return store.Task{}, err
	}
	return s.SubmitGeneration(ctx, sessionToken, SubmitGenerationInput{
		Prompt:     existing.Prompt,
		Ratio:      existing.Ratio,
		Resolution: existing.Resolution,
		ImageNum:   existing.ImageNum,
	})
}

func (s *Service) RecoverRunningTasks(ctx context.Context) error {
	tasks, err := s.repo.ListRecoverableTasks(ctx)
	if err != nil {
		return err
	}

	for _, task := range tasks {
		if task.ParentMessageID != "" && task.Status == store.TaskStatusRunning {
			go s.resumeTask(context.Background(), task)
			continue
		}

		go s.runTask(context.Background(), task)
	}

	return nil
}

func (s *Service) runTask(ctx context.Context, task store.Task) {
	generateResp, err := s.gateway.Generate(ctx, task.VisitorHeader, upstream.GenerateRequest{
		Prompt:     task.Prompt,
		Ratio:      task.Ratio,
		Resolution: task.Resolution,
		ImageNum:   task.ImageNum,
	})
	if err != nil {
		s.failTask(task.ID, task, err)
		return
	}

	if err := s.repo.UpdateTaskStart(ctx, task.ID, generateResp.ParentMessageID); err != nil {
		s.failTask(task.ID, task, err)
		return
	}

	running, err := s.repo.GetTask(ctx, task.ID)
	if err != nil {
		s.failTask(task.ID, task, err)
		return
	}
	s.publish(task.ID, TaskEvent{Type: "running", Task: running})

	events, err := s.gateway.Stream(ctx, task.VisitorHeader, generateResp.ParentMessageID)
	if err != nil {
		s.failTask(task.ID, running, err)
		return
	}

	s.handleStreamEvents(ctx, running, generateResp.ParentMessageID, events)
}

func (s *Service) resumeTask(ctx context.Context, task store.Task) {
	events, err := s.gateway.Stream(ctx, task.VisitorHeader, task.ParentMessageID)
	if err != nil {
		s.failTask(task.ID, task, err)
		return
	}

	s.handleStreamEvents(ctx, task, task.ParentMessageID, events)
}

func (s *Service) handleStreamEvents(ctx context.Context, task store.Task, parentMessageID string, events []upstream.StreamEvent) {
	for _, event := range events {
		if event.Result == "success" && event.ImageURL != "" {
			if err := s.repo.UpdateTaskResult(ctx, store.UpdateTaskResultParams{
				TaskID:          task.ID,
				Status:          store.TaskStatusSucceeded,
				ParentMessageID: parentMessageID,
				MessageID:       event.MessageID,
				ResultURL:       event.ImageURL,
				FinishedAt:      time.Now().UTC(),
				IsPublic:        true,
			}); err != nil {
				s.failTask(task.ID, task, err)
				return
			}

			succeeded, err := s.repo.GetTask(ctx, task.ID)
			if err != nil {
				s.failTask(task.ID, task, err)
				return
			}

			s.publish(task.ID, TaskEvent{Type: "result", Task: succeeded})
			s.publish(task.ID, TaskEvent{Type: "done", Task: succeeded})
			return
		}
	}

	s.failTask(task.ID, task, errors.New("rita stream ended without result"))
}

func (s *Service) failTask(taskID string, fallback store.Task, cause error) {
	_ = s.repo.UpdateTaskResult(context.Background(), store.UpdateTaskResultParams{
		TaskID:       taskID,
		Status:       store.TaskStatusFailed,
		ErrorMessage: cause.Error(),
		FinishedAt:   time.Now().UTC(),
	})

	failed, err := s.repo.GetTask(context.Background(), taskID)
	if err != nil {
		fallback.Status = store.TaskStatusFailed
		fallback.ErrorMessage = cause.Error()
		s.publish(taskID, TaskEvent{Type: "failed", Task: fallback})
		return
	}

	s.publish(taskID, TaskEvent{Type: "failed", Task: failed})
	s.publish(taskID, TaskEvent{Type: "done", Task: failed})
}

func (s *Service) publish(taskID string, event TaskEvent) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	for ch := range s.subscribers[taskID] {
		select {
		case ch <- event:
		default:
		}
	}
}

func randomHex(bytesCount int) (string, error) {
	raw := make([]byte, bytesCount)
	if _, err := rand.Read(raw); err != nil {
		return "", fmt.Errorf("generate random hex: %w", err)
	}
	return hex.EncodeToString(raw), nil
}

func (s *Service) newVisitorHeader() (string, error) {
	visitorID, err := randomHex(16)
	if err != nil {
		return "", err
	}

	return upstream.BuildVisitorIDHeader(visitorID, s.cfg.VisitorSecret)
}
