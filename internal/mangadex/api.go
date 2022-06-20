package mangadex

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"

	"golang.org/x/time/rate"
)

const (
	apiURL       = "https://api.mangadex.org"
	mangadexTime = "2006-01-02T15:04:05"
)

var (
	c      = &http.Client{}
	rl     = rate.NewLimiter(5, 1)
	homeRl = rate.NewLimiter(0.55, 1)
)

func get(ctx context.Context, endpoint string, query url.Values) (*http.Response, error) {
	address := fmt.Sprintf("%s/%s", apiURL, endpoint)
	if query != nil {
		address += "?" + query.Encode()
	}

	req, err := http.NewRequestWithContext(ctx, "GET", address, nil)
	if err != nil {
		return nil, err
	}

	err = rl.Wait(ctx)
	if err != nil {
		return nil, err
	}

	return c.Do(req)
}

//  result pops up in API responses

type result struct {
	Result string `json:"result"`
	Errors []struct {
		ID     string `json:"id"`
		Status int    `json:"status"`
		Title  string `json:"title"`
		Detail string `json:"detail"`
	} `json:"errors"`
}

func (r result) errored() bool {
	return r.Result == "error"
}

func (r result) err() error {
	msg := "mangadex error"
	for _, e := range r.Errors {
		msg += fmt.Sprintf(": %s - %s", e.Title, e.Detail)
	}

	return errors.New(msg)
}
