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

func TestSerialize(t *testing.T) {
	var buf = protocol.Buffer{Bytes: make([]byte, 2048)}
	if _, e := bncs.Write(&buf, &bncs.AuthInfoReq{LocalIP: net.IP([]byte{0, 0})}, bncs.Encoding{}); e != protocol.ErrInvalidIP4 {
		t.Fatal("ErrInvalidIP4 expected")
	}
}

func TestDeserialize(t *testing.T) {
	if _, _, e := bncs.Deserialize([]byte{0, 255, 4, 0}, bncs.Encoding{}); e != bncs.ErrNoProtocolSig {
		t.Fatal("ErrNoProtocolSig expected if no protocol signature")
	}
	if _, _, e := bncs.Deserialize([]byte{bncs.ProtocolSig, 255}, bncs.Encoding{}); e != bncs.ErrNoProtocolSig {
		t.Fatal("ErrNoProtocolSig expected if no size")
	}
	if _, _, e := bncs.Deserialize([]byte{bncs.ProtocolSig, 3, 0}, bncs.Encoding{}); e != bncs.ErrNoProtocolSig {
		t.Fatal("ErrNoProtocolSig expected if size < 4")
	}
	if _, _, e := bncs.Deserialize([]byte{bncs.ProtocolSig, 255, 255, 0}, bncs.Encoding{}); e != bncs.ErrInvalidPacketSize {
		t.Fatal("ErrUnexpectedEOF expected if bytes invalid size", e)
	}
	if _, _, e := bncs.Read(&protocol.Buffer{Bytes: []byte{bncs.ProtocolSig, 255, 255, 0}}, bncs.Encoding{}); e != io.ErrUnexpectedEOF {
		t.Fatal("ErrUnexpectedEOF expected if reader invalid size", e)
	}
}
