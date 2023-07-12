package archive

import (
	"fmt"
	"path/filepath"
	"strings"
)

type Type int

const (
	Zip Type = iota
	Rar

	Invalid = -1
)

func (t Type) String() string {
	return [...]string{"zip", "rar"}[t]
}

func (t Type) MimeType() string {
	return [...]string{"application/zip", "application/x-rar"}[t]
}

func InferType(path string) (Type, error) {
	ext := filepath.Ext(path)
	ext = strings.TrimPrefix(ext, ".")
	ext = strings.ToLower(ext)

	switch ext {
	case "zip", "cbz":
		return Zip, nil
	case "rar", "cbr":
		return Rar, nil
	}

	return Invalid, fmt.Errorf("invalid archive type: '%s'", ext)
}
