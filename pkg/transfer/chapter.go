package transfer

import "time"

type Chapter interface {
	ID() string
	Title() string
	ScanlationGroup() string
	PublishedAt() time.Time
	Pages() int
	Volume() string
	Chapter() string
}
