// Author:  Niels A.D.
// Project: gowarcraft3 (https://github.com/nielsAD/gowarcraft3)
// License: Mozilla Public License, v2.0

package network

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
	Opt         []EventArg
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

// EventEmitter is an event emitter based on argument types
// For every type, a listener can register callbacks. Callbacks will be fired in reverse order of registration.
// The structure is thread-safe and functions can be called from multiple goroutines at the same time.
type EventEmitter struct {
	id       uint32
	hanmutex sync.RWMutex
	handlers map[string][]eventHandler
	emask    uint32
	epool    [16]Event
}

// Emitter is the interface that wraps the basic Fire method
type Emitter interface {
	Fire(a EventArg, o ...EventArg) bool
}

// Listener is the interface that wraps the basic On/Once methods
type Listener interface {
	On(a EventArg, h EventHandler) EventID
	Once(a EventArg, h EventHandler) EventID
	Off(id EventID)
	OffAll(a EventArg)
}

func topic(a EventArg) string {
	if a == nil {
		return "*"
	}
	return reflect.TypeOf(a).String()
}

func (e *EventEmitter) addHandler(a EventArg, h EventHandler, once bool) EventID {
	var ht = topic(a)

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
func (e *EventEmitter) On(a EventArg, h EventHandler) EventID {
	return e.addHandler(a, h, false)
}

// Once an event of type a is fired, call handler h once
func (e *EventEmitter) Once(a EventArg, h EventHandler) EventID {
	return e.addHandler(a, h, true)
}

// Off stops id from listening to future events
func (e *EventEmitter) Off(id EventID) {
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
func (e *EventEmitter) OffAll(a EventArg) {
	var ht = topic(a)

	e.hanmutex.Lock()
	e.handlers[ht] = e.handlers[ht][:0]
	e.hanmutex.Unlock()
}

func (e *EventEmitter) offOnce(ht string, minID uint32, maxID uint32) {
	e.hanmutex.Lock()
	var end = 0
	var arr = append([]eventHandler(nil), e.handlers[ht]...)

	for i := 0; i < len(arr); i++ {
		if arr[i].id > maxID || arr[i].id < minID || !arr[i].once {
			arr[end] = arr[i]
			end++
		}
	}

	e.handlers[ht] = arr[:end]
	e.hanmutex.Unlock()
}

// Maintain a simple free list
func (e *EventEmitter) newEvent() (*Event, uint32) {
	var m = atomic.LoadUint32(&e.emask)
	var b = uint32(bits.TrailingZeros16(uint16(^m)))

	if b < uint32(len(e.epool)) && atomic.CompareAndSwapUint32(&e.emask, m, m|(1<<b)) {
		return &e.epool[b], b
	}

	return &Event{}, 0xFF
}

func (e *EventEmitter) freeEvent(b uint32) {
	if b >= uint32(len(e.epool)) {
		return
	}

	// Clear
	e.epool[b] = Event{}

	for {
		var m = atomic.LoadUint32(&e.emask)
		if atomic.CompareAndSwapUint32(&e.emask, m, m&^(1<<b)) {
			break
		}
	}
}

func (e *EventEmitter) fire(ht string, arr []eventHandler, ev *Event) bool {
	if len(arr) == 0 {
		return false
	}

	var prevent bool
	var once bool
	var minID = uint32(0xFFFFFFFF)

	for i := 0; i < len(arr); i++ {
		var eh = arr[i]
		eh.fun(ev)

		minID = eh.id
		if eh.once {
			once = true
		}
		if ev.preventNext {
			prevent = true
			break
		}
	}

	if once {
		e.offOnce(ht, minID, arr[0].id)
	}

	return prevent
}

// CatchAll
var ca = topic(nil)

// Fire new event of type a
func (e *EventEmitter) Fire(a EventArg, o ...EventArg) bool {
	var ht = topic(a)

	e.hanmutex.RLock()
	var arr1 = e.handlers[ca][:]
	var arr2 = e.handlers[ht][:]
	e.hanmutex.RUnlock()

	if len(arr1) == 0 && len(arr2) == 0 {
		return false
	}

	var ev, eid = e.newEvent()
	ev.Arg = a
	ev.Opt = o

	var prevent = e.fire(ca, arr1, ev)
	if !prevent && ht != ca {
		prevent = e.fire(ht, arr2, ev)
	}

	e.freeEvent(eid)

	return prevent
}
