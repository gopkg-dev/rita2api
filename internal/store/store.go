package store

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"time"

	_ "modernc.org/sqlite"
)

type TaskStatus string

var (
	ErrTaskNotFound    = errors.New("task not found")
	ErrSessionNotFound = errors.New("session not found")
)

const (
	TaskStatusQueued    TaskStatus = "queued"
	TaskStatusRunning   TaskStatus = "running"
	TaskStatusSucceeded TaskStatus = "succeeded"
	TaskStatusFailed    TaskStatus = "failed"
)

type Store struct {
	db *sql.DB
}

type AnonymousSession struct {
	ID            int64
	Token         string
	VisitorHeader string
	CreatedAt     time.Time
	LastSeenAt    time.Time
}

type Task struct {
	ID              string
	SessionID       int64
	VisitorHeader   string
	Prompt          string
	Ratio           string
	Resolution      string
	ImageNum        int
	Status          TaskStatus
	ParentMessageID string
	MessageID       string
	ResultURL       string
	ErrorMessage    string
	IsPublic        bool
	CreatedAt       time.Time
	UpdatedAt       time.Time
	FinishedAt      sql.NullTime
}

type CreateTaskParams struct {
	SessionID       int64
	VisitorHeader   string
	Prompt          string
	Ratio           string
	Resolution      string
	ImageNum        int
	Status          TaskStatus
	ParentMessageID string
	IsPublic        bool
}

type UpdateTaskResultParams struct {
	TaskID          string
	Status          TaskStatus
	ParentMessageID string
	MessageID       string
	ResultURL       string
	ErrorMessage    string
	FinishedAt      time.Time
	IsPublic        bool
}

type TaskList struct {
	Total int
	Items []Task
}

func OpenSQLite(dsn string) (*Store, error) {
	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(1)

	store := &Store{db: db}
	if err := store.migrate(context.Background()); err != nil {
		_ = db.Close()
		return nil, err
	}

	return store, nil
}

func (s *Store) Close() error {
	return s.db.Close()
}

func (s *Store) migrate(ctx context.Context) error {
	statements := []string{
		`CREATE TABLE IF NOT EXISTS anonymous_sessions (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			token TEXT NOT NULL UNIQUE,
			visitor_header TEXT NOT NULL,
			created_at TEXT NOT NULL,
			last_seen_at TEXT NOT NULL
		);`,
		`CREATE TABLE IF NOT EXISTS generation_tasks (
			id TEXT PRIMARY KEY,
			session_id INTEGER NOT NULL,
			visitor_header TEXT NOT NULL DEFAULT '',
			prompt TEXT NOT NULL,
			ratio TEXT NOT NULL,
			resolution TEXT NOT NULL,
			image_num INTEGER NOT NULL,
			status TEXT NOT NULL,
			parent_message_id TEXT NOT NULL DEFAULT '',
			message_id TEXT NOT NULL DEFAULT '',
			result_url TEXT NOT NULL DEFAULT '',
			error_message TEXT NOT NULL DEFAULT '',
			is_public INTEGER NOT NULL DEFAULT 0,
			created_at TEXT NOT NULL,
			updated_at TEXT NOT NULL,
			finished_at TEXT,
			FOREIGN KEY(session_id) REFERENCES anonymous_sessions(id)
		);`,
	}

	for _, stmt := range statements {
		if _, err := s.db.ExecContext(ctx, stmt); err != nil {
			return err
		}
	}

	if _, err := s.db.ExecContext(
		ctx,
		`ALTER TABLE generation_tasks ADD COLUMN visitor_header TEXT NOT NULL DEFAULT ''`,
	); err != nil && !strings.Contains(err.Error(), "duplicate column name") {
		return err
	}

	return nil
}

func (s *Store) CreateAnonymousSession(ctx context.Context, token, visitorHeader string) (AnonymousSession, error) {
	now := time.Now().UTC()
	res, err := s.db.ExecContext(
		ctx,
		`INSERT INTO anonymous_sessions (token, visitor_header, created_at, last_seen_at) VALUES (?, ?, ?, ?)`,
		token,
		visitorHeader,
		now.Format(time.RFC3339Nano),
		now.Format(time.RFC3339Nano),
	)
	if err != nil {
		return AnonymousSession{}, err
	}

	id, err := res.LastInsertId()
	if err != nil {
		return AnonymousSession{}, err
	}

	return AnonymousSession{
		ID:            id,
		Token:         token,
		VisitorHeader: visitorHeader,
		CreatedAt:     now,
		LastSeenAt:    now,
	}, nil
}

