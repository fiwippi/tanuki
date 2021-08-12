package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/fiwippi/tanuki/internal/fse"
	"github.com/fiwippi/tanuki/internal/pretty"
	"github.com/fiwippi/tanuki/pkg/mangadex"
)

var ErrDownloadCancelled = errors.New("download cancelled")

type Download struct {
	Manga   string            `json:"manga"`
	Chapter *mangadex.Chapter `json:"chapter"`

	Status      DownloadStatus `json:"status"`
	CurrentPage int            `json:"current_page"`
	TotalPages  int            `json:"total_pages"`

	StartTime time.Time `json:"-"`
	TimeTaken string    `json:"time_taken"`
}

// Constructor

func NewDownload(manga string, ch *mangadex.Chapter) *Download {
	return &Download{
		Manga:       manga,
		Chapter:     ch,
		Status:      Queued,
		CurrentPage: 0,
		TotalPages:  len(ch.Attributes.Data),
	}
}

// Time now

func (d *Download) Time() {
	d.TimeTaken = pretty.Duration(time.Now().Sub(d.StartTime))
}

// Download state

func (d *Download) Start() {
	d.Status = Started
	d.TimeTaken = ""
	d.StartTime = time.Now()
}

func (d *Download) Finish() {
	d.Status = Finished
	d.Time()
}

func (d *Download) FinishExists() {
	d.Status = Exists
	d.CurrentPage = d.TotalPages
	d.Time()
}

func (d *Download) FinishFailed() {
	d.Status = Failed
	d.Time()
}

func (d *Download) FinishCancelled() {
	d.Status = Cancelled
	d.Time()
}

func (d *Download) IsFinished() bool {
	return d.Status.Finished()
}

// String processing

func (d *Download) Key() string {
	return fmt.Sprintf("%s-%s-%s", d.Manga, d.chapterNum(), d.volumeNum())
}

func (d *Download) Filepath() string {
	title := fse.Sanitise(d.Manga)
	vol := fse.Sanitise(d.volumeNum())
	chapter := fse.Sanitise(d.chapterNum())

	var fp string
	if len(vol) > 0 {
		fp = fmt.Sprintf("%s/Vol. %s/Ch. %s.cbz", title, vol, chapter)
	} else {
		fp = fmt.Sprintf("%s/Ch. %s.cbz", title, chapter)
	}

	return fp
}

func (d *Download) chapterNum() string {
	return d.Chapter.Attributes.Chapter
}

func (d *Download) volumeNum() string {
	return d.Chapter.Attributes.Volume
}

// JSON

func UnmarshalDownload(data []byte) *Download {
	var d Download
	err := json.Unmarshal(data, &d)
	if err != nil {
		panic(err)
	}
	return &d
}
