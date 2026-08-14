package todo

import (
	"crypto/rand"
	"fmt"
	"time"
)

type Status string

const (
	StatusPending    Status = "pending"
	StatusInProgress Status = "in_progress"
	StatusCompleted  Status = "completed"
)

type Todo struct {
	ID          string     `json:"id"`
	Title       string     `json:"title"`
	Description string     `json:"description"`
	Status      Status     `json:"status"`
	DueDate     *time.Time `json:"due_date,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
	Version     int        `json:"version"`
}

type CreateInput struct {
	Title       string     `json:"title"`
	Description string     `json:"description"`
	Status      Status     `json:"status"`
	DueDate     *time.Time `json:"due_date"`
}

type UpdateInput struct {
	ID          string
	Title       string     `json:"title"`
	Description string     `json:"description"`
	Status      Status     `json:"status"`
	DueDate     *time.Time `json:"due_date"`
	Version     int        `json:"version"`
}

type ListFilter struct {
	Status *Status
	Limit  int
	Cursor *Cursor
	Sort   Sort
}

type ListResult struct {
	Todos      []Todo  `json:"todos"`
	NextCursor *string `json:"next_cursor,omitempty"`
}

type Sort string

const (
	SortCreatedAtDesc Sort = "created_at_desc"
	SortCreatedAtAsc  Sort = "created_at_asc"
)

func NewUUID() (id string, err error) {
	var randomBytes [16]byte
	if _, err := rand.Read(randomBytes[:]); err != nil {
		return "", fmt.Errorf("read random uuid bytes: %w", err)
	}

	randomBytes[6] = (randomBytes[6] & 0x0f) | 0x40
	randomBytes[8] = (randomBytes[8] & 0x3f) | 0x80

	id = fmt.Sprintf(
		"%08x-%04x-%04x-%04x-%012x",
		randomBytes[0:4],
		randomBytes[4:6],
		randomBytes[6:8],
		randomBytes[8:10],
		randomBytes[10:16],
	)

	return id, nil
}

func ValidStatus(status Status) (ok bool) {
	switch status {
	case StatusPending, StatusInProgress, StatusCompleted:
		return true
	default:
		return false
	}
}
