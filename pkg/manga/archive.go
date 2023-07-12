package manga

import (
	"context"
	"database/sql/driver"
	"encoding/json"
	"io"

	"github.com/mholt/archiver/v4"

	"github.com/fiwippi/tanuki/internal/archive"
	"github.com/fiwippi/tanuki/internal/sqlutil"
)

type Archive struct {
	Title string       `json:"title"` // Filename without the extension
	Path  string       `json:"path"`  // Filepath on the file system
	Type  archive.Type `json:"type"`  // File format i.e. ZIP/RAR
}

// Extraction

func (a *Archive) Walk(ctx context.Context, handler archiver.FileHandler) error {
	return a.Type.Walk(ctx, a.Path, handler)
}

func (a *Archive) Extract(ctx context.Context, pathInArchive string) (io.Reader, int64, error) {
	return a.Type.Extract(ctx, a.Path, pathInArchive)
}

// Marshaling

func (a Archive) Value() (driver.Value, error) {
	return json.Marshal(a)
}

func (a *Archive) Scan(src interface{}) error {
	return sqlutil.ScanJSON(src, a)
}
