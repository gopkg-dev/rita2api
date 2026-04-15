package httpapi

import (
	"context"
	"embed"
	"encoding/json"
	"errors"
	"io/fs"
	"net/http"
	"path"
	"strconv"
	"strings"
	"time"

	"rita.ai/internal/app"
	"rita.ai/internal/store"
)

//go:embed webdist/*
var embeddedStatic embed.FS

type ServerConfig struct {
	CookieName string
}

type Server struct {
	service *app.Service
	cfg     ServerConfig
	mux     *http.ServeMux
}

type apiResponse struct {
	Data  any    `json:"data"`
	Error string `json:"error,omitempty"`
}

type taskDTO struct {
	ID              string `json:"id"`
	Prompt          string `json:"prompt"`
	Ratio           string `json:"ratio"`
	Resolution      string `json:"resolution"`
	ImageNum        int    `json:"imageNum"`
	Status          string `json:"status"`
	ParentMessageID string `json:"parentMessageId"`
	MessageID       string `json:"messageId"`
	ResultURL       string `json:"resultUrl"`
	ErrorMessage    string `json:"errorMessage"`
	IsPublic        bool   `json:"isPublic"`
	CreatedAt       string `json:"createdAt"`
	UpdatedAt       string `json:"updatedAt"`
	FinishedAt      string `json:"finishedAt"`
}

func NewServer(service *app.Service, cfg ServerConfig) *Server {
	if cfg.CookieName == "" {
		cfg.CookieName = "rita_session"
	}

	s := &Server{
		service: service,
		cfg:     cfg,
		mux:     http.NewServeMux(),
	}
	s.routes()
	return s
}

func (s *Server) Handler() http.Handler {
	return s.mux
}

func (s *Server) routes() {
	s.mux.HandleFunc("GET /api/v1/bootstrap", s.handleBootstrap)
	s.mux.HandleFunc("POST /api/v1/sessions/anonymous", s.handleAnonymousSession)
	s.mux.HandleFunc("POST /api/v1/generations", s.handleCreateGeneration)
	s.mux.HandleFunc("GET /api/v1/history", s.handleHistory)
	s.mux.HandleFunc("GET /api/v1/gallery", s.handleGallery)
	s.mux.HandleFunc("GET /api/v1/generations/", s.handleGenerationRoutes)
	s.mux.HandleFunc("POST /api/v1/generations/", s.handleGenerationRoutes)
	s.mux.Handle("/", s.handleStatic())
}

func (s *Server) handleBootstrap(w http.ResponseWriter, r *http.Request) {
	session, _ := s.ensureSession(w, r)
	s.writeJSON(w, http.StatusOK, apiResponse{
		Data: map[string]any{
			"session": map[string]any{
				"token": session.Token,
			},
			"defaults": map[string]any{
				"ratio":      "1:1",
				"resolution": "1K",
				"imageNum":   1,
			},
			"gallery": map[string]any{
				"enabled": true,
			},
		},
	})
}

func (s *Server) handleAnonymousSession(w http.ResponseWriter, r *http.Request) {
	session, created := s.ensureSession(w, r)
	status := http.StatusOK
	if created {
		status = http.StatusCreated
	}

	s.writeJSON(w, status, apiResponse{
		Data: map[string]any{
			"token": session.Token,
		},
	})
}

