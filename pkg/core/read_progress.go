package core

import (
	"errors"
	"fmt"
	"sync"
)

var ErrSeriesNotExist = errors.New("series does not exist")

type ReadProgress struct {
	Current int `json:"current"`
	Total   int `json:"total"`
}

func NewReadProgress(total int) *ReadProgress {
	return &ReadProgress{Current: 0, Total: total}
}

func (rp *ReadProgress) String() string {
	return fmt.Sprintf("%.2f%%", rp.Percent())
}

func (rp *ReadProgress) Set(n int) {
	if n > rp.Total {
		n = rp.Total
	}
	rp.Current = n
}

func (rp *ReadProgress) SetRead() {
	rp.Current = rp.Total
}

func (rp *ReadProgress) SetUnread() {
	rp.Current = 0
}

func (rp *ReadProgress) IsDone() bool {
	return rp.Current == rp.Total
}

func (rp *ReadProgress) Percent() float64 {
	if rp.Current == 0 && rp.Total == 0 {
		return 0
	}

	percent := (float64(rp.Current) / float64(rp.Total)) * 100
	if percent < 0 {
		percent = 0
	} else if percent > 100 {
		percent = 100
	}

	return percent
}

type ProgressTracker struct {
	Tracker map[string]map[string]*ReadProgress `json:"tracker"`
	M       sync.RWMutex                        `json:"m"`
}

func NewProgressTracker() *ProgressTracker {
	return &ProgressTracker{
		Tracker: make(map[string]map[string]*ReadProgress),
		M:       sync.RWMutex{},
	}
}

func (p *ProgressTracker) AddSeries(series string) {
	p.M.Lock()
	defer p.M.Unlock()
	p.Tracker[series] = make(map[string]*ReadProgress)
}

func (p *ProgressTracker) AddEntry(series, entry string, total int) {
	p.M.Lock()
	defer p.M.Unlock()
	if p.Tracker[series] == nil {
		p.Tracker[series] = make(map[string]*ReadProgress)
	}
	p.Tracker[series][entry] = NewReadProgress(total)
}

func (p *ProgressTracker) Delete(series, entry string) {
	p.M.Lock()
	defer p.M.Unlock()
	if p.Tracker[series] != nil {
		delete(p.Tracker[series], entry)
	}
}

func (p *ProgressTracker) ProgressEntry(series, entry string) *ReadProgress {
	p.M.RLock()
	defer p.M.RUnlock()

	if p.Tracker[series] != nil {
		return p.Tracker[series][entry]
	}
	return nil
}

func (p *ProgressTracker) SetSeriesAllRead(series string) error {
	p.M.RLock()
	defer p.M.RUnlock()

	if !p.HasSeries(series) {
		return ErrSeriesNotExist
	}

	for _, v := range p.Tracker[series] {
		v.SetRead()
	}
	return nil
}

func (p *ProgressTracker) SetSeriesAllUnread(series string) error {
	p.M.RLock()
	defer p.M.RUnlock()

	if !p.HasSeries(series) {
		return ErrSeriesNotExist
	}

	for _, v := range p.Tracker[series] {
		v.SetUnread()
	}
	return nil
}

func (p *ProgressTracker) SeriesEntriesNum(series string) (int, error) {
	p.M.RLock()
	defer p.M.RUnlock()

	if !p.HasSeries(series) {
		return -1, ErrSeriesNotExist
	}

	return len(p.Tracker[series]), nil
}

func (p *ProgressTracker) ProgressSeries(series string) *ReadProgress {
	p.M.RLock()
	defer p.M.RUnlock()

	if !p.HasSeries(series) {
		return nil
	}

	progress := &ReadProgress{}
	for _, v := range p.Tracker[series] {
		progress.Current += v.Current
		progress.Total += v.Total
	}
	return progress
}

func (p *ProgressTracker) HasSeries(series string) bool {
	p.M.RLock()
	defer p.M.RUnlock()

	return p.Tracker[series] != nil
}

func (p *ProgressTracker) HasEntry(series, entry string) bool {
	p.M.RLock()
	defer p.M.RUnlock()

	return p.Tracker[series] != nil && p.Tracker[series][entry] != nil
}