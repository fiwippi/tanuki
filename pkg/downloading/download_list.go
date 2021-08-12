package downloading

import (
	"sync"

	"github.com/fiwippi/tanuki/pkg/store/entities/api"
)

type DownloadList struct {
	l []*api.Download
	m *sync.Mutex
}

func NewDownloadList() *DownloadList {
	return &DownloadList{
		l: make([]*api.Download, 0),
		m: &sync.Mutex{},
	}
}

func (dl *DownloadList) Add(d *api.Download) {
	dl.m.Lock()
	defer dl.m.Unlock()

	dl.l = append(dl.l, d)
}

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

func (dl *DownloadList) List() []*api.Download {
	dl.m.Lock()
	defer dl.m.Unlock()

	p := downloadsPool.Get()
	p = append(p, dl.l...)
	return p
}

func (dl *DownloadList) Cancel() {
	dl.m.Lock()
	defer dl.m.Unlock()

	for i := range dl.l {
		dl.l[i].Status = api.Cancelled
		dl.l[i] = nil
	}
	dl.l = dl.l[:0]
}
