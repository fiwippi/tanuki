package mangadex

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/fiwippi/tanuki/internal/archive"
	"github.com/fiwippi/tanuki/internal/image"
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
	fileName := fmt.Sprintf("%s.%s", padInt(i+1, padding), imgType)
	err = z.Write(fileName, resp.Body)
	if err != nil {
		return err
	}
	return nil
}

func padInt(num int, p int) string {
	var s = fmt.Sprintf("%d", num)
	var zeroNum = p - len(s)

	var sb strings.Builder
	for i := 0; i < zeroNum; i++ {
		sb.WriteRune('0')
	}
	sb.WriteString(s)

	return sb.String()
}
