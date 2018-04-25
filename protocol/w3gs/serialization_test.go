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

func TestSerializePacket(t *testing.T) {
	var buf = protocol.Buffer{Bytes: make([]byte, 2048)}
	if _, e := w3gs.SerializePacket(&buf, &w3gs.Join{InternalAddr: protocol.SockAddr{IP: net.IP([]byte{0, 0})}}); e != protocol.ErrInvalidIP4 {
		t.Fatal("ErrInvalidIP4 expected")
	}
}

func TestDeserializePacket(t *testing.T) {
	if _, _, e := w3gs.DeserializePacket(&protocol.Buffer{Bytes: []byte{0, 255, 4, 0}}); e != w3gs.ErrNoProtocolSig {
		t.Fatal("ErrNoProtocolSig expected if no protocol signature")
	}
	if _, _, e := w3gs.DeserializePacket(&protocol.Buffer{Bytes: []byte{w3gs.ProtocolSig, 255}}); e != w3gs.ErrNoProtocolSig {
		t.Fatal("ErrNoProtocolSig expected if no size")
	}
	if _, _, e := w3gs.DeserializePacket(&protocol.Buffer{Bytes: []byte{w3gs.ProtocolSig, 3, 0}}); e != w3gs.ErrNoProtocolSig {
		t.Fatal("ErrNoProtocolSig expected if size < 4")
	}
	if _, _, e := w3gs.DeserializePacket(&protocol.Buffer{Bytes: []byte{w3gs.ProtocolSig, 255, 255, 0}}); e != io.ErrUnexpectedEOF {
		t.Fatal("ErrUnexpectedEOF expected if invalid size", e)
	}

	var buf = protocol.Buffer{Bytes: make([]byte, 8096)}
	buf.WriteUInt8At(0, w3gs.ProtocolSig)
	buf.WriteUInt8At(1, w3gs.PidSlotInfoJoin)
	buf.WriteUInt16At(2, 8)
	if _, _, e := w3gs.DeserializePacket(&buf); e != w3gs.ErrInvalidPacketSize {
		t.Fatal("ErrInvalidPacketSize expected if invalid data")
	}

	buf.WriteUInt8At(0, w3gs.ProtocolSig)
	buf.WriteUInt8At(1, w3gs.PidSlotInfoJoin)
	buf.WriteUInt16At(2, 6144)
	if _, _, e := w3gs.DeserializePacket(&buf); e != w3gs.ErrBufferTooSmall {
		t.Fatal("ErrBufferTooSmall expected if packet size exceeds buffer")
	}
}

func BenchmarkSerializePacket(b *testing.B) {
	var pkt = w3gs.SlotInfo{
		Slots: sd,
	}

	var bbuf w3gs.SerializationBuffer
	var w = &protocol.Buffer{}

	w3gs.SerializePacketWithBuffer(w, &bbuf, &pkt)

	b.SetBytes(int64(w.Size()))
	b.ResetTimer()

	for n := 0; n < b.N; n++ {
		w.Truncate()
		w3gs.SerializePacketWithBuffer(w, &bbuf, &pkt)
	}
}

func BenchmarkDeserializePacket(b *testing.B) {
	var pkt = w3gs.SlotInfo{
		Slots: sd,
	}

	var pbuf = protocol.Buffer{Bytes: make([]byte, 0, 2048)}
	pkt.Serialize(&pbuf)

	b.SetBytes(int64(pbuf.Size()))
	b.ResetTimer()

	var bbuf w3gs.DeserializationBuffer
	var r = &protocol.Buffer{}
	for n := 0; n < b.N; n++ {
		r.Bytes = pbuf.Bytes
		w3gs.DeserializePacketWithBuffer(r, &bbuf)
	}
}
