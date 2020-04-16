// Author:  Niels A.D.
// Project: gowarcraft3 (https://github.com/nielsAD/gowarcraft3)
// License: Mozilla Public License, v2.0

package w3g

import (
	"io"

	"github.com/nielsAD/gowarcraft3/protocol"
)

// RecordEncoder keeps amortized allocs at 0 for repeated Record.Serialize calls
// Byte slices are valid until the next Serialize() call
type RecordEncoder struct {
	Encoding
	buf protocol.Buffer
}

// NewRecordEncoder initialization
func NewRecordEncoder(e Encoding) *RecordEncoder {
	return &RecordEncoder{
		Encoding: e,
	}
}

// Serialize record and returns its byte representation.
// Result is valid until the next Serialize() call.
func (enc *RecordEncoder) Serialize(r Record) ([]byte, error) {
	enc.buf.Truncate()
	if err := r.Serialize(&enc.buf, &enc.Encoding); err != nil {
		return nil, err
	}
	return enc.buf.Bytes, nil
}

// Write serializes r and writes it to w.
func (enc *RecordEncoder) Write(w io.Writer, r Record) (int, error) {
	b, err := enc.Serialize(r)
	if err != nil {
		return 0, err
	}

	return w.Write(b)
}

// RecordDecoder keeps amortized allocs at 0 for repeated Record.Deserialize calls.
type RecordDecoder struct {
	Encoding
	RecordFactory
	buf protocol.Buffer
}

// NewRecordDecoder initialization
func NewRecordDecoder(e Encoding, f RecordFactory) *RecordDecoder {
	return &RecordDecoder{
		Encoding:      e,
		RecordFactory: f,
	}
}

// Deserialize reads exactly one record from b and returns it in the proper (deserialized) record type.
func (dec *RecordDecoder) Deserialize(b []byte) (Record, int, error) {
	dec.buf.Reset(b)

	var size = dec.buf.Size()
	if size < 2 {
		return nil, 0, io.ErrUnexpectedEOF
	}

	var fac = dec.RecordFactory
	if fac == nil {
		fac = DefaultFactory
	}

	var rec = fac.NewRecord(b[0], &dec.Encoding)
	if rec == nil {
		return nil, 0, ErrUnknownRecord
	}

	var err = rec.Deserialize(&dec.buf, &dec.Encoding)

	var n = size - dec.buf.Size()
	if err != nil {
		return nil, n, err
	}

	return rec, n, nil
}

// Peeker is the interface that wraps basic Peek and Discard methods.
//
// There is no way to determine record size without reading the full record,
// so it's impossible to know how many bytes should be read from an io.Reader.
// Instead, we peek a certain amount of bytes, deserialize the record, and
// then discard the actual amount of bytes read.
//
// This interface is implemented by bufio.Reader, for example.
type Peeker interface {
	Peek(n int) ([]byte, error)
	Discard(n int) (discarded int, err error)
}

// Read exactly one record from r and returns it in the proper (deserialized) record type.
func (dec *RecordDecoder) Read(r Peeker) (Record, int, error) {
	var peek = 1024
	var skip = 0

	for {
		bytes, peekErr := r.Peek(peek)
		if len(bytes) == 0 {
			return nil, 0, peekErr
		}

		// Skip padding
		if bytes[0] == 0 {
			var n = 1
			var l = len(bytes)
			for n < l && bytes[n] == 0 {
				n++
			}

			d, err := r.Discard(n)
			skip += d

			if err != nil {
				return nil, skip, err
			}

			continue
		}

		rec, n, err := dec.Deserialize(bytes)
		switch err {
		case nil:
			d, err := r.Discard(n)
			return rec, skip + d, err
		case io.ErrShortBuffer:
			if peekErr != nil && peekErr != io.EOF {
				return nil, skip, peekErr
			}
			if len(bytes) < peek {
				return nil, skip, io.ErrUnexpectedEOF
			}
			peek *= 2
		default:
			return nil, skip, err
		}
	}
}

// SerializeRecord serializes r and returns its byte representation.
func SerializeRecord(r Record, e Encoding) ([]byte, error) {
	return NewRecordEncoder(e).Serialize(r)
}

// DeserializeRecord reads exactly one record from b and returns it in the proper (deserialized) record type.
func DeserializeRecord(b []byte, e Encoding) (Record, int, error) {
	return NewRecordDecoder(e, nil).Deserialize(b)
}

// ReadRecord reads one record from r and returns it in the proper (deserialized) record type.
func ReadRecord(r Peeker, e Encoding) (Record, int, error) {
	return NewRecordDecoder(e, nil).Read(r)
}

// WriteRecord serializes r and writes it to w.
func WriteRecord(w io.Writer, r Record, e Encoding) (int, error) {
	return NewRecordEncoder(e).Write(w, r)
}
