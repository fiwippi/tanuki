package tanuki

import (
	"github.com/fiwippi/tanuki/internal/fse"
)

// paths contains directory/file paths which 
// tanuki uses in order to navigate to the 
// right file
type paths struct {
	DB      string // Path to the database
	Log     string // Where tanuki should log to
	Library string // Where tanuki stores uploaded/downloaded manga
}

//
func defaultPaths() paths {
	return paths{
		DB:      "./data/tanuki.db",
		Log:     "./data/tanuki.log",
		Library: "./library",
	}
}

// EnsureExist ensures that the required directories for the paths exist
func (p paths) EnsureExist() error {
	if err := fse.EnsureFileDir(p.DB); err != nil {
		return err
	} else if err := fse.EnsureFileDir(p.Log); err != nil {
		return err
	} else if err := fse.EnsureDir(p.Library); err != nil {
		return err
	}

	return nil
}