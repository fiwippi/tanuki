package mangadex

import "fmt"

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
