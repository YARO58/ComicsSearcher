package core

import "context"

type Searcher interface {
	Search(ctx context.Context, query string, limit int) ([]Comics, int, error)
	ISearch(ctx context.Context, query string, limit int) ([]Comics, int, error)
}

type Words interface {
	Norm(ctx context.Context, phrase string) ([]string, error)
}

type DB interface {
	Get(ctx context.Context, id int) (Comics, error)
	IDs(context.Context) ([]int, error)
}

type UpdateIndex interface {
	UpdateIndex(ctx context.Context)
}
