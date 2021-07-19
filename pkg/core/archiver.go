package core

import (
	"compress/flate"

	"github.com/mholt/archiver/v3"
)

func DefaultZipArchiver() *archiver.Zip {
	a := archiver.NewZip()
	a.MkdirAll = true
	a.SelectiveCompression = true
	a.CompressionLevel = flate.BestSpeed
	return a
}

func DefaultRarArchiver() *archiver.Rar {
	a := archiver.NewRar()
	a.MkdirAll = true
	return a
}