func (s *Store) GetSessionByToken(ctx context.Context, token string) (AnonymousSession, error) {
	return s.getSession(
		ctx,
		`SELECT id, token, visitor_header, created_at, last_seen_at
		 FROM anonymous_sessions WHERE token = ?`,
		token,
	)
}

func (s *Store) GetSessionByID(ctx context.Context, sessionID int64) (AnonymousSession, error) {
	return s.getSession(
		ctx,
		`SELECT id, token, visitor_header, created_at, last_seen_at
		 FROM anonymous_sessions WHERE id = ?`,
		sessionID,
	)
}

func (s *Store) getSession(ctx context.Context, query string, arg any) (AnonymousSession, error) {
	row := s.db.QueryRowContext(
		ctx, query, arg,
	)

	var (
		session               AnonymousSession
		createdAt, lastSeenAt string
	)
	if err := row.Scan(
		&session.ID,
		&session.Token,
		&session.VisitorHeader,
		&createdAt,
		&lastSeenAt,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return AnonymousSession{}, ErrSessionNotFound
		}
		return AnonymousSession{}, err
	}

	var err error
	session.CreatedAt, err = time.Parse(time.RFC3339Nano, createdAt)
	if err != nil {
		return AnonymousSession{}, err
	}
	session.LastSeenAt, err = time.Parse(time.RFC3339Nano, lastSeenAt)
	if err != nil {
		return AnonymousSession{}, err
	}

	return session, nil
}

func (s *Store) TouchSession(ctx context.Context, sessionID int64) error {
	_, err := s.db.ExecContext(
		ctx,
		`UPDATE anonymous_sessions SET last_seen_at = ? WHERE id = ?`,
		time.Now().UTC().Format(time.RFC3339Nano),
		sessionID,
	)
	return err
}

