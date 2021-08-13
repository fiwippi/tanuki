package downloading

import (
	"sync"

	"github.com/fiwippi/tanuki/pkg/store/entities/api"
)

// DownloadList keeps track of the queued and active
// downloads in the download manager
type DownloadList struct {
	l []*api.Download // Downloads list
	m *sync.Mutex     // Mutex
}

func NewDownloadList() *DownloadList {
	return &DownloadList{
		l: make([]*api.Download, 0),
		m: &sync.Mutex{},
	}
}

// Add adds a download to the list
func (dl *DownloadList) Add(d *api.Download) {
	dl.m.Lock()
	defer dl.m.Unlock()

	dl.l = append(dl.l, d)
}

// Remove removes a download from the list
func (dl *DownloadList) Remove(d *api.Download) {
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
func (dl *DownloadList) Has(d *api.Download) bool {
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
func (dl *DownloadList) List() []*api.Download {
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
		dl.l[i].Status = api.DownloadCancelled
		dl.l[i] = nil
	}
	dl.l = dl.l[:0]
}
