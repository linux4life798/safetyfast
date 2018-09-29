package safetyfast

import "sync"

// LockedContext provides an AtomicContext that utilizes any sync.Locker.
type LockedContext struct {
	lock sync.Locker
}

// NewLockedContext creates a LockedContext that uses lock as the sync method.
func NewLockedContext(lock sync.Locker) *LockedContext {
	c := new(LockedContext)
	c.lock = lock
	return c
}

// Atomic executes commiter atomically with respect to other commiters
// launched from this context.
//go:nosplit
func (c *LockedContext) Atomic(commiter func()) {
	c.lock.Lock()
	commiter()
	c.lock.Unlock()
}
