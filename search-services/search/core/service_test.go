package core

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"testing"

	"github.com/stretchr/testify/assert"
)

type mockWords struct {
	normFunc func(ctx context.Context, phrase string) ([]string, error)
}

func (m *mockWords) Norm(ctx context.Context, phrase string) ([]string, error) {
	return m.normFunc(ctx, phrase)
}

type mockDB struct {
	getFunc func(ctx context.Context, id int) (Comics, error)
	idsFunc func(ctx context.Context) ([]int, error)
}

func (m *mockDB) Get(ctx context.Context, id int) (Comics, error) {
	return m.getFunc(ctx, id)
}

func (m *mockDB) IDs(ctx context.Context) ([]int, error) {
	return m.idsFunc(ctx)
}

func newTestLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}

func TestService_Search(t *testing.T) {
	tests := []struct {
		name           string
		query          string
		limit          int
		normFunc       func(ctx context.Context, phrase string) ([]string, error)
		getFunc        func(ctx context.Context, id int) (Comics, error)
		idsFunc        func(ctx context.Context) ([]int, error)
		expectedComics []Comics
		expectedCount  int
		expectedErr    error
	}{
		{
			name:  "successful search",
			query: "test query",
			limit: 2,
			normFunc: func(ctx context.Context, phrase string) ([]string, error) {
				return []string{"test", "query"}, nil
			},
			getFunc: func(ctx context.Context, id int) (Comics, error) {
				return Comics{
					ID:          id,
					URL:         "http://example.com",
					Description: "test query description",
				}, nil
			},
			idsFunc: func(ctx context.Context) ([]int, error) {
				return []int{2, 3}, nil
			},
			expectedComics: []Comics{
				{
					ID:          2,
					URL:         "http://example.com",
					Description: "test query description",
				},
				{
					ID:          3,
					URL:         "http://example.com",
					Description: "test query description",
				},
			},
			expectedCount: 2,
			expectedErr:   nil,
		},
		{
			name:  "invalid limit",
			query: "test query",
			limit: 0,
			normFunc: func(ctx context.Context, phrase string) ([]string, error) {
				return []string{"test", "query"}, nil
			},
			getFunc: func(ctx context.Context, id int) (Comics, error) {
				return Comics{}, nil
			},
			idsFunc: func(ctx context.Context) ([]int, error) {
				return []int{}, nil
			},
			expectedComics: nil,
			expectedCount:  0,
			expectedErr:    ErrBadArguments,
		},
		{
			name:  "normalization error",
			query: "test query",
			limit: 2,
			normFunc: func(ctx context.Context, phrase string) ([]string, error) {
				return nil, errors.New("normalization error")
			},
			getFunc: func(ctx context.Context, id int) (Comics, error) {
				return Comics{}, nil
			},
			idsFunc: func(ctx context.Context) ([]int, error) {
				return []int{}, nil
			},
			expectedComics: nil,
			expectedCount:  0,
			expectedErr:    errors.New("failed to normalize words"),
		},
		{
			name:  "get ids error",
			query: "test query",
			limit: 2,
			normFunc: func(ctx context.Context, phrase string) ([]string, error) {
				return []string{"test", "query"}, nil
			},
			getFunc: func(ctx context.Context, id int) (Comics, error) {
				return Comics{}, nil
			},
			idsFunc: func(ctx context.Context) ([]int, error) {
				return nil, errors.New("get ids error")
			},
			expectedComics: nil,
			expectedCount:  0,
			expectedErr:    errors.New("failed to get ids"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := &mockDB{
				getFunc: tt.getFunc,
				idsFunc: tt.idsFunc,
			}
			index := NewIndex(newTestLogger(), db)
			service := NewService(
				newTestLogger(),
				db,
				&mockWords{
					normFunc: tt.normFunc,
				},
				index,
			)

			comics, count, err := service.Search(context.Background(), tt.query, tt.limit)

			if tt.expectedErr != nil {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedErr.Error())
			} else {
				assert.NoError(t, err)
				assert.ElementsMatch(t, tt.expectedComics, comics)
				assert.Equal(t, tt.expectedCount, count)
			}
		})
	}
}

func TestService_ISearch(t *testing.T) {
	tests := []struct {
		name           string
		query          string
		limit          int
		normFunc       func(ctx context.Context, phrase string) ([]string, error)
		getFunc        func(ctx context.Context, id int) (Comics, error)
		idsFunc        func(ctx context.Context) ([]int, error)
		expectedComics []Comics
		expectedCount  int
		expectedErr    error
	}{
		{
			name:  "successful isearch",
			query: "test query",
			limit: 2,
			normFunc: func(ctx context.Context, phrase string) ([]string, error) {
				return []string{"test", "query"}, nil
			},
			getFunc: func(ctx context.Context, id int) (Comics, error) {
				return Comics{
					ID:          id,
					URL:         "http://example.com",
					Description: "test query description",
				}, nil
			},
			idsFunc: func(ctx context.Context) ([]int, error) {
				return []int{2, 3}, nil
			},
			expectedComics: []Comics{
				{
					ID:          2,
					URL:         "http://example.com",
					Description: "test query description",
				},
				{
					ID:          3,
					URL:         "http://example.com",
					Description: "test query description",
				},
			},
			expectedCount: 2,
			expectedErr:   nil,
		},
		{
			name:  "invalid limit",
			query: "test query",
			limit: 0,
			normFunc: func(ctx context.Context, phrase string) ([]string, error) {
				return []string{"test", "query"}, nil
			},
			getFunc: func(ctx context.Context, id int) (Comics, error) {
				return Comics{}, nil
			},
			idsFunc: func(ctx context.Context) ([]int, error) {
				return []int{}, nil
			},
			expectedComics: nil,
			expectedCount:  0,
			expectedErr:    ErrBadArguments,
		},
		{
			name:  "normalization error",
			query: "test query",
			limit: 2,
			normFunc: func(ctx context.Context, phrase string) ([]string, error) {
				return nil, errors.New("normalization error")
			},
			getFunc: func(ctx context.Context, id int) (Comics, error) {
				return Comics{}, nil
			},
			idsFunc: func(ctx context.Context) ([]int, error) {
				return []int{}, nil
			},
			expectedComics: nil,
			expectedCount:  0,
			expectedErr:    errors.New("failed to normalize words"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := &mockDB{
				getFunc: tt.getFunc,
				idsFunc: tt.idsFunc,
			}
			index := NewIndex(newTestLogger(), db)
			service := NewService(
				newTestLogger(),
				db,
				&mockWords{
					normFunc: tt.normFunc,
				},
				index,
			)

			index.UpdateIndex(context.Background())

			comics, count, err := service.ISearch(context.Background(), tt.query, tt.limit)

			if tt.expectedErr != nil {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedErr.Error())
			} else {
				assert.NoError(t, err)
				assert.ElementsMatch(t, tt.expectedComics, comics)
				assert.Equal(t, tt.expectedCount, count)
			}
		})
	}
}