func (s *Server) handleCreateGeneration(w http.ResponseWriter, r *http.Request) {
	session, _ := s.ensureSession(w, r)

	var input struct {
		Prompt     string `json:"prompt"`
		Ratio      string `json:"ratio"`
		Resolution string `json:"resolution"`
		ImageNum   int    `json:"imageNum"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		s.writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	task, err := s.service.SubmitGeneration(r.Context(), session.Token, app.SubmitGenerationInput{
		Prompt:     input.Prompt,
		Ratio:      input.Ratio,
		Resolution: input.Resolution,
		ImageNum:   input.ImageNum,
	})
	if err != nil {
		s.writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	s.writeJSON(w, http.StatusAccepted, apiResponse{Data: toTaskDTO(task)})
}

func (s *Server) handleHistory(w http.ResponseWriter, r *http.Request) {
	session, _ := s.ensureSession(w, r)
	page, limit := readPagination(r)
	history, err := s.service.ListHistory(r.Context(), session.Token, page, limit)
	if err != nil {
		s.writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	s.writeJSON(w, http.StatusOK, apiResponse{Data: toTaskListDTO(history)})
}

func (s *Server) handleGallery(w http.ResponseWriter, r *http.Request) {
	page, limit := readPagination(r)
	gallery, err := s.service.ListGallery(r.Context(), page, limit)
	if err != nil {
		s.writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	s.writeJSON(w, http.StatusOK, apiResponse{Data: toTaskListDTO(gallery)})
}

func (s *Server) handleGenerationRoutes(w http.ResponseWriter, r *http.Request) {
	taskPath := strings.TrimPrefix(r.URL.Path, "/api/v1/generations/")
	parts := strings.Split(strings.Trim(taskPath, "/"), "/")
	if len(parts) == 0 || parts[0] == "" {
		http.NotFound(w, r)
		return
	}

	taskID := parts[0]
	if len(parts) == 1 && r.Method == http.MethodGet {
		s.handleGetTask(w, r, taskID)
		return
	}

	if len(parts) == 2 && parts[1] == "stream" && r.Method == http.MethodGet {
		s.handleTaskStream(w, r, taskID)
		return
	}

	if len(parts) == 2 && parts[1] == "retry" && r.Method == http.MethodPost {
		s.handleRetryTask(w, r, taskID)
		return
	}

	http.NotFound(w, r)
}

func (s *Server) handleGetTask(w http.ResponseWriter, r *http.Request, taskID string) {
	task, err := s.service.GetTask(r.Context(), taskID)
	if err != nil {
		status := http.StatusInternalServerError
		if errors.Is(err, store.ErrTaskNotFound) {
			status = http.StatusNotFound
		}
		s.writeError(w, status, err.Error())
		return
	}
	s.writeJSON(w, http.StatusOK, apiResponse{Data: toTaskDTO(task)})
}

func (s *Server) handleRetryTask(w http.ResponseWriter, r *http.Request, taskID string) {
	session, _ := s.ensureSession(w, r)
	task, err := s.service.RetryTask(r.Context(), session.Token, taskID)
	if err != nil {
		s.writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	s.writeJSON(w, http.StatusAccepted, apiResponse{Data: toTaskDTO(task)})
}

func (s *Server) handleTaskStream(w http.ResponseWriter, r *http.Request, taskID string) {
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	flusher, ok := w.(http.Flusher)
	if !ok {
		s.writeError(w, http.StatusInternalServerError, "streaming unsupported")
		return
	}

	stream, cancel := s.service.SubscribeTask(taskID)
	defer cancel()

	ctx := r.Context()
	for {
		select {
		case <-ctx.Done():
			return
		case event := <-stream:
			payload, err := json.Marshal(map[string]any{
				"type": event.Type,
				"task": toTaskDTO(event.Task),
			})
			if err != nil {
				return
			}
			_, _ = w.Write([]byte("data: "))
			_, _ = w.Write(payload)
			_, _ = w.Write([]byte("\n\n"))
			flusher.Flush()
			if event.Type == "done" || event.Type == string(store.TaskStatusSucceeded) || event.Type == string(store.TaskStatusFailed) {
				return
			}
		}
	}
}

func (s *Server) ensureSession(w http.ResponseWriter, r *http.Request) (store.AnonymousSession, bool) {
	var token string
	if cookie, err := r.Cookie(s.cfg.CookieName); err == nil {
		token = cookie.Value
	}

	session, created, err := s.service.EnsureSession(r.Context(), token)
	if err != nil {
		session, created, _ = s.service.EnsureSession(context.Background(), "")
	}

	http.SetCookie(w, &http.Cookie{
		Name:     s.cfg.CookieName,
		Value:    session.Token,
		Path:     "/",
		Expires:  time.Now().Add(365 * 24 * time.Hour),
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})

	return session, created
}

func (s *Server) handleStatic() http.Handler {
	sub, err := fs.Sub(embeddedStatic, "webdist")
	if err != nil {
		return http.NotFoundHandler()
	}
	fileServer := http.FileServer(http.FS(sub))
	indexHTML, err := fs.ReadFile(sub, "index.html")
	if err != nil {
		return http.NotFoundHandler()
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/api/") {
			http.NotFound(w, r)
			return
		}

		if ext := path.Ext(r.URL.Path); ext == "" {
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write(indexHTML)
			return
		}

		fileServer.ServeHTTP(w, r)
	})
}

func (s *Server) writeJSON(w http.ResponseWriter, status int, payload apiResponse) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

func (s *Server) writeError(w http.ResponseWriter, status int, message string) {
	s.writeJSON(w, status, apiResponse{Error: message})
}

func readPagination(r *http.Request) (int, int) {
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 20
	}
	return page, limit
}

func toTaskDTO(task store.Task) taskDTO {
	finishedAt := ""
	if task.FinishedAt.Valid {
		finishedAt = task.FinishedAt.Time.Format(time.RFC3339)
	}

	return taskDTO{
		ID:              task.ID,
		Prompt:          task.Prompt,
		Ratio:           task.Ratio,
		Resolution:      task.Resolution,
		ImageNum:        task.ImageNum,
		Status:          string(task.Status),
		ParentMessageID: task.ParentMessageID,
		MessageID:       task.MessageID,
		ResultURL:       task.ResultURL,
		ErrorMessage:    task.ErrorMessage,
		IsPublic:        task.IsPublic,
		CreatedAt:       task.CreatedAt.Format(time.RFC3339),
		UpdatedAt:       task.UpdatedAt.Format(time.RFC3339),
		FinishedAt:      finishedAt,
	}
}

func toTaskListDTO(list store.TaskList) map[string]any {
	items := make([]taskDTO, 0, len(list.Items))
	for _, item := range list.Items {
		items = append(items, toTaskDTO(item))
	}

	return map[string]any{
		"total": list.Total,
		"items": items,
	}
}
