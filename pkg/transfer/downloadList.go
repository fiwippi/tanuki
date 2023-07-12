package transfer

import (
	"sync"

	"github.com/fiwippi/tanuki/pkg/mangadex"
)

// DownloadList keeps track of the queued and active
// downloads in the download manager
type DownloadList struct {
	m *sync.Mutex          // Mutex
	l []*mangadex.Download // Downloads list
}

func NewDownloadList() *DownloadList {
	return &DownloadList{
		l: make([]*mangadex.Download, 0),
		m: &sync.Mutex{},
	}
}

// Add adds a download to the list
func (dl *DownloadList) Add(d *mangadex.Download) {
	dl.m.Lock()
	defer dl.m.Unlock()

	dl.l = append(dl.l, d)
}

// Remove removes a download from the list
func (dl *DownloadList) Remove(d *mangadex.Download) {
	dl.m.Lock()
	defer dl.m.Unlock()

	i := 0
	for _, v := range dl.l {
		if v != d {
			dl.l[i] = v
			i++
		}
	}

	dl.l = dl.l[:i]
}

// Has returns whether the list has a given download
func (dl *DownloadList) Has(d *mangadex.Download) bool {
	dl.m.Lock()
	defer dl.m.Unlock()

	for _, v := range dl.l {
		if v == d {
			return true
		}
	}
	return false
}

// List returns a slice copy of the list
func (dl *DownloadList) List() []*mangadex.Download {
	dl.m.Lock()
	defer dl.m.Unlock()

	p := downloadsPool.Get()
	p = append(p, dl.l...)
	return p
}

// Cancel cancels all currently running downloads
// and removes them from the list
func (dl *DownloadList) Cancel() {
	dl.m.Lock()
	defer dl.m.Unlock()

	for i := range dl.l {
		dl.l[i].Cancel()
		dl.l[i] = nil
	}
	dl.l = dl.l[:0]
}
