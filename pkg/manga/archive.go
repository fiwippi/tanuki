package manga

import (
	"context"
	"database/sql/driver"
	"encoding/json"
	"io"

	"github.com/mholt/archiver/v4"

	"github.com/fiwippi/tanuki/internal/platform/archive"
	"github.com/fiwippi/tanuki/internal/platform/dbutil"
	"github.com/fiwippi/tanuki/internal/platform/fse"
)

type Archive struct {
	Title string       `json:"title"` // Filename without the extension
	Path  string       `json:"path"`  // Filepath on the file system
	Type  archive.Type `json:"type"`  // File format i.e. ZIP/RAR
}

func (a *Archive) Exists() bool {
	return fse.Exists(a.Path)
}

func (a *Archive) Walk(ctx context.Context, fh archiver.FileHandler) error {
	return a.Type.Walk(ctx, a.Path, fh)
}

func (a *Archive) ReaderForFile(fp string) (io.Reader, int64, error) {
	r, size, err := a.Type.Extract(context.Background(), a.Path, fp)
	return r, size, err
}

func (a Archive) Value() (driver.Value, error) {
	return json.Marshal(a)
}

func (a *Archive) Scan(src interface{}) error {
	return dbutil.ScanJSON(src, a)
}
