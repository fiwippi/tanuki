package mangadex

import (
	"context"
	"encoding/json"
	"errors"
	"net/url"
	"strconv"
)

func SearchManga(ctx context.Context, title string, limit int) ([]Listing, error) {
	if limit < 0 {
		return nil, errors.New("limit cannot be below zero")
	}

	// Build and execute query
	q := url.Values{}
	q.Add("title", title)
	q.Add("limit", strconv.Itoa(limit))
	q.Add("includes[]", "cover_art")
	resp, err := get(ctx, "manga", q)
	if err != nil {
		return nil, err
	}

	// Decode response
	r := struct {
		result
		Data []listingData `json:"data"`
	}{}
	err = json.NewDecoder(resp.Body).Decode(&r)
	if err != nil {
		return nil, err
	}
	if r.errored() {
		return nil, r.err()
	}

	// Make listing
	listings := make([]Listing, 0)
	for _, d := range r.Data {
		listings = append(listings, d.makeListing())
	}
	return listings, nil
}
