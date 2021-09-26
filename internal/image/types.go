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

// Type represents an image format
type Type int

const (
	PNG Type = iota
	JPEG
	GIF
	WEBP
	TIFF
	BMP
)

// MimeType returns the image mimetype, used for sending the image over the web
func (t Type) MimeType() string {
	return [...]string{"image/png", "image/jpeg", "image/gif", "image/webp", "image/tiff", "image/bmp"}[t]
}

func (t Type) String() string {
	return [...]string{"png", "jpg", "gif", "webp", "tiff", "bmp"}[t]
}

// Decode decodes an image given its type
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
	}

	panic(fmt.Sprintf("invalid image type: '%d'", t))
}

// InferType attempts to guess a files image type based
// on its filepath and fails if not possible
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
