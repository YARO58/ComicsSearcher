package grpc

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	wordspb "yadro.com/course/proto/words"
)

type MockNormalizer struct {
	normFunc func(phrase string) []string
}

func (m MockNormalizer) Norm(phrase string) []string {
	if m.normFunc != nil {
		return m.normFunc(phrase)
	}
	return []string{}
}

func TestServer_Ping(t *testing.T) {
	normalizer := MockNormalizer{}
	server := New(normalizer)

	resp, err := server.Ping(context.Background(), &emptypb.Empty{})
	require.NoError(t, err)
	assert.Nil(t, resp)
}

func TestServer_Norm(t *testing.T) {
	tests := []struct {
		name      string
		request   *wordspb.WordsRequest
		normFunc  func(string) []string
		want      *wordspb.WordsReply
		wantError bool
		errorCode codes.Code
	}{
		{
			name:    "successful normalization",
			request: &wordspb.WordsRequest{Phrase: "hello world"},
			normFunc: func(phrase string) []string {
				return []string{"hello", "world"}
			},
			want: &wordspb.WordsReply{
				Words: []string{"hello", "world"},
			},
		},
		{
			name:    "empty phrase",
			request: &wordspb.WordsRequest{Phrase: ""},
			normFunc: func(phrase string) []string {
				return []string{}
			},
			want: &wordspb.WordsReply{
				Words: []string{},
			},
		},
		{
			name: "phrase too long",
			request: &wordspb.WordsRequest{
				Phrase: func() string {
					s := make([]byte, maxPhraseLen+1)
					for i := range s {
						s[i] = 'a'
					}
					return string(s)
				}(),
			},
			wantError: true,
			errorCode: codes.ResourceExhausted,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			normalizer := MockNormalizer{normFunc: tt.normFunc}
			server := New(normalizer)

			resp, err := server.Norm(context.Background(), tt.request)

			if tt.wantError {
				require.Error(t, err)
				st, ok := status.FromError(err)
				require.True(t, ok)
				assert.Equal(t, tt.errorCode, st.Code())
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.want.Words, resp.Words)
		})
	}
}
