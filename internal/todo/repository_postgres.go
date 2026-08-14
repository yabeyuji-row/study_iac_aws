package todo

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresRepository struct {
	pool *pgxpool.Pool
}

func NewPostgresRepository(pool *pgxpool.Pool) (repository *PostgresRepository) {
	return &PostgresRepository{pool: pool}
}

func (repository *PostgresRepository) Create(
	ctx context.Context,
	todo Todo,
) (created Todo, err error) {
	query := `
insert into todos (
    id,
    title,
    description,
    status,
    due_date,
    created_at,
    updated_at,
    version
) values (
    $1, $2, $3, $4, $5, $6, $7, $8
)
returning id, title, description, status, due_date, created_at, updated_at,
          version
`

	err = repository.pool.QueryRow(
		ctx,
		query,
		todo.ID,
		todo.Title,
		todo.Description,
		todo.Status,
		todo.DueDate,
		todo.CreatedAt,
		todo.UpdatedAt,
		todo.Version,
	).Scan(
		&created.ID,
		&created.Title,
		&created.Description,
		&created.Status,
		&created.DueDate,
		&created.CreatedAt,
		&created.UpdatedAt,
		&created.Version,
	)
	if err != nil {
		return created, fmt.Errorf("insert todo: %w", err)
	}

	created = normalizeTodo(created)

	return created, nil
}

func (repository *PostgresRepository) List(
	ctx context.Context,
	filter ListFilter,
) (result ListResult, err error) {
	args := []any{}
	clauses := []string{}

	if filter.Status != nil {
		args = append(args, *filter.Status)
		clauses = append(clauses, fmt.Sprintf("status = $%d", len(args)))
	}

	if filter.Cursor != nil {
		args = append(args, filter.Cursor.CreatedAt, filter.Cursor.ID)
		createdAtIndex := len(args) - 1
		idIndex := len(args)
		operator := "<"
		if filter.Sort == SortCreatedAtAsc {
			operator = ">"
		}
		clauses = append(
			clauses,
			fmt.Sprintf(
				"(created_at, id) %s ($%d, $%d)",
				operator,
				createdAtIndex,
				idIndex,
			),
		)
	}

	limit := filter.Limit
	args = append(args, limit+1)
	limitIndex := len(args)

	orderDirection := "desc"
	if filter.Sort == SortCreatedAtAsc {
		orderDirection = "asc"
	}

	whereSQL := ""
	if len(clauses) > 0 {
		whereSQL = "where " + strings.Join(clauses, " and ")
	}

	query := fmt.Sprintf(`
select id, title, description, status, due_date, created_at, updated_at, version
from todos
%s
order by created_at %s, id %s
limit $%d
`, whereSQL, orderDirection, orderDirection, limitIndex)

	rows, err := repository.pool.Query(ctx, query, args...)
	if err != nil {
		return result, fmt.Errorf("select todos: %w", err)
	}
	defer rows.Close()

	todos := make([]Todo, 0, limit)
	for rows.Next() {
		var todo Todo
		if err := rows.Scan(
			&todo.ID,
			&todo.Title,
			&todo.Description,
			&todo.Status,
			&todo.DueDate,
			&todo.CreatedAt,
			&todo.UpdatedAt,
			&todo.Version,
		); err != nil {
			return result, fmt.Errorf("scan todo: %w", err)
		}
		todos = append(todos, normalizeTodo(todo))
	}
	if err := rows.Err(); err != nil {
		return result, fmt.Errorf("iterate todos: %w", err)
	}

	if len(todos) > limit {
		lastTodo := todos[limit-1]
		nextCursor, err := EncodeCursor(Cursor{
			CreatedAt: lastTodo.CreatedAt,
			ID:        lastTodo.ID,
		})
		if err != nil {
			return result, fmt.Errorf("encode next cursor: %w", err)
		}
		result.NextCursor = &nextCursor
		todos = todos[:limit]
	}

	result.Todos = todos

	return result, nil
}

func (repository *PostgresRepository) Get(
	ctx context.Context,
	id string,
) (todo Todo, err error) {
	query := `
select id, title, description, status, due_date, created_at, updated_at, version
from todos
where id = $1
`

	err = repository.pool.QueryRow(ctx, query, id).Scan(
		&todo.ID,
		&todo.Title,
		&todo.Description,
		&todo.Status,
		&todo.DueDate,
		&todo.CreatedAt,
		&todo.UpdatedAt,
		&todo.Version,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return todo, ErrTodoNotFound
	}
	if err != nil {
		return todo, fmt.Errorf("select todo: %w", err)
	}

	todo = normalizeTodo(todo)

	return todo, nil
}

func (repository *PostgresRepository) Update(
	ctx context.Context,
	input UpdateInput,
) (todo Todo, err error) {
	now := time.Now().UTC()
	query := `
update todos
set title = $2,
    description = $3,
    status = $4,
    due_date = $5,
    updated_at = $6,
    version = version + 1
where id = $1 and version = $7
returning id, title, description, status, due_date, created_at, updated_at,
          version
`

	err = repository.pool.QueryRow(
		ctx,
		query,
		input.ID,
		input.Title,
		input.Description,
		input.Status,
		input.DueDate,
		now,
		input.Version,
	).Scan(
		&todo.ID,
		&todo.Title,
		&todo.Description,
		&todo.Status,
		&todo.DueDate,
		&todo.CreatedAt,
		&todo.UpdatedAt,
		&todo.Version,
	)
	if err == nil {
		return normalizeTodo(todo), nil
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return todo, fmt.Errorf("update todo: %w", err)
	}

	exists, err := repository.exists(ctx, input.ID)
	if err != nil {
		return todo, fmt.Errorf("check todo exists after update miss: %w", err)
	}
	if !exists {
		return todo, ErrTodoNotFound
	}

	return todo, ErrVersionConflict
}

func (repository *PostgresRepository) Delete(ctx context.Context, id string) (err error) {
	commandTag, err := repository.pool.Exec(ctx, "delete from todos where id = $1", id)
	if err != nil {
		return fmt.Errorf("delete todo: %w", err)
	}
	if commandTag.RowsAffected() == 0 {
		return ErrTodoNotFound
	}

	return nil
}

func (repository *PostgresRepository) exists(
	ctx context.Context,
	id string,
) (exists bool, err error) {
	err = repository.pool.QueryRow(
		ctx,
		"select exists(select 1 from todos where id = $1)",
		id,
	).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("select todo exists: %w", err)
	}

	return exists, nil
}

func normalizeTodo(todo Todo) (normalized Todo) {
	todo.CreatedAt = todo.CreatedAt.UTC()
	todo.UpdatedAt = todo.UpdatedAt.UTC()
	todo.DueDate = normalizeOptionalTime(todo.DueDate)

	return todo
}
