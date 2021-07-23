package core

import (
	"fmt"
	"image"
	"os"

	"github.com/fiwippi/tanuki/internal/fse"
)

// Cover is a filepath to an image file which
// represents the cover of the manga entry
type Cover struct {
	Fp        string    `json:"file"` // Filepath
	ImageType ImageType `json:"image_type"`
}

func (c *Cover) String() string {
	return fmt.Sprintf("CoverImage::Filepath: %s, ImageType: %s", c.Fp, c.ImageType)
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

func (c *Cover) Image() (image.Image, error) {
	f, err := os.Open(c.Fp)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	img, err := c.ImageType.Decode(f)
	if err != nil {
		return nil, err
	}

	return img, nil
}

func (c *Cover) Thumbnail() ([]byte, error) {
	img, err := c.Image()
	if err != nil {
		return nil, err
	}

	return EncodeJPEG(thumbnail(img))
}
