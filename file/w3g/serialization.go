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

// SerializeRecord serializes r and returns its byte representation.
func SerializeRecord(r Record, e Encoding) ([]byte, error) {
	return NewRecordEncoder(e).Serialize(r)
}

// WriteRecord serializes r and writes it to w.
func WriteRecord(w io.Writer, r Record, e Encoding) (int, error) {
	return NewRecordEncoder(e).Write(w, r)
}

// RecordDecoder keeps amortized allocs at 0 for repeated Record.Deserialize calls.
// Records are valid until the next Deserialize() call
type RecordDecoder struct {
	Encoding
	buf protocol.Buffer

	gameInfo       GameInfo
	playerInfo     PlayerInfo
	playerLeft     PlayerLeft
	slotInfo       SlotInfo
	countDownStart CountDownStart
	countDownEnd   CountDownEnd
	gameStart      GameStart
	timeSlot       TimeSlot
	chatMessage    ChatMessage
	timeSlotAck    TimeSlotAck
	desync         Desync
	endTimer       EndTimer
}

// NewRecordDecoder initialization
func NewRecordDecoder(e Encoding) *RecordDecoder {
	return &RecordDecoder{
		Encoding: e,
	}
}

// Deserialize reads exactly one record from b and returns it in the proper (deserialized) record type.
// Result is valid until the next Deserialize() call.
func (dec *RecordDecoder) Deserialize(b []byte) (Record, int, error) {
	dec.buf.Reset(b)

	var size = dec.buf.Size()
	if size < 2 {
		return nil, 0, io.ErrUnexpectedEOF
	}

	var rec Record
	var err error

	// Explicitly call deserialize on type instead of interface for compiler optimizations
	switch b[0] {
	case RidGameInfo:
		err = dec.gameInfo.Deserialize(&dec.buf, &dec.Encoding)
		rec = &dec.gameInfo
	case RidPlayerInfo:
		err = dec.playerInfo.Deserialize(&dec.buf, &dec.Encoding)
		rec = &dec.playerInfo
	case RidPlayerLeft:
		err = dec.playerLeft.Deserialize(&dec.buf, &dec.Encoding)
		rec = &dec.playerLeft
	case RidSlotInfo:
		err = dec.slotInfo.Deserialize(&dec.buf, &dec.Encoding)
		rec = &dec.slotInfo
	case RidCountDownStart:
		err = dec.countDownStart.Deserialize(&dec.buf, &dec.Encoding)
		rec = &dec.countDownStart
	case RidCountDownEnd:
		err = dec.countDownEnd.Deserialize(&dec.buf, &dec.Encoding)
		rec = &dec.countDownEnd
	case RidGameStart:
		err = dec.gameStart.Deserialize(&dec.buf, &dec.Encoding)
		rec = &dec.gameStart
	case RidTimeSlot, RidTimeSlot2:
		err = dec.timeSlot.Deserialize(&dec.buf, &dec.Encoding)
		rec = &dec.timeSlot
	case RidChatMessage:
		if dec.GameVersion == 0 || dec.GameVersion > 2 {
			err = dec.chatMessage.Deserialize(&dec.buf, &dec.Encoding)
			rec = &dec.chatMessage
			break
		}
		fallthrough
	case RidTimeSlotAck:
		err = dec.timeSlotAck.Deserialize(&dec.buf, &dec.Encoding)
		rec = &dec.timeSlotAck
	case RidDesync:
		err = dec.desync.Deserialize(&dec.buf, &dec.Encoding)
		rec = &dec.desync
	case RidEndTimer:
		err = dec.endTimer.Deserialize(&dec.buf, &dec.Encoding)
		rec = &dec.endTimer
	default:
		err = ErrUnknownRecord
	}

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

			if d, err := r.Discard(n); err != nil {
				return nil, d, err
			}
			continue
		}

		rec, n, err := dec.Deserialize(bytes)
		switch err {
		case nil:
			if d, err := r.Discard(n); err != nil {
				return nil, d, err
			}
			return rec, n, nil
		case io.ErrShortBuffer:
			if peekErr != nil {
				return nil, 0, peekErr
			}
			if len(bytes) < peek {
				return nil, 0, io.ErrUnexpectedEOF
			}
			peek *= 2
		default:
			return nil, 0, err
		}
	}
}

// DeserializeRecord reads exactly one record from b and returns it in the proper (deserialized) record type.
func DeserializeRecord(b []byte, e Encoding) (Record, int, error) {
	return NewRecordDecoder(e).Deserialize(b)
}

// ReadRecord reads one record from r and returns it in the proper (deserialized) record type.
func ReadRecord(r Peeker, e Encoding) (Record, int, error) {
	return NewRecordDecoder(e).Read(r)
}
