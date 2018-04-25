// Author:  Niels A.D.
// Project: gowarcraft3 (https://github.com/nielsAD/gowarcraft3)
// License: Mozilla Public License, v2.0

package bncs_test

import (
	"io"
	"net"
	"testing"

	"github.com/nielsAD/gowarcraft3/protocol"
	"github.com/nielsAD/gowarcraft3/protocol/bncs"
)

func TestSerializePacket(t *testing.T) {
	var buf = protocol.Buffer{Bytes: make([]byte, 2048)}
	if _, e := bncs.SerializePacket(&buf, &bncs.AuthInfoReq{LocalIP: net.IP([]byte{0, 0})}); e != protocol.ErrInvalidIP4 {
		t.Fatal("ErrInvalidIP4 expected")
	}
}

func TestDeserializeClientPacket(t *testing.T) {
	if _, _, e := bncs.DeserializeClientPacket(&protocol.Buffer{Bytes: []byte{0, 255, 4, 0}}); e != bncs.ErrNoProtocolSig {
		t.Fatal("ErrNoProtocolSig expected if no protocol signature")
	}
	if _, _, e := bncs.DeserializeClientPacket(&protocol.Buffer{Bytes: []byte{bncs.ProtocolSig, 255}}); e != bncs.ErrNoProtocolSig {
		t.Fatal("ErrNoProtocolSig expected if no size")
	}
	if _, _, e := bncs.DeserializeClientPacket(&protocol.Buffer{Bytes: []byte{bncs.ProtocolSig, 3, 0}}); e != bncs.ErrNoProtocolSig {
		t.Fatal("ErrNoProtocolSig expected if size < 4")
	}
	if _, _, e := bncs.DeserializeClientPacket(&protocol.Buffer{Bytes: []byte{bncs.ProtocolSig, 255, 255, 0}}); e != io.ErrUnexpectedEOF {
		t.Fatal("ErrUnexpectedEOF expected if invalid size", e)
	}

	var buf = protocol.Buffer{Bytes: make([]byte, 8096)}
	buf.WriteUInt8At(0, bncs.ProtocolSig)
	buf.WriteUInt8At(1, bncs.PidAuthAccountLogon)
	buf.WriteUInt16At(2, 8)
	if _, _, e := bncs.DeserializeClientPacket(&buf); e != bncs.ErrInvalidPacketSize {
		t.Fatal("ErrInvalidPacketSize expected if invalid data")
	}

	buf.WriteUInt8At(0, bncs.ProtocolSig)
	buf.WriteUInt8At(1, bncs.PidAuthAccountLogon)
	buf.WriteUInt16At(2, 6144)
	if _, _, e := bncs.DeserializeClientPacket(&buf); e != bncs.ErrBufferTooSmall {
		t.Fatal("ErrBufferTooSmall expected if packet size exceeds buffer")
	}
}

func TestDeserializeServerPacket(t *testing.T) {
	if _, _, e := bncs.DeserializeServerPacket(&protocol.Buffer{Bytes: []byte{0, 255, 4, 0}}); e != bncs.ErrNoProtocolSig {
		t.Fatal("ErrNoProtocolSig expected if no protocol signature")
	}
	if _, _, e := bncs.DeserializeServerPacket(&protocol.Buffer{Bytes: []byte{bncs.ProtocolSig, 255}}); e != bncs.ErrNoProtocolSig {
		t.Fatal("ErrNoProtocolSig expected if no size")
	}
	if _, _, e := bncs.DeserializeServerPacket(&protocol.Buffer{Bytes: []byte{bncs.ProtocolSig, 3, 0}}); e != bncs.ErrNoProtocolSig {
		t.Fatal("ErrNoProtocolSig expected if size < 4")
	}
	if _, _, e := bncs.DeserializeServerPacket(&protocol.Buffer{Bytes: []byte{bncs.ProtocolSig, 255, 255, 0}}); e != io.ErrUnexpectedEOF {
		t.Fatal("ErrUnexpectedEOF expected if invalid size", e)
	}

	var buf = protocol.Buffer{Bytes: make([]byte, 8096)}
	buf.WriteUInt8At(0, bncs.ProtocolSig)
	buf.WriteUInt8At(1, bncs.PidAuthAccountLogon)
	buf.WriteUInt16At(2, 8)
	if _, _, e := bncs.DeserializeServerPacket(&buf); e != bncs.ErrInvalidPacketSize {
		t.Fatal("ErrInvalidPacketSize expected if invalid data")
	}

	buf.WriteUInt8At(0, bncs.ProtocolSig)
	buf.WriteUInt8At(1, bncs.PidAuthAccountLogon)
	buf.WriteUInt16At(2, 6144)
	if _, _, e := bncs.DeserializeServerPacket(&buf); e != bncs.ErrBufferTooSmall {
		t.Fatal("ErrBufferTooSmall expected if packet size exceeds buffer")
	}
}
