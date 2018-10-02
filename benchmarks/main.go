// This is a contrived benchmark that caters to transaction memory primitives.
package main

import (
	"encoding/csv"
	"flag"
	"fmt"
	"math/rand"
	"os"
	"runtime"
	"strings"
	"sync"
	"time"

	rtm "github.com/0xmjk/go-tsx-rtm"
	"github.com/linux4life798/safetyfast"
)

type testlock struct {
	name string
	m    sync.Locker
}

var syncprimitives = map[string]testlock{
	"systemmutex": {
		name: "SystemMutex",
		m:    new(sync.Mutex),
	},
	"spinmutex": {
		name: "SpinMutex",
		m:    new(safetyfast.SpinMutex),
	},
	"spinhlemutex": {
		name: "SpinHLEMutex",
		m:    new(safetyfast.SpinHLEMutex),
	},
	"spinrtm": {
		name: "SpinRTMWithPause",
		m:    new(sync.Mutex),
	},
	"spinrtmnopause": {
		name: "SpinRTMNoPause",
		m:    new(sync.Mutex),
	},
	"spinrtmwithlibrary": {
		name: "SpinRTMWithLibrary",
		m:    new(sync.Mutex),
	},
}

type BinTouchCounter struct {
	bins []int32
}

func NewBinTouchCounter(numBins int) *BinTouchCounter {
	btc := new(BinTouchCounter)
	btc.bins = make([]int32, numBins)
	return btc
}

func (btc *BinTouchCounter) Resize(numBins int) {
	if numBins > cap(btc.bins) {
		// allocate new
		oldbins := btc.bins
		btc.bins = make([]int32, numBins)
		copy(btc.bins, oldbins)
	} else {
		// resize using slice
		btc.bins = btc.bins[0:numBins]
	}
}

func (btc *BinTouchCounter) Clear() {
	for i := range btc.bins {
		btc.bins[i] = 0
	}
}

func (btc *BinTouchCounter) TotalSum() uint64 {
	var sum uint64
	for i := range btc.bins {
		sum += uint64(btc.bins[i])
	}
	return sum
}

func (btc *BinTouchCounter) Touch(binIndex int) {
	if len(btc.bins) == 0 {
		return
	}
	binIndex = binIndex % len(btc.bins)
	btc.bins[binIndex]++
}

func GoRoutine(wg *sync.WaitGroup, values *RandValues, btc *BinTouchCounter, m sync.Locker) {
	vals := values.GetAll()
	for _, v := range vals {
		index := int(v.(int32))
		m.Lock()
		btc.Touch(index)
		m.Unlock()
	}
	wg.Done()
}

func GoRoutineRTMWithPause(wg *sync.WaitGroup, values *RandValues, btc *BinTouchCounter, m *safetyfast.SpinMutex) {
	vals := values.GetAll()
	for _, v := range vals {
		index := int(v.(int32))
	retry:
		if status := rtm.TxBegin(); status == rtm.TxBeginStarted {
			if m.IsLocked() {
				rtm.TxAbort()
			}
			btc.Touch(index)
			rtm.TxEnd()
		} else {
			if status&(rtm.TxAbortRetry|rtm.TxAbortConflict) != 0 {
				safetyfast.Pause()
				goto retry
			}
			// if status&rtm.TxAbortCapacity != 0 {
			// 	runtime.Breakpoint()
			// }
			m.Lock()
			btc.Touch(index)
			m.Unlock()
		}
	}
	wg.Done()
}

func GoRoutineRTMNoPause(wg *sync.WaitGroup, values *RandValues, btc *BinTouchCounter, m sync.Locker, fallback *bool) {
	vals := values.GetAll()
	for _, v := range vals {
		index := int(v.(int32))
	retry:
		if status := rtm.TxBegin(); status == rtm.TxBeginStarted {
			if *fallback {
				rtm.TxAbort()
			}
			btc.Touch(index)
			rtm.TxEnd()
		} else {
			if status&(rtm.TxAbortRetry|rtm.TxAbortConflict) != 0 {
				goto retry
			}
			m.Lock()
			*fallback = true
			btc.Touch(index)
			*fallback = false
			m.Unlock()
		}
	}
	wg.Done()
}

// GoRoutineRTMWithLibrary exercises RTMContext
func GoRoutineRTMWithLibrary(wg *sync.WaitGroup, values *RandValues, btc *BinTouchCounter, r *safetyfast.RTMContext) {
	vals := values.GetAll()
	for _, v := range vals {
		index := int(v.(int32))

		r.Atomic(func() {
			btc.Touch(index)
		})

	}
	wg.Done()
}

var FlagLockType string
var FlagCSV bool
var FlagNumBinStart int64
var FlagNumBinEnd int64
var FlagPlotFileName string

