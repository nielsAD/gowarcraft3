// Author:  Niels A.D.
// Project: gowarcraft3 (https://github.com/nielsAD/gowarcraft3)
// License: Mozilla Public License, v2.0

// Package bncs implements the old Battle.net chat protocol for Warcraft III.
// The protocol is used for multiple classic games, but this package only
// implements the small set of packets required for Warcraft III to log
// on and enter chat.
//
// Based on protocol documentation by https://bnetdocs.org/
//
// Each packet type is mapped to a struct type that implements the Packet
// interface. To deserialize from a binary stream, use bncs.Read().
//
// This package tries to keep ammortized heap memory allocations to 0.
//
// General serialization format:
//
//    (UINT8)  Protocol signature (0xFF)
//    (UINT8)  Packet type ID
//    (UINT16) Packet size
//    [Packet Data]
//
package bncs

import (
	"github.com/nielsAD/gowarcraft3/protocol"
	"github.com/nielsAD/gowarcraft3/protocol/w3gs"
)

// Packet interface.
type Packet interface {
	Serialize(buf *protocol.Buffer, enc *Encoding) error
	Deserialize(buf *protocol.Buffer, enc *Encoding) error
}

// Encoding options for (de)serialization
type Encoding struct {
	w3gs.Encoding
}
