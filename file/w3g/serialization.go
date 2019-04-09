// Author:  Niels A.D.
// Project: gowarcraft3 (https://github.com/nielsAD/gowarcraft3)
// License: Mozilla Public License, v2.0

package w3g

import (
	"io"

	"github.com/nielsAD/gowarcraft3/protocol"
)

// SerializationBuffer is used by SerializeRecordWithBuffer to bring amortized allocs to 0 for repeated calls
type SerializationBuffer = protocol.Buffer

// SerializeRecordWithBuffer serializes r and writes it to w.
func SerializeRecordWithBuffer(w io.Writer, b *SerializationBuffer, r Record) (int, error) {
	b.Truncate()
	if err := r.Serialize(b); err != nil {
		return 0, err
	}
	return w.Write(b.Bytes)
}

// SerializeRecord serializes p and writes it to w.
func SerializeRecord(w io.Writer, r Record) (int, error) {
	return SerializeRecordWithBuffer(w, &SerializationBuffer{}, r)
}

// DeserializationBuffer is used by DeserializeRecordWithBuffer to bring amortized allocs to 0 for repeated calls
type DeserializationBuffer struct {
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

// DeserializeRecordWithBufferRaw reads exactly one record from buf and returns it in the proper (deserialized) record type.
func DeserializeRecordWithBufferRaw(r []byte, b *DeserializationBuffer) (Record, int, error) {
	var pbuf = protocol.Buffer{Bytes: r}
	if pbuf.Size() < 2 {
		return nil, 0, io.ErrUnexpectedEOF
	}

	var rec Record
	var err error

	// Explicitly call deserialize on type instead of interface for compiler optimizations
	switch r[0] {
	case RidGameInfo:
		err = b.gameInfo.Deserialize(&pbuf)
		rec = &b.gameInfo
	case RidPlayerInfo:
		err = b.playerInfo.Deserialize(&pbuf)
		rec = &b.playerInfo
	case RidPlayerLeft:
		err = b.playerLeft.Deserialize(&pbuf)
		rec = &b.playerLeft
	case RidSlotInfo:
		err = b.slotInfo.Deserialize(&pbuf)
		rec = &b.slotInfo
	case RidCountDownStart:
		err = b.countDownStart.Deserialize(&pbuf)
		rec = &b.countDownStart
	case RidCountDownEnd:
		err = b.countDownEnd.Deserialize(&pbuf)
		rec = &b.countDownEnd
	case RidGameStart:
		err = b.gameStart.Deserialize(&pbuf)
		rec = &b.gameStart
	case RidTimeSlot, RidTimeSlot2:
		err = b.timeSlot.Deserialize(&pbuf)
		rec = &b.timeSlot
	case RidChatMessage:
		err = b.chatMessage.Deserialize(&pbuf)
		rec = &b.chatMessage
		if err == nil || r[1] != 4 {
			break
		}

		// On patch version <= 1.02 this is a TimeSlotAck, so try
		// fallthrough if deserialization as ChatMessage failed.
		pbuf.Bytes = r
		fallthrough
	case RidTimeSlotAck:
		err = b.timeSlotAck.Deserialize(&pbuf)
		rec = &b.timeSlotAck
	case RidDesync:
		err = b.desync.Deserialize(&pbuf)
		rec = &b.desync
	case RidEndTimer:
		err = b.endTimer.Deserialize(&pbuf)
		rec = &b.endTimer
	default:
		err = ErrUnknownRecord
	}

	var n = len(r) - len(pbuf.Bytes)
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

// DeserializeRecordWithBuffer reads exactly one record from r and returns it in the proper (deserialized) record type.
func DeserializeRecordWithBuffer(r Peeker, b *DeserializationBuffer) (Record, int, error) {
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

		rec, n, err := DeserializeRecordWithBufferRaw(bytes, b)
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

// DeserializeRecordRaw reads exactly one record from r and returns it in the proper (deserialized) record type.
func DeserializeRecordRaw(r []byte) (Record, int, error) {
	return DeserializeRecordWithBufferRaw(r, &DeserializationBuffer{})
}

// DeserializeRecord reads exactly one record from r and returns it in the proper (deserialized) record type.
func DeserializeRecord(r Peeker) (Record, int, error) {
	return DeserializeRecordWithBuffer(r, &DeserializationBuffer{})
}
