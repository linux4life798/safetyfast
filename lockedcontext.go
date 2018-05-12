package safetyfast

import "sync"

type LockedContext struct {
	lock sync.Locker
}

func NewLockedContext(lock sync.Locker) *LockedContext {
	c := new(LockedContext)
	c.lock = lock
	return c
}

//go:nosplit
func (c *LockedContext) Atomic(commiter func()) {
	c.lock.Lock()
	commiter()
	c.lock.Unlock()
}
