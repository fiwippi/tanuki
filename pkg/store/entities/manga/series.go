package manga

import (
	"fmt"
	"github.com/fiwippi/tanuki/internal/archive"
	"github.com/fiwippi/tanuki/internal/fse"
	"io/fs"
	"path/filepath"
	"sync"
)

// ParsedSeries is a collection of ParsedEntry volumes/chapters
type ParsedSeries struct {
	Title   string         // Title of the ParsedSeries
	Entries []*ParsedEntry // Slice of all entries in the series
}

func ParseSeriesFolder(dir string) (*ParsedSeries, error) {
	var errors error
	series := &ParsedSeries{}
	entries := make([]*ParsedEntry, 0)
	errorQueue := make(chan error, 1)
	entriesQueue := make(chan *ParsedEntry, 1)

	// Set the series title
	series.Title = fse.Filename(dir)

	// Set the entries
	order := 1
	wg := sync.WaitGroup{}

	go func() {
		for e := range errorQueue {
			errors = fmt.Errorf("%w, %s", errors, e)
		}
	}()

	go func() {
		for e := range entriesQueue {
			entries = append(entries, e)
			wg.Done()
		}
	}()

	filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if !d.IsDir() {
			// We want to avoid parsing non-archive files like cover images
			if _, err := archive.InferType(path); err == nil {
				wg.Add(1)
				go func(o int, p string) {
					m, err := ParseArchive(p)
					if err != nil {
						errorQueue <- err
						return
					}
					m.Order = o
					entriesQueue <- m
				}(order, path)
				order += 1
			}
		}
		return nil
	})

	wg.Wait()
	close(errorQueue)
	close(entriesQueue)
	series.Entries = entries

	return series, errors
}
