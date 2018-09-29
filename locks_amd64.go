// +build 386 amd64

package safetyfast

import (
	"runtime"
	"sync/atomic"
)

// Pause executes the PAUSE x86 instruction
func Pause()

// Lock1XCHG8 will atomically write 1 to val while returning the old value.
// The size of val must be 8 bits.
func Lock1XCHG8(val *int8) (old int8)

// Lock1XCHG32 will atomically write 1 to val while returning the old value.
// The size of val must be 32 bits.
func Lock1XCHG32(val *int32) (old int32)

// Lock1XCHG64 will atomically write 1 to val while returning the old value.
// The size of val must be 64 bits.
func Lock1XCHG64(val *int64) (old int64)

// SpinLock implements a basic spin lock that waits forever until the lock
// can be acquired.
// It assumes the lock has been acquired when it successfully writes a 1 to val,
// while the original value was 0.
// This implementation spins on a read only and then uses the XCHG instruction
// in order to claim the lock. The loop makes use of the PAUSE hint instruction.
func SpinLock(val *int32)

// SpinCountLock implements a basic spin lock that tries to acquire the lock
// only attempts times.
// It assumes the lock has been acquired when it successfully writes a 1 to val,
// while the original value was 0.
// This implementation spins on a read only and then uses the XCHG instruction
// in order to claim the lock. The loop makes use of the PAUSE hint instruction.
func SpinCountLock(val, attempts *int32)

// HLETryLock attempts only once to acquire the lock by writing a 1 to val
// using HLE primitives. This function returns a 0 if the lock was acquired.
func HLETryLock(val *int32) int32

// HLESpinLock repeatedly tries to set val to 1 using Intel HLE and XCHG.
// It is implemented as a spin lock that makes use of the PAUSE and
// LOCK XCHG instructions.
// Please note that this function will never return unless the lock is acquired.
// This means a deadlock will occur if the holder of the lock is descheduled
// by the goruntime.
// Please use HLESpinCountLock to limit the spins and manually invoke
// runtime.Gosched periodically, insted.
func HLESpinLock(val *int32)

// HLESpinCountLock tries to set val to 1 at most attempts times using Intel HLE.
// It is implemented as a spin lock that decrements attempts for each attempt.
// The spin operation makes use of the PAUSE and XACQUIRE LOCK XCHG instructions.
// If attempts is 0 when the function returns, the lock was not acquired and
// the spin lock gave up.
// Please note that attempts must be greater 0 when called.
func HLESpinCountLock(val, attempts *int32)

// HLEUnlock writes a 0 to val to indicate the lock has been released
// using HLE primitives
func HLEUnlock(val *int32)

// LockAttempts sets how many times the spin loop is willing to try to
// fetching the lock.
const LockAttempts = int32(200)

// Using Golang's builtin atomics
func SpinLockAtomics(val *int32) {
	for {
		// Spin on simple read
		for *val != 0 {
			// ASM hint for spin loop
			// Pause()
		}
		if atomic.SwapInt32(val, 1) == 0 {
			break
		}
	}
}

// Using similar custom ASM
func SpinLockASM(val *int32) {
	for {
		// Spin on simple read
		for *val != 0 {
			// ASM hint for spin loop
			Pause()
		}
		if Lock1XCHG32(val) == 0 {
			break
		}
	}
}

type SpinMutexBasic struct {
	val int32
}

func (m *SpinMutexBasic) Lock() {
	for !atomic.CompareAndSwapInt32(&m.val, 0, 1) {
		Pause()
	}
}

func (m *SpinMutexBasic) Unlock() {
	m.val = 0
}

type SpinMutex int32

// Fastest
func (m *SpinMutex) Lock() {
	for {
		var attempts int32 = LockAttempts
		SpinCountLock((*int32)(m), &attempts)
		if attempts > 0 {
			// if attempts < LockAttempts {
			// 	fmt.Println(attempts)
			// }
			return
		}
		// fmt.Println("Gosched")
		runtime.Gosched()
	}
}

func (m *SpinMutex) Unlock() {
	*m = 0
}

func (m *SpinMutex) IsLocked() bool {
	return *m == 1
}

type SpinMutexASM int32

func (m *SpinMutexASM) Lock() {
	for {
		// Spin on simple read
		for *m != 0 {
			// ASM hint for spin loop
			Pause()
		}
		if Lock1XCHG32((*int32)(m)) == 0 {
			break
		}
	}
	// for atomic.SwapInt32(&m.val, 1) != 0 {
	// 	// Spin on simple read
	// 	for m.val != 0 {
	// 		// ASM hint for spin loop
	// 		Pause()
	// 	}
	// }
}

func (m *SpinMutexASM) Unlock() {
	*m = 0
}

type SpinHLEMutex int32

func (m *SpinHLEMutex) Lock() {
	// HLESpinLock((*int32)(m))
	for {
		var attempts int32 = LockAttempts
		HLESpinCountLock((*int32)(m), &attempts)

		// If
		if attempts > 0 {
			// if attempts < LockAttempts {
			// 	fmt.Println(attempts)
			// }
			return
		}
		runtime.Gosched()
	}
}

func (m *SpinHLEMutex) Unlock() {
	HLEUnlock((*int32)(m))
}

type RTMMutex struct {
	val int32
	// fallback bool
	// lock     sync.Mutex
}

func (m *RTMMutex) Lock() {
	// if status := rtm.TxBegin(); status == rtm.TxBeginStarted {
	// 	rtm.TxEnd()
	// } else {
	// 	if (status & rtm.TxAbortRetry) == 0 {
	// 		fmt.Println("Said not to retry")
	// 	}

	// }

	for {
		for m.val != 0 {
			Pause()
		}
		if atomic.CompareAndSwapInt32(&m.val, 0, 1) {
			break
		}
	}
}

func (m *RTMMutex) Unlock() {
	m.val = 0
}
