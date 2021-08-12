package sync

import "sync"

type Controller struct {
	paused bool
	cond   *sync.Cond
}

func NewController() *Controller {
	m := &sync.Mutex{}
	cond := sync.NewCond(m)

	return &Controller{
		paused: false,
		cond:   cond,
	}
}

func (c *Controller) Paused() bool {
	return c.paused
}

func (c *Controller) Pause() {
	c.cond.L.Lock()
	c.paused = true
	c.cond.L.Unlock()
}

func (c *Controller) Resume() {
	c.cond.L.Lock()
	c.paused = false
	c.cond.L.Unlock()
	c.cond.Broadcast()
}

func (c *Controller) WaitIfPaused() {
	c.cond.L.Lock()
	if c.paused {
		c.cond.Wait()
	}
	c.cond.L.Unlock()
}
