package db

import (
	bolt "go.etcd.io/bbolt"

	"github.com/fiwippi/tanuki/internal/sets"
	"github.com/fiwippi/tanuki/pkg/api"
	"github.com/fiwippi/tanuki/pkg/auth"
	"github.com/fiwippi/tanuki/pkg/core")

type SeriesListBucket struct {
	*bolt.Bucket
}

func (b *SeriesListBucket) getSeries(sid []byte) (*SeriesBucket, error) {
	bucket := b.Bucket.Bucket(sid)
	if bucket == nil {
		return nil, ErrSeriesNotExist
	}

	return &SeriesBucket{Bucket: bucket}, nil
}

func (b *SeriesListBucket) GetSeries(sid string) (*SeriesBucket, error) {
	return b.getSeries([]byte(sid))
}

func (b *SeriesListBucket) GetFirstEntry(seriesHash string) (*MangaBucket, error) {
	seriesBucket, err := b.GetSeries(seriesHash)
	if err != nil {
		return nil, err
	}

	key := []byte(seriesBucket.Data()[0].Hash)
	mangaBucket := seriesBucket.GetEntry(key)
	if mangaBucket == nil {
		return nil, ErrMangaNotExist
	}

	return mangaBucket, nil
}

func (b *SeriesListBucket) GetEntry(seriesHash, entryHash string) (*MangaBucket, error) {
	seriesBucket, err := b.GetSeries(seriesHash)
	if err != nil {
		return nil, err
	}

	mangaBucket := seriesBucket.GetEntry([]byte(entryHash))
	if mangaBucket == nil {
		return nil, ErrMangaNotExist
	}

	return mangaBucket, nil
}

func (b *SeriesListBucket) AddSeries(s *core.Series, manga []*core.Manga) error {
	if len(manga) == 0 {
		return ErrMangaNotExist
	}

	// Create the bucket for the series
	seriesHash := auth.HashSHA1(s.Title)
	newBucket, err := b.Bucket.CreateBucketIfNotExists([]byte(seriesHash))
	if err != nil {
		return err
	}
	seriesBucket := &SeriesBucket{newBucket}

	// Set the series title
	if err := seriesBucket.SetTitle(s.Title); err != nil {
		return err
	}
	// Only set series tags if one doesn't already exist
	t := seriesBucket.Tags()
	if t == nil {
		if err := seriesBucket.SetTags(sets.NewSet()); err != nil {
			return err
		}
	}
	// Create a new series metadata entry if it doesn't exist
	meta := seriesBucket.Metadata()
	if meta == nil {
		newMetadata := core.NewSeriesMetadata()
		newMetadata.Title = s.Title
		if err := seriesBucket.SetMetadata(newMetadata); err != nil {
			return err
		}
	}
	// Create new cover entry if it doesn't exist
	cover := seriesBucket.Cover()
	if cover == nil {
		if err := seriesBucket.SetCover(&core.Cover{}); err != nil {
			return err
		}
	}

	// Ensure the bucket for entries exists
	_, err = seriesBucket.CreateBucketIfNotExists(keySeriesEntries)
	if err != nil {
		return err
	}
	// AddEntry the manga entries
	seriesData := make(api.SeriesEntries, 0)
	for _, m := range manga {
		entryHash := auth.HashSHA1(m.Title)
		if e := seriesBucket.GetEntry([]byte(entryHash)); e != nil {
			if e.Archive().ModTime != m.Archive.ModTime {
				err := seriesBucket.DeleteEntry([]byte(entryHash))
				if err != nil {
					return err
				}
			}
		}

		err := seriesBucket.AddEntry(m)
		if err != nil {
			return err
		}

		// AddEntry the series data
		e := &api.SeriesEntry{
			Hash:  auth.HashSHA1(m.Title),
			Title: m.Title,
			Pages: len(m.Pages),
			Path:  m.Archive.Path,
		}
		seriesData = append(seriesData, e)
	}
	if err := seriesBucket.SetData(seriesData); err != nil {
		return err
	}

	return nil
}

func (b *SeriesListBucket) ForEachSeries(f func(hash string, b *SeriesBucket) error) error {
	return b.Bucket.ForEach(func(k, v []byte) error {
		if v == nil {
			s, _ := b.getSeries(k)
			err := f(string(k), s)
			if err != nil {
				return err
			}
		}
		return nil
	})
}