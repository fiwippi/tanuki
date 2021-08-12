package mangadex

import (
	"fmt"
	"net/http"
	"time"

	"github.com/fiwippi/tanuki/internal/archive"
	"github.com/fiwippi/tanuki/internal/image"
	"github.com/fiwippi/tanuki/internal/sync"
)

type Chapter struct {
	ID         string            `json:"id"`
	Attributes ChapterAttributes `json:"attributes"`
}

type ChapterAttributes struct {
	Chapter     string   `json:"chapter"`
	Hash        string   `json:"hash"`
	Data        []string `json:"data"`
	Title       string   `json:"title"`
	Volume      string   `json:"volume"`
	PublishedAt string   `json:"publishAt"`
}

type Chapters []*Chapter

func (c *Client) CreateChapterArchive(ch *Chapter, homeUrl string, forChapter func(i int) error, cont *sync.Controller) (*archive.ZipFile, error) {
	z, err := archive.NewZipFile()
	if err != nil {
		return nil, err
	}
	defer z.CloseWriter()

	lastModified, err := time.Parse(time.RFC3339, ch.Attributes.PublishedAt)
	if err != nil {
		return nil, err
	}

	for i, p := range ch.Attributes.Data {
		cont.WaitIfPaused()
		err := forChapter(i + 1)
		if err != nil {
			return nil, err
		}

		// Format the Mangadex@Home url
		url := fmt.Sprintf("%s/data/%s/%s", homeUrl, ch.Attributes.Hash, p)

		// Create the API request
		r, err := http.NewRequest("GET", url, nil)
		if err != nil {
			return nil, err
		}

		// Send the API request
		resp, err := c.sendRequest(r)
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()

		// Save the image to the archive
		// 1. Create the filename
		// 2. Create the file info
		// 3. Save the file
		imgType, err := image.InferType(p)
		if err != nil {
			return nil, err
		}
		fileName := fmt.Sprintf("%d.%s", i+1, imgType)

		fi := NewPageInfo(fileName, resp.ContentLength, lastModified)

		err = z.Write(fileName, fi, resp.Body)
		if err != nil {
			return nil, err
		}
	}

	return z, nil
}
