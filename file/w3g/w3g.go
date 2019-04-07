// Author:  Niels A.D.
// Project: gowarcraft3 (https://github.com/nielsAD/gowarcraft3)
// License: Mozilla Public License, v2.0

// Package w3g implements a decoder and encoder for w3g files.
//
// Format:
//
//    size/type | Description
//   -----------+-----------------------------------------------------------
//    28 chars  | zero terminated string "Warcraft III recorded game\0x1A\0"
//     1 dword  | fileoffset of first compressed data block (header size)
//              |  0x40 for WarCraft III with patch <= v1.06
//              |  0x44 for WarCraft III patch >= 1.07 and TFT replays
//     1 dword  | overall size of compressed file
//     1 dword  | replay header version:
//              |  0x00 for WarCraft III with patch <= 1.06
//              |  0x01 for WarCraft III patch >= 1.07 and TFT replays
//     1 dword  | overall size of decompressed data (excluding header)
//     1 dword  | number of compressed data blocks in file
//
//   * replay header version 0x00:
//        1  word  | unknown (always zero so far)
//        1  word  | version number (corresponds to patch 1.xx)
//   * replay header version 0x01:
//        1 dword  | version identifier string reading:
//                 |  'WAR3' for WarCraft III Classic
//                 |  'W3XP' for WarCraft III Expansion Set 'The Frozen Throne'
//        1 dword  | version number (corresponds to patch 1.xx so far)
//
//     1  word  | build number
//     1  word  | flags
//              |   0x0000 for single player games
//              |   0x8000 for multiplayer games (LAN or Battle.net)
//     1 dword  | replay length in msec
//     1 dword  | CRC32 checksum for the header
//              | (the checksum is calculated for the complete header
//              |  including this field which is set to zero)
//
//   For each data block:
//     1  word  | size n of compressed data block (excluding header)
//     1  word  | size of decompressed data block (currently 8k)
//     1  word  | CRC checksum for the header
//     1  word  | CRC checksum for the compressed block
//     n bytes  | compressed data (using zlib)
//
package w3g

import (
	"bufio"
	"bytes"
	"hash/crc32"
	"io"
	"io/ioutil"
	"os"

	"github.com/nielsAD/gowarcraft3/protocol"
	"github.com/nielsAD/gowarcraft3/protocol/w3gs"
)

// Record interface.
type Record interface {
	Serialize(buf *protocol.Buffer) error
	Deserialize(buf *protocol.Buffer) error
}

// Header for a Warcraft III recorded game file
type Header struct {
	GameVersion  w3gs.GameVersion
	BuildNumber  uint16
	DurationMS   uint32
	SinglePlayer bool
}

// Replay information for  Warcraft III recorded game
type Replay struct {
	Header
	GameInfo
	SlotInfo
	Players []*PlayerInfo
	Records []Record
}

// Open a w3g file
func Open(name string) (*Replay, error) {
	f, err := os.Open(name)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var b = bufio.NewReaderSize(f, 8192)
	if _, err := FindHeader(b); err != nil {
		return nil, ErrBadFormat
	}

	rep, err := Decode(b)
	return rep, err
}

// Save a w3g file
func (r *Replay) Save(name string) error {
	f, err := os.Create(name)
	if err != nil {
		return err
	}
	defer f.Close()

	return r.Encode(f)
}

// Encode to w
func (r *Replay) Encode(w io.Writer) error {
	e, err := NewEncoder(w)
	if err != nil {
		return err
	}

	if _, err := e.WriteRecords(&r.GameInfo, &r.SlotInfo); err != nil {
		return err
	}
	for _, p := range r.Players {
		if p.ID == r.HostPlayer.ID {
			// Skip host
			continue
		}
		if _, err := e.WriteRecord(p); err != nil {
			return err
		}
	}
	if _, err := e.WriteRecords(&CountDownStart{}, &CountDownEnd{}, &GameStart{}); err != nil {
		return err
	}

	if _, err := e.WriteRecords(r.Records...); err != nil {
		return err
	}

	e.Header = r.Header
	return e.Close()
}

// FindHeader in r
func FindHeader(r Peeker) (int, error) {
	var n = 0

	for {
		b, err := r.Peek(2048)
		if len(b) == 0 {
			return n, err
		}

		var idx = bytes.Index(b, []byte(Signature))
		var del = idx
		if del < 0 {
			del = 2048 - len(Signature) + 1
		}
		nn, err := r.Discard(del)
		n += nn

		if idx >= 0 || err != nil {
			return n, err
		}
	}
}

// DecodeHeader a w3g file, returns header and a Decompressor to read compressed records
func DecodeHeader(r io.Reader) (*Header, *Decompressor, int, error) {
	var buf [68]byte
	var hdr Header

	n, err := io.ReadFull(r, buf[:64])
	if err != nil {
		return nil, nil, n, err
	}

	var pbuf = protocol.Buffer{Bytes: buf[:]}
	if s, err := pbuf.ReadCString(); err != nil {
		return nil, nil, n, err
	} else if s != Signature {
		return nil, nil, n, ErrBadFormat
	}

	var sizeHeader = pbuf.ReadUInt32()

	// File size
	pbuf.Skip(4)

	var headerVersion = pbuf.ReadUInt32()
	switch headerVersion {
	case 0:
	case 1:
		nn, err := io.ReadFull(r, buf[64:68])
		n += nn
		if err != nil {
			return nil, nil, n, err
		}

	default:
		return nil, nil, n, ErrUnexpectedConst
	}

	var sizeBlocks = pbuf.ReadUInt32()
	var numBlocks = pbuf.ReadUInt32()

	switch headerVersion {
	case 0:
		if pbuf.ReadUInt16() != 0 {
			return nil, nil, n, ErrUnexpectedConst
		}
		hdr.GameVersion.Product = w3gs.ProductROC
		hdr.GameVersion.Version = uint32(pbuf.ReadUInt16())
	case 1:
		hdr.GameVersion.DeserializeContent(&pbuf)
	}

	hdr.BuildNumber = pbuf.ReadUInt16()
	hdr.SinglePlayer = pbuf.ReadUInt16() == 0
	hdr.DurationMS = pbuf.ReadUInt32()

	var crc = pbuf.ReadUInt32()
	buf[n-4], buf[n-3], buf[n-2], buf[n-1] = 0, 0, 0, 0
	if crc != uint32(crc32.ChecksumIEEE(buf[0:n])) {
		return nil, nil, n, ErrInvalidChecksum
	}

	if uint32(n) > sizeHeader {
		return nil, nil, n, ErrBadFormat
	}

	// Skip to start of data section
	nn, err := io.CopyN(ioutil.Discard, r, int64(sizeHeader-uint32(n)))
	n += int(nn)
	if err != nil {
		return nil, nil, n, err
	}

	return &hdr, NewDecompressor(r, numBlocks, sizeBlocks), n, err
}

