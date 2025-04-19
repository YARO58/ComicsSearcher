package search

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/emptypb"

	"yadro.com/course/api/core"
	searchpb "yadro.com/course/proto/search"
)

type mockSearchClient struct {
	searchpb.SearchClient
	pingFunc    func(ctx context.Context, req *emptypb.Empty, opts ...grpc.CallOption) (*emptypb.Empty, error)
	searchFunc  func(ctx context.Context, req *searchpb.SearchRequest, opts ...grpc.CallOption) (*searchpb.ComicsResponse, error)
	isearchFunc func(ctx context.Context, req *searchpb.ISearchRequest, opts ...grpc.CallOption) (*searchpb.ComicsResponse, error)
}

func (m *mockSearchClient) Ping(ctx context.Context, req *emptypb.Empty, opts ...grpc.CallOption) (*emptypb.Empty, error) {
	return m.pingFunc(ctx, req, opts...)
}

func (m *mockSearchClient) Search(ctx context.Context, req *searchpb.SearchRequest, opts ...grpc.CallOption) (*searchpb.ComicsResponse, error) {
	return m.searchFunc(ctx, req, opts...)
}

func (m *mockSearchClient) ISearch(ctx context.Context, req *searchpb.ISearchRequest, opts ...grpc.CallOption) (*searchpb.ComicsResponse, error) {
	return m.isearchFunc(ctx, req, opts...)
}

func newTestLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}

func TestNewClient(t *testing.T) {
	tests := []struct {
		name      string
		address   string
		expected  *Client
		expectErr bool
	}{
		{
			name:     "valid address",
			address:  "localhost:50051",
			expected: &Client{log: newTestLogger()},
		},
		{
			name:      "empty address",
			address:   "",
			expected:  &Client{log: newTestLogger()},
			expectErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, err := NewClient(tt.address, newTestLogger())
			if tt.expectErr {
				assert.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.NotNil(t, client)
			assert.NotNil(t, client.client)
		})
	}
}

func TestClient_Ping(t *testing.T) {
	tests := []struct {
		name        string
		pingFunc    func(ctx context.Context, req *emptypb.Empty, opts ...grpc.CallOption) (*emptypb.Empty, error)
		expectedErr error
	}{
		{
			name: "successful ping",
			pingFunc: func(ctx context.Context, req *emptypb.Empty, opts ...grpc.CallOption) (*emptypb.Empty, error) {
				return &emptypb.Empty{}, nil
			},
			expectedErr: nil,
		},
		{
			name: "grpc error",
			pingFunc: func(ctx context.Context, req *emptypb.Empty, opts ...grpc.CallOption) (*emptypb.Empty, error) {
				return nil, errors.New("grpc error")
			},
			expectedErr: errors.New("grpc error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := &Client{
				log: newTestLogger(),
				client: &mockSearchClient{
					pingFunc: tt.pingFunc,
				},
			}

			err := client.Ping(context.Background())

			if tt.expectedErr != nil {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedErr.Error())
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestClient_Search(t *testing.T) {
	tests := []struct {
		name          string
		searchFunc    func(ctx context.Context, req *searchpb.SearchRequest, opts ...grpc.CallOption) (*searchpb.ComicsResponse, error)
		phrase        string
		limit         int
		expected      []core.Comics
		expectedTotal int
		expectedErr   error
	}{
		{
			name: "successful search",
			searchFunc: func(ctx context.Context, req *searchpb.SearchRequest, opts ...grpc.CallOption) (*searchpb.ComicsResponse, error) {
				return &searchpb.ComicsResponse{
					Items: []*searchpb.Comics{
						{
							Id:  1,
							Url: "http://example.com/1",
						},
						{
							Id:  2,
							Url: "http://example.com/2",
						},
					},
					Total: 2,
				}, nil
			},
			phrase: "test phrase",
			limit:  10,
			expected: []core.Comics{
				{
					ID:  1,
					URL: "http://example.com/1",
				},
				{
					ID:  2,
					URL: "http://example.com/2",
				},
			},
			expectedTotal: 2,
			expectedErr:   nil,
		},
		{
			name: "empty results",
			searchFunc: func(ctx context.Context, req *searchpb.SearchRequest, opts ...grpc.CallOption) (*searchpb.ComicsResponse, error) {
				return &searchpb.ComicsResponse{
					Items: []*searchpb.Comics{},
					Total: 0,
				}, nil
			},
			phrase:        "test phrase",
			limit:         10,
			expected:      []core.Comics{},
			expectedTotal: 0,
			expectedErr:   nil,
		},
		{
			name: "grpc error",
			searchFunc: func(ctx context.Context, req *searchpb.SearchRequest, opts ...grpc.CallOption) (*searchpb.ComicsResponse, error) {
				return nil, errors.New("grpc error")
			},
			phrase:        "test phrase",
			limit:         10,
			expected:      nil,
			expectedTotal: 0,
			expectedErr:   errors.New("grpc error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := &Client{
				log: newTestLogger(),
				client: &mockSearchClient{
					searchFunc: tt.searchFunc,
				},
			}

			comics, total, err := client.Search(context.Background(), tt.phrase, tt.limit)

			if tt.expectedErr != nil {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedErr.Error())
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, comics)
				assert.Equal(t, tt.expectedTotal, total)
			}
		})
	}
}

func TestClient_ISearch(t *testing.T) {
	tests := []struct {
		name          string
		isearchFunc   func(ctx context.Context, req *searchpb.ISearchRequest, opts ...grpc.CallOption) (*searchpb.ComicsResponse, error)
		phrase        string
		limit         int
		expected      []core.Comics
		expectedTotal int
		expectedErr   error
	}{
		{
			name: "successful isearch",
			isearchFunc: func(ctx context.Context, req *searchpb.ISearchRequest, opts ...grpc.CallOption) (*searchpb.ComicsResponse, error) {
				return &searchpb.ComicsResponse{
					Items: []*searchpb.Comics{
						{
							Id:  1,
							Url: "http://example.com/1",
						},
						{
							Id:  2,
							Url: "http://example.com/2",
						},
					},
					Total: 2,
				}, nil
			},
			phrase: "test phrase",
			limit:  10,
			expected: []core.Comics{
				{
					ID:  1,
					URL: "http://example.com/1",
				},
				{
					ID:  2,
					URL: "http://example.com/2",
				},
			},
			expectedTotal: 2,
			expectedErr:   nil,
		},
		{
			name: "empty results",
			isearchFunc: func(ctx context.Context, req *searchpb.ISearchRequest, opts ...grpc.CallOption) (*searchpb.ComicsResponse, error) {
				return &searchpb.ComicsResponse{
					Items: []*searchpb.Comics{},
					Total: 0,
				}, nil
			},
			phrase:        "test phrase",
			limit:         10,
			expected:      []core.Comics{},
			expectedTotal: 0,
			expectedErr:   nil,
		},
		{
			name: "grpc error",
			isearchFunc: func(ctx context.Context, req *searchpb.ISearchRequest, opts ...grpc.CallOption) (*searchpb.ComicsResponse, error) {
				return nil, errors.New("grpc error")
			},
			phrase:        "test phrase",
			limit:         10,
			expected:      nil,
			expectedTotal: 0,
			expectedErr:   errors.New("grpc error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := &Client{
				log: newTestLogger(),
				client: &mockSearchClient{
					isearchFunc: tt.isearchFunc,
				},
			}

			comics, total, err := client.ISearch(context.Background(), tt.phrase, tt.limit)

			if tt.expectedErr != nil {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedErr.Error())
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, comics)
				assert.Equal(t, tt.expectedTotal, total)
			}
		})
	}
}
