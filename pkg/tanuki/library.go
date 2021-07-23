package tanuki

import (
	"os"
	"path/filepath"
	"sync"

	"github.com/rs/zerolog/log"

	"github.com/fiwippi/tanuki/internal/fse"
	"github.com/fiwippi/tanuki/pkg/core"
)

func ScanLibrary() error {
	wg := sync.WaitGroup{}

	series := make([]*core.ParsedSeries, 0)
	seriesQueue := make(chan *core.ParsedSeries, 1)

	go func() {
		for s := range seriesQueue {
			series = append(series, s)
			wg.Done()
		}
	}()

	items, err := os.ReadDir(conf.Paths.Library)
	if err != nil {
		return err
	}

	for _, item := range items {
		if item.IsDir() {
			name := item.Name()
			wg.Add(1)
			go func() {
				fp := filepath.Join(conf.Paths.Library, name)

				// Check for .tanuki folder and delete it if empty
				tanukiFp := fp + "/.tanuki"
				if fse.Exists(tanukiFp) {
					err := fse.DeleteDirIfEmpty(tanukiFp)
					if err != nil {
						log.Debug().Err(err).Str("fp", tanukiFp).Msg("failed to delete dir")
						wg.Done()
						return
					}
				}

				// Parse the series
				s, errors := core.ParseSeriesFolder(fp)
				if !errors.Empty() {
					log.Error().Err(errors).Str("fp", fp).Msg("failed to parse series folder")
					wg.Done()
					return
				}
				seriesQueue <- s
			}()
		}
	}
	wg.Wait()

	errors := db.PopulateCatalog(series)
	if !errors.Empty() {
		log.Error().Err(errors).Msg("failed to populate series db")
	}

	return db.GenerateThumbnails(false)
}
