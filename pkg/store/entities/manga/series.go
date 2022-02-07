package manga

import (
	"io/fs"
	"path/filepath"
	"sort"
	"sync"

	"github.com/fvbommel/sortorder"
	"github.com/rs/zerolog/log"

	"github.com/fiwippi/tanuki/internal/archive"
	"github.com/fiwippi/tanuki/internal/errors"
	"github.com/fiwippi/tanuki/internal/fse"
)

// ParsedSeries is a collection of ParsedEntry volumes/chapters
type ParsedSeries struct {
	Title   string         // Title of the ParsedSeries
	Entries []*ParsedEntry // Slice of all entries in the series
}

func ParseSeriesFolder(dir string) (*ParsedSeries, error) {
	var errs error
	series := &ParsedSeries{}
	entries := make([]*ParsedEntry, 0)
	errorQueue := make(chan error, 1)
	entriesQueue := make(chan *ParsedEntry, 1)

	// Set the series title
	series.Title = fse.Filename(dir)

	// Set the entries
	wg := sync.WaitGroup{}

	go func() {
		for e := range errorQueue {
			errs = errors.Wrap(errs, e)
			wg.Done()
		}
	}()

	go func() {
		for e := range entriesQueue {
			entries = append(entries, e)
			wg.Done()
		}
	}()

	err := filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if !d.IsDir() {
			// We want to avoid parsing non-archive files like cover images
			_, err = archive.InferType(path)
			if err != nil {
				log.Trace().Err(errs).Str("fp", path).Msg("file is not archive")
				return nil
			}

			// Parse the archive
			wg.Add(1)
			go func(p string) {
				m, err := ParseArchive(p)
				if err != nil {
					errorQueue <- err
					return
				}
				entriesQueue <- m
			}(path)
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	wg.Wait()
	close(errorQueue)
	close(entriesQueue)

	// Sort the entries list and add the correct order for each one
	sort.SliceStable(entries, func(i, j int) bool {
		return sortorder.NaturalLess(entries[i].Archive.Title, entries[j].Archive.Title)
	})
	for i := range entries {
		entries[i].Order = i + 1
	}
	series.Entries = entries

	return series, errs
}
