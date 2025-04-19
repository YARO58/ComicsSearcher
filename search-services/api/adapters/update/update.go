package update

import (
	"context"
	"errors"
	"log/slog"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/types/known/emptypb"
	"yadro.com/course/api/core"
	updatepb "yadro.com/course/proto/update"
)

type Client struct {
	log    *slog.Logger
	client updatepb.UpdateClient
}

func NewClient(address string, log *slog.Logger) (*Client, error) {
	conn, err := grpc.NewClient(address, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}
	return &Client{
		client: updatepb.NewUpdateClient(conn),
		log:    log,
	}, nil
}

func (c Client) Ping(ctx context.Context) error {
	_, err := c.client.Ping(ctx, &emptypb.Empty{})
	if err != nil {
		c.log.Error("cannot ping update service", "error", err)
		return err
	}

	c.log.Debug("pinged update service")
	return nil
}

func (c Client) Status(ctx context.Context) (core.UpdateStatus, error) {
	status, err := c.client.Status(ctx, &emptypb.Empty{})
	if err != nil {
		c.log.Error("cannot get update service status", "error", err)
		return core.StatusUpdateUnknown, err
	}

	switch status.Status {
	case updatepb.Status_STATUS_RUNNING:
		return core.StatusUpdateRunning, nil
	case updatepb.Status_STATUS_IDLE:
		return core.StatusUpdateIdle, nil
	}

	return core.StatusUpdateUnknown, errors.New("unknown status")
}

func (c Client) Stats(ctx context.Context) (core.UpdateStats, error) {
	stats, err := c.client.Stats(ctx, &emptypb.Empty{})
	if err != nil {
		c.log.Error("cannot get update service stats", "error", err)
		return core.UpdateStats{}, err
	}

	return core.UpdateStats{
		WordsTotal:    int(stats.WordsTotal),
		WordsUnique:   int(stats.WordsUnique),
		ComicsTotal:   int(stats.ComicsTotal),
		ComicsFetched: int(stats.ComicsFetched),
	}, nil

}

func (c Client) Update(ctx context.Context) error {
	_, err := c.client.Update(ctx, &emptypb.Empty{})
	if err != nil {
		c.log.Error("cannot update update service", "error", err)
		return err
	}

	return nil
}

func (c Client) Drop(ctx context.Context) error {
	_, err := c.client.Drop(ctx, &emptypb.Empty{})
	if err != nil {
		c.log.Error("cannot drop update service", "error", err)
		return err
	}

	return nil
}
