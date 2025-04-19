package words

import (
	"context"
	"log/slog"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/types/known/emptypb"
	wordspb "yadro.com/course/proto/words"
)

type Client struct {
	log    *slog.Logger
	client wordspb.WordsClient
}

func NewClient(address string, log *slog.Logger) (*Client, error) {
	conn, err := grpc.NewClient(address, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Error("cannot create grpc client", "error", err)
		return nil, err
	}

	return &Client{
		log:    log,
		client: wordspb.NewWordsClient(conn),
	}, nil
}

func (c Client) Norm(ctx context.Context, phrase string) ([]string, error) {
	response, err := c.client.Norm(ctx, &wordspb.WordsRequest{Phrase: phrase})
	if err != nil {
		c.log.Error("cannot normalize phrase", "error", err, "phrase", phrase)
		return nil, err
	}

	c.log.Debug("got response from words service", "words", response.Words)

	return response.Words, nil
}

func (c Client) Ping(ctx context.Context) error {
	_, err := c.client.Ping(ctx, &emptypb.Empty{})
	if err != nil {
		c.log.Error("cannot ping words service", "error", err)
		return err
	}

	c.log.Debug("pinged words service")
	return nil
}
