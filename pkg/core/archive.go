package core

import (
	"bytes"
	"errors"
	"fmt"
	"image"
	"io"
	"io/ioutil"
	"os"
	"strings"
	"time"

	"github.com/mholt/archiver/v3"
	"github.com/nfnt/resize"

	"github.com/fiwippi/tanuki/internal/fse"
)

type Archive struct {
	Path    string      `json:"path"`  // Path to the archive on the filesystem
	Type    ArchiveType `json:"type"`  // What file format is the archive e.g. zip/rar
	Cover   *Cover      `json:"cover"` // Link to the embedded cover in the archive
	ModTime time.Time   `json:"mod_time"`
}

func (a *Archive) FilenameWithExt() string {
	return fmt.Sprintf("%s.%s", fse.Filename(a.Path), a.Type.String())
}

func (a *Archive) Exists() bool {
	return fse.Exists(a.Path)
}

func (a *Archive) Walk(f func(f archiver.File) error) error {
	return a.Type.Walker().Walk(a.Path,  f)
}

func (a *Archive) File(fp string) (archiver.File, error) {
	// Attempt to find the file
	var file archiver.File
	err := a.Type.Walker().Walk(a.Path, func(f archiver.File) error {
		if strings.HasSuffix(fp, f.Name()) {
			file.FileInfo = f.FileInfo
			file.Header = f.Header
			data, err := ioutil.ReadAll(f)
			if err != nil {
				return err
			}
			file.ReadCloser = archiver.ReadFakeCloser{Reader: bytes.NewReader(data)}
		}
		return nil
	})

	// If there was an error retrieving the file or
	// the file was not found then return an error
	if err != nil {
		return archiver.File{}, err
	} else if file == (archiver.File{}) {
		return archiver.File{}, ErrArchiveFileNotFound
	}
	return file, nil
}

func (a *Archive) FileReader(fp string) (io.Reader, int64, error) {
	// Attempt to find the file
	var r io.Reader
	var size int64
	err := a.Type.Walker().Walk(a.Path, func(f archiver.File) error {
		if strings.HasSuffix(fp, f.Name()) {
			data, err := ioutil.ReadAll(f)
			if err != nil {
				return err
			}
			r = bytes.NewReader(data)
			size = f.Size()
		}
		return nil
	})

	// If there was an error retrieving the file or
	// the file was not found then return an error
	if err != nil {
		return nil, 0, err
	} else if size == 0 {
		return nil, 0, ErrArchiveFileNotFound
	}
	return r, size, nil
}

func (a *Archive) CoverImage() (image.Image, error){
	if a.Cover == nil {
		return nil, errors.New("cover is nil")
	}

	r, _, err := a.FileReader(a.Cover.Fp)
	if err != nil {
		return nil, err
	}

	img, err := a.Cover.ImageType.Decode(r)
	if err != nil {
		return nil, err
	}
	return img, nil
}

func (a *Archive) CoverBytes() ([]byte, error){
	img, err := a.CoverImage()
	if err != nil {
		return nil, err
	}

	return EncodeJPEG(img)
}

func (a *Archive) Thumbnail() ([]byte, error){
	// Get the cover image
	img, err := a.CoverImage()
	if err != nil {
		return nil, err
	}

	// Create thumbnail
	thumb := resize.Thumbnail(300, 300, img, resize.Lanczos2)
	return EncodeJPEG(thumb)
}

func (a *Archive) Filesize() int64 {
	fi, err := os.Stat(a.Path)
	if err != nil {
		return 0
	}
	return fi.Size()
}

func (a *Archive) FilesizeMib() float64 {
	s := a.Filesize()
	if s > 0 {
		return float64(s) / 1024 / 1024
	}
	return 0
}
