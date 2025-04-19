package words

import (
	"context"
	"errors"
	"io"
	"testing"

	"log/slog"

	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/emptypb"
	wordspb "yadro.com/course/proto/words"
)

// Моки для protobuf типов
type WordsRequest struct {
	Phrase string
}

type WordsResponse struct {
	Words []string
}

// Empty - пустое сообщение для ping
type Empty struct{}

type mockWordsClient struct {
	wordspb.WordsClient
	normFunc func(ctx context.Context, in *wordspb.WordsRequest, opts ...grpc.CallOption) (*wordspb.WordsReply, error)
	pingFunc func(ctx context.Context, in *emptypb.Empty, opts ...grpc.CallOption) (*emptypb.Empty, error)
}

func (m *mockWordsClient) Norm(ctx context.Context, in *wordspb.WordsRequest, opts ...grpc.CallOption) (*wordspb.WordsReply, error) {
	if m.normFunc != nil {
		return m.normFunc(ctx, in, opts...)
	}
	return nil, nil
}

func (m *mockWordsClient) Ping(ctx context.Context, in *emptypb.Empty, opts ...grpc.CallOption) (*emptypb.Empty, error) {
	if m.pingFunc != nil {
		return m.pingFunc(ctx, in, opts...)
	}
	return nil, nil
}

func TestClient_Norm(t *testing.T) {
	tests := []struct {
		name      string
		phrase    string
		words     []string
		normError error
		wantError bool
	}{
		{
			name:      "successful normalization",
			phrase:    "test phrase",
			words:     []string{"test", "phrase"},
			wantError: false,
		},
		{
			name:      "empty phrase",
			phrase:    "",
			words:     []string{},
			wantError: false,
		},
		{
			name:      "normalization error",
			phrase:    "test phrase",
			normError: errors.New("normalization error"),
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := slog.New(slog.NewTextHandler(io.Discard, &slog.HandlerOptions{Level: slog.LevelDebug}))
			mockClient := &mockWordsClient{
				normFunc: func(ctx context.Context, in *wordspb.WordsRequest, opts ...grpc.CallOption) (*wordspb.WordsReply, error) {
					if tt.normError != nil {
						return nil, tt.normError
					}
					return &wordspb.WordsReply{Words: tt.words}, nil
				},
			}

			client := &Client{
				log:    logger,
				client: mockClient,
			}

			words, err := client.Norm(context.Background(), tt.phrase)
			if tt.wantError {
				assert.Error(t, err)
				assert.Nil(t, words)
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, tt.words, words)
		})
	}
}

func TestClient_Ping(t *testing.T) {
	tests := []struct {
		name      string
		pingError error
		wantError bool
	}{
		{
			name:      "successful ping",
			wantError: false,
		},
		{
			name:      "ping error",
			pingError: errors.New("ping error"),
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := slog.New(slog.NewTextHandler(io.Discard, &slog.HandlerOptions{Level: slog.LevelDebug}))
			mockClient := &mockWordsClient{
				pingFunc: func(ctx context.Context, in *emptypb.Empty, opts ...grpc.CallOption) (*emptypb.Empty, error) {
					if tt.pingError != nil {
						return nil, tt.pingError
					}
					return &emptypb.Empty{}, nil
				},
			}

			client := &Client{
				log:    logger,
				client: mockClient,
			}

			err := client.Ping(context.Background())
			if tt.wantError {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
		})
	}
}
