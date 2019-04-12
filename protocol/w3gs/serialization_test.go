// Author:  Niels A.D.
// Project: gowarcraft3 (https://github.com/nielsAD/gowarcraft3)
// License: Mozilla Public License, v2.0

package w3gs_test

import (
	"io"
	"net"
	"testing"

	"github.com/nielsAD/gowarcraft3/protocol"
	"github.com/nielsAD/gowarcraft3/protocol/w3gs"
)

func TestSerialize(t *testing.T) {
	if _, e := w3gs.Write(&protocol.Buffer{}, &w3gs.Join{InternalAddr: protocol.SockAddr{IP: net.IP([]byte{0, 0})}}, w3gs.Encoding{}); e != protocol.ErrInvalidIP4 {
		t.Fatal("ErrInvalidIP4 expected")
	}
}

func TestDeserialize(t *testing.T) {
	if _, _, e := w3gs.Deserialize([]byte{0, 255, 4, 0}, w3gs.Encoding{}); e != w3gs.ErrNoProtocolSig {
		t.Fatal("ErrNoProtocolSig expected if no protocol signature")
	}
	if _, _, e := w3gs.Deserialize([]byte{w3gs.ProtocolSig, 255}, w3gs.Encoding{}); e != w3gs.ErrNoProtocolSig {
		t.Fatal("ErrNoProtocolSig expected if no size")
	}
	if _, _, e := w3gs.Deserialize([]byte{w3gs.ProtocolSig, 3, 0}, w3gs.Encoding{}); e != w3gs.ErrNoProtocolSig {
		t.Fatal("ErrNoProtocolSig expected if size < 4")
	}
	if _, _, e := w3gs.Deserialize([]byte{w3gs.ProtocolSig, 255, 255, 0}, w3gs.Encoding{}); e != w3gs.ErrInvalidPacketSize {
		t.Fatal("ErrInvalidPacketSize expected if bytes invalid size", e)
	}
	if _, _, e := w3gs.Read(&protocol.Buffer{Bytes: []byte{w3gs.ProtocolSig, 255, 255, 0}}, w3gs.Encoding{}); e != io.ErrUnexpectedEOF {
		t.Fatal("ErrUnexpectedEOF expected if reader invalid size", e)
	}
}

func BenchmarkEncoder(b *testing.B) {
	var pkt = w3gs.SlotInfo{
		Slots: sd,
	}

	var e = w3gs.NewEncoder(w3gs.Encoding{})
	var w = &protocol.Buffer{}

	e.Write(w, &pkt)

	b.SetBytes(int64(w.Size()))
	b.ResetTimer()

	for n := 0; n < b.N; n++ {
		w.Truncate()
		e.Write(w, &pkt)
	}
}

func BenchmarkDecoder(b *testing.B) {
	var pkt = w3gs.SlotInfo{
		Slots: sd,
	}

	var input = protocol.Buffer{}
	pkt.Serialize(&input, &w3gs.Encoding{})

	b.SetBytes(int64(input.Size()))
	b.ResetTimer()

	var d = w3gs.NewDecoder(w3gs.Encoding{}, w3gs.NewFactoryCache(w3gs.DefaultFactory))
	var r = &protocol.Buffer{}
	for n := 0; n < b.N; n++ {
		r.Reset(input.Bytes)
		d.Read(r)
	}
}
