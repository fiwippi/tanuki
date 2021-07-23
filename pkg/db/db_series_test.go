package db

import (
	"github.com/fiwippi/tanuki/pkg/core"
	"os"
	"testing"
)

func TestDB_SaveSeries(t *testing.T) {
	s, err := core.ParseSeriesFolder(os.Getenv("SERIES_PATH"))
	if err != nil {
		t.Error(err)
	}

	err = testDb.PopulateCatalog([]*core.ParsedSeries{s})
	if err != nil {
		t.Error(err)
	}
}

func TestDB_CreateThumbnails(t *testing.T) {
	s, err := core.ParseSeriesFolder(os.Getenv("SERIES_PATH"))
	if err != nil {
		t.Error(err)
	}

	err = testDb.PopulateCatalog([]*core.ParsedSeries{s})
	if err != nil {
		t.Error(err)
	}

	errors := testDb.GenerateThumbnails(false)
	if errors != nil {
		t.Error(err)
	}
}
