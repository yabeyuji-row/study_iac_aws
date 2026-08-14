package todo

import "errors"

var (
	ErrInvalidInput     = errors.New("invalid input")
	ErrInvalidCursor    = errors.New("invalid cursor")
	ErrTodoNotFound     = errors.New("todo was not found")
	ErrVersionConflict  = errors.New("todo version conflict")
	ErrInvalidJSON      = errors.New("invalid json")
	ErrUnsupportedRoute = errors.New("unsupported route")
)
