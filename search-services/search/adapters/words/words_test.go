package words

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	wordspb "yadro.com/course/proto/words"
)

type mockWordsClient struct {
	wordspb.WordsClient
	normFunc func(ctx context.Context, req *wordspb.WordsRequest, opts ...grpc.CallOption) (*wordspb.WordsReply, error)
}

func (m *mockWordsClient) Norm(ctx context.Context, req *wordspb.WordsRequest, opts ...grpc.CallOption) (*wordspb.WordsReply, error) {
	return m.normFunc(ctx, req, opts...)
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

func TestClient_Norm(t *testing.T) {
	tests := []struct {
		name        string
		phrase      string
		normFunc    func(ctx context.Context, req *wordspb.WordsRequest, opts ...grpc.CallOption) (*wordspb.WordsReply, error)
		expected    []string
		expectedErr error
	}{
		{
			name:   "successful normalization",
			phrase: "test phrase",
			normFunc: func(ctx context.Context, req *wordspb.WordsRequest, opts ...grpc.CallOption) (*wordspb.WordsReply, error) {
				return &wordspb.WordsReply{
					Words: []string{"test", "phrase"},
				}, nil
			},
			expected:    []string{"test", "phrase"},
			expectedErr: nil,
		},
		{
			name:   "empty response",
			phrase: "test phrase",
			normFunc: func(ctx context.Context, req *wordspb.WordsRequest, opts ...grpc.CallOption) (*wordspb.WordsReply, error) {
				return &wordspb.WordsReply{
					Words: []string{},
				}, nil
			},
			expected:    []string{},
			expectedErr: nil,
		},
		{
			name:   "grpc error",
			phrase: "test phrase",
			normFunc: func(ctx context.Context, req *wordspb.WordsRequest, opts ...grpc.CallOption) (*wordspb.WordsReply, error) {
				return nil, errors.New("grpc error")
			},
			expected:    nil,
			expectedErr: errors.New("failed to normalize words"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := &Client{
				log: newTestLogger(),
				client: &mockWordsClient{
					normFunc: tt.normFunc,
				},
			}

			words, err := client.Norm(context.Background(), tt.phrase)

			if tt.expectedErr != nil {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedErr.Error())
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, words)
			}
		})
	}
}
