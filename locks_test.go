package safetyfast

import (
	"math/rand"
	"runtime"
	"testing"
	"time"
)

func TestLock1XCHG8(t *testing.T) {
	var x int8
	ret := Lock1XCHG8(&x)
	if ret != 0 {
		t.Errorf("LockXCHG returned %v instead of 0", ret)
	}
	if x != 1 {
		t.Errorf("LockXCHG set x to %v instead of 1", x)
	}
}

func TestLock1XCHG32(t *testing.T) {
	var x int32
	ret := Lock1XCHG32(&x)
	if ret != 0 {
		t.Errorf("LockXCHG returned %v instead of 0", ret)
	}
	if x != 1 {
		t.Errorf("LockXCHG set x to %v instead of 1", x)
	}
}

func TestLock1XCHG64(t *testing.T) {
	var x int64
	ret := Lock1XCHG64(&x)
	if ret != 0 {
		t.Errorf("LockXCHG returned %v instead of 0", ret)
	}
	if x != 1 {
		t.Errorf("LockXCHG set x to %v instead of 1", x)
	}
}

func TestHLELock(t *testing.T) {
	var x int32
	t.Run("Default", func(t *testing.T) {
		ret := HLETryLock(&x)
		if ret != 0 {
			t.Errorf("HLETryLock returned %v instead of 0", ret)
		}
		if x != 1 {
			t.Errorf("HLETryLock set x to %v instead of 1", x)
		}
		HLEUnlock(&x)
		if x != 0 {
			t.Errorf("HLEUnlock set x to %v instead of 0", x)
		}
	})
	t.Run("2 Attempts", func(t *testing.T) {
		// Try to acquire when already acquired
		x = 1
		ret := HLETryLock(&x)
		if ret != 1 {
			t.Errorf("HLETryLock returned %v instead of 1", ret)
		}
		if x != 1 {
			t.Errorf("HLETryLock set x to %v instead of 1", x)
		}

		// No acquire when it has been released
		x = 0
		ret = HLETryLock(&x)
		if ret != 0 {
			t.Errorf("HLETryLock returned %v instead of 0", ret)
		}
		if x != 1 {
			t.Errorf("HLETryLock set x to %v instead of 1", x)
		}

		HLEUnlock(&x)
		if x != 0 {
			t.Errorf("HLEUnlock set x to %v instead of 0", x)
		}
	})
}

func TestHLESpinLock(t *testing.T) {
	var x int32
	t.Run("Default", func(t *testing.T) {
		HLESpinLock(&x)
		if x != 1 {
			t.Errorf("HLESpinLock set x to %v instead of 1", x)
		}
		HLEUnlock(&x)
		if x != 0 {
			t.Errorf("HLEUnlock set x to %v instead of 0", x)
		}
	})

	t.Run("Waiting", func(t *testing.T) {
		var released bool

		oldmaxprocs := runtime.GOMAXPROCS(2)

		x = 1
		go func() {
			randsrc := rand.NewSource(int64(time.Now().Second()))
			r := rand.New(randsrc)
			delayMS := r.Int31n(250)
			time.Sleep(time.Millisecond * time.Duration(delayMS))
			released = true
			// definitely race - but minimal test
			x = 0
		}()

		HLESpinLock(&x)
		if x != 1 {
			t.Errorf("HLESpinLock set x to %v instead of 1", x)
		}
		if !released {
			t.Error("HLESpinLock returned before sleeping goroutine released x")
		}
		HLEUnlock(&x)
		if x != 0 {
			t.Errorf("HLEUnlock set x to %v instead of 0", x)
		}

		runtime.GOMAXPROCS(oldmaxprocs)
	})
}
