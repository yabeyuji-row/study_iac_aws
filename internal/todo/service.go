package todo

import (
	"context"
	"fmt"
	"strings"
	"time"
)

type Service struct {
	repository Repository
	now        func() time.Time
}

func NewService(repository Repository, now func() time.Time) (service *Service) {
	return &Service{
		repository: repository,
		now:        now,
	}
}

func (service *Service) Create(
	ctx context.Context,
	input CreateInput,
) (created Todo, err error) {
	input, err = validateCreateInput(input)
	if err != nil {
		return created, err
	}

	id, err := NewUUID()
	if err != nil {
		return created, fmt.Errorf("create todo id: %w", err)
	}

	now := service.now().UTC()
	todo := Todo{
		ID:          id,
		Title:       input.Title,
		Description: input.Description,
		Status:      input.Status,
		DueDate:     normalizeOptionalTime(input.DueDate),
		CreatedAt:   now,
		UpdatedAt:   now,
		Version:     1,
	}

	created, err = service.repository.Create(ctx, todo)
	if err != nil {
		return created, fmt.Errorf("create todo: %w", err)
	}

	return created, nil
}

func (service *Service) List(
	ctx context.Context,
	filter ListFilter,
) (result ListResult, err error) {
	filter, err = validateListFilter(filter)
	if err != nil {
		return result, err
	}

	result, err = service.repository.List(ctx, filter)
	if err != nil {
		return result, fmt.Errorf("list todos: %w", err)
	}

	return result, nil
}

func (service *Service) Get(ctx context.Context, id string) (todo Todo, err error) {
	if strings.TrimSpace(id) == "" {
		return todo, fmt.Errorf("%w: id is required", ErrInvalidInput)
	}

	todo, err = service.repository.Get(ctx, id)
	if err != nil {
		return todo, fmt.Errorf("get todo: %w", err)
	}

	return todo, nil
}

func (service *Service) Update(
	ctx context.Context,
	input UpdateInput,
) (updated Todo, err error) {
	input, err = validateUpdateInput(input)
	if err != nil {
		return updated, err
	}

	updated, err = service.repository.Update(ctx, input)
	if err != nil {
		return updated, fmt.Errorf("update todo: %w", err)
	}

	return updated, nil
}

func (service *Service) Delete(ctx context.Context, id string) (err error) {
	if strings.TrimSpace(id) == "" {
		return fmt.Errorf("%w: id is required", ErrInvalidInput)
	}

	if err := service.repository.Delete(ctx, id); err != nil {
		return fmt.Errorf("delete todo: %w", err)
	}

	return nil
}

func validateCreateInput(input CreateInput) (validated CreateInput, err error) {
	input.Title = strings.TrimSpace(input.Title)
	if input.Status == "" {
		input.Status = StatusPending
	}

	if err := validateTodoFields(input.Title, input.Description, input.Status); err != nil {
		return validated, err
	}

	input.DueDate = normalizeOptionalTime(input.DueDate)

	return input, nil
}

func validateUpdateInput(input UpdateInput) (validated UpdateInput, err error) {
	input.ID = strings.TrimSpace(input.ID)
	input.Title = strings.TrimSpace(input.Title)
	if input.ID == "" {
		return validated, fmt.Errorf("%w: id is required", ErrInvalidInput)
	}
	if input.Version < 1 {
		return validated, fmt.Errorf("%w: version is required", ErrInvalidInput)
	}
	if err := validateTodoFields(input.Title, input.Description, input.Status); err != nil {
		return validated, err
	}

	input.DueDate = normalizeOptionalTime(input.DueDate)

	return input, nil
}

func validateTodoFields(title string, description string, status Status) (err error) {
	if title == "" {
		return fmt.Errorf("%w: title is required", ErrInvalidInput)
	}
	if len(title) > 200 {
		return fmt.Errorf("%w: title must be 200 characters or fewer", ErrInvalidInput)
	}
	if len(description) > 2000 {
		return fmt.Errorf(
			"%w: description must be 2000 characters or fewer",
			ErrInvalidInput,
		)
	}
	if !ValidStatus(status) {
		return fmt.Errorf("%w: status is invalid", ErrInvalidInput)
	}

	return nil
}

func validateListFilter(filter ListFilter) (validated ListFilter, err error) {
	if filter.Limit == 0 {
		filter.Limit = 20
	}
	if filter.Limit < 1 || filter.Limit > 100 {
		return validated, fmt.Errorf("%w: limit is invalid", ErrInvalidInput)
	}
	if filter.Status != nil && !ValidStatus(*filter.Status) {
		return validated, fmt.Errorf("%w: status is invalid", ErrInvalidInput)
	}
	if filter.Sort == "" {
		filter.Sort = SortCreatedAtDesc
	}
	if filter.Sort != SortCreatedAtDesc && filter.Sort != SortCreatedAtAsc {
		return validated, fmt.Errorf("%w: sort is invalid", ErrInvalidInput)
	}

	return filter, nil
}

func normalizeOptionalTime(value *time.Time) (normalized *time.Time) {
	if value == nil {
		return nil
	}
	utcValue := value.UTC()

	return &utcValue
}
