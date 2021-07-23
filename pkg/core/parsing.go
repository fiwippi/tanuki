package core

import (
	"archive/zip"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/mholt/archiver/v3"

	"github.com/fiwippi/tanuki/internal/fse"
)

func ParseArchive(fp string) (*ParsedEntry, error) {
	// Ensure valid archive
	at, err := GetArchiveType(fp)
	if err != nil {
		return nil, err
	}
	absFp, err := filepath.Abs(fp)
	if err != nil {
		return nil, err
	}
	aStats, err := os.Stat(absFp)
	if err != nil {
		return nil, err
	}
	a := &Archive{
		Path:    absFp,
		Type:    at,
		Cover:   &Cover{},
		ModTime: aStats.ModTime(),
	}

	// Parse the archive into a EntryProgress struct
	e := newEntry()
	e.Archive = a

	// Iterate through the files in the archive, try to parse metadata
	err = a.Walk(func(f archiver.File) error {
		if !f.IsDir() && !strings.HasPrefix(f.Name(), ".") {
			// Validate page has correct image type
			t, err := GetImageType(filepath.Ext(f.Name()))
			if err != nil {
				return err
			}

			// Allow one cover page
			if strings.HasPrefix(strings.ToLower(f.Name()), "cover") {
				e.Archive.Cover.Fp = f.Name()
				e.Archive.Cover.ImageType = t
			} else {
				// If image isn't a cover page then parse it as a proper page
				page := &Page{
					ImageType: t,
				}

				// If zip we need the file header to get the absolute filepath
				// but with rar calling f.Name() already gives it to us
				switch a.Type {
				case ZipArchive:
					page.Path = f.Header.(zip.FileHeader).Name
				case RarArchive:
					page.Path = f.Name()
				}

				//
				e.Pages = append(e.Pages, page)
			}
		}

		return nil
	})

	// Add the rest of the metadata
	e.Archive.Title = fse.Filename(fp)
	e.Metadata.Title = e.Archive.Title

	// First item (which is a page) is guaranteed to be first in order
	// of pages since all cbz/cbr archive are ordered beforehand, so
	// we default to using the first page as cover if one could not be set
	if e.Archive.Cover.Fp == "" {
		e.Archive.Cover.Fp = e.Pages[0].Path
		e.Archive.Cover.ImageType = e.Pages[0].ImageType
	}

	return e, nil
}

func ParseSeriesFolder(dir string) (*ParsedSeries, ErrorSlice) {
	series := &ParsedSeries{}
	entries := make([]*ParsedEntry, 0)
	errors := NewErrorSlice()
	errorQueue := make(chan error, 1)
	entriesQueue := make(chan *ParsedEntry, 1)

	// Set the series title
	series.Title = fse.Filename(dir)

	// Set the entries
	order := 1
	wg := sync.WaitGroup{}

	go func() {
		for e := range errorQueue {
			errors = append(errors, e)
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
			if _, err := GetArchiveType(path); err == nil {
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
