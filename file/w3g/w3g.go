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
	"io"
	"os"

	"github.com/nielsAD/gowarcraft3/protocol/w3gs"
)

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
	e, err := NewEncoder(w, r.Encoding())
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
			if v.Type != w3gs.MsgChatExtra {
				break
			}

			var cpy = *v
			res.Records = append(res.Records, &cpy)
		case *PlayerLeft:
			var cpy = *v
			res.Records = append(res.Records, &cpy)
		case *Desync:
			var cpy = *v
			cpy.PlayersInState = append(([]byte)(nil), cpy.PlayersInState...)

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
