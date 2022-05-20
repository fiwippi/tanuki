package mangadex

import (
	"time"

	"golang.org/x/time/rate"
)

var chapterRL = rate.NewLimiter(60/40, 1)

type chapter struct {
	id              string
	title           string
	externalURL     string
	scanlationGroup string
	publishedAt     time.Time
	pages           int
	volume          string
	chapter         string
}

func (c chapter) ID() string {
	return c.id
}

func (c chapter) Title() string {
	return c.title
}

func (c chapter) ScanlationGroup() string {
	return c.scanlationGroup
}

func (c chapter) PublishedAt() time.Time {
	return c.publishedAt
}

func (c chapter) Pages() int {
	return c.pages
}

func (c chapter) Volume() string {
	return c.volume
}

func (c chapter) Chapter() string {
	return c.chapter
}
