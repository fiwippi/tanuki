package manga

import (
	"database/sql/driver"
	"encoding/json"

	"github.com/fiwippi/tanuki/internal/sqlutil"
)

var exists = struct{}{}

type Tags struct {
	m map[string]struct{}
}

func NewTags() *Tags {
	return &Tags{m: map[string]struct{}{}}
}

func (t *Tags) Add(values ...string) {
	for _, v := range values {
		t.m[v] = exists
	}
}

func (t *Tags) Combine(tags *Tags) {
	for tag := range tags.m {
		t.Add(tag)
	}
}

func (t *Tags) Has(value string) bool {
	_, c := t.m[value]
	return c
}

func (t *Tags) Empty() bool {
	return len(t.m) == 0
}

func (t Tags) List() []string {
	list := make([]string, 0, len(t.m))
	for item := range t.m {
		list = append(list, item)
	}
	return list
}

// Representation

func (t Tags) MarshalJSON() ([]byte, error) {
	return json.Marshal(t.List())
}

func (t *Tags) UnmarshalJSON(b []byte) error {
	if b == nil {
		// No set exists
		return nil
	}

	// A set exists which may be empty
	var list []string
	err := json.Unmarshal(b, &list)
	if err != nil {
		return err
	}

	*t = *NewTags()
	for _, v := range list {
		t.Add(v)
	}
	return nil
}

func (t Tags) Value() (driver.Value, error) {
	return json.Marshal(t)
}

func (t *Tags) Scan(src interface{}) error {
	return sqlutil.ScanJSON(src, t)
}
