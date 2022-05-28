package image

import (
	"fmt"
	"path/filepath"
	"strings"
)

// TODO remove all these types and only have a function which checks for valid images

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

// TODO switch this func to just validate the extension and dont return the type
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
