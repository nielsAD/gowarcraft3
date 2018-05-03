// Author:  Niels A.D.
// Project: gowarcraft3 (https://github.com/nielsAD/gowarcraft3)
// License: Mozilla Public License, v2.0

package mock_test

import (
	"testing"

	"github.com/nielsAD/gowarcraft3/mock"
)

func TestEmitterSingle(t *testing.T) {
	var e mock.Emitter

	// Empty
	e.Fire("Hello")

	var fired bool
	var id = e.On("", func(ev *mock.Event) {
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
	e.Once(false, func(ev *mock.Event) {
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
	var e mock.Emitter
	var fired int

	e.On(42, func(ev *mock.Event) {
		t.Fatal("Should never be called")
	})

	e.On(42, func(ev *mock.Event) {
		fired++
		ev.PreventNext()
	})

	e.Fire(1)
	if fired != 1 {
		t.Fatal("Expected 1 callback, got", fired)
	}

	var h = func(ev *mock.Event) {
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
	var e mock.Emitter
	var fired bool

	e.On("", func(ev *mock.Event) {
		e.Once("", func(ev *mock.Event) {
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

func benchmarkEmitter(b *testing.B, numListeners int) {
	var e mock.Emitter
	for i := 0; i < numListeners; i++ {
		e.On("", func(ev *mock.Event) {})
	}
	for n := 0; n < b.N; n++ {
		e.Fire("")
	}
}

func BenchmarkEmitter0(b *testing.B) { benchmarkEmitter(b, 0) }
func BenchmarkEmitter1(b *testing.B) { benchmarkEmitter(b, 1) }
func BenchmarkEmitter4(b *testing.B) { benchmarkEmitter(b, 4) }