// Decode a w3g file
func Decode(r io.Reader) (*Replay, error) {
	hdr, data, _, err := DecodeHeader(r)
	if err != nil {
		return nil, err
	}

	var res = Replay{Header: *hdr}
	if err := data.ForEach(func(r Record) error {
		switch v := r.(type) {
		case *GameInfo:
			res.GameInfo = *v
			res.Players = []*PlayerInfo{&res.GameInfo.HostPlayer}
		case *SlotInfo:
			res.SlotInfo = *v
		case *PlayerInfo:
			var cpy = *v
			res.Players = append(res.Players, &cpy)
		case *TimeSlot:
			var cpy = *v

			cpy.Actions = nil
			for _, a := range v.Actions {
				cpy.Actions = append(cpy.Actions, w3gs.PlayerAction{
					PlayerID: a.PlayerID,
					Data:     append(([]byte)(nil), a.Data...),
				})
			}

			res.Records = append(res.Records, &cpy)
		case *ChatMessage:
			var cpy = *v
			res.Records = append(res.Records, &cpy)
		case *PlayerLeft:
			var cpy = *v
			res.Records = append(res.Records, &cpy)
		case *Desync:
			var cpy = *v
			res.Records = append(res.Records, &cpy)
		case *EndTimer:
			var cpy = *v
			res.Records = append(res.Records, &cpy)
		}
		return nil
	}); err != nil {
		return nil, err
	}

	if len(res.SlotInfo.Slots) == 0 {
		for i, p := range res.Players {
			res.SlotInfo.NumPlayers++
			res.SlotInfo.Slots = append(res.SlotInfo.Slots, w3gs.SlotData{
				PlayerID:       uint8(i + 1),
				DownloadStatus: 100,
				SlotStatus:     w3gs.SlotOccupied,
				Computer:       false,
				Team:           uint8(i % 2),
				Color:          uint8(i),
				Race:           p.Race,
				ComputerType:   w3gs.ComputerNormal,
				Handicap:       100,
			})
		}
	}

	return &res, nil
}

// Encoder compresses records and updates header on close
type Encoder struct {
	Header
	*BufferedCompressor

	b protocol.Buffer
	w io.Writer
}

// NewEncoder for replay file
func NewEncoder(w io.Writer) (*Encoder, error) {
	var res = Encoder{
		w: w,
	}

	if _, ok := w.(io.Seeker); ok {
		// Write placeholder for header
		var h [68]byte
		if _, err := w.Write(h[:]); err != nil {
			return nil, err
		}

		res.BufferedCompressor = NewBufferedCompressor(w)
	} else {
		res.BufferedCompressor = NewBufferedCompressor(&res.b)
	}

	return &res, nil
}

// Close writer, flush data, and update header.
// Does not close underlying writer.
func (e *Encoder) Close() error {
	if err := e.BufferedCompressor.Close(); err != nil {
		return err
	}

	var buf [68]byte
	var pbuf = protocol.Buffer{Bytes: buf[:0]}
	pbuf.WriteCString(Signature)
	pbuf.WriteUInt32(68)
	pbuf.WriteUInt32(e.SizeWritten + 68)
	pbuf.WriteUInt32(1)
	pbuf.WriteUInt32(e.SizeTotal)
	pbuf.WriteUInt32(e.NumBlocks)
	e.GameVersion.SerializeContent(&pbuf)
	pbuf.WriteUInt16(e.BuildNumber)
	if e.SinglePlayer {
		pbuf.WriteUInt16(0x0000)
	} else {
		pbuf.WriteUInt16(0x8000)
	}
	pbuf.WriteUInt32(e.DurationMS)
	pbuf.WriteUInt32(0)
	pbuf.WriteUInt32At(64, crc32.ChecksumIEEE(pbuf.Bytes))

	s, seeker := e.w.(io.Seeker)
	if seeker {
		// Seek to beginning
		if _, err := s.Seek(-int64(e.BufferedCompressor.SizeWritten+68), io.SeekCurrent); err != nil {
			return err
		}
		// Overwrite header
		if n, err := e.w.Write(pbuf.Bytes); err != nil {
			s.Seek(-int64(e.BufferedCompressor.SizeWritten+68-uint32(n)), io.SeekCurrent)
			return err
		}
		// Seek to end
		if _, err := s.Seek(int64(e.BufferedCompressor.SizeWritten), io.SeekCurrent); err != nil {
			return err
		}
		return nil
	}

	if _, err := e.w.Write(pbuf.Bytes); err != nil {
		return err
	}
	if _, err := e.w.Write(e.b.Bytes); err != nil {
		return err
	}

	return nil
}
