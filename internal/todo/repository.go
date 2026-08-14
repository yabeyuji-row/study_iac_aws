package todo

import "context"

type Repository interface {
	Create(ctx context.Context, todo Todo) (created Todo, err error)
	List(ctx context.Context, filter ListFilter) (result ListResult, err error)
	Get(ctx context.Context, id string) (todo Todo, err error)
	Update(ctx context.Context, input UpdateInput) (todo Todo, err error)
	Delete(ctx context.Context, id string) (err error)
}
