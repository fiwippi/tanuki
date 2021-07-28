package manga

import (
	"encoding/json"
	"os"

	"github.com/fiwippi/tanuki/internal/fse"
	"github.com/fiwippi/tanuki/internal/image"
)

// Cover is a filepath to an image file which
// represents the cover of the manga entry
type Cover struct {
	Fp        string     `json:"file"`       // Filepath
	ImageType image.Type `json:"image_type"` // What type is the image, e.g. PNG, JPEG, ...
}

func (c *Cover) ExistsOnFS() bool {
	return fse.Exists(c.Fp)
}

func (c *Cover) ReadFile() ([]byte, error) {
	d, err := os.ReadFile(c.Fp)
	if err != nil {
		return nil, err
	}
	return d, nil
}

func (c *Cover) ThumbnailFile() ([]byte, error) {
	f, err := os.Open(c.Fp)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	img, err := c.ImageType.Decode(f)
	if err != nil {
		return nil, err
	}

	return image.EncodeThumbnail(img)
}

func UnmarshalCover(data []byte) *Cover {
	if data == nil {
		return nil
	}

	var s Cover
	err := json.Unmarshal(data, &s)
	if err != nil {
		panic(err)
	}
	return &s
}
