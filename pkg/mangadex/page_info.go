package mangadex

import (
	"os"
	"time"
)

type PageInfo struct {
	name    string
	size    int64
	modTime time.Time
}

func NewPageInfo(name string, size int64, modTime time.Time) *PageInfo {
	return &PageInfo{
		name:    name,
		size:    size,
		modTime: modTime,
	}
}

func (pi *PageInfo) Name() string {
	return pi.name
}

func (pi *PageInfo) Size() int64 {
	return pi.size
}

func (pi *PageInfo) Mode() os.FileMode {
	return 0666
}

func (pi *PageInfo) ModTime() time.Time {
	return pi.modTime
}

func (pi *PageInfo) IsDir() bool {
	return false
}

func (pi *PageInfo) Sys() interface{} {
	return nil
}
