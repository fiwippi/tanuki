package tanuki

import (
	"os"
	"path/filepath"
	"sync"

	"github.com/rs/zerolog/log"

	"github.com/fiwippi/tanuki/internal/fse"
	"github.com/fiwippi/tanuki/pkg/auth"
	"github.com/fiwippi/tanuki/pkg/core"
)

func ScanLibrary() error {
	wg := sync.WaitGroup{}

	items, err := os.ReadDir(conf.Paths.Library)
	if err != nil {
		return err
	}
	//fmt.Println("ITEMS")
	//for _, v := range items {
	//	fmt.Println("-", v.Name())
	//}

	for _, item := range items {
		if item.IsDir() {
			name := item.Name()
			wg.Add(1)
			go func() {
				fp := filepath.Join(conf.Paths.Library, name)
				s, m, err := core.ParseSeriesFolder(fp)
				if err != nil {
					log.Error().Err(err).Str("filepath", fp).Msg("failed to parse series folder")
					wg.Done()
					return
				}

				err = db.SaveSeries(s, m)
				if err != nil {
					log.Error().Err(err).Str("filepath", fp).Msg("failed to save series")
					wg.Done()
					return
				}

				// If the archive has changed, ensure the progress total pages matches the current total
				for _, e := range m {
					db.EnsureValidSeriesProgress(auth.HashSHA1(s.Title), auth.HashSHA1(e.Title), len(e.Pages))
				}

				// Delete empty .tanuki folder
				tanukiFp := fp + "/.tanuki"
				if fse.Exists(tanukiFp) {
					err := fse.DeleteDirIfEmpty(tanukiFp)
					if err != nil {
						log.Debug().Err(err).Str("fp", tanukiFp).Msg("failed to delete dir")
						wg.Done()
						return
					}
				}

				wg.Done()
			}()
		}
	}
	wg.Wait()

	return db.GenerateThumbnails(false)
}
