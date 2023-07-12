package image

import (
	"bytes"
	"errors"
	"fmt"
	"image"
	"image/gif"
	"image/jpeg"
	"image/png"
	"io"

	"github.com/nfnt/resize"
	"golang.org/x/image/bmp"
	"golang.org/x/image/tiff"
	"golang.org/x/image/webp"
)

func (t Type) Decode(r io.Reader) (image.Image, error) {
	switch t {
	case PNG, JPEG, GIF, WEBP, TIFF, BMP:
		// Try generic decoding first because even if the
		// file extension is .png for example, the actual
		// image might not be
		img, _, err := image.Decode(r)
		if err == nil {
			return img, err
		}

		// Generic decoding doesn't always work so if the
		// image format is unrecognised we know we try again
		// with specific encoding
		switch t {
		case PNG:
			return png.Decode(r)
		case JPEG:
			return jpeg.Decode(r)
		case GIF:
			return gif.Decode(r)
		case WEBP:
			return webp.Decode(r)
		case TIFF:
			return tiff.Decode(r)
		case BMP:
			return bmp.Decode(r)
		}
	case Invalid:
		return nil, errors.New("cannot decode invalid image")
	}

	panic(fmt.Sprintf("invalid image type: '%d'", t))
}

func (t Type) EncodeThumbnail(img image.Image, width, height uint) ([]byte, error) {
	// Create thumbnail
	thumb := resize.Thumbnail(width, height, img, resize.Bicubic)

	// Encode thumbnail and write to buffer
	buf := bytes.NewBuffer(nil)
	err := jpeg.Encode(buf, thumb, &jpeg.Options{Quality: 70})
	if err != nil {
		return nil, err
	}

	// Return data from buffer
	return buf.Bytes(), nil
}
