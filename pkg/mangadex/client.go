// Package mangadex implements a client which can download chapters from Mangadex
package mangadex

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"time"

	"golang.org/x/time/rate"
)

const ApiUrl = "https://api.mangadex.org"

type Client struct {
	client *http.Client
	rl     *rate.Limiter
}

func NewClient() *Client {
	return &Client{
		client: &http.Client{Timeout: 60 * time.Second},
		rl:     rate.NewLimiter(rate.Every(time.Second/5), 1), // 5 requests per second
	}
}

func formatUrl(endpoint string) string {
	return fmt.Sprintf("%s/%s", ApiUrl, endpoint)
}

func (c *Client) Request(method, endpoint string, payload io.Reader) (*http.Response, error) {
	return c.fmtAndSend(method, endpoint, payload)
}

func (c *Client) fmtAndSend(method, endpoint string, payload io.Reader) (*http.Response, error) {
	r, err := http.NewRequest(method, formatUrl(endpoint), payload)
	if err != nil {
		return nil, err
	}
	return c.sendRequest(r)
}

func (c *Client) sendRequest(r *http.Request) (*http.Response, error) {
	ctx := context.Background()
	err := c.rl.Wait(ctx)
	if err != nil {
		return nil, err
	}

	resp, err := c.client.Do(r)
	if err != nil {
		return nil, err
	}

	return resp, nil
}
