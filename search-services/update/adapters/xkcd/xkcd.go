package xkcd

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"yadro.com/course/update/core"
)

type Client struct {
	log    *slog.Logger
	client http.Client
	url    string
}

func NewClient(url string, timeout time.Duration, log *slog.Logger) (*Client, error) {
	if url == "" {
		return nil, fmt.Errorf("empty base url specified")
	}
	return &Client{
		client: http.Client{Timeout: timeout},
		log:    log,
		url:    url,
	}, nil
}

func (c Client) Get(ctx context.Context, id int) (core.XKCDInfo, error) {
	if id == 404 {
		c.log.Debug("skipping special comic 404")
		return core.XKCDInfo{
			ID:          id,
			Description: "Not found",
			Title:       "404",
		}, nil
	}
	resp, err := c.client.Get(c.url + fmt.Sprintf("/%d/info.0.json", id))
	if err != nil {
		c.log.Error("failed to get comic", "id", id, "error", err)
		return core.XKCDInfo{}, core.ErrNotFound
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		c.log.Error("unexpected status code", "status", resp.StatusCode)
		return core.XKCDInfo{}, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	info := struct {
		ID         int    `json:"num"`
		URL        string `json:"img"`
		Title      string `json:"title"`
		Alt        string `json:"alt"`
		SafeTitle  string `json:"safe_title"`
		Transcript string `json:"transcript"`
	}{}
	if err := json.NewDecoder(resp.Body).Decode(&info); err != nil {
		c.log.Error("failed to decode comic", "error", err)
		return core.XKCDInfo{}, fmt.Errorf("failed to decode comic: %w", err)
	}

	c.log.Debug("got comic", "id", id, "title", info.Title)
	return core.XKCDInfo{
		ID:          info.ID,
		URL:         info.URL,
		Title:       info.Title,
		Description: info.Transcript + " " + info.Alt + " " + info.SafeTitle,
	}, nil
}

func (c Client) LastID(ctx context.Context) (int, error) {
	resp, err := c.client.Get(c.url + "/info.0.json")
	if err != nil {
		c.log.Error("failed to get last id", "error", err)
		return 0, fmt.Errorf("failed to get last id: %w", err)
	}
	defer resp.Body.Close()

	var info struct {
		Num int `json:"num"`
	}

	err = json.NewDecoder(resp.Body).Decode(&info)
	if err != nil {
		c.log.Error("failed to decode last id", "error", err)
		return 0, fmt.Errorf("failed to decode last id: %w", err)
	}

	c.log.Debug("last id", "id", info.Num)

	return info.Num, nil
}
