// Author:  Niels A.D.
// Project: gowarcraft3 (https://github.com/nielsAD/gowarcraft3)
// License: Mozilla Public License, v2.0

package network

import "sync"

// RWMutex implements a read-preferring readers–writer lock
type RWMutex struct {
	sync.Mutex
	rcount uint32
	rmutex sync.Mutex
}

// RLock acquires a readers-lock
func (m *RWMutex) RLock() {
	m.rmutex.Lock()
	m.rcount++
	if m.rcount == 1 {
		m.Lock()
	}
	m.rmutex.Unlock()
}

// RUnlock decrements the readers-lock
func (m *RWMutex) RUnlock() {
	m.rmutex.Lock()
	if m.rcount == 0 {
		panic("rwlock: rcount < 0")
	}
	m.rcount--
	if m.rcount == 0 {
		m.Unlock()
	}
	m.rmutex.Unlock()
}
