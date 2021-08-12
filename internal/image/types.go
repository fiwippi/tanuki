package image

import (
	"fmt"
	"image"
	"image/gif"
	"image/jpeg"
	"image/png"
	"io"
	"path/filepath"
	"strings"

	"golang.org/x/image/bmp"
	"golang.org/x/image/tiff"
	"golang.org/x/image/webp"
)

type Type int

const (
	PNG Type = iota
	JPEG
	GIF
	WEBP
	TIFF
	BMP
)

func (t Type) MimeType() string {
	return [...]string{"image/png", "image/jpeg", "image/gif", "image/webp", "image/tiff", "image/bmp"}[t]
}

func (t Type) String() string {
	return [...]string{"png", "jpg", "gif", "webp", "tiff", "bmp"}[t]
}

func (t Type) Decode(r io.Reader) (image.Image, error) {
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

	panic(fmt.Sprintf("invalid image type: '%d'", t))
}

func InferType(fp string) (Type, error) {
	ext := filepath.Ext(fp)
	ext = strings.TrimPrefix(ext, ".")
	ext = strings.ToLower(ext)

	switch ext {
	case "png":
		return PNG, nil
	case "jpeg", "jpg":
		return JPEG, nil
	case "gif":
		return GIF, nil
	case "webp":
		return WEBP, nil
	case "tiff":
		return TIFF, nil
	case "bmp":
		return BMP, nil
	}

	return -1, fmt.Errorf("invalid image type: '%s'", ext)
}
