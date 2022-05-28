// Package image provides functionality to create thumbnails, decode
// and recognise images
package image

import (
	"bytes"
	"image"
	_ "image/gif"
	"image/jpeg"
	_ "image/jpeg"
	_ "image/png"

	"github.com/nfnt/resize"
	_ "golang.org/x/image/bmp"
	_ "golang.org/x/image/tiff"
	_ "golang.org/x/image/webp"
)

var jpegOption = &jpeg.Options{Quality: 70}

func EncodeThumbnail(data []byte) ([]byte, error) {
	img, _, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		return nil, err
	}

	thumb := resize.Thumbnail(300, 300, img, resize.Bicubic)
	buf := bytes.NewBuffer(nil)
	err = jpeg.Encode(buf, thumb, jpegOption)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
