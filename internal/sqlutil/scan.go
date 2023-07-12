package sqlutil

import (
	"encoding/json"
	"errors"
)

func ScanJSON(data interface{}, v any) error {
	b, ok := data.([]byte)
	if !ok {
		return errors.New("invalid []byte assertion")
	}

	return json.Unmarshal(b, v)
}
