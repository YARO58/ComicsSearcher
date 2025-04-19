package words

import (
	"context"
	"fmt"
	"log/slog"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	wordspb "yadro.com/course/proto/words"
)

type Client struct {
	log    *slog.Logger
	client wordspb.WordsClient
}

func NewClient(address string, log *slog.Logger) (*Client, error) {
	conn, err := grpc.NewClient(address, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}
	return &Client{
		client: wordspb.NewWordsClient(conn),
		log:    log,
	}, nil
}

func (c Client) Norm(ctx context.Context, phrase string) ([]string, error) {
	response, err := c.client.Norm(ctx, &wordspb.WordsRequest{Phrase: phrase})
	if err != nil {
		c.log.Error("failed to normalize words", "error", err)
		return nil, fmt.Errorf("failed to normalize words: %w", err)
	}

	c.log.Debug("words normalized", "words", response.Words)

	return response.Words, nil
}
