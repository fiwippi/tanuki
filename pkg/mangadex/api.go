package mangadex

import (
	"context"
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
	c         = &http.Client{}
	defaultRL = rate.NewLimiter(5, 1)   // 5 events per second
	atHomeRL  = rate.NewLimiter(0.6, 1) // 0.6 events per second -> approx. 40 events per minute
)

func GetCover(ctx context.Context, endpoint string) (*http.Response, error) {
	// Create the request
	address := fmt.Sprintf("https://uploads.mangadex.org/covers/%s", endpoint)
	req, err := http.NewRequestWithContext(ctx, "GET", address, nil)
	if err != nil {
		return nil, err
	}

	// Rate limit the request if needed and then perform it
	err = defaultRL.Wait(ctx)
	if err != nil {
		return nil, err
	}
	return c.Do(req)
}

func get(ctx context.Context, endpoint string, query url.Values) (*http.Response, error) {
	// Create the request
	address := fmt.Sprintf("%s/%s", apiURL, endpoint)
	if query != nil {
		address += "?" + query.Encode()
	}
	req, err := http.NewRequestWithContext(ctx, "GET", address, nil)
	if err != nil {
		return nil, err
	}

	// Rate limit the request if needed and then perform it
	err = defaultRL.Wait(ctx)
	if err != nil {
		return nil, err
	}
	return c.Do(req)
}
