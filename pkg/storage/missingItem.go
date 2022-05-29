package storage

type MissingItem struct {
	Type  string `json:"type"`
	Title string `json:"title"`
	Path  string `json:"path"`
}
