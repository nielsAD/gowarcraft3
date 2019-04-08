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
	if _, _, e := w3g.DeserializeRecordRaw([]byte{255}); e != io.ErrUnexpectedEOF {
		t.Fatal("ErrUnexpectedEOF expected if buffer size < 2")
	}
	if _, _, e := w3g.DeserializeRecordRaw([]byte{255, 0, 0, 0}); e != w3g.ErrUnknownRecord {
		t.Fatal("ErrUnknownRecord expected if invalid record ID")
	}
	if _, _, e := w3g.DeserializeRecordRaw([]byte{w3g.RidCountDownEnd, 0, 0}); e != io.ErrShortBuffer {
		t.Fatal("ErrShortBuffer expected if buffer too short")
	}
}

func BenchmarkSerializePacket(b *testing.B) {
	var bbuf w3g.SerializationBuffer
	var w = &protocol.Buffer{}

	w3g.SerializeRecordWithBuffer(w, &bbuf, &ts)

	b.SetBytes(int64(w.Size()))
	b.ResetTimer()

	for n := 0; n < b.N; n++ {
		w.Truncate()
		w3g.SerializeRecordWithBuffer(w, &bbuf, &ts)
	}
}

func BenchmarkDeserializePacket(b *testing.B) {
	var pbuf protocol.Buffer
	ts.Serialize(&pbuf)

	b.SetBytes(int64(pbuf.Size()))
	b.ResetTimer()

	var bbuf w3g.DeserializationBuffer
	for n := 0; n < b.N; n++ {
		w3g.DeserializeRecordWithBufferRaw(pbuf.Bytes, &bbuf)
	}
}
