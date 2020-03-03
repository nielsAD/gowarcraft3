// Author:  Niels A.D.
// Project: gowarcraft3 (https://github.com/nielsAD/gowarcraft3)
// License: Mozilla Public License, v2.0

package w3g_test

import (
	"bytes"
	"io"
	"math"
	"reflect"
	"testing"

	"github.com/nielsAD/gowarcraft3/file/w3g"
	"github.com/nielsAD/gowarcraft3/protocol"
	"github.com/nielsAD/gowarcraft3/protocol/w3gs"
)

func TestBlockCompressor(t *testing.T) {
	var ref [20480]byte
	for i := range ref {
		ref[i] = byte(i)
	}

	var b protocol.Buffer
	var c = w3g.NewBlockCompressor(&b, w3g.Encoding{})
	for i := 0; i < 10; i++ {
		n, err := c.Write(ref[i*2048 : (i+1)*2048])
		if err != nil {
			t.Fatal(err)
		}
		if n != 2048 {
			t.Fatalf("%d: Expected c.Write() to be 2048, but got %d", i, n)
		}
	}
	if c.NumBlocks != 10 {
		t.Fatalf("Expected c.NumBlocks to be 10, but got %d", c.NumBlocks)
	}
	if c.SizeTotal != 20480 {
		t.Fatalf("Expected c.SizeTotal to be 20480, but got %d", c.SizeTotal)
	}
	if c.SizeWritten != uint32(b.Size()) {
		t.Fatalf("Expected c.SizeWritten to be %d, but got %d", b.Size(), c.SizeWritten)
	}

	var buf [2048]byte
	var d = w3g.NewDecompressor(&b, w3g.Encoding{}, nil, c.NumBlocks, math.MaxUint32)
	for i := 0; i < 10; i++ {
		n, err := d.Read(buf[:])
		if err != nil {
			t.Fatal(err)
		}
		if n != 2048 {
			t.Fatalf("%d: Expected d.Read() to be 2048, but got %d", i, n)
		}
		if !bytes.Equal(buf[:], ref[i*2048:(i+1)*2048]) {
			t.Fatalf("%d: Bytes not equal", i)
		}
	}

	if d.SizeRead != c.SizeWritten {
		t.Fatalf("Expected d.SizeRead to be c.SizeWritten, but got %d != %d", d.SizeRead, c.SizeWritten)
	}
	if n, err := d.Read(buf[:]); err != io.EOF || n != 0 {
		t.Fatalf("Expected d.Read() to be EOF, but got %s", err.Error())
	}
}

func TestCompressor(t *testing.T) {
	var b protocol.Buffer
	var c = w3g.NewCompressor(&b, w3g.Encoding{})
	for i := 0; i < 100; i++ {
		if _, err := c.WriteRecord(&w3g.TimeSlot{TimeSlot: w3gs.TimeSlot{
			TimeIncrementMS: uint16(i),
			Actions:         ts.Actions,
		}}); err != nil {
			t.Fatal(err)
		}
	}
	if err := c.Close(); err != nil {
		t.Fatal(err)
	}

	if c.NumBlocks != 3 {
		t.Fatalf("Expected 3 blocks, but got %d", c.NumBlocks)
	}

	var i = 0
	var d = w3g.NewDecompressor(&b, w3g.Encoding{}, nil, c.NumBlocks, c.SizeTotal)
	if err := d.ForEach(func(r w3g.Record) error {
		s, ok := r.(*w3g.TimeSlot)
		if !ok {
			t.Fatal("Expected TimeSlot")
		}
		if s.TimeIncrementMS != uint16(i) || !reflect.DeepEqual(s.Actions, ts.Actions) {
			t.Fatal("Corrupt data")
		}
		i++
		return nil
	}); err != nil {
		t.Fatal(err)
	}

	if i != 100 {
		t.Fatalf("Expected 100 records, but got %d", i)
	}
	if d.SizeTotal != 0 {
		t.Fatalf("Expected nothing left to read, but got %d", d.SizeTotal)
	}
	if d.SizeRead != c.SizeWritten {
		t.Fatalf("Expected d.SizeRead to be c.SizeWritten, but got %d != %d", d.SizeRead, c.SizeWritten)
	}
}

func BenchmarkCompress(b *testing.B) {
	var ref [8196]byte
	for i := range ref {
		ref[i] = byte(i)
	}

	var w protocol.Buffer
	var c = w3g.NewBlockCompressor(&w, w3g.Encoding{})
	c.Write(ref[:])

	b.SetBytes(int64(len(ref)))
	b.ResetTimer()

	for n := 0; n < b.N; n++ {
		w.Truncate()
		c.Write(ref[:])
	}
}

func BenchmarkDecompress(b *testing.B) {
	var ref [8196]byte
	for i := range ref {
		ref[i] = byte(i)
	}

	var w protocol.Buffer
	var c = w3g.NewBlockCompressor(&w, w3g.Encoding{})
	c.Write(ref[:])

	var r protocol.Buffer
	var d = w3g.NewDecompressor(&r, w3g.Encoding{}, nil, c.NumBlocks, c.SizeTotal)

	b.SetBytes(int64(len(ref)))
	b.ResetTimer()

	for n := 0; n < b.N; n++ {
		r.Reset(w.Bytes)
		d.NumBlocks = c.NumBlocks
		d.SizeTotal = c.SizeTotal
		d.Read(ref[:])
	}
}
