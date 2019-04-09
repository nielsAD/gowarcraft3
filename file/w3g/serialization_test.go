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
	if _, _, e := w3g.DeserializeRecordBytes([]byte{255}); e != io.ErrUnexpectedEOF {
		t.Fatal("ErrUnexpectedEOF expected if buffer size < 2")
	}
	if _, _, e := w3g.DeserializeRecordBytes([]byte{255, 0, 0, 0}); e != w3g.ErrUnknownRecord {
		t.Fatal("ErrUnknownRecord expected if invalid record ID")
	}
	if _, _, e := w3g.DeserializeRecordBytes([]byte{w3g.RidCountDownEnd, 0, 0}); e != io.ErrShortBuffer {
		t.Fatal("ErrShortBuffer expected if buffer too short")
	}
}

// 420MB/s
func BenchmarkSerializePacket(b *testing.B) {
	var e = w3g.NewRecordEncoder(w3g.Stream{})
	var w = &protocol.Buffer{}

	e.Serialize(w, &ts)

	b.SetBytes(int64(w.Size()))
	b.ResetTimer()

	for n := 0; n < b.N; n++ {
		w.Truncate()
		e.Serialize(w, &ts)
	}
}

// 700MB
func BenchmarkDeserializePacket(b *testing.B) {
	var bytes w3g.Stream
	ts.Serialize(&bytes)

	b.SetBytes(int64(bytes.Size()))
	b.ResetTimer()

	var d = w3g.NewRecordDecoder(w3g.Stream{})
	for n := 0; n < b.N; n++ {
		d.DeserializeBytes(bytes.Bytes)
	}
}
