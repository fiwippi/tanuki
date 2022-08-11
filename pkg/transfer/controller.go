package transfer

import "sync"

// controller implements the ability to tell multiple goroutines
// to wait or resume based on a pause condition. At appropriate times
// goroutines should check if they should pause using the WaitIfPaused()
// method.
type controller struct {
	cond   *sync.Cond
	paused bool
}

// newController returns a new controller to synchronise goroutines.
// The controller starts unpaused by default
func newController() *controller {
	return &controller{
		paused: false,
		cond:   sync.NewCond(&sync.Mutex{}),
	}
}

// Paused returns whether the controller is currently pausing goroutines
func (c *controller) Paused() bool {
	c.cond.L.Lock()
	defer c.cond.L.Unlock()

	return c.paused
}

// Pause sets the controller to a paused state
func (c *controller) Pause() {
	c.cond.L.Lock()
	c.paused = true
	c.cond.L.Unlock()
}

// Resume wakes up all goroutines waiting on the controller
// and continues their execution
func (c *controller) Resume() {
	c.cond.L.Lock()
	c.paused = false
	c.cond.L.Unlock()
	c.cond.Broadcast()
}

// WaitIfPaused causes the goroutine to wait if it is paused,
// until the Resume() function is called to continue execution
func (c *controller) WaitIfPaused() {
	c.cond.L.Lock()
	if c.paused {
		c.cond.Wait()
	}
	c.cond.L.Unlock()
}
