// Author:  Niels A.D.
// Project: gowarcraft3 (https://github.com/nielsAD/gowarcraft3)
// License: Mozilla Public License, v2.0

package w3g

import (
	"io"
)

// RecordEncoder keeps amortized allocs at 0 for repeated Record.Serialize calls
type RecordEncoder struct {
	s Stream
}

// NewRecordEncoder initialization
func NewRecordEncoder(s Stream) *RecordEncoder {
	return &RecordEncoder{s: s}
}

// Serialize record
func (enc *RecordEncoder) Serialize(w io.Writer, r Record) (int, error) {
	enc.s.Truncate()
	if err := r.Serialize(&enc.s); err != nil {
		return 0, err
	}
	return w.Write(enc.s.Bytes)
}

// SerializeRecord serializes p and writes it to w.
func SerializeRecord(w io.Writer, r Record) (int, error) {
	return (&RecordEncoder{}).Serialize(w, r)
}

// RecordDecoder keeps amortized allocs at 0 for repeated Record.Deserialize calls
// Records are valid until the next Deserialize() call
type RecordDecoder struct {
	s Stream

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
func NewRecordDecoder(s Stream) *RecordDecoder {
	return &RecordDecoder{s: s}
}

// DeserializeBytes reads exactly one record from b and returns it in the proper (deserialized) record type.
func (dec *RecordDecoder) DeserializeBytes(b []byte) (Record, int, error) {
	dec.s.Bytes = b

	var size = dec.s.Size()
	if size < 2 {
		return nil, 0, io.ErrUnexpectedEOF
	}

	var rec Record
	var err error

	// Explicitly call deserialize on type instead of interface for compiler optimizations
	switch b[0] {
	case RidGameInfo:
		err = dec.gameInfo.Deserialize(&dec.s)
		rec = &dec.gameInfo
	case RidPlayerInfo:
		err = dec.playerInfo.Deserialize(&dec.s)
		rec = &dec.playerInfo
	case RidPlayerLeft:
		err = dec.playerLeft.Deserialize(&dec.s)
		rec = &dec.playerLeft
	case RidSlotInfo:
		err = dec.slotInfo.Deserialize(&dec.s)
		rec = &dec.slotInfo
	case RidCountDownStart:
		err = dec.countDownStart.Deserialize(&dec.s)
		rec = &dec.countDownStart
	case RidCountDownEnd:
		err = dec.countDownEnd.Deserialize(&dec.s)
		rec = &dec.countDownEnd
	case RidGameStart:
		err = dec.gameStart.Deserialize(&dec.s)
		rec = &dec.gameStart
	case RidTimeSlot, RidTimeSlot2:
		err = dec.timeSlot.Deserialize(&dec.s)
		rec = &dec.timeSlot
	case RidChatMessage:
		if dec.s.ProtocolVersion == 0 || dec.s.ProtocolVersion > 2 {
			err = dec.chatMessage.Deserialize(&dec.s)
			rec = &dec.chatMessage
			break
		}
		fallthrough
	case RidTimeSlotAck:
		err = dec.timeSlotAck.Deserialize(&dec.s)
		rec = &dec.timeSlotAck
	case RidDesync:
		err = dec.desync.Deserialize(&dec.s)
		rec = &dec.desync
	case RidEndTimer:
		err = dec.endTimer.Deserialize(&dec.s)
		rec = &dec.endTimer
	default:
		err = ErrUnknownRecord
	}

	var n = size - dec.s.Size()
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

// Deserialize reads exactly one record from r and returns it in the proper (deserialized) record type.
func (dec *RecordDecoder) Deserialize(r Peeker) (Record, int, error) {
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

		rec, n, err := dec.DeserializeBytes(bytes)
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

// DeserializeRecordBytes reads exactly one record from b and returns it in the proper (deserialized) record type.
func DeserializeRecordBytes(b []byte) (Record, int, error) {
	return (&RecordDecoder{}).DeserializeBytes(b)
}

// DeserializeRecord reads exactly one record from r and returns it in the proper (deserialized) record type.
func DeserializeRecord(r Peeker) (Record, int, error) {
	return (&RecordDecoder{}).Deserialize(r)
}
