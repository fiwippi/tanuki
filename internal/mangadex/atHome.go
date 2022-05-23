package mangadex

import (
	"fmt"
	"strconv"

	"github.com/fiwippi/tanuki/internal/platform/archive"
	"github.com/fiwippi/tanuki/internal/platform/image"
	"github.com/fiwippi/tanuki/internal/platform/pretty"
)

type atHomeURLData struct {
	result
	BaseUrl string `json:"baseUrl"`
	Chapter struct {
		Hash string   `json:"hash"`
		Data []string `json:"data"`
	} `json:"chapter"`
}

func (h atHomeURLData) Invalid() bool {
	a := h.Chapter.Hash == ""
	b := len(h.Chapter.Data) == 0
	return a || b
}

func (h atHomeURLData) WritePage(i int, p string, z *archive.ZipFile) error {
	resp, err := c.Get(fmt.Sprintf("%s/data/%s/%s", h.BaseUrl, h.Chapter.Hash, p))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 && resp.StatusCode <= 500 {
		return fmt.Errorf("error retrieving page from chapter: %d", resp.StatusCode)
	}

	imgType, err := image.InferType(p)
	if err != nil {
		return err
	}

	padding := len(strconv.Itoa(len(h.Chapter.Data)))
	fileName := fmt.Sprintf("%s.%s", pretty.Padded(i+1, padding), imgType)
	err = z.Write(fileName, resp.Body)
	if err != nil {
		return err
	}
	return nil
}
