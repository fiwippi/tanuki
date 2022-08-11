package manga

import "github.com/fiwippi/tanuki/internal/platform/dbutil"

type Subscription struct {
	SID                 string            `json:"sid" db:"sid"`
	Title               string            `json:"title" db:"title"`
	MdexUUID            dbutil.NullString `json:"mangadex_uuid" db:"mangadex_uuid"`
	MdexLastPublishedAt dbutil.Time       `json:"mangadex_last_published_at" db:"mangadex_last_published_at"`
}
