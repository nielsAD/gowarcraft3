// Author:  Niels A.D.
// Project: gowarcraft3 (https://github.com/nielsAD/gowarcraft3)
// License: Mozilla Public License, v2.0

package w3g_test

import (
	"io"
	"testing"

	"github.com/nielsAD/gowarcraft3/file/w3g"
	"github.com/nielsAD/gowarcraft3/protocol"
)

func TestDeserializeRecord(t *testing.T) {
	if _, _, e := w3g.DeserializeRecord([]byte{255}, w3g.Encoding{}); e != io.ErrUnexpectedEOF {
		t.Fatal("ErrUnexpectedEOF expected if buffer size < 2")
	}
	if _, _, e := w3g.DeserializeRecord([]byte{255, 0, 0, 0}, w3g.Encoding{}); e != w3g.ErrUnknownRecord {
		t.Fatal("ErrUnknownRecord expected if invalid record ID")
	}
	if _, _, e := w3g.DeserializeRecord([]byte{w3g.RidCountDownEnd, 0, 0}, w3g.Encoding{}); e != io.ErrShortBuffer {
		t.Fatal("ErrShortBuffer expected if buffer too short")
	}
}

func BenchmarkSerializePacket(b *testing.B) {
	var e = w3g.NewRecordEncoder(w3g.Encoding{})
	var w = &protocol.Buffer{}

	e.Write(w, &ts)

	b.SetBytes(int64(w.Size()))
	b.ResetTimer()

	for n := 0; n < b.N; n++ {
		w.Truncate()
		e.Write(w, &ts)
	}
}

func BenchmarkDeserializePacket(b *testing.B) {
	var input = protocol.Buffer{}
	ts.Serialize(&input, &w3g.Encoding{})

	b.SetBytes(int64(input.Size()))
	b.ResetTimer()

	var d = w3g.NewRecordDecoder(w3g.Encoding{})
	for n := 0; n < b.N; n++ {
		d.Deserialize(input.Bytes)
	}
}
