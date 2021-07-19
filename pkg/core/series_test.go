package core

import (
	"os"
	"path/filepath"
	"testing"
)

func TestParseSeriesFolder(t *testing.T) {
	// Set the entries
	dir := os.Getenv("SERIES_PATH")
	files, err := os.ReadDir(dir)
	if err != nil {
		t.Error(err)
	}

	for _, f := range files {
		_, _, err := ParseSeriesFolder(filepath.Join(dir, f.Name()))
		if err != nil {
			t.Error(err)
		}
	}
}