func (s *Store) CreateTask(ctx context.Context, params CreateTaskParams) (Task, error) {
	if params.SessionID == 0 {
		return Task{}, errors.New("session id is required")
	}

	if params.Prompt == "" {
		return Task{}, errors.New("prompt is required")
	}

	now := time.Now().UTC()
	taskID, err := newTaskID()
	if err != nil {
		return Task{}, err
	}
	_, err = s.db.ExecContext(
		ctx,
		`INSERT INTO generation_tasks (
			id, session_id, visitor_header, prompt, ratio, resolution, image_num, status,
			parent_message_id, is_public, created_at, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		taskID,
		params.SessionID,
		params.VisitorHeader,
		params.Prompt,
		params.Ratio,
		params.Resolution,
		params.ImageNum,
		string(params.Status),
		params.ParentMessageID,
		boolToInt(params.IsPublic),
		now.Format(time.RFC3339Nano),
		now.Format(time.RFC3339Nano),
	)
	if err != nil {
		return Task{}, err
	}

	return s.GetTask(ctx, taskID)
}

func (s *Store) UpdateTaskResult(ctx context.Context, params UpdateTaskResultParams) error {
	_, err := s.db.ExecContext(
		ctx,
		`UPDATE generation_tasks
		 SET status = ?, parent_message_id = ?, message_id = ?, result_url = ?,
		     error_message = ?, finished_at = ?, is_public = ?, updated_at = ?
		 WHERE id = ?`,
		string(params.Status),
		params.ParentMessageID,
		params.MessageID,
		params.ResultURL,
		params.ErrorMessage,
		params.FinishedAt.UTC().Format(time.RFC3339Nano),
		boolToInt(params.IsPublic),
		time.Now().UTC().Format(time.RFC3339Nano),
		params.TaskID,
	)
	return err
}

func (s *Store) UpdateTaskStart(ctx context.Context, taskID, parentMessageID string) error {
	_, err := s.db.ExecContext(
		ctx,
		`UPDATE generation_tasks
		 SET status = ?, parent_message_id = ?, updated_at = ?
		 WHERE id = ?`,
		string(TaskStatusRunning),
		parentMessageID,
		time.Now().UTC().Format(time.RFC3339Nano),
		taskID,
	)
	return err
}

func (s *Store) GetTask(ctx context.Context, taskID string) (Task, error) {
	row := s.db.QueryRowContext(ctx, `SELECT
			id, session_id, visitor_header, prompt, ratio, resolution, image_num, status,
			parent_message_id, message_id, result_url, error_message, is_public,
			created_at, updated_at, finished_at
		FROM generation_tasks WHERE id = ?`, taskID)
	return scanTask(row)
}

func (s *Store) ListHistory(ctx context.Context, sessionID int64, page, limit int) (TaskList, error) {
	return s.listTasks(ctx, `WHERE session_id = ?`, sessionID, page, limit)
}

func (s *Store) ListGallery(ctx context.Context, page, limit int) (TaskList, error) {
	return s.listTasks(ctx, `WHERE is_public = 1 AND status = ?`, string(TaskStatusSucceeded), page, limit)
}

func (s *Store) ListRecoverableTasks(ctx context.Context) ([]Task, error) {
	rows, err := s.db.QueryContext(
		ctx,
		`SELECT id, session_id, visitor_header, prompt, ratio, resolution, image_num, status,
		        parent_message_id, message_id, result_url, error_message, is_public,
		        created_at, updated_at, finished_at
		 FROM generation_tasks
		 WHERE status IN (?, ?)`,
		string(TaskStatusQueued),
		string(TaskStatusRunning),
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tasks []Task
	for rows.Next() {
		task, err := scanTaskRows(rows)
		if err != nil {
			return nil, err
		}
		tasks = append(tasks, task)
	}

	return tasks, rows.Err()
}

func (s *Store) listTasks(ctx context.Context, whereClause string, filter any, page, limit int) (TaskList, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 20
	}

	var total int
	countQuery := `SELECT COUNT(*) FROM generation_tasks ` + whereClause
	if err := s.db.QueryRowContext(ctx, countQuery, filter).Scan(&total); err != nil {
		return TaskList{}, err
	}

	offset := (page - 1) * limit
	query := `SELECT id, session_id, visitor_header, prompt, ratio, resolution, image_num, status,
		parent_message_id, message_id, result_url, error_message, is_public,
		created_at, updated_at, finished_at
		FROM generation_tasks ` + whereClause + ` ORDER BY created_at DESC LIMIT ? OFFSET ?`
	rows, err := s.db.QueryContext(ctx, query, filter, limit, offset)
	if err != nil {
		return TaskList{}, err
	}
	defer rows.Close()

	items := make([]Task, 0, limit)
	for rows.Next() {
		task, err := scanTaskRows(rows)
		if err != nil {
			return TaskList{}, err
		}
		items = append(items, task)
	}

	return TaskList{Total: total, Items: items}, rows.Err()
}

type scanner interface {
	Scan(dest ...any) error
}

func scanTask(row scanner) (Task, error) {
	var (
		task          Task
		status        string
		isPublic      int
		createdAt     string
		updatedAt     string
		finishedAtRaw sql.NullString
	)

	if err := row.Scan(
		&task.ID,
		&task.SessionID,
		&task.VisitorHeader,
		&task.Prompt,
		&task.Ratio,
		&task.Resolution,
		&task.ImageNum,
		&status,
		&task.ParentMessageID,
		&task.MessageID,
		&task.ResultURL,
		&task.ErrorMessage,
		&isPublic,
		&createdAt,
		&updatedAt,
		&finishedAtRaw,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return Task{}, ErrTaskNotFound
		}
		return Task{}, err
	}

	task.Status = TaskStatus(status)
	task.IsPublic = isPublic == 1

	var err error
	task.CreatedAt, err = time.Parse(time.RFC3339Nano, createdAt)
	if err != nil {
		return Task{}, err
	}

	task.UpdatedAt, err = time.Parse(time.RFC3339Nano, updatedAt)
	if err != nil {
		return Task{}, err
	}

	if finishedAtRaw.Valid && finishedAtRaw.String != "" {
		finishedAt, err := time.Parse(time.RFC3339Nano, finishedAtRaw.String)
		if err != nil {
			return Task{}, err
		}
		task.FinishedAt = sql.NullTime{Time: finishedAt, Valid: true}
	}

	return task, nil
}

func scanTaskRows(rows *sql.Rows) (Task, error) {
	return scanTask(rows)
}

func boolToInt(value bool) int {
	if value {
		return 1
	}
	return 0
}

func newTaskID() (string, error) {
	var raw [8]byte
	if _, err := rand.Read(raw[:]); err != nil {
		return "", fmt.Errorf("generate task id: %w", err)
	}

	return "task_" + hex.EncodeToString(raw[:]), nil
}
