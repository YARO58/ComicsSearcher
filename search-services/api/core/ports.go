package core

import "context"

type Normalizer interface {
	Norm(context.Context, string) ([]string, error)
}

type Pinger interface {
	Ping(context.Context) error
}

type Searcher interface {
	Search(context.Context, string, int) ([]Comics, int, error)
	ISearch(context.Context, string, int) ([]Comics, int, error)
}

type Updater interface {
	Update(context.Context) error
	Stats(context.Context) (UpdateStats, error)
	Status(context.Context) (UpdateStatus, error)
	Drop(context.Context) error
}

type Authenticator interface {
	GenerateToken(username, password string) (string, error)
	ValidateToken(token string) error
}

type RateLimiter interface {
	Allow() bool
	Wait()
}

type ConcurrencyLimiter interface {
	Acquire() bool
	Release()
}
