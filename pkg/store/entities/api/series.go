package api

import (
	"encoding/json"
	"github.com/fiwippi/tanuki/internal/date"
)

type Series struct {
	Order        int        `json:"order"`
	Hash         string     `json:"hash"`
	Title        string     `json:"title"`
	Entries      int        `json:"entries"`
	Tags         []string   `json:"tags"`
	Author       string     `json:"author"`
	DateReleased *date.Date `json:"date_released"`
}

type Catalog []*Series

func UnmarshalCatalog(data []byte) Catalog {
	var s Catalog
	err := json.Unmarshal(data, &s)
	if err != nil {
		panic(err)
	}
	return s
}
