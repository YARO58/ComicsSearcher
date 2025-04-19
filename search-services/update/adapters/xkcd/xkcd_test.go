package xkcd

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"yadro.com/course/update/core"
)

func newTestLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}

func TestNewClient(t *testing.T) {
	tests := []struct {
		name    string
		url     string
		wantErr bool
	}{
		{
			name:    "valid url",
			url:     "http://example.com",
			wantErr: false,
		},
		{
			name:    "empty url",
			url:     "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, err := NewClient(tt.url, time.Second, newTestLogger())
			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, client)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, client)
				assert.Equal(t, tt.url, client.url)
			}
		})
	}
}

func TestClient_Get(t *testing.T) {
	tests := []struct {
		name        string
		id          int
		handler     http.HandlerFunc
		expected    core.XKCDInfo
		expectedErr error
	}{
		{
			name: "successful get",
			id:   1,
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				_, err := io.WriteString(w, `{
					"num": 1,
					"img": "http://example.com/1.png",
					"title": "Test Comic",
					"alt": "Alt text",
					"safe_title": "Safe Title",
					"transcript": "Transcript"
				}`)
				if err != nil {
					t.Fatal(err)
				}
			},
			expected: core.XKCDInfo{
				ID:          1,
				URL:         "http://example.com/1.png",
				Title:       "Test Comic",
				Description: "Transcript Alt text Safe Title",
			},
			expectedErr: nil,
		},
		{
			name: "comic 404",
			id:   404,
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusNotFound)
			},
			expected: core.XKCDInfo{
				ID:          404,
				Description: "Not found",
				Title:       "404",
			},
			expectedErr: nil,
		},
		{
			name: "not found error",
			id:   999,
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusNotFound)
			},
			expected:    core.XKCDInfo{},
			expectedErr: errors.New("unexpected status code: 404"),
		},
		{
			name: "invalid json",
			id:   1,
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				_, err := io.WriteString(w, `invalid json`)
				if err != nil {
					t.Fatal(err)
				}
			},
			expected:    core.XKCDInfo{},
			expectedErr: errors.New("failed to decode comic"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(tt.handler)
			defer server.Close()

			client, err := NewClient(server.URL, time.Second, newTestLogger())
			assert.NoError(t, err)

			info, err := client.Get(context.Background(), tt.id)

			if tt.expectedErr != nil {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedErr.Error())
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, info)
			}
		})
	}
}

func TestClient_LastID(t *testing.T) {
	tests := []struct {
		name        string
		handler     http.HandlerFunc
		expected    int
		expectedErr error
	}{
		{
			name: "successful get last id",
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				_, err := io.WriteString(w, `{"num": 100}`)
				if err != nil {
					t.Fatal(err)
				}
			},
			expected:    100,
			expectedErr: nil,
		},
		{
			name: "invalid json",
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				_, err := io.WriteString(w, `invalid json`)
				if err != nil {
					t.Fatal(err)
				}
			},
			expected:    0,
			expectedErr: errors.New("failed to decode last id"),
		},
		{
			name: "server error",
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusInternalServerError)
			},
			expected:    0,
			expectedErr: errors.New("failed to decode last id: EOF"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(tt.handler)
			defer server.Close()

			client, err := NewClient(server.URL, time.Second, newTestLogger())
			assert.NoError(t, err)

			id, err := client.LastID(context.Background())

			if tt.expectedErr != nil {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedErr.Error())
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, id)
			}
		})
	}
}
