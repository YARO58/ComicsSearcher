package update

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
	updatepb "yadro.com/course/proto/update"
)

type mockUpdateClient struct {
	updatepb.UpdateClient
	pingFunc   func(ctx context.Context, req *emptypb.Empty, opts ...grpc.CallOption) (*emptypb.Empty, error)
	statusFunc func(ctx context.Context, req *emptypb.Empty, opts ...grpc.CallOption) (*updatepb.StatusReply, error)
	statsFunc  func(ctx context.Context, req *emptypb.Empty, opts ...grpc.CallOption) (*updatepb.StatsReply, error)
	updateFunc func(ctx context.Context, req *emptypb.Empty, opts ...grpc.CallOption) (*emptypb.Empty, error)
	dropFunc   func(ctx context.Context, req *emptypb.Empty, opts ...grpc.CallOption) (*emptypb.Empty, error)
}

func (m *mockUpdateClient) Ping(ctx context.Context, req *emptypb.Empty, opts ...grpc.CallOption) (*emptypb.Empty, error) {
	return m.pingFunc(ctx, req, opts...)
}

func (m *mockUpdateClient) Status(ctx context.Context, req *emptypb.Empty, opts ...grpc.CallOption) (*updatepb.StatusReply, error) {
	return m.statusFunc(ctx, req, opts...)
}

func (m *mockUpdateClient) Stats(ctx context.Context, req *emptypb.Empty, opts ...grpc.CallOption) (*updatepb.StatsReply, error) {
	return m.statsFunc(ctx, req, opts...)
}

func (m *mockUpdateClient) Update(ctx context.Context, req *emptypb.Empty, opts ...grpc.CallOption) (*emptypb.Empty, error) {
	return m.updateFunc(ctx, req, opts...)
}

func (m *mockUpdateClient) Drop(ctx context.Context, req *emptypb.Empty, opts ...grpc.CallOption) (*emptypb.Empty, error) {
	return m.dropFunc(ctx, req, opts...)
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
				client: &mockUpdateClient{
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

func TestClient_Status(t *testing.T) {
	tests := []struct {
		name        string
		statusFunc  func(ctx context.Context, req *emptypb.Empty, opts ...grpc.CallOption) (*updatepb.StatusReply, error)
		expected    core.UpdateStatus
		expectedErr error
	}{
		{
			name: "running status",
			statusFunc: func(ctx context.Context, req *emptypb.Empty, opts ...grpc.CallOption) (*updatepb.StatusReply, error) {
				return &updatepb.StatusReply{
					Status: updatepb.Status_STATUS_RUNNING,
				}, nil
			},
			expected:    core.StatusUpdateRunning,
			expectedErr: nil,
		},
		{
			name: "idle status",
			statusFunc: func(ctx context.Context, req *emptypb.Empty, opts ...grpc.CallOption) (*updatepb.StatusReply, error) {
				return &updatepb.StatusReply{
					Status: updatepb.Status_STATUS_IDLE,
				}, nil
			},
			expected:    core.StatusUpdateIdle,
			expectedErr: nil,
		},
		{
			name: "unknown status",
			statusFunc: func(ctx context.Context, req *emptypb.Empty, opts ...grpc.CallOption) (*updatepb.StatusReply, error) {
				return &updatepb.StatusReply{
					Status: updatepb.Status_STATUS_UNSPECIFIED,
				}, nil
			},
			expected:    core.StatusUpdateUnknown,
			expectedErr: errors.New("unknown status"),
		},
		{
			name: "grpc error",
			statusFunc: func(ctx context.Context, req *emptypb.Empty, opts ...grpc.CallOption) (*updatepb.StatusReply, error) {
				return nil, errors.New("grpc error")
			},
			expected:    core.StatusUpdateUnknown,
			expectedErr: errors.New("grpc error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := &Client{
				log: newTestLogger(),
				client: &mockUpdateClient{
					statusFunc: tt.statusFunc,
				},
			}

			status, err := client.Status(context.Background())

			if tt.expectedErr != nil {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedErr.Error())
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, status)
			}
		})
	}
}

func TestClient_Stats(t *testing.T) {
	tests := []struct {
		name        string
		statsFunc   func(ctx context.Context, req *emptypb.Empty, opts ...grpc.CallOption) (*updatepb.StatsReply, error)
		expected    core.UpdateStats
		expectedErr error
	}{
		{
			name: "successful stats",
			statsFunc: func(ctx context.Context, req *emptypb.Empty, opts ...grpc.CallOption) (*updatepb.StatsReply, error) {
				return &updatepb.StatsReply{
					WordsTotal:    100,
					WordsUnique:   50,
					ComicsTotal:   10,
					ComicsFetched: 5,
				}, nil
			},
			expected: core.UpdateStats{
				WordsTotal:    100,
				WordsUnique:   50,
				ComicsTotal:   10,
				ComicsFetched: 5,
			},
			expectedErr: nil,
		},
		{
			name: "grpc error",
			statsFunc: func(ctx context.Context, req *emptypb.Empty, opts ...grpc.CallOption) (*updatepb.StatsReply, error) {
				return nil, errors.New("grpc error")
			},
			expected:    core.UpdateStats{},
			expectedErr: errors.New("grpc error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := &Client{
				log: newTestLogger(),
				client: &mockUpdateClient{
					statsFunc: tt.statsFunc,
				},
			}

			stats, err := client.Stats(context.Background())

			if tt.expectedErr != nil {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedErr.Error())
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, stats)
			}
		})
	}
}

func TestClient_Update(t *testing.T) {
	tests := []struct {
		name        string
		updateFunc  func(ctx context.Context, req *emptypb.Empty, opts ...grpc.CallOption) (*emptypb.Empty, error)
		expectedErr error
	}{
		{
			name: "successful update",
			updateFunc: func(ctx context.Context, req *emptypb.Empty, opts ...grpc.CallOption) (*emptypb.Empty, error) {
				return &emptypb.Empty{}, nil
			},
			expectedErr: nil,
		},
		{
			name: "grpc error",
			updateFunc: func(ctx context.Context, req *emptypb.Empty, opts ...grpc.CallOption) (*emptypb.Empty, error) {
				return nil, errors.New("grpc error")
			},
			expectedErr: errors.New("grpc error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := &Client{
				log: newTestLogger(),
				client: &mockUpdateClient{
					updateFunc: tt.updateFunc,
				},
			}

			err := client.Update(context.Background())

			if tt.expectedErr != nil {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedErr.Error())
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestClient_Drop(t *testing.T) {
	tests := []struct {
		name        string
		dropFunc    func(ctx context.Context, req *emptypb.Empty, opts ...grpc.CallOption) (*emptypb.Empty, error)
		expectedErr error
	}{
		{
			name: "successful drop",
			dropFunc: func(ctx context.Context, req *emptypb.Empty, opts ...grpc.CallOption) (*emptypb.Empty, error) {
				return &emptypb.Empty{}, nil
			},
			expectedErr: nil,
		},
		{
			name: "grpc error",
			dropFunc: func(ctx context.Context, req *emptypb.Empty, opts ...grpc.CallOption) (*emptypb.Empty, error) {
				return nil, errors.New("grpc error")
			},
			expectedErr: errors.New("grpc error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := &Client{
				log: newTestLogger(),
				client: &mockUpdateClient{
					dropFunc: tt.dropFunc,
				},
			}

			err := client.Drop(context.Background())

			if tt.expectedErr != nil {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedErr.Error())
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
