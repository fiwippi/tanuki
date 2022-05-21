package transfer

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/fiwippi/tanuki/internal/fse"
	"github.com/fiwippi/tanuki/internal/mangadex"
	"github.com/fiwippi/tanuki/internal/pretty"
)

type Download struct {
	MangaTitle string           `json:"manga_title"`
	Chapter    mangadex.Chapter `json:"chapter"`

	cancelFn    func()
	Status      DownloadStatus `json:"status"`
	CurrentPage int            `json:"current_page"`
	TotalPages  int            `json:"total_pages"`
	TimeTaken   string         `json:"time_taken"`
}

func (d Download) String() string {
	return d.Filepath()
}

func newDownload(mangaTitle string, ch mangadex.Chapter) *Download {
	return &Download{
		MangaTitle:  mangaTitle,
		Chapter:     ch,
		Status:      DownloadQueued,
		CurrentPage: 0,
		TotalPages:  ch.Pages,
	}
}

func (d *Download) Filepath() string {
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
		d.TimeTaken = pretty.Duration(time.Now().Sub(start))
	}()

	// If the download already exists then finish and exit
	path := fmt.Sprintf("%s/%s", libraryPath, d.Filepath())
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
	z, err := d.Chapter.DownloadZip(ctx, progress)
	if err != nil {
		d.Status = DownloadFailed
		if errors.Is(err, context.Canceled) {
			d.Status = DownloadCancelled
		}
		return err
	}

	if err = fse.EnsureWriteFile(path, z.Data(), 0666); err != nil {
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
