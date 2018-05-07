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

func NewRTMContexDefault() *RTMContext {
	return &RTMContext{
		lock: new(sync.Mutex),
	}
}

func NewRTMContex(l sync.Locker) *RTMContext {
	return &RTMContext{
		lock: l,
	}
}
func (r *RTMContext) CapacityAborts() uint64 {
	return r.capacityaborts
}

func (r *RTMContext) Commit(commiter func()) {
retry:
	if status := rtm.TxBegin(); status == rtm.TxBeginStarted {
		if r.fallback != 0 {
			rtm.TxAbort()
		}
		commiter()
		rtm.TxEnd()
	} else {
		if status&(rtm.TxAbortRetry|rtm.TxAbortConflict) != 0 {
			// safetyfast.Pause()
			goto retry
		}
		if status&rtm.TxAbortCapacity != 0 {
			atomic.AddUint64(&r.capacityaborts, 1)
		}
		// r.fallback = 1
		r.lock.Lock()
		atomic.SwapInt32(&r.fallback, 1)
		commiter()
		r.fallback = 0
		r.lock.Unlock()
		// r.fallback = 0
	}
}
