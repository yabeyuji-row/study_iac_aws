package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"study_iac_aws/internal/todo"
)

func TestRunRequiresDatabaseURL(t *testing.T) {
	t.Setenv("DATABASE_URL", "")
	t.Setenv("HTTP_ADDR", "")

	err := run(context.Background())
	if err == nil {
		t.Fatal("run() error = nil, want error")
	}
	if !strings.Contains(err.Error(), "load config") {
		t.Fatalf("run() error = %v, want load config context", err)
	}
}

func TestOperationalEndpoints(t *testing.T) {
	todoHandler := todo.NewHandler(todo.NewService(fakeRepository{}, time.Now))
	echoServer := newEchoServer(todoHandler, fakePinger{}, buildInfo{
		Version:   "test-version",
		Commit:    "test-commit",
		BuildTime: "test-build-time",
	})

	tests := []struct {
		name       string
		path       string
		wantStatus int
		wantBody   string
	}{
		{
			name:       "health",
			path:       "/health",
			wantStatus: http.StatusOK,
			wantBody:   `"status":"ok"`,
		},
		{
			name:       "healthz",
			path:       "/healthz",
			wantStatus: http.StatusOK,
			wantBody:   `"status":"ok"`,
		},
		{
			name:       "ready",
			path:       "/ready",
			wantStatus: http.StatusOK,
			wantBody:   `"status":"ready"`,
		},
		{
			name:       "readyz",
			path:       "/readyz",
			wantStatus: http.StatusOK,
			wantBody:   `"status":"ready"`,
		},
		{
			name:       "version",
			path:       "/version",
			wantStatus: http.StatusOK,
			wantBody:   `"version":"test-version"`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request := httptest.NewRequest(http.MethodGet, tt.path, nil)
			recorder := httptest.NewRecorder()

			echoServer.ServeHTTP(recorder, request)

			if recorder.Code != tt.wantStatus {
				t.Fatalf("status = %d, body = %s", recorder.Code, recorder.Body.String())
			}
			if !strings.Contains(recorder.Body.String(), tt.wantBody) {
				t.Fatalf("body = %s, want %s", recorder.Body.String(), tt.wantBody)
			}
		})
	}
}

func TestReadyReturnsUnavailableWhenDatabasePingFails(t *testing.T) {
	todoHandler := todo.NewHandler(todo.NewService(fakeRepository{}, time.Now))
	echoServer := newEchoServer(
		todoHandler,
		fakePinger{err: errors.New("database unavailable")},
		buildInfo{},
	)

	request := httptest.NewRequest(http.MethodGet, "/readyz", nil)
	recorder := httptest.NewRecorder()

	echoServer.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusServiceUnavailable {
		t.Fatalf("status = %d, body = %s", recorder.Code, recorder.Body.String())
	}
	if !strings.Contains(recorder.Body.String(), `"status":"unready"`) {
		t.Fatalf("body = %s, want unready status", recorder.Body.String())
	}
}

func TestRequestLogMiddlewareEmitsMetricFields(t *testing.T) {
	var logs bytes.Buffer
	previousLogger := slog.Default()
	slog.SetDefault(slog.New(slog.NewJSONHandler(&logs, nil)))
	t.Cleanup(func() {
		slog.SetDefault(previousLogger)
	})

	todoHandler := todo.NewHandler(todo.NewService(fakeRepository{}, time.Now))
	echoServer := newEchoServer(todoHandler, fakePinger{}, buildInfo{})

	request := httptest.NewRequest(http.MethodGet, "/healthz?ignored=true", nil)
	request.RemoteAddr = "192.0.2.10:12345"
	request.Header.Set(echoHeaderXRequestID, "request-1")
	recorder := httptest.NewRecorder()

	echoServer.ServeHTTP(recorder, request)

	var entry map[string]any
	if err := json.Unmarshal(logs.Bytes(), &entry); err != nil {
		t.Fatalf("parse log entry: %v; log = %s", err, logs.String())
	}

	assertLogField(t, entry, "event", "http_request_completed")
	assertLogField(t, entry, "method", http.MethodGet)
	assertLogField(t, entry, "path", "/healthz")
	assertLogField(t, entry, "request_id", "request-1")
	if got := entry["status"]; got != float64(http.StatusOK) {
		t.Fatalf("status = %v, want %d", got, http.StatusOK)
	}
	if _, ok := entry["duration_ms"].(float64); !ok {
		t.Fatalf("duration_ms = %T, want number", entry["duration_ms"])
	}
}

func assertLogField(t *testing.T, entry map[string]any, key string, want string) {
	t.Helper()

	if got := entry[key]; got != want {
		t.Fatalf("%s = %v, want %q", key, got, want)
	}
}

const echoHeaderXRequestID = "X-Request-Id"

type fakePinger struct {
	err error
}

func (pinger fakePinger) Ping(ctx context.Context) (err error) {
	return pinger.err
}

type fakeRepository struct{}

func (repository fakeRepository) Create(
	ctx context.Context,
	todoItem todo.Todo,
) (created todo.Todo, err error) {
	return todoItem, nil
}

func (repository fakeRepository) List(
	ctx context.Context,
	filter todo.ListFilter,
) (result todo.ListResult, err error) {
	return todo.ListResult{}, nil
}

func (repository fakeRepository) Get(
	ctx context.Context,
	id string,
) (todoItem todo.Todo, err error) {
	return todo.Todo{}, todo.ErrTodoNotFound
}

func (repository fakeRepository) Update(
	ctx context.Context,
	input todo.UpdateInput,
) (todoItem todo.Todo, err error) {
	return todo.Todo{}, todo.ErrTodoNotFound
}

func (repository fakeRepository) Delete(ctx context.Context, id string) (err error) {
	return todo.ErrTodoNotFound
}
