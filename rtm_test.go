package safetyfast

import (
	"fmt"
	"math/rand"
	"runtime"
	"sync"
	"testing"
	"time"

	"github.com/intel-go/cpuid"
)

func TestRTMContext(t *testing.T) {
	const numConcurGoRoutines = 8
	const arrayLength = 100
	const numIterations = 5000000

	if !cpuid.HasExtendedFeature(cpuid.RTM) {
		// Let's not fail for Travis-CI
		fmt.Println("The CPU does not support Intel RTM - Skipping RTM Test!")
		t.Log("The CPU does not support Intel RTM - Skipping RTM Test!")
		// t.Fatal("The CPU does not support Intel RTM")
		return
	}

	oldmaxprocs := runtime.GOMAXPROCS(numConcurGoRoutines)

	var wg sync.WaitGroup
	var arr = make([]int, arrayLength)
	var c = NewRTMContexDefault()

	routine := func() {
		randsrc := rand.NewSource(int64(time.Now().Second()))
		r := rand.New(randsrc)

		for i := 0; i < numIterations; i++ {
			index := r.Int() % len(arr)
			c.Atomic(func() {
				arr[index]++
			})
		}
		wg.Done()
	}

	wg.Add(numConcurGoRoutines)
	for i := 0; i < numConcurGoRoutines; i++ {
		go routine()
	}
	wg.Wait()

	var sum int
	for _, v := range arr {
		sum += v
	}

	expected := numIterations * numConcurGoRoutines
	t.Logf("Array Length=%d | NumberGoRoutines=%d | NumIterations=%d", len(arr), numConcurGoRoutines, numIterations)
	t.Logf("ArraySum=%d | Expected=%d", sum, expected)
	t.Logf("CapacityAborts=%d", c.CapacityAborts())
	if sum != expected {
		t.Fatalf("Sum result is %d, but we expected %d", sum, expected)
	}

	runtime.GOMAXPROCS(oldmaxprocs)
}
