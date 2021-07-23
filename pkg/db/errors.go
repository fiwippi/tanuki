package db

import "errors"

var (
	ErrUserExists      = errors.New("user exists")
	ErrUserNotExist    = errors.New("user does not exist")
	ErrNotEnoughAdmins = errors.New("unsuccessful since there would be no admins")

	ErrSeriesNotExist          = errors.New("series does not exist")
	ErrProgressNotExist        = errors.New("catalog progress does not exist")
	ErrPageNotExist            = errors.New("page does not exist")
	ErrCatalogNotExist         = errors.New("catalog does not exist")
	ErrCatalogEntryNotExist    = errors.New("catalog entry does not exist")
	ErrEntryNotExist           = errors.New("entry does not exist")
	ErrEntriesMetadataNotExist = errors.New("entries metadata do not exist")
	ErrEntryMetadataNotExist   = errors.New("entry metadata do not exist")
	ErrCoverEmpty              = errors.New("cover file has no content")
)
