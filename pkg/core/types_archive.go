package core

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/mholt/archiver/v3"
)

type ArchiveType int

const (
	ZipArchive ArchiveType = iota
	RarArchive
)

func (at ArchiveType) MimeType() string {
	return [...]string{"application/zip", "application/x-rar"}[at]
}

func (at ArchiveType) String() string {
	return [...]string{"zip", "rar"}[at]
}

func (at ArchiveType) Walker() archiver.Walker {
	switch at {
	case ZipArchive:
		return DefaultZipArchiver()
	case RarArchive:
		return DefaultRarArchiver()
	}

	panic(fmt.Sprintf("invalid archive type: '%d'", at))
}

func GetArchiveType(fp string) (ArchiveType, error){
	ext := filepath.Ext(fp)
	ext = strings.TrimPrefix(ext, ".")
	ext = strings.ToLower(ext)

	switch ext {
	case "zip", "cbz":
		return ZipArchive, nil
	case "rar", "cbr":
		return RarArchive, nil
	}

	return -1, fmt.Errorf("invalid archive type: '%s'", ext)
}

