package image

import (
	"bytes"
	"github.com/nfnt/resize"
	"image"
	"image/jpeg"
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
