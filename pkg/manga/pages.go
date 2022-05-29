package manga

import (
	"database/sql/driver"
	"encoding/json"

	"github.com/fiwippi/tanuki/internal/platform/dbutil"
)

type Pages []string

func (p Pages) Total() int {
	return len(p)
}

func (p Pages) Value() (driver.Value, error) {
	return json.Marshal(p)
}

func (p *Pages) Scan(src interface{}) error {
	return dbutil.ScanJSON(src, p)
}
