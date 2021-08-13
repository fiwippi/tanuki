package mangadex

import "fmt"

// Error represents an error returned from Mangadex
// when attempting to retrieve a Mangadex@Home url
type Error struct {
	ID     string `json:"id"`
	Status int    `json:"status"`
	Title  string `json:"title"`
	Detail string `json:"detail"`
}

func (e Error) Error() string {
	return fmt.Sprintf("|%d| %s: %s", e.Status, e.Title, e.Detail)
}

type Errors []Error
