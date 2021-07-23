package core

import (
	"github.com/nfnt/resize"
	"image"
)

func thumbnail(image image.Image) image.Image {
	return resize.Thumbnail(300, 300, image, resize.Bicubic)
}
