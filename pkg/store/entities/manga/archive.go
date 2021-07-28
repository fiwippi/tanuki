package manga

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/fiwippi/tanuki/internal/archive"
	"github.com/fiwippi/tanuki/internal/image"
	"io"
	"io/ioutil"
	"os"
	"strings"
	"time"

	"github.com/fiwippi/tanuki/internal/fse"
	"github.com/mholt/archiver/v3"
)

var (
	ErrCoverNotExist       = errors.New("cover does not exist")
	ErrArchiveFileNotFound = errors.New("file not found in archive")
)

type Archive struct {
	Title   string       `json:"title"`    // Title of the archive (filename without the extension)
	Path    string       `json:"path"`     // Path to the archive on the filesystem
	Type    archive.Type `json:"type"`     // What file format is the archive e.g. zip/rar
	Cover   *Cover       `json:"cover"`    // Link to the embedded cover in the archive
	ModTime time.Time    `json:"mod_time"` // Modified time of the archive
}

func (a *Archive) FilenameWithExt() string {
	return fmt.Sprintf("%s.%s", a.Title, a.Type.String())
}

func (a *Archive) Exists() bool {
	return fse.Exists(a.Path)
}

func (a *Archive) Walk(f func(f archiver.File) error) error {
	return a.Type.Walker().Walk(a.Path, f)
}

func (a *Archive) ReaderForFile(fp string) (io.Reader, int64, error) {
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

func (a *Archive) CoverFile() ([]byte, error) {
	if a.Cover == nil {
		return nil, ErrCoverNotExist
	}

	r, _, err := a.ReaderForFile(a.Cover.Fp)
	if err != nil {
		return nil, err
	}

	buf, err := ioutil.ReadAll(r)
	if err != nil {
		return nil, err
	}
	return buf, nil
}

func (a *Archive) ThumbnailFile() ([]byte, error) {
	// Get the cover image
	if a.Cover == nil {
		return nil, ErrCoverNotExist
	}

	r, _, err := a.ReaderForFile(a.Cover.Fp)
	if err != nil {
		return nil, err
	}

	img, err := a.Cover.ImageType.Decode(r)
	if err != nil {
		return nil, err
	}

	return image.EncodeThumbnail(img)
}

func (a *Archive) Filesize() float64 {
	fi, err := os.Stat(a.Path)
	if err != nil {
		return 0
	}

	s := fi.Size()
	if s > 0 {
		return float64(s) / 1024 / 1024
	}
	return 0
}

func UnmarshalArchive(data []byte) *Archive {
	var s Archive
	err := json.Unmarshal(data, &s)
	if err != nil {
		panic(err)
	}
	return &s
}
