[![GoDoc](https://godoc.org/github.com/linux4life798/safetyfast?status.svg)](https://godoc.org/github.com/linux4life798/safetyfast)

# SafetyFast - Put thread-safety first, with the performance of safety last.

This is a Go library that implements synchronization primitives over
[Intel TSX][wikipedia-tsx] (hardware transactional primitives).

```shell
go get github.com/linux4life798/safetyfast
```

Checkout the [SafetyFast Project Page](http://craighesling.com/project/safetyfast).

# Examples

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
