// Author:  Niels A.D.
// Project: gowarcraft3 (https://github.com/nielsAD/gowarcraft3)
// License: Mozilla Public License, v2.0

package w3g_test

import (
	"bytes"
	"io"
	"math"
	"testing"

	"github.com/nielsAD/gowarcraft3/file/w3g"
	"github.com/nielsAD/gowarcraft3/protocol"
)

func TestCompress(t *testing.T) {
	var ref [20480]byte
	for i := range ref {
		ref[i] = byte(i)
	}

	var b protocol.Buffer
	var c = w3g.NewCompressor(&b)
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
	var d = w3g.NewDecompressor(&b, c.NumBlocks, math.MaxUint32)
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
