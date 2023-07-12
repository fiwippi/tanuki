package mangadex

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"time"

	"github.com/fiwippi/tanuki/internal/sqlutil"
)

type listingData struct {
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
}

func (ld listingData) makeListing() Listing {
	var coverURL string
	for _, rel := range ld.Relationships {
		if rel.Type == "cover_art" {
			coverURL = fmt.Sprintf("https://uploads.mangadex.org/covers/%s/%s", ld.ID, rel.Attributes.FileName)
			break
		}
	}

	return Listing{
		ID:            ld.ID,
		Title:         ld.Attributes.Title.English,
		Description:   ld.Attributes.Description.English,
		CoverURL:      coverURL,
		SmallCoverURL: coverURL + ".256.jpg",
		Year:          ld.Attributes.Year,
	}
}

type Listing struct {
	ID            string `json:"id"`
	Title         string `json:"title"`
	Description   string `json:"description"`
	CoverURL      string `json:"cover_url"`
	SmallCoverURL string `json:"small_cover_url"`
	Year          int    `json:"year"`
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
			PublishedAt:     sqlutil.Time(d.Attributes.PublishedAt),
			Pages:           d.Attributes.Pages,
			VolumeNo:        d.Attributes.Volume,
			ChapterNo:       d.Attributes.Chapter,
		}

		chapters = append(chapters, ch)
	}

	return chapters, nil
}
