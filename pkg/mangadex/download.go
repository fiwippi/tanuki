package mangadex

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/fiwippi/tanuki/internal/fse"
)

type Download struct {
	MangaTitle string  `json:"manga_title" db:"manga_title"`
	Chapter    Chapter `json:"chapter" db:"chapter"`

	cancelFn    func()
	Status      DownloadStatus `json:"status" db:"status"`
	CurrentPage int            `json:"current_page" db:"current_page"`
	TotalPages  int            `json:"total_pages" db:"total_pages"`
	TimeTaken   int64          `json:"time_taken" db:"time_taken"`
	Subscribe   bool           `json:"subscribe"`
}

func (d Download) String() string {
	return d.filepath()
}

func (d *Download) filepath() string {
	title := fse.Sanitise(d.MangaTitle)
	vol := fse.Sanitise(d.Chapter.VolumeNo)
	chapter := fse.Sanitise(d.Chapter.ChapterNo)

	var fp string
	if vol != "" {
		fp = fmt.Sprintf("%s/Vol. %s/Ch. %s.cbz", title, vol, chapter)
	} else if chapter != "" {
		fp = fmt.Sprintf("%s/Ch. %s.cbz", title, chapter)
	} else {
		fp = fmt.Sprintf("%s/%s.cbz", title, fse.Sanitise(d.Chapter.Title))
	}

	return fp
}

func (d *Download) Run(ctx context.Context, libraryPath string) error {
	ctx, d.cancelFn = context.WithCancel(ctx)
	d.Status = DownloadStarted
	start := time.Now()
	defer func() {
		d.TimeTaken = time.Since(start).Milliseconds()
	}()

	// If the download already exists then finish and exit
	path := fmt.Sprintf("%s/%s", libraryPath, d.filepath())
	if fse.Exists(path) {
		d.Status = DownloadExists
		return nil
	}

	// Otherwise, download the ZIP archive and save it to the disk
	progress := make(chan int)
	go func() {
		for p := range progress {
			d.CurrentPage = p
		}
	}()
	z, err := d.Chapter.downloadZip(ctx, progress)
	if err != nil {
		d.Status = DownloadFailed
		if errors.Is(err, context.Canceled) {
			d.Status = DownloadCancelled
		}
		return err
	}
	defer z.Close()

	parentDir := filepath.Dir(path)
	err = fse.CreateDirs(parentDir)
	if err != nil {
		return err
	}

	if err = os.WriteFile(path, z.Bytes(), 0666); err != nil {
		d.Status = DownloadFailed
		return err
	}

	d.Status = DownloadFinished
	return nil
}

func (d *Download) Cancel() {
	if d.cancelFn != nil {
		d.cancelFn()
	}
	d.Status = DownloadCancelled
}
