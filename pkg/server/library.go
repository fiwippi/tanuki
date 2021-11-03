package server

import (
	"os"
	"path/filepath"

	"github.com/rs/zerolog/log"

	"github.com/fiwippi/tanuki/internal/fse"
	"github.com/fiwippi/tanuki/pkg/store/entities/manga"
)

func (s *Server) ScanLibrary() error {
	series := make([]*manga.ParsedSeries, 0)

	items, err := os.ReadDir(s.Conf.Paths.Library)
	if err != nil {
		return err
	}

	for _, item := range items {
		if item.IsDir() {
			fp := filepath.Join(s.Conf.Paths.Library, item.Name())

			// Check for .tanuki folder and delete it if empty
			tanukiFp := fp + "/.tanuki"
			if fse.Exists(tanukiFp) {
				err := fse.DeleteDirIfEmpty(tanukiFp)
				if err != nil {
					log.Debug().Err(err).Str("fp", tanukiFp).Msg("failed to delete dir")
					continue
				}
			}

			// Parse the series
			s, err := manga.ParseSeriesFolder(fp)
			if err != nil {
				log.Error().Err(err).Str("fp", fp).Msg("failed to parse series folder")
				continue
			}
			series = append(series, s)
		}
	}

	err = s.Store.PopulateCatalog(series)
	if err != nil {
		log.Error().Err(err).Msg("failed to populate series db")
	}

	return nil
}
