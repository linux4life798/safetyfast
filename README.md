[![GoDoc](https://godoc.org/github.com/linux4life798/safetyfast?status.svg)](https://godoc.org/github.com/linux4life798/safetyfast)
[![Go Report Card](https://goreportcard.com/badge/github.com/linux4life798/safetyfast)](https://goreportcard.com/report/github.com/linux4life798/safetyfast)

# SafetyFast - Put thread-safety first, with the performance of safety last.

This is a Go library that implements synchronization primitives over
[Intel TSX][wikipedia-tsx] (hardware transactional primitives).

```shell
go get github.com/linux4life798/safetyfast
```

Checkout the [SafetyFast Project Page](http://craighesling.com/project/safetyfast).

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
m := map[string]int{
    "word1": 0,
    "word2": 0,
    "word3": 0,
    "word4": 0,
}

c := NewRTMContexDefault()
c.Atomic(func() {
    // Action to be done transactionally
    count := m["word1"]
    m["word1"] = count + 1
})
```

## Using HLE

```go
m := map[string]int{
    "word1": 0,
    "word2": 0,
    "word3": 0,
    "word4": 0,
}

var lock SpinHLEMutex
lock.Lock()
// Action to be done transactionally
count := m["word1"]
m["word1"] = count + 1
lock.Unlock()
```

[wikipedia-tsx]: https://en.wikipedia.org/wiki/Transactional_Synchronization_Extensions
