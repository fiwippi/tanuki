package image

import (
	"bytes"
	"image"
	"image/jpeg"

	"github.com/nfnt/resize"
)

var jpegOption = &jpeg.Options{Quality: 70}

func EncodeThumbnail(img image.Image) ([]byte, error) {
	thumb := resize.Thumbnail(300, 300, img, resize.Bicubic)
	buf := bytes.NewBuffer(nil)
	err := jpeg.Encode(buf, thumb, jpegOption)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
