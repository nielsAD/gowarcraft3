// Author:  Niels A.D.
// Project: gowarcraft3 (https://github.com/nielsAD/gowarcraft3)
// License: Mozilla Public License, v2.0

package mock

import (
	"math/bits"
	"reflect"
	"sync"
	"sync/atomic"
)

// EventArg identifies an event (and with it the type of its Arg)
type EventArg interface{}

// EventID identifies a single event handler
type EventID struct {
	ht string
	id uint32
}

// Event structure passed to event handlers
type Event struct {
	Arg         EventArg
	preventNext bool
}

// PreventNext prevents any other handlers from being called for this event
func (e *Event) PreventNext() {
	e.preventNext = true
}

// EventHandler callback function
type EventHandler func(*Event)

// Internal eventHandler struct
type eventHandler = struct {
	id   uint32
	once bool
	fun  EventHandler
}

// Emitter is an event emitter based on argument types
// For every type, a listener can register callbacks. Callbacks will be fired in reverse order of registration.
// The structure is thread-safe and functions can be called from multiple goroutines at the same time.
type Emitter struct {
	id       uint32
	hanmutex sync.RWMutex
	handlers map[string][]eventHandler
	emask    uint32
	epool    [16]Event
}

func (e *Emitter) addHandler(a EventArg, h EventHandler, once bool) EventID {
	var ht = reflect.TypeOf(a).String()

	e.hanmutex.Lock()
	e.id++
	var id = e.id

	if e.handlers == nil {
		e.handlers = make(map[string][]eventHandler)
	}
	e.handlers[ht] = append([]eventHandler{eventHandler{
		id:   id,
		once: once,
		fun:  h,
	}}, e.handlers[ht]...)
	e.hanmutex.Unlock()

	return EventID{
		ht: ht,
		id: id,
	}
}

// On an event of type a is, call handler h
func (e *Emitter) On(a EventArg, h EventHandler) EventID {
	return e.addHandler(a, h, false)
}

// Once an event of type a is fired, call handler h once
func (e *Emitter) Once(a EventArg, h EventHandler) EventID {
	return e.addHandler(a, h, true)
}

// Off stops id from listening to future events
func (e *Emitter) Off(id EventID) {
	e.hanmutex.Lock()
	var end = 0
	var arr = append([]eventHandler(nil), e.handlers[id.ht]...)

	for i := 0; i < len(arr); i++ {
		if arr[i].id != id.id {
			arr[end] = arr[i]
			end++
		}
	}

	e.handlers[id.ht] = arr[:end]
	e.hanmutex.Unlock()
}

// OffAll clears the current listeners for events of type a
func (e *Emitter) OffAll(a EventArg) {
	var ht = reflect.TypeOf(a).String()

	e.hanmutex.Lock()
	e.handlers[ht] = e.handlers[ht][:0]
	e.hanmutex.Unlock()
}

func (e *Emitter) offOnce(ht string, minID uint32) {
	e.hanmutex.Lock()
	var end = 0
	var arr = append([]eventHandler(nil), e.handlers[ht]...)

	for i := 0; i < len(arr); i++ {
		if arr[i].id < minID {
			break
		}
		if !arr[i].once {
			arr[end] = arr[i]
			end++
		}
	}

	e.handlers[ht] = arr[:end]
	e.hanmutex.Unlock()
}

// Maintain a simple free list
func (e *Emitter) newEvent() (*Event, uint32) {
	var m = e.emask
	var b = uint32(bits.TrailingZeros16(uint16(^m)))

	if b < uint32(len(e.epool)) && atomic.CompareAndSwapUint32(&e.emask, m, m|(1<<b)) {
		return &e.epool[b], b
	}

	return &Event{}, 0xFF
}

func (e *Emitter) freeEvent(b uint32) {
	if b >= uint32(len(e.epool)) {
		return
	}

	// Clear
	e.epool[b] = Event{}

	for {
		var m = e.emask
		if atomic.CompareAndSwapUint32(&e.emask, m, m&^(1<<b)) {
			break
		}
	}
}

// Fire new event of type a
func (e *Emitter) Fire(a EventArg) {
	var ht = reflect.TypeOf(a).String()

	e.hanmutex.RLock()
	var arr = e.handlers[ht][:]
	e.hanmutex.RUnlock()

	if len(arr) == 0 {
		return
	}

	var once bool
	var minID = uint32(0xFFFFFFFF)

	var ev, eid = e.newEvent()
	ev.Arg = a

	for i := 0; i < len(arr); i++ {
		var eh = arr[i]
		eh.fun(ev)

		minID = eh.id
		if eh.once {
			once = true
		}
		if ev.preventNext {
			break
		}
	}

	e.freeEvent(eid)

	if once {
		e.offOnce(ht, minID)
	}
}
