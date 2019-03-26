// Author:  Niels A.D.
// Project: gowarcraft3 (https://github.com/nielsAD/gowarcraft3)
// License: Mozilla Public License, v2.0

package network_test

import (
	"sync"
	"sync/atomic"
	"testing"

	"github.com/nielsAD/gowarcraft3/network"
)

func TestEmitterSingle(t *testing.T) {
	var e network.EventEmitter

	// Empty
	e.Fire("Hello")

	var fired bool
	var id = e.On("", func(ev *network.Event) {
		s, ok := ev.Arg.(string)
		if !ok {
			t.Fatal("Expected string")
		}
		if s != "World" {
			t.Fatal("Expected World")
		}
		fired = true
	})

	// Fire something else
	e.Fire(123)
	if fired {
		t.Fatal("World fired too early")
	}

	// Should fire an event
	e.Fire("World")
	if !fired {
		t.Fatal("World not fired")
	}

	// Add second handler
	e.Once(false, func(ev *network.Event) {
		b, ok := ev.Arg.(bool)
		if !ok {
			t.Fatal("Expected bool")
		}
		fired = b
	})

	// Should fire again
	fired = false
	e.Fire("World")
	if !fired {
		t.Fatal("World not fired 2")
	}

	e.Off(id)

	// Should not fire again
	fired = false
	e.Fire("World")
	if fired {
		t.Fatal("World fired while off")
	}

	// Fire second handler once
	e.Fire(false)
	if fired {
		t.Fatal("False not fired")
	}

	// Should not fire again
	e.Fire(true)
	if fired {
		t.Fatal("True fired while once")
	}
}

func TestEmitterMulti(t *testing.T) {
	var e network.EventEmitter
	var fired int

	e.On(42, func(ev *network.Event) {
		t.Fatal("Should never be called")
	})

	e.On(42, func(ev *network.Event) {
		fired++
		ev.PreventNext()
	})

	e.Fire(1)
	if fired != 1 {
		t.Fatal("Expected 1 callback, got", fired)
	}

	var h = func(ev *network.Event) {
		fired++
	}

	e.On(42, h)
	e.Once(42, h)
	var id = e.On(42, h)
	e.Once(42, h)

	fired = 0
	e.Fire(1)
	if fired != 5 {
		t.Fatal("Expected 5 callbacks, got", fired)
	}

	fired = 0
	e.Fire(1)
	if fired != 3 {
		t.Fatal("Expected 3 callbacks, got", fired)
	}

	e.Off(id)

	fired = 0
	e.Fire(1)
	if fired != 2 {
		t.Fatal("Expected 2 callbacks, got", fired)
	}

	e.OffAll(42)

	fired = 0
	e.Fire(1)
	if fired != 0 {
		t.Fatal("Expected 0 callbacks, got", fired)
	}
}

func TestEmitterRecursive(t *testing.T) {
	var e network.EventEmitter
	var fired bool

	e.On("", func(ev *network.Event) {
		e.Once("", func(ev *network.Event) {
			fired = true
			ev.PreventNext()
		})
		e.Fire("Foo")
	})

	e.Fire("Bar")
	if !fired {
		t.Fatal("Recursive callback not fired")
	}
}

func TestLiteralTopic(t *testing.T) {
	var e network.EventEmitter
	var fired bool

	e.On(network.Topic("hello world"), func(ev *network.Event) {
		fired = true
	})

	e.Fire("Foo")
	if fired {
		t.Fatal("Topic fired too early")
	}

	e.Fire(network.Topic("hello world"))
	if !fired {
		t.Fatal("Topic not fired")
	}
}

func TestCatchAll(t *testing.T) {
	var e network.EventEmitter
	var fired int

	e.On(nil, func(ev *network.Event) { fired++ })
	e.On("", func(ev *network.Event) { fired++ })
	e.Once(nil, func(ev *network.Event) { fired++ })
	e.Once("", func(ev *network.Event) { fired++ })

	e.Fire(123)
	e.Fire("Foo")
	e.Fire(456)
	e.Fire("Bar")

	if fired != 8 {
		t.Fatal("Expected 8 callbacks, got", fired)
	}
}

func TestGoroutines(t *testing.T) {
	var c uint32

	var e network.EventEmitter
	e.On(uint32(0), func(ev *network.Event) {
		atomic.AddUint32(&c, ev.Arg.(uint32))
	})

	var wg sync.WaitGroup
	for i := 0; i < 32; i++ {
		wg.Add(1)
		go func() {
			for i := 0; i < 1000; i++ {
				e.Fire(uint32(1))
			}
			wg.Done()
		}()
	}

	wg.Wait()

	if atomic.LoadUint32(&c) != 32*1000 {
		t.Fatal("Result invalid")
	}
}

func benchmarkEmitter(b *testing.B, numListeners int) {
	var e network.EventEmitter
	for i := 0; i < numListeners; i++ {
		e.On("", func(ev *network.Event) {})
	}
	for n := 0; n < b.N; n++ {
		e.Fire("")
	}
}

func BenchmarkEmitter0(b *testing.B) { benchmarkEmitter(b, 0) }
func BenchmarkEmitter1(b *testing.B) { benchmarkEmitter(b, 1) }
func BenchmarkEmitter4(b *testing.B) { benchmarkEmitter(b, 4) }
