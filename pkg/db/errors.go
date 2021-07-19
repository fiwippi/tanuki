package db

import "errors"

var (
	ErrUserExists = errors.New("user exists")
	ErrUserNotExist = errors.New("user does not exist")
	ErrNotEnoughAdmins = errors.New("unsuccessful since there would be no admins")

	ErrSeriesNotExist = errors.New("series does not exist")
	ErrProgressCount = errors.New("number of entries in series and series progress do not match up")
	ErrMangaNotExist = errors.New("series does not exist")
	ErrCoverEmpty = errors.New("cover file has no content")
)

