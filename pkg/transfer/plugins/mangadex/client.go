package mangadex

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"golang.org/x/time/rate"

	"github.com/fiwippi/tanuki/pkg/transfer"
)

// TODO support manga on external sites

const apiURL = "https://api.mangadex.org"

func New() transfer.Plugin {
	return &client{
		http: &http.Client{},
		rl:   rate.NewLimiter(5, 1),
	}
}

type client struct {
	http *http.Client
	rl   *rate.Limiter
}

func (c *client) get(ctx context.Context, endpoint string, query url.Values) (*http.Response, error) {
	address := fmt.Sprintf("%s/%s", apiURL, endpoint)
	if query != nil {
		address += "?" + query.Encode()
	}

	req, err := http.NewRequestWithContext(ctx, "GET", address, nil)
	if err != nil {
		return nil, err
	}

	err = c.rl.Wait(ctx)
	if err != nil {
		return nil, err
	}

	return c.http.Do(req)
}

func (c *client) Search(ctx context.Context, title string, limit int) ([]transfer.Listing, error) {
	if limit < 0 {
		return nil, errors.New("limit cannot be below zero")
	}

	q := url.Values{}
	q.Add("title", title)
	q.Add("limit", strconv.Itoa(limit))
	q.Add("includes[]", "cover_art")

	resp, err := c.get(ctx, "manga", q)
	if err != nil {
		return nil, err
	}

	r := struct {
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

	listings := make([]transfer.Listing, 0)
	for _, d := range r.Data {
		var coverURL string
		for _, rel := range d.Relationships {
			if rel.Type == "cover_art" {
				coverURL = fmt.Sprintf("https://uploads.mangadex.org/covers/%s/%s", d.ID, rel.Attributes.FileName)
				break
			}
		}

		l := transfer.Listing{
			ID:            d.ID,
			Title:         d.Attributes.Title.English,
			Description:   d.Attributes.Description.English,
			CoverURL:      coverURL,
			SmallCoverURL: coverURL + ".256.jpg",
			Year:          d.Attributes.Year,
		}

		listings = append(listings, l)
	}

	return listings, nil
}

func (c *client) ViewChapters(ctx context.Context, l transfer.Listing) ([]transfer.Chapter, error) {
	q := url.Values{}
	q.Add("offset", "0")
	q.Add("limit", "5")
	q.Add("translatedLanguage[]", "en")
	q.Add("order[chapter]", "desc")
	q.Add("includes[]", "scanlation_group")

	resp, err := c.get(ctx, "manga/"+l.ID+"/feed", q)
	if err != nil {
		return nil, err
	}

	r := struct {
		Data []struct {
			ID         string `json:"id"`
			Attributes struct {
				Volume      string    `json:"volume"`
				Chapter     string    `json:"chapter"`
				Title       string    `json:"title"`
				ExternalURL string    `json:"externalURL"`
				PublishedAt time.Time `json:"publishAt"`
				Pages       int       `json:"pages"`
			} `json:"attributes"`
			Relationships []struct {
				ID         string `json:"id"`
				Type       string `json:"type"`
				Attributes struct {
					Name string `json:"name"`
				} `json:"attributes"`
			} `json:"relationships"`
		} `json:"data"`
	}{}

	err = json.NewDecoder(resp.Body).Decode(&r)
	if err != nil {
		return nil, err
	}

	chapters := make([]transfer.Chapter, 0)
	for _, d := range r.Data {
		var scanG string
		for _, rel := range d.Relationships {
			if rel.Type == "scanlation_group" {
				scanG = rel.Attributes.Name
				break
			}
		}

		ch := chapter{
			id:              d.ID,
			title:           d.Attributes.Title,
			scanlationGroup: scanG,
			publishedAt:     d.Attributes.PublishedAt,
			pages:           d.Attributes.Pages,
			volume:          d.Attributes.Volume,
			chapter:         d.Attributes.Chapter,
		}

		chapters = append(chapters, ch)
	}

	return chapters, nil
}

func (c *client) Download(ctx context.Context, ch transfer.Chapter, folder string) error {
	return nil
}
