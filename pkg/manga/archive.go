package manga

import (
	"context"
	"time"

	"github.com/mholt/archiver/v4"

	"github.com/fiwippi/tanuki/internal/platform/archive"
	"github.com/fiwippi/tanuki/internal/platform/fse"
)

type Archive struct {
	Title   string       `json:"title"`    // Filename without the extension
	Path    string       `json:"path"`     // Filepath on the file system
	Type    archive.Type `json:"type"`     // File format i.e. ZIP/RAR
	ModTime time.Time    `json:"mod_time"` // Modified time of the file
}

func (a *Archive) Exists() bool {
	return fse.Exists(a.Path)
}

func (a *Archive) Walk(ctx context.Context, fh archiver.FileHandler) error {
	return a.Type.Walk(ctx, a.Path, fh)
}