func init() {
	flag.StringVar(&FlagLockType, "lock", "all", "SystemMutex | SpinMutex | SpinHLEMutex | SpinRTM | SpinRTMNoPause | SpinRTMWithLibrary | all")
	flag.BoolVar(&FlagCSV, "csv", false, "Indicates if the output should be CSV format")
	flag.Int64Var(&FlagNumBinStart, "binsstart", 1, "")
	flag.Int64Var(&FlagNumBinEnd, "binsend", 1000000000, "")
	flag.StringVar(&FlagPlotFileName, "plot", "", "The name of the file to save the plot in eg. \"plot.svg\"")
}

func main() {
	flag.Parse()
	numGoRoutines := runtime.GOMAXPROCS(-1)
	numOpsPerGoRoutine := 480000

	csvout := csv.NewWriter(os.Stdout)
	defer csvout.Flush()
	if FlagCSV {
		csvout.Write([]string{"LockMethod", "NumBins", "NumTotalOperations", "TotalRuntime-ms", "RuntimePerOp-ns"})
	}

	rand.Seed(int64(1234))
	values := make([]*RandValues, numGoRoutines)
	for i := range values {
		values[i] = NewRandValues()
	}

	for i := range values {
		values[i].Clear().AddSparseInt32(numOpsPerGoRoutine)
	}

	var wg sync.WaitGroup

	var testlocks []testlock = make([]testlock, 0)

	for _, primitive := range strings.Split(strings.ToLower(FlagLockType), ",") {
		if primitive == "all" {
			for _, p := range syncprimitives {
				testlocks = append(testlocks, p)
			}
			break
		}

		p, ok := syncprimitives[primitive]
		if !ok {
			fmt.Fprintf(os.Stderr, "LockType \"%s\" invalid\n", primitive)
			os.Exit(1)
		}
		testlocks = append(testlocks, p)
	}

	plot := NewPerfPlot()

	for _, l := range testlocks {
		var avgCount int64
		var avgMs float64
		m := l.m

		btc := NewBinTouchCounter(int(FlagNumBinEnd * 2))

		for numBins := FlagNumBinStart; numBins < FlagNumBinEnd; numBins *= 2 {
			btc.Resize(int(numBins))
			btc.Clear()
			var rtmfallback bool
			r := safetyfast.NewRTMContexDefault()

			wg.Add(numGoRoutines)
			start := time.Now()
			if l.name == "SpinRTMWithPause" {
				var m safetyfast.SpinMutex
				for gid := 0; gid < numGoRoutines; gid++ {
					go GoRoutineRTMWithPause(&wg, values[gid], btc, &m)
				}
			} else if l.name == "SpinRTMNoPause" {
				for gid := 0; gid < numGoRoutines; gid++ {
					go GoRoutineRTMNoPause(&wg, values[gid], btc, m.(*sync.Mutex), &rtmfallback)
				}
			} else if l.name == "SpinRTMWithLibrary" {
				for gid := 0; gid < numGoRoutines; gid++ {
					go GoRoutineRTMWithLibrary(&wg, values[gid], btc, r)
				}
			} else {
				for gid := 0; gid < numGoRoutines; gid++ {
					go GoRoutine(&wg, values[gid], btc, m)
				}
			}
			wg.Wait()
			dur := time.Since(start)
			// fmt.Println("CapacityAborts:", r.CapacityAborts())

			totalOps := (int64(numGoRoutines) * int64(numOpsPerGoRoutine))
			ns := float64(dur.Nanoseconds()) / float64(totalOps)
			ms := dur.Seconds() * float64(1000.0)
			if !FlagCSV {
				fmt.Printf("[%16s][%9d bins] %4.6f ms - %.6f ns/op\n", l.name, numBins, ms, ns)
			} else {
				csvout.Write([]string{l.name, fmt.Sprint(numBins), fmt.Sprint(totalOps), fmt.Sprint(ms), fmt.Sprint(ns)})
				csvout.Flush()
			}

			if FlagPlotFileName != "" {
				plot.AddMetric(l.name, numBins, dur/time.Duration(totalOps))
			}

			s := btc.TotalSum()
			if s != uint64(totalOps) {
				fmt.Fprintf(os.Stderr, "ERROR: The sum did not match num total ops. sum = %v | total = %v\n", s, totalOps)
			}
			avgCount++
			avgMs += ms
			runtime.GC()
			runtime.Gosched()
		}

		if !FlagCSV {
			avg := avgMs / float64(avgCount)
			fmt.Printf("[%16s] avg = %f ms\n", FlagLockType, avg)
		}
	}

	if FlagPlotFileName != "" {
		plot.Plot(FlagPlotFileName, "Number of Bins", "Runtime (ns)", fmt.Sprintf("%s Performance", FlagLockType))
		OpenPlot(FlagPlotFileName)
	}

}
