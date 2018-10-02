// +build amd64

package safetyfast

import (
	"sync"
	"sync/atomic"

	rtm "github.com/0xmjk/go-tsx-rtm"
)

// RTMContext holds the shared state for the fallback path if the RTM
// transaction fails
type RTMContext struct {
	fallback       int32
	lock           sync.Locker
	capacityaborts uint64
}

// NewRTMContexDefault creates an AtomicContext that tries to use Intel RTM,
// but can fallback to using the native sync.Mutex.
func NewRTMContexDefault() *RTMContext {
	return &RTMContext{
		lock: new(sync.Mutex),
	}
}

// NewRTMContex creates an AtomicContext that tries to use Intel RTM,
// but can fallback to using the provided sync.Locker.
func NewRTMContex(l sync.Locker) *RTMContext {
	return &RTMContext{
		lock: l,
	}
}

// CapacityAborts returns the number of aborts that were due to cache capacity.
// If you see lots of capacity aborts, this means the commiter function
// if touching too many memory locations and is unlikely to be reaping any gains
// from using an RTMContext.
func (r *RTMContext) CapacityAborts() uint64 {
	return r.capacityaborts
}

// Atomic executes the commiter in an atomic fasion.
//go:nosplit
func (r *RTMContext) Atomic(commiter func()) {
retry:
	if status := rtm.TxBegin(); status == rtm.TxBeginStarted {
		// Since the system lock does not have any way to check it's status
		if r.fallback != 0 {
			rtm.TxAbort()
		}
		commiter()
		rtm.TxEnd()
	} else {
		if status&(rtm.TxAbortRetry /*|rtm.TxAbortConflict*/) != 0 {
			// safetyfast.Pause()
			goto retry
		}
		// The following lines should be commented out to achieve top performance.
		if status&rtm.TxAbortCapacity != 0 {
			atomic.AddUint64(&r.capacityaborts, 1)
		}

		r.lock.Lock()
		SetAndFence32(&r.fallback)
		commiter()
		r.fallback = 0
		r.lock.Unlock()

	}
}
