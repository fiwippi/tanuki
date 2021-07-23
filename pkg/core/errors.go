package core

import (
	"errors"
	"strings"
)

var (
	ErrArchiveFileNotFound = errors.New("file not found in archive")
	ErrEntryNotExist       = errors.New("entry does not exist")
)

type ErrorSlice []error

func NewErrorSlice() ErrorSlice {
	return make([]error, 0)
}

func (es ErrorSlice) Empty() bool {
	return len(es) == 0
}

func (es ErrorSlice) Error() string {
	var sb strings.Builder
	for _, e := range es {
		sb.WriteString(e.Error())
	}
	return sb.String()
}
