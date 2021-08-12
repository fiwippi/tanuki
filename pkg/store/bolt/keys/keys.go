package keys

type Key []byte

var (
	// Related to the Catalog/Series/Entry buckets

	Tags            = Key("tags")
	Title           = Key("title")
	Cover           = Key("cover")
	Pages           = Key("pages")
	Order           = Key("order")
	Archive         = Key("archive")
	Catalog         = Key("catalog")
	ModTime         = Key("modtime")
	Metadata        = Key("metadata")
	Thumbnail       = Key("thumbnail")
	EntriesData     = Key("entries-data")
	EntriesMetadata = Key("entries-metadata")

	// Related to the User bucket

	Type     = Key("type")
	Username = Key("username")
	Password = Key("password")
	Progress = Key("progress")

	// Related to the root buckets

	Users     = Key("users")
	Downloads = Key("downloads")
	// Catalog key should be here but is specified earlier

)
