package grpc

import (
	"context"
	"log/slog"
	"strconv"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	wordspb "yadro.com/course/proto/words"
	"yadro.com/course/words/core"
)

const (
	maxPhraseLen    = 20000
	maxShutdownTime = 10 * time.Second
)

type Server struct {
	wordspb.UnimplementedWordsServer
	words core.Normalizer
}

func New(words core.Normalizer) *Server {
	return &Server{
		words: words,
	}
}

func (s *Server) Ping(_ context.Context, in *emptypb.Empty) (*emptypb.Empty, error) {
	slog.Info("pinged words service")
	return nil, nil
}

func (s *Server) Norm(_ context.Context, in *wordspb.WordsRequest) (*wordspb.WordsReply, error) {
	if len(in.GetPhrase()) > maxPhraseLen {
		slog.Error("phrase is large than max phrase length", "phrase", in.GetPhrase(), "max phrase length", maxPhraseLen)
		return nil, status.Error(
			codes.ResourceExhausted,
			"phrase is large than "+strconv.Itoa(maxPhraseLen),
		)
	}
	return &wordspb.WordsReply{
		Words: s.words.Norm(in.GetPhrase()),
	}, nil
}
