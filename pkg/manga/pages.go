package manga

import (
	"database/sql/driver"
	"encoding/json"

	"github.com/fiwippi/tanuki/internal/platform/dbutil"
	"github.com/fiwippi/tanuki/internal/platform/image"
)

type Page struct {
	Path string     `json:"path"`
	Type image.Type `json:"type"`
}

type Pages []Page

func (p Pages) Total() int {
	return len(p)
}

func (p Pages) Value() (driver.Value, error) {
	return json.Marshal(p)
}

func (p *Pages) Scan(src interface{}) error {
	return dbutil.ScanJSON(src, p)
}
