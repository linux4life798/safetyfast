[![GoDoc](https://godoc.org/github.com/linux4life798/safetyfast?status.svg)](https://godoc.org/github.com/linux4life798/safetyfast)
[![Go Report Card](https://goreportcard.com/badge/github.com/linux4life798/safetyfast)](https://goreportcard.com/report/github.com/linux4life798/safetyfast)
[![codecov](https://codecov.io/gh/linux4life798/safetyfast/branch/master/graph/badge.svg)](https://codecov.io/gh/linux4life798/safetyfast)
[![Build Status](https://travis-ci.org/linux4life798/safetyfast.svg?branch=master)](https://travis-ci.org/linux4life798/safetyfast)

# SafetyFast - Put thread-safety first, with the performance of safety last.

This is a Go library that implements synchronization primitives over
[Intel TSX][wikipedia-tsx] (hardware transactional primitives).

```shell
go get github.com/linux4life798/safetyfast
```

Checkout the [SafetyFast Project Page](http://craighesling.com/project/safetyfast).

# Benchmarking

The following plot shows the number of milliseconds it took for 8 goroutines
to increments 480000 random elements (per goroutine) of an array of ints.
The x axis denotes how large (and therefore sparse) the array was.
The synchronization primitive used during the increment is indicated as
a series/line.

![Performance Graph](benchmarks/output-craigmobileworkstation.svg)

Note that, as the array size increases, the likelihood of two goroutines
touching the same element at the same instance decreases.
This is why we see such a dramatic increase in speed, when using either
the HLE or RTM style synchronization primitive.

The `SystemMutex` is just `sync.Mutex`.

It is also worth observing that the performance started to degrade towards the
very large array sizes. This is most likely due to a cache size limitation.

# Snippets

## Using RTM

```go
m := map[string]int{
    "word1": 0,
}

c := NewRTMContexDefault()
c.Atomic(func() {
    // Action to be done transactionally
    m["word1"] = m["word1"] + 1
})
```

# Using HLE

```go
m := map[string]int{
    "word1": 0,
}

var lock safetyfast.SpinHLEMutex
lock.Lock()
// Action to be done transactionally
m["word1"] = m["word1"] + 1
lock.Unlock()
```

# Examples

## Checking for HLE and RTM CPU support
It is necessary to check that the CPU you are using support Intel RTM and/or
Intel HLE instruction sets, since safetyfast does not check.
This can be accomplished by using the Intel provided `cpuid` package, as shown
below.

```go
import (
  "github.com/intel-go/cpuid"
)

func main() {
	if !cpuid.HasExtendedFeature(cpuid.RTM) {
		panic("The CPU does not support Intel RTM")
	}

	if !cpuid.HasExtendedFeature(cpuid.HLE) {
		panic("The CPU does not support Intel HLE")
	}
}

```

## Using RTM

```go
package main

import (
    "fmt"
    "sync"
    "github.com/linux4life798/safetyfast"
)

func main() {
    m := map[string]int{
        "word1": 0,
        "word2": 0,
    }

    c := safetyfast.NewRTMContexDefault()
    var wg sync.WaitGroup

    wg.Add(2)
    go c.Atomic(func() {
        // Action to be done transactionally
        m["word1"] = m["word1"] + 1
        wg.Done()
    })
    go c.Atomic(func() {
        // Action to be done transactionally
        m["word1"] = m["word1"] + 1
        wg.Done()
    })
    wg.Wait()

    fmt.Println("word1 =", m["word1"])
}
```

## Using HLE

```go
package main

import (
    "fmt"
    "sync"
    "github.com/linux4life798/safetyfast"
)

func main() {
    m := map[string]int{
        "word1": 0,
        "word2": 0,
    }

    var lock safetyfast.SpinHLEMutex
    var wg sync.WaitGroup

    wg.Add(2)
    go func() {
        lock.Lock()
        // Action to be done transactionally
        m["word1"] = m["word1"] + 1
        lock.Unlock()
        wg.Done()
    }()
    go func() {
        lock.Lock()
        // Action to be done transactionally
        m["word1"] = m["word1"] + 1
        lock.Unlock()
        wg.Done()
    }()
    wg.Wait()

    fmt.Println("word1 =", m["word1"])
}
```

[wikipedia-tsx]: https://en.wikipedia.org/wiki/Transactional_Synchronization_Extensions
