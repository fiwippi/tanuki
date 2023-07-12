package image

import (
	"fmt"
	"path/filepath"
	"strings"
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

	Invalid Type = -1
)

func (t Type) String() string {
	if t == -1 {
		return "invalid"
	}
	return [...]string{"png", "jpg", "gif", "webp", "tiff", "bmp"}[t]
}

func (t Type) MimeType() string {
	if t == -1 {
		return ""
	}
	return [...]string{"image/png", "image/jpeg", "image/gif", "image/webp", "image/tiff", "image/bmp"}[t]
}

func InferType(path string) (Type, error) {
	ext := filepath.Ext(path)
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
