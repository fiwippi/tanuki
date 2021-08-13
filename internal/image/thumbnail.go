// Package image provides functionality to create thumbnails, decode
// and recognise images
package image

import (
	"bytes"
	"image"
	"image/jpeg"

	"github.com/nfnt/resize"
)

var jpegOption = &jpeg.Options{Quality: 70}

// EncodeThumbnail encodes a given image into a JPEG thumbnail given
// maximum dimensions and then returns its byte contents
func EncodeThumbnail(img image.Image, maxWidth, maxHeight uint) ([]byte, error) {
	thumb := resize.Thumbnail(maxWidth, maxHeight, img, resize.Bicubic)
	buf := bytes.NewBuffer(nil)
	err := jpeg.Encode(buf, thumb, jpegOption)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
