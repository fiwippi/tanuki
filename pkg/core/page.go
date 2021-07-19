package core

import "fmt"

type Page struct {
	IsCover   bool      `json:"cover"`      // Is the page a cover page
	ImageType ImageType `json:"image_type"` // Image encoding e.g. ".png"
	Path      string    `json:"path"`       // Path to the file in the archive
}

func (p *Page) String() string {
	return fmt.Sprintf("Page::CoverImage: %v, Format: %s, Path: %s", p.IsCover, p.ImageType, p.Path)
}