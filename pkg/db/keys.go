package db

var (
	// Series/Entries related
	keyTags            = []byte("tags")
	keyTitle           = []byte("title")
	keyCover           = []byte("cover")
	keyPages           = []byte("pages")
	keyOrder           = []byte("order")
	keyArchive         = []byte("archive")
	keyCatalog         = []byte("catalog")
	keyModTime         = []byte("modtime")
	keyThumbnail       = []byte("thumbnail")
	keyEntriesData     = []byte("entries-data")
	keyEntriesMetadata = []byte("entries-metadata")

	// User related
	keyType     = []byte("type")
	keyUsername = []byte("username")
	keyPassword = []byte("password")
	keyProgress = []byte("progress")
)
