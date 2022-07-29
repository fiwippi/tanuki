package mangadex

import (
	"context"
	"encoding/json"
	"net/url"
	"time"

	"github.com/fiwippi/tanuki/internal/platform/dbutil"
)

type Listing struct {
	ID            string
	Title         string
	Description   string
	CoverURL      string
	SmallCoverURL string
	Year          int
}

func (l Listing) ListChapters(ctx context.Context) ([]Chapter, error) {
	return NewChapters(ctx, l.ID, time.Time{})
}

func (l Listing) NewChapters(ctx context.Context, since time.Time) ([]Chapter, error) {
	return NewChapters(ctx, l.ID, since)
}

func NewChapters(ctx context.Context, id string, since time.Time) ([]Chapter, error) {
	q := url.Values{}
	q.Add("offset", "0")
	q.Add("limit", "500")
	q.Add("translatedLanguage[]", "en")
	q.Add("order[chapter]", "desc")
	q.Add("includes[]", "scanlation_group")
	if since != (time.Time{}) {
		q.Add("publishAtSince", since.Format(mangadexTime))
	}

	resp, err := get(ctx, "manga/"+id+"/feed", q)
	if err != nil {
		return nil, err
	}

	r := struct {
		result
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

	if r.errored() {
		return nil, r.err()
	}

	chapters := make([]Chapter, 0)
	for _, d := range r.Data {
		var scanG string
		for _, rel := range d.Relationships {
			if rel.Type == "scanlation_group" {
				scanG = rel.Attributes.Name
				break
			}
		}

		ch := Chapter{
			ID:              d.ID,
			SeriesID:        id,
			Title:           d.Attributes.Title,
			ScanlationGroup: scanG,
			PublishedAt:     dbutil.Time(d.Attributes.PublishedAt),
			Pages:           d.Attributes.Pages,
			VolumeNo:        d.Attributes.Volume,
			ChapterNo:       d.Attributes.Chapter,
		}

		chapters = append(chapters, ch)
	}

	return chapters, nil
}
