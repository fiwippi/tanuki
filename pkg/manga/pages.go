package manga

import (
	"database/sql/driver"
	"encoding/json"

	"github.com/fiwippi/tanuki/internal/image"
	"github.com/fiwippi/tanuki/internal/sqlutil"
)

type Page struct {
	Path string     `json:"path"`
	Type image.Type `json:"type"`
}

type Pages []Page

// Representation

func (p Pages) Value() (driver.Value, error) {
	return json.Marshal(p)
}

func (p *Pages) Scan(src interface{}) error {
	return sqlutil.ScanJSON(src, p)
}
