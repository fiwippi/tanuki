package db

import (
	"github.com/fiwippi/tanuki/pkg/core"
	"os"
	"testing"
)

func TestDB_SaveSeries(t *testing.T) {
	s, m, err := core.ParseSeriesFolder(os.Getenv("SERIES_PATH"))
	if err != nil {
		t.Error(err)
	}

	err = testDb.SaveSeries(s, m)
	if err != nil {
		t.Error(err)
	}
}

func TestDB_CreateThumbnails(t *testing.T) {
	dir := os.Getenv("SERIES_PATH")
	s, m, err := core.ParseSeriesFolder(dir)
	if err != nil {
		t.Error(err)
	}

	err = testDb.SaveSeries(s, m)
	if err != nil {
		t.Error(err)
	}

	err = testDb.GenerateThumbnails(false)
	if err != nil {
		t.Error(err)
	}
}

func TestDB_GenerateSeriesList(t *testing.T) {
	dir := os.Getenv("SERIES_PATH")
	s, m, err := core.ParseSeriesFolder(dir)
	if err != nil {
		t.Error(err)
	}

	err = testDb.SaveSeries(s, m)
	if err != nil {
		t.Error(err)
	}
}