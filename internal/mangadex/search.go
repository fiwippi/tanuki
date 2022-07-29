package mangadex

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"strconv"
)

func SearchManga(ctx context.Context, title string, limit int) ([]Listing, error) {
	if limit < 0 {
		return nil, errors.New("limit cannot be below zero")
	}

	q := url.Values{}
	q.Add("title", title)
	q.Add("limit", strconv.Itoa(limit))
	q.Add("includes[]", "cover_art")

	resp, err := get(ctx, "manga", q)
	if err != nil {
		return nil, err
	}

	r := struct {
		result
		Data []struct {
			ID         string `json:"id"`
			Attributes struct {
				Title struct {
					English string `json:"en"`
				} `json:"title"`
				Description struct {
					English string `json:"en"`
				} `json:"description"`
				Year int `json:"year"`
			} `json:"attributes"`
			Relationships []struct {
				ID         string `json:"id"`
				Type       string `json:"type"`
				Attributes struct {
					FileName string `json:"fileName"`
				} `json:"attributes"`
			} `json:"relationships"`
		} `json:"data"`
	}{}

	err = json.NewDecoder(resp.Body).Decode(&r)
	if err != nil {
		return nil, err
	}

	if r.errored() {
		return nil, r.err()
	}

	listings := make([]Listing, 0)
	for _, d := range r.Data {
		var coverURL string
		for _, rel := range d.Relationships {
			if rel.Type == "cover_art" {
				coverURL = fmt.Sprintf("https://uploads.mangadex.org/covers/%s/%s", d.ID, rel.Attributes.FileName)
				break
			}
		}

		listings = append(listings, Listing{
			ID:            d.ID,
			Title:         d.Attributes.Title.English,
			Description:   d.Attributes.Description.English,
			CoverURL:      coverURL,
			SmallCoverURL: coverURL + ".256.jpg",
			Year:          d.Attributes.Year,
		})
	}

	return listings, nil
}
