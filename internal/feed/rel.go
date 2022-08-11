package feed

type Relation string

const (
	RelSelf        Relation = "self"
	RelStart       Relation = "start"
	RelSearch      Relation = "search"
	RelCover       Relation = "http://opds-spec.org/image"
	RelThumbnail   Relation = "http://opds-spec.org/image/thumbnail"
	RelAcquisition Relation = "http://opds-spec.org/acquisition"
	RelPageStream  Relation = "http://vaemendis.net/opds-pse/stream"
)
