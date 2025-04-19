package search

import (
	"context"
	"log/slog"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/types/known/emptypb"
	"yadro.com/course/api/core"
	searchpb "yadro.com/course/proto/search"
)

type Client struct {
	log    *slog.Logger
	client searchpb.SearchClient
}

func NewClient(address string, log *slog.Logger) (*Client, error) {
	conn, err := grpc.NewClient(address, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}
	return &Client{
		log:    log,
		client: searchpb.NewSearchClient(conn),
	}, nil
}

func (c Client) Ping(ctx context.Context) error {
	_, err := c.client.Ping(ctx, &emptypb.Empty{})
	if err != nil {
		c.log.Error("cannot ping search service", "error", err)
		return err
	}

	c.log.Debug("pinged search service")
	return nil
}

func (c Client) Search(ctx context.Context, phrase string, limit int) ([]core.Comics, int, error) {
	response, err := c.client.Search(ctx, &searchpb.SearchRequest{
		Query: phrase,
		Limit: int64(limit),
	})
	if err != nil {
		c.log.Error("cannot search comics", "error", err)
		return nil, 0, err
	}

	comics := make([]core.Comics, 0, len(response.Items))
	for _, item := range response.Items {
		comics = append(comics, core.Comics{
			ID:  item.Id,
			URL: item.Url,
		})
	}

	return comics, int(response.Total), nil
}

func (c Client) ISearch(ctx context.Context, phrase string, limit int) ([]core.Comics, int, error) {
	response, err := c.client.ISearch(ctx, &searchpb.ISearchRequest{
		Query: phrase,
		Limit: int64(limit),
	})
	if err != nil {
		c.log.Error("cannot Isearch comics", "error", err)
		return nil, 0, err
	}

	comics := make([]core.Comics, 0, len(response.Items))
	for _, item := range response.Items {
		comics = append(comics, core.Comics{
			ID:  item.Id,
			URL: item.Url,
		})
	}
	return comics, int(response.Total), nil
}
