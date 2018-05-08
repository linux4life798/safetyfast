package safetyfast

import "sync"

type SystemMutex struct {
	m sync.Mutex
}

func (m *SystemMutex) Lock() {
	m.m.Lock()
}

func (m *SystemMutex) Unlock() {
	m.m.Unlock()
}

type NoMutex struct{}

func (m *NoMutex) Lock() {
}

func (m *NoMutex) Unlock() {
}
