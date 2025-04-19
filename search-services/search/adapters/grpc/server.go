package grpc

import (
	"context"
	"log/slog"

	"google.golang.org/protobuf/types/known/emptypb"
	searchpb "yadro.com/course/proto/search"
	"yadro.com/course/search/core"
)

type Server struct {
	searchpb.UnimplementedSearchServer
	service core.Searcher
}

func NewServer(service core.Searcher) *Server {
	return &Server{service: service}
}

func (s *Server) Ping(_ context.Context, _ *emptypb.Empty) (*emptypb.Empty, error) {
	slog.Info("pinged search service")
	return nil, nil
}

func (s *Server) Search(ctx context.Context, req *searchpb.SearchRequest) (*searchpb.ComicsResponse, error) {
	comics, total, err := s.service.Search(ctx, req.Query, int(req.Limit))
	if err != nil {
		return nil, err
	}

	pbComics := make([]*searchpb.Comics, len(comics))
	for i, c := range comics {
		pbComics[i] = &searchpb.Comics{
			Id:  int64(c.ID),
			Url: c.URL,
		}
	}

	return &searchpb.ComicsResponse{
		Items: pbComics,
		Total: int64(total),
	}, nil
}

func (s *Server) ISearch(ctx context.Context, req *searchpb.ISearchRequest) (*searchpb.ComicsResponse, error) {
	comics, total, err := s.service.ISearch(ctx, req.Query, int(req.Limit))
	if err != nil {
		return nil, err
	}

	pbComics := make([]*searchpb.Comics, len(comics))
	for i, c := range comics {
		pbComics[i] = &searchpb.Comics{
			Id:  int64(c.ID),
			Url: c.URL,
		}
	}

	return &searchpb.ComicsResponse{
		Items: pbComics,
		Total: int64(total),
	}, nil
}
