package todo

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/labstack/echo/v4"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) (handler *Handler) {
	return &Handler{service: service}
}

func (handler *Handler) Register(group *echo.Group) {
	group.POST("/todos", handler.createTodo)
	group.GET("/todos", handler.listTodos)
	group.GET("/todos/:todo_id", handler.getTodo)
	group.PUT("/todos/:todo_id", handler.updateTodo)
	group.DELETE("/todos/:todo_id", handler.deleteTodo)
}

func (handler *Handler) createTodo(echoContext echo.Context) (err error) {
	request := echoContext.Request()
	var input CreateInput
	if err := decodeJSON(request, &input); err != nil {
		return writeError(echoContext, http.StatusBadRequest, err)
	}

	ctx, cancel := context.WithTimeout(request.Context(), 3*time.Second)
	defer cancel()

	todo, err := handler.service.Create(ctx, input)
	if err != nil {
		return writeMappedError(echoContext, err)
	}

	return echoContext.JSON(http.StatusCreated, todo)
}

func (handler *Handler) listTodos(echoContext echo.Context) (err error) {
	request := echoContext.Request()
	filter, err := parseListFilter(request)
	if err != nil {
		return writeError(echoContext, http.StatusBadRequest, err)
	}

	ctx, cancel := context.WithTimeout(request.Context(), 3*time.Second)
	defer cancel()

	result, err := handler.service.List(ctx, filter)
	if err != nil {
		return writeMappedError(echoContext, err)
	}

	return echoContext.JSON(http.StatusOK, result)
}

func (handler *Handler) getTodo(echoContext echo.Context) (err error) {
	request := echoContext.Request()
	id := echoContext.Param("todo_id")

	ctx, cancel := context.WithTimeout(request.Context(), 3*time.Second)
	defer cancel()

	todo, err := handler.service.Get(ctx, id)
	if err != nil {
		return writeMappedError(echoContext, err)
	}

	return echoContext.JSON(http.StatusOK, todo)
}

func (handler *Handler) updateTodo(echoContext echo.Context) (err error) {
	request := echoContext.Request()
	id := echoContext.Param("todo_id")

	var input UpdateInput
	if err := decodeJSON(request, &input); err != nil {
		return writeError(echoContext, http.StatusBadRequest, err)
	}
	input.ID = id

	ctx, cancel := context.WithTimeout(request.Context(), 3*time.Second)
	defer cancel()

	todo, err := handler.service.Update(ctx, input)
	if err != nil {
		return writeMappedError(echoContext, err)
	}

	return echoContext.JSON(http.StatusOK, todo)
}

func (handler *Handler) deleteTodo(echoContext echo.Context) (err error) {
	request := echoContext.Request()
	id := echoContext.Param("todo_id")

	ctx, cancel := context.WithTimeout(request.Context(), 3*time.Second)
	defer cancel()

	if err := handler.service.Delete(ctx, id); err != nil {
		return writeMappedError(echoContext, err)
	}

	return echoContext.NoContent(http.StatusNoContent)
}

func parseListFilter(request *http.Request) (filter ListFilter, err error) {
	query := request.URL.Query()
	for key := range query {
		switch key {
		case "status", "limit", "cursor", "sort":
		default:
			return filter, fmt.Errorf("%w: unknown query parameter", ErrInvalidInput)
		}
	}

	if value := query.Get("status"); value != "" {
		status := Status(value)
		filter.Status = &status
	}
	if value := query.Get("limit"); value != "" {
		limit, err := strconv.Atoi(value)
		if err != nil {
			return filter, fmt.Errorf("%w: limit must be an integer", ErrInvalidInput)
		}
		filter.Limit = limit
	}
	if value := query.Get("cursor"); value != "" {
		cursor, err := DecodeCursor(value)
		if err != nil {
			return filter, err
		}
		filter.Cursor = &cursor
	}
	if value := query.Get("sort"); value != "" {
		filter.Sort = Sort(value)
	}

	return filter, nil
}

func decodeJSON(request *http.Request, value any) (err error) {
	decoder := json.NewDecoder(request.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(value); err != nil {
		return fmt.Errorf("%w: %v", ErrInvalidJSON, err)
	}
	if err := decoder.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		return fmt.Errorf("%w: multiple json values", ErrInvalidJSON)
	}

	return nil
}

type errorResponse struct {
	Error responseError `json:"error"`
}

type responseError struct {
	Code      string `json:"code"`
	Message   string `json:"message"`
	RequestID string `json:"request_id"`
}

func writeMappedError(
	echoContext echo.Context,
	err error,
) (echoErr error) {
	switch {
	case errors.Is(err, ErrInvalidInput), errors.Is(err, ErrInvalidCursor),
		errors.Is(err, ErrInvalidJSON):
		return writeError(echoContext, http.StatusBadRequest, err)
	case errors.Is(err, ErrTodoNotFound):
		return writeError(echoContext, http.StatusNotFound, err)
	case errors.Is(err, ErrVersionConflict):
		return writeError(echoContext, http.StatusConflict, err)
	default:
		return writeError(echoContext, http.StatusInternalServerError, err)
	}
}

func writeError(
	echoContext echo.Context,
	status int,
	err error,
) (echoErr error) {
	request := echoContext.Request()
	code := errorCode(status, err)
	message := safeMessage(status, err)
	return echoContext.JSON(status, errorResponse{
		Error: responseError{
			Code:      code,
			Message:   message,
			RequestID: request.Header.Get("X-Request-ID"),
		},
	})
}

func errorCode(status int, err error) (code string) {
	switch {
	case errors.Is(err, ErrInvalidJSON):
		return "INVALID_JSON"
	case errors.Is(err, ErrInvalidCursor):
		return "INVALID_CURSOR"
	case errors.Is(err, ErrInvalidInput):
		return "VALIDATION_ERROR"
	case errors.Is(err, ErrTodoNotFound):
		return "TODO_NOT_FOUND"
	case errors.Is(err, ErrVersionConflict):
		return "VERSION_CONFLICT"
	case status == http.StatusNotFound:
		return "NOT_FOUND"
	default:
		return "INTERNAL_ERROR"
	}
}

func safeMessage(status int, err error) (message string) {
	if status >= 500 {
		return "internal server error"
	}
	if errors.Is(err, ErrTodoNotFound) {
		return ErrTodoNotFound.Error()
	}
	if errors.Is(err, ErrVersionConflict) {
		return ErrVersionConflict.Error()
	}

	return err.Error()
}
