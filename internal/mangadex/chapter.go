package mangadex

import (
	"context"
	"database/sql/driver"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/fiwippi/tanuki/internal/platform/archive"
	"github.com/fiwippi/tanuki/internal/platform/dbutil"
)

type Chapter struct {
	ID              string      `json:"id"`
	SeriesID        string      `json:"series_id"`
	Title           string      `json:"title"`
	ScanlationGroup string      `json:"scanlation_group"`
	PublishedAt     dbutil.Time `json:"published_at"`
	Pages           int         `json:"pages"`
	VolumeNo        string      `json:"volume_no"`
	ChapterNo       string      `json:"chapter_no"`
}

func (ch Chapter) getHomeURL(ctx context.Context) (atHomeURLData, error) {
	resp, err := get(ctx, fmt.Sprintf("at-home/server/%s", ch.ID), nil)
	if err != nil {
		return atHomeURLData{}, err
	}
	defer resp.Body.Close()

	var data atHomeURLData
	err = json.NewDecoder(resp.Body).Decode(&data)
	if err != nil {
		return atHomeURLData{}, err
	}
	if data.errored() {
		return atHomeURLData{}, data.err()
	}

	return data, nil
}

func (ch Chapter) downloadZip(ctx context.Context, progress chan<- int) (*archive.ZipFile, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	if progress != nil {
		defer close(progress)
	}

	z, err := archive.NewZipFile()
	if err != nil {
		return nil, err
	}
	defer z.Close()

	home, err := ch.getHomeURL(ctx)
	if err != nil {
		return nil, err
	}
	if home.Invalid() {
		return nil, errors.New("no pages or no id exist for chapter")
	}

	// Download each page and write it to the archive
	for i, p := range home.Chapter.Data {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		if progress != nil {
			progress <- i + 1
		}

		err := homeRl.Wait(ctx)
		if err != nil {
			return nil, err
		}

		err = home.WritePage(i, p, z)
		if err != nil {
			return nil, err
		}
	}

	return z, nil
}

func (ch Chapter) CreateDownload(mangaTitle string, createSubscription bool) *Download {
	return &Download{
		MangaTitle:  mangaTitle,
		Chapter:     ch,
		Status:      DownloadQueued,
		CurrentPage: 0,
		TotalPages:  ch.Pages,
		Subscribe:   createSubscription,
	}
}

func (ch Chapter) Value() (driver.Value, error) {
	return json.Marshal(ch)
}

func (ch *Chapter) Scan(src interface{}) error {
	return dbutil.ScanJSON(src, ch)
}
