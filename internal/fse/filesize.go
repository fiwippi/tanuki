package fse

import "os"

func FilesizeMiB(path string) float64 {
	fi, err := os.Stat(path)
	if err != nil {
		return 0
	}

	s := float64(fi.Size())
	if s > 0 {
		return s / 1024 / 1024
	}
	return s
}
