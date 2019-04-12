// Author:  Niels A.D.
// Project: gowarcraft3 (https://github.com/nielsAD/gowarcraft3)
// License: Mozilla Public License, v2.0

package w3gs

import (
	"io"

	"github.com/nielsAD/gowarcraft3/protocol"
)

// Encoder keeps amortized allocs at 0 for repeated Packet.Serialize calls.
type Encoder struct {
	Encoding
	buf protocol.Buffer
}

// NewEncoder initialization
func NewEncoder(e Encoding) *Encoder {
	return &Encoder{
		Encoding: e,
	}
}

// Serialize packet and returns its byte representation.
// Result is valid until the next Serialize() call.
func (enc *Encoder) Serialize(p Packet) ([]byte, error) {
	enc.buf.Truncate()
	if err := p.Serialize(&enc.buf, &enc.Encoding); err != nil {
		return nil, err
	}
	return enc.buf.Bytes, nil
}

// Write serializes p and writes it to w.
func (enc *Encoder) Write(w io.Writer, p Packet) (int, error) {
	b, err := enc.Serialize(p)
	if err != nil {
		return 0, err
	}

	return w.Write(b)
}

// Decoder keeps amortized allocs at 0 for repeated Packet.Deserialize calls.
type Decoder struct {
	Encoding
	PacketFactory
	bufRaw protocol.Buffer
	bufDes protocol.Buffer
}

// NewDecoder initialization
func NewDecoder(e Encoding, f PacketFactory) *Decoder {
	return &Decoder{
		Encoding:      e,
		PacketFactory: f,
	}
}

// Deserialize reads exactly one packet from b and returns it in the proper (deserialized) packet type.
func (dec *Decoder) Deserialize(b []byte) (Packet, int, error) {
	dec.bufDes.Reset(b)

	var size = dec.bufDes.Size()
	if size < 4 || b[0] != ProtocolSig {
		return nil, 0, ErrNoProtocolSig
	}

	var fac = dec.PacketFactory
	if fac == nil {
		fac = DefaultFactory
	}

	var pkt = fac.NewPacket(b[1], &dec.Encoding)
	if pkt == nil {
		return nil, 0, ErrNoFactory
	}

	var err = pkt.Deserialize(&dec.bufDes, &dec.Encoding)

	var n = size - dec.bufDes.Size()
	if err != nil {
		return nil, n, err
	}

	return pkt, n, nil
}

// ReadRaw reads exactly one packet from r and returns its raw bytes.
// Result is valid until the next ReadRaw() call.
func (dec *Decoder) ReadRaw(r io.Reader) ([]byte, int, error) {
	dec.bufRaw.Truncate()

	if n, err := dec.bufRaw.ReadSizeFrom(r, 4); err != nil {
		if err == io.ErrUnexpectedEOF {
			err = ErrNoProtocolSig
		}

		return nil, int(n), err
	}

	if dec.bufRaw.Bytes[0] != ProtocolSig {
		return nil, 4, ErrNoProtocolSig
	}

	var size = int(uint16(dec.bufRaw.Bytes[3])<<8 | uint16(dec.bufRaw.Bytes[2]))
	if size < 4 {
		return nil, 4, ErrNoProtocolSig
	}

	if n, err := dec.bufRaw.ReadSizeFrom(r, size-4); err != nil {
		if err == io.EOF {
			err = io.ErrUnexpectedEOF
		}
		return nil, int(n) + 4, err
	}

	return dec.bufRaw.Bytes, size, nil
}

// Read exactly one packet from r and returns it in the proper (deserialized) packet type.
func (dec *Decoder) Read(r io.Reader) (Packet, int, error) {
	b, n, err := dec.ReadRaw(r)
	if err != nil {
		return nil, n, err
	}

	p, m, err := dec.Deserialize(b)
	if err != nil {
		return nil, n, err
	}
	if m != n {
		return nil, n, ErrInvalidPacketSize
	}

	return p, n, nil
}

// Serialize serializes p and returns its byte representation.
func Serialize(p Packet, e Encoding) ([]byte, error) {
	return NewEncoder(e).Serialize(p)
}

// Deserialize reads exactly one packet from b and returns it in the proper (deserialized) packet type.
func Deserialize(b []byte, e Encoding) (Packet, int, error) {
	return NewDecoder(e, nil).Deserialize(b)
}

// Read exactly one packet from r and returns it in the proper (deserialized) packet type.
func Read(r io.Reader, e Encoding) (Packet, int, error) {
	return NewDecoder(e, nil).Read(r)
}

// Write serializes p and writes it to w.
func Write(w io.Writer, p Packet, e Encoding) (int, error) {
	return NewEncoder(e).Write(w, p)
}
