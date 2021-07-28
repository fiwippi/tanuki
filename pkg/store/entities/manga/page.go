package manga

import (
	"encoding/json"
	"github.com/fiwippi/tanuki/internal/image"
)

type Page struct {
	IsCover   bool       `json:"cover"`      // Is the page a cover page // TODO remove this
	ImageType image.Type `json:"image_type"` // Image encoding e.g. ".png"
	Path      string     `json:"path"`       // Path to the file in the archive
}

func UnmarshalPage(data []byte) *Page {
	var s Page
	err := json.Unmarshal(data, &s)
	if err != nil {
		panic(err)
	}
	return &s
}
