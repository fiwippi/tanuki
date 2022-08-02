package feed

type Type string

const (
	Navigation  Type = "application/atom+xml;profile=opds-catalog;kind=navigation"
	Acquisition Type = "application/atom+xml;profile=opds-catalog;kind=acquisition"
)
