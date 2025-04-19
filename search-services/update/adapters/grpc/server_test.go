package grpc

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"

	updatepb "yadro.com/course/proto/update"
	"yadro.com/course/update/core"
)

type mockUpdater struct {
	status core.ServiceStatus
	stats  core.ServiceStats
	err    error
}

func (m *mockUpdater) Update(ctx context.Context) error {
	return m.err
}

func (m *mockUpdater) Stats(ctx context.Context) (core.ServiceStats, error) {
	return m.stats, m.err
}

func (m *mockUpdater) Status(ctx context.Context) core.ServiceStatus {
	return m.status
}

func (m *mockUpdater) Drop(ctx context.Context) error {
	return m.err
}

func TestServer_Ping(t *testing.T) {
	server := NewServer(&mockUpdater{})
	resp, err := server.Ping(context.Background(), &emptypb.Empty{})
	assert.NoError(t, err)
	assert.Nil(t, resp)
}

func TestServer_Status(t *testing.T) {
	tests := []struct {
		name           string
		serviceStatus  core.ServiceStatus
		expectedStatus updatepb.Status
		expectedError  error
	}{
		{
			name:           "running status",
			serviceStatus:  core.StatusRunning,
			expectedStatus: updatepb.Status_STATUS_RUNNING,
			expectedError:  nil,
		},
		{
			name:           "idle status",
			serviceStatus:  core.StatusIdle,
			expectedStatus: updatepb.Status_STATUS_IDLE,
			expectedError:  nil,
		},
		{
			name:           "unknown status",
			serviceStatus:  core.ServiceStatus("unknown"),
			expectedStatus: updatepb.Status_STATUS_UNSPECIFIED,
			expectedError:  status.Error(codes.Internal, "unknown status"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := NewServer(&mockUpdater{status: tt.serviceStatus})
			resp, err := server.Status(context.Background(), &emptypb.Empty{})

			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedError.Error(), err.Error())
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, tt.expectedStatus, resp.Status)
		})
	}
}

func TestServer_Update(t *testing.T) {
	tests := []struct {
		name          string
		serviceError  error
		expectedError error
	}{
		{
			name:          "successful update",
			serviceError:  nil,
			expectedError: nil,
		},
		{
			name:          "already exists error",
			serviceError:  core.ErrAlreadyExists,
			expectedError: status.Error(codes.AlreadyExists, "update already in progress"),
		},
		{
			name:          "internal error",
			serviceError:  errors.New("some error"),
			expectedError: status.Error(codes.Internal, "failed to update"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := NewServer(&mockUpdater{err: tt.serviceError})
			resp, err := server.Update(context.Background(), &emptypb.Empty{})

			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedError.Error(), err.Error())
				return
			}

			assert.NoError(t, err)
			assert.NotNil(t, resp)
		})
	}
}

func TestServer_Stats(t *testing.T) {
	tests := []struct {
		name          string
		stats         core.ServiceStats
		serviceError  error
		expectedError error
	}{
		{
			name: "successful stats",
			stats: core.ServiceStats{
				DBStats: core.DBStats{
					WordsTotal:    100,
					WordsUnique:   50,
					ComicsFetched: 5,
				},
				ComicsTotal: 10,
			},
			serviceError:  nil,
			expectedError: nil,
		},
		{
			name: "service error",
			stats: core.ServiceStats{
				DBStats: core.DBStats{},
			},
			serviceError:  errors.New("some error"),
			expectedError: status.Error(codes.Internal, "failed to get stats"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := NewServer(&mockUpdater{stats: tt.stats, err: tt.serviceError})
			resp, err := server.Stats(context.Background(), &emptypb.Empty{})

			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedError.Error(), err.Error())
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, tt.stats.WordsTotal, int(resp.WordsTotal))
			assert.Equal(t, tt.stats.WordsUnique, int(resp.WordsUnique))
			assert.Equal(t, tt.stats.ComicsTotal, int(resp.ComicsTotal))
			assert.Equal(t, tt.stats.ComicsFetched, int(resp.ComicsFetched))
		})
	}
}

func TestServer_Drop(t *testing.T) {
	tests := []struct {
		name          string
		serviceError  error
		expectedError error
	}{
		{
			name:          "successful drop",
			serviceError:  nil,
			expectedError: nil,
		},
		{
			name:          "service error",
			serviceError:  errors.New("some error"),
			expectedError: status.Error(codes.Internal, "failed to drop data"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := NewServer(&mockUpdater{err: tt.serviceError})
			resp, err := server.Drop(context.Background(), &emptypb.Empty{})

			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedError.Error(), err.Error())
				return
			}

			assert.NoError(t, err)
			assert.NotNil(t, resp)
		})
	}
}
