package mangadex

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/fiwippi/tanuki/internal/archive"
	"github.com/fiwippi/tanuki/internal/image"
	"github.com/fiwippi/tanuki/internal/pretty"
	"github.com/fiwippi/tanuki/internal/sync"
)

type Chapter struct {
	ID         string            `json:"id"`
	Attributes ChapterAttributes `json:"attributes"`
}

type ChapterAttributes struct {
	Chapter     string `json:"chapter"`
	Pages       int    `json:"pages"`
	Title       string `json:"title"`
	Volume      string `json:"volume"`
	PublishedAt string `json:"publishAt"`
}

type Chapters []*Chapter

func (c *Client) CreateChapterArchive(ch *Chapter, data *HomeURLData, forChapter func(i int) error, cont *sync.Controller) (*archive.ZipFile, error) {
	z, err := archive.NewZipFile()
	if err != nil {
		return nil, err
	}
	defer z.CloseWriter()

	lastModified, err := time.Parse(time.RFC3339, ch.Attributes.PublishedAt)
	if err != nil {
		return nil, err
	}

	padding := len(strconv.Itoa(len(data.Chapter.Data)))

	for i, p := range data.Chapter.Data {
		cont.WaitIfPaused()
		err := forChapter(i + 1)
		if err != nil {
			return nil, err
		}

		// Format the Mangadex@Home url
		url := fmt.Sprintf("%s/data/%s/%s", data.BaseUrl, data.Chapter.Hash, p)

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
		if resp.StatusCode >= 400 && resp.StatusCode <= 500 {
			return nil, fmt.Errorf("error retrieving page from chapter: %d", resp.StatusCode)
		}

		// Save the image to the archive
		// 1. Create the filename
		// 2. Create the file info
		// 3. Save the file
		imgType, err := image.InferType(p)
		if err != nil {
			return nil, err
		}
		fileName := fmt.Sprintf("%s.%s", pretty.Padded(i+1, padding), imgType)

		fi := NewPageInfo(fileName, resp.ContentLength, lastModified)

		err = z.Write(fileName, fi, resp.Body)
		if err != nil {
			return nil, err
		}
	}

	return z, nil
}
