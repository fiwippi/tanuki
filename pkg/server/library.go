package server

import (
	"github.com/fiwippi/tanuki/internal/fse"
	"github.com/fiwippi/tanuki/pkg/store/entities/manga"
	"github.com/rs/zerolog/log"
	"os"
	"path/filepath"
	"sync"
)

func (s *Server) ScanLibrary() error {
	wg := sync.WaitGroup{}

	series := make([]*manga.ParsedSeries, 0)
	seriesQueue := make(chan *manga.ParsedSeries, 1)

	go func() {
		for s := range seriesQueue {
			series = append(series, s)
			wg.Done()
		}
	}()

	items, err := os.ReadDir(s.Conf.Paths.Library)
	if err != nil {
		return err
	}

	for _, item := range items {
		if item.IsDir() {
			name := item.Name()
			wg.Add(1)
			go func() {
				fp := filepath.Join(s.Conf.Paths.Library, name)

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
				s, err := manga.ParseSeriesFolder(fp)
				if err != nil {
					log.Error().Err(err).Str("fp", fp).Msg("failed to parse series folder")
					wg.Done()
					return
				}
				seriesQueue <- s
			}()
		}
	}
	wg.Wait()

	err = s.Store.PopulateCatalog(series)
	if err != nil {
		log.Error().Err(err).Msg("failed to populate series db")
	}

	return s.Store.GenerateThumbnails(false)
}
