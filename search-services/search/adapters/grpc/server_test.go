package grpc

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/types/known/emptypb"
	searchpb "yadro.com/course/proto/search"
	"yadro.com/course/search/core"
)

type mockSearcher struct {
	searchFunc  func(ctx context.Context, query string, limit int) ([]core.Comics, int, error)
	isearchFunc func(ctx context.Context, query string, limit int) ([]core.Comics, int, error)
}

func (m *mockSearcher) Search(ctx context.Context, query string, limit int) ([]core.Comics, int, error) {
	return m.searchFunc(ctx, query, limit)
}

func (m *mockSearcher) ISearch(ctx context.Context, query string, limit int) ([]core.Comics, int, error) {
	return m.isearchFunc(ctx, query, limit)
}

func TestServer_Ping(t *testing.T) {
	server := NewServer(&mockSearcher{})

	resp, err := server.Ping(context.Background(), &emptypb.Empty{})

	assert.NoError(t, err)
	assert.Nil(t, resp)
}

func TestServer_Search(t *testing.T) {
	tests := []struct {
		name           string
		query          string
		limit          int64
		searchFunc     func(ctx context.Context, query string, limit int) ([]core.Comics, int, error)
		expectedComics []*searchpb.Comics
		expectedTotal  int64
		expectedErr    error
	}{
		{
			name:  "successful search",
			query: "test query",
			limit: 2,
			searchFunc: func(ctx context.Context, query string, limit int) ([]core.Comics, int, error) {
				return []core.Comics{
					{
						ID:          1,
						URL:         "http://example.com/1",
						Description: "test description 1",
					},
					{
						ID:          2,
						URL:         "http://example.com/2",
						Description: "test description 2",
					},
				}, 2, nil
			},
			expectedComics: []*searchpb.Comics{
				{
					Id:  1,
					Url: "http://example.com/1",
				},
				{
					Id:  2,
					Url: "http://example.com/2",
				},
			},
			expectedTotal: 2,
			expectedErr:   nil,
		},
		{
			name:  "search error",
			query: "test query",
			limit: 2,
			searchFunc: func(ctx context.Context, query string, limit int) ([]core.Comics, int, error) {
				return nil, 0, errors.New("search error")
			},
			expectedComics: nil,
			expectedTotal:  0,
			expectedErr:    errors.New("search error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := NewServer(&mockSearcher{
				searchFunc: tt.searchFunc,
			})

			resp, err := server.Search(context.Background(), &searchpb.SearchRequest{
				Query: tt.query,
				Limit: tt.limit,
			})

			if tt.expectedErr != nil {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedErr.Error())
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedComics, resp.Items)
				assert.Equal(t, tt.expectedTotal, resp.Total)
			}
		})
	}
}

func TestServer_ISearch(t *testing.T) {
	tests := []struct {
		name           string
		query          string
		limit          int64
		isearchFunc    func(ctx context.Context, query string, limit int) ([]core.Comics, int, error)
		expectedComics []*searchpb.Comics
		expectedTotal  int64
		expectedErr    error
	}{
		{
			name:  "successful isearch",
			query: "test query",
			limit: 2,
			isearchFunc: func(ctx context.Context, query string, limit int) ([]core.Comics, int, error) {
				return []core.Comics{
					{
						ID:          1,
						URL:         "http://example.com/1",
						Description: "test description 1",
					},
					{
						ID:          2,
						URL:         "http://example.com/2",
						Description: "test description 2",
					},
				}, 2, nil
			},
			expectedComics: []*searchpb.Comics{
				{
					Id:  1,
					Url: "http://example.com/1",
				},
				{
					Id:  2,
					Url: "http://example.com/2",
				},
			},
			expectedTotal: 2,
			expectedErr:   nil,
		},
		{
			name:  "isearch error",
			query: "test query",
			limit: 2,
			isearchFunc: func(ctx context.Context, query string, limit int) ([]core.Comics, int, error) {
				return nil, 0, errors.New("isearch error")
			},
			expectedComics: nil,
			expectedTotal:  0,
			expectedErr:    errors.New("isearch error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := NewServer(&mockSearcher{
				isearchFunc: tt.isearchFunc,
			})

			resp, err := server.ISearch(context.Background(), &searchpb.ISearchRequest{
				Query: tt.query,
				Limit: tt.limit,
			})

			if tt.expectedErr != nil {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedErr.Error())
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedComics, resp.Items)
				assert.Equal(t, tt.expectedTotal, resp.Total)
			}
		})
	}
}
