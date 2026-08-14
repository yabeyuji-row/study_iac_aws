package todo

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/labstack/echo/v4"
)

func newTestEcho(handler *Handler) (echoServer *echo.Echo) {
	echoServer = echo.New()
	handler.Register(echoServer.Group("/v1"))

	return echoServer
}

func TestHandlerCreateTodo(t *testing.T) {
	repository := newFakeRepository()
	service := NewService(repository, time.Now)
	handler := NewHandler(service)
	echoServer := newTestEcho(handler)

	body := bytes.NewBufferString(`{"title":"learn Go","status":"pending"}`)
	request := httptest.NewRequest(http.MethodPost, "/v1/todos", body)
	recorder := httptest.NewRecorder()

	echoServer.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusCreated {
		t.Fatalf("status = %d, body = %s", recorder.Code, recorder.Body.String())
	}

	var response Todo
	if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}
	if response.Title != "learn Go" {
		t.Fatalf("Title = %q", response.Title)
	}
}

func TestHandlerInvalidJSON(t *testing.T) {
	service := NewService(newFakeRepository(), time.Now)
	handler := NewHandler(service)
	echoServer := newTestEcho(handler)

	request := httptest.NewRequest(
		http.MethodPost,
		"/v1/todos",
		bytes.NewBufferString(`{"title":`),
	)
	recorder := httptest.NewRecorder()

	echoServer.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, body = %s", recorder.Code, recorder.Body.String())
	}
}

func TestHandlerNotFound(t *testing.T) {
	service := NewService(newFakeRepository(), time.Now)
	handler := NewHandler(service)
	echoServer := newTestEcho(handler)

	request := httptest.NewRequest(
		http.MethodGet,
		"/v1/todos/018f1a54-3c5b-4ef3-9ec4-8ff8e32180b8",
		nil,
	)
	recorder := httptest.NewRecorder()

	echoServer.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusNotFound {
		t.Fatalf("status = %d, body = %s", recorder.Code, recorder.Body.String())
	}
}

func TestHandlerVersionConflict(t *testing.T) {
	repository := newFakeRepository()
	service := NewService(repository, time.Now)
	handler := NewHandler(service)
	echoServer := newTestEcho(handler)

	created, err := service.Create(context.Background(), CreateInput{
		Title:  "learn Go",
		Status: StatusPending,
	})
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	_, err = service.Update(context.Background(), UpdateInput{
		ID:      created.ID,
		Title:   "new",
		Status:  StatusInProgress,
		Version: created.Version,
	})
	if err != nil {
		t.Fatalf("Update() error = %v", err)
	}

	body := bytes.NewBufferString(
		`{"title":"stale","status":"completed","version":1}`,
	)
	request := httptest.NewRequest(
		http.MethodPut,
		"/v1/todos/"+created.ID,
		body,
	)
	recorder := httptest.NewRecorder()

	echoServer.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusConflict {
		t.Fatalf("status = %d, body = %s", recorder.Code, recorder.Body.String())
	}
}
