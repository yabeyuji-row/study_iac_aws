package todo

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"
)

type fakeRepository struct {
	todos map[string]Todo
}

func newFakeRepository() (repository *fakeRepository) {
	return &fakeRepository{todos: map[string]Todo{}}
}

func (repository *fakeRepository) Create(
	ctx context.Context,
	todo Todo,
) (created Todo, err error) {
	_ = ctx
	repository.todos[todo.ID] = todo

	return todo, nil
}

func (repository *fakeRepository) List(
	ctx context.Context,
	filter ListFilter,
) (result ListResult, err error) {
	_ = ctx
	_ = filter
	for _, todo := range repository.todos {
		result.Todos = append(result.Todos, todo)
	}

	return result, nil
}

func (repository *fakeRepository) Get(
	ctx context.Context,
	id string,
) (todo Todo, err error) {
	_ = ctx
	todo, ok := repository.todos[id]
	if !ok {
		return todo, ErrTodoNotFound
	}

	return todo, nil
}

func (repository *fakeRepository) Update(
	ctx context.Context,
	input UpdateInput,
) (todo Todo, err error) {
	_ = ctx
	current, ok := repository.todos[input.ID]
	if !ok {
		return todo, ErrTodoNotFound
	}
	if current.Version != input.Version {
		return todo, ErrVersionConflict
	}

	current.Title = input.Title
	current.Description = input.Description
	current.Status = input.Status
	current.DueDate = input.DueDate
	current.Version++
	repository.todos[input.ID] = current

	return current, nil
}

func (repository *fakeRepository) Delete(ctx context.Context, id string) (err error) {
	_ = ctx
	if _, ok := repository.todos[id]; !ok {
		return ErrTodoNotFound
	}
	delete(repository.todos, id)

	return nil
}

func TestServiceCreateTrimsTitleAndDefaultsStatus(t *testing.T) {
	repository := newFakeRepository()
	service := NewService(repository, func() time.Time {
		return time.Date(2026, 8, 3, 1, 2, 3, 0, time.FixedZone("JST", 9*3600))
	})

	created, err := service.Create(context.Background(), CreateInput{
		Title:       "  Terraformを学ぶ  ",
		Description: "VPCを作る",
	})
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	if created.Title != "Terraformを学ぶ" {
		t.Fatalf("Title = %q", created.Title)
	}
	if created.Status != StatusPending {
		t.Fatalf("Status = %q", created.Status)
	}
	if created.Version != 1 {
		t.Fatalf("Version = %d", created.Version)
	}
	if created.CreatedAt.Location() != time.UTC {
		t.Fatalf("CreatedAt location = %v", created.CreatedAt.Location())
	}
}

func TestServiceCreateValidation(t *testing.T) {
	service := NewService(newFakeRepository(), time.Now)

	tests := []struct {
		name  string
		input CreateInput
	}{
		{
			name:  "missing title",
			input: CreateInput{Status: StatusPending},
		},
		{
			name: "title too long",
			input: CreateInput{
				Title:  strings.Repeat("a", 201),
				Status: StatusPending,
			},
		},
		{
			name: "description too long",
			input: CreateInput{
				Title:       "ok",
				Description: strings.Repeat("a", 2001),
				Status:      StatusPending,
			},
		},
		{
			name: "invalid status",
			input: CreateInput{
				Title:  "ok",
				Status: "blocked",
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := service.Create(context.Background(), test.input)
			if !errors.Is(err, ErrInvalidInput) {
				t.Fatalf("Create() error = %v, want ErrInvalidInput", err)
			}
		})
	}
}

func TestServiceUpdateVersionConflict(t *testing.T) {
	repository := newFakeRepository()
	service := NewService(repository, time.Now)

	created, err := service.Create(context.Background(), CreateInput{
		Title:  "first",
		Status: StatusPending,
	})
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	_, err = service.Update(context.Background(), UpdateInput{
		ID:          created.ID,
		Title:       "updated",
		Description: "",
		Status:      StatusInProgress,
		Version:     created.Version,
	})
	if err != nil {
		t.Fatalf("first Update() error = %v", err)
	}

	_, err = service.Update(context.Background(), UpdateInput{
		ID:          created.ID,
		Title:       "stale",
		Description: "",
		Status:      StatusCompleted,
		Version:     created.Version,
	})
	if !errors.Is(err, ErrVersionConflict) {
		t.Fatalf("second Update() error = %v, want ErrVersionConflict", err)
	}
}
