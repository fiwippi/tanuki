package core

import (
	"fmt"
	"image"
	"io"
	"io/ioutil"
	"os"

	"github.com/nfnt/resize"

	"github.com/fiwippi/tanuki/internal/fse"
)

// Cover is a filepath to an image file which
// represents the cover of the manga entry
type Cover struct {
	Fp               string    `json:"file"` // Filepath
	ImageType        ImageType `json:"image_type"`
}

func (c *Cover) String() string {
	return fmt.Sprintf("CoverImage::Filepath: %s, ImageType: %s", c.Fp, c.ImageType)
}

func (c *Cover) ExistsOnFS() bool {
	return fse.Exists(c.Fp)
}

func (c *Cover) FromFS() ([]byte, error) {
	d, err := ioutil.ReadFile(c.Fp)
	if err != nil {
		return nil, err
	}
	return d, nil
}

func (c *Cover) Reader() (io.Reader, error) {
	f, err := os.Open(c.Fp)
	if err != nil {
		return nil, err
	}
	return f, nil
}

func (c *Cover) ImageFromFS() (image.Image, error) {
	r, err := c.Reader()
	if err != nil {
		return nil, err
	}

	img, err := c.ImageType.Decode(r)
	if err != nil {
		return nil, err
	}

	return img, nil
}

func (c *Cover) ThumbnailFromFS() ([]byte, error) {
	img, err := c.ImageFromFS()
	if err != nil {
		return nil, err
	}

	thumb := resize.Thumbnail(300, 300, img, resize.Lanczos2)
	return EncodeJPEG(thumb)
}