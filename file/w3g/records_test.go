// Author:  Niels A.D.
// Project: gowarcraft3 (https://github.com/nielsAD/gowarcraft3)
// License: Mozilla Public License, v2.0
package w3g_test

import (
	"bytes"
	"reflect"
	"testing"

	"github.com/nielsAD/gowarcraft3/file/w3g"
	"github.com/nielsAD/gowarcraft3/protocol"
	"github.com/nielsAD/gowarcraft3/protocol/w3gs"
)

func TestRecords(t *testing.T) {
	var types = []w3g.Record{
		&w3g.GameInfo{},
		&w3g.GameInfo{
			HostPlayer: w3g.PlayerInfo{
				ID:          1,
				Name:        "Niels",
				Race:        w3gs.RaceHuman,
				JoinCounter: 666,
			},
			GameName: "niels (Local Game)",
			GameSettings: w3gs.GameSettings{
				GameSettingFlags: w3gs.SettingSpeedNormal,
				MapWidth:         1,
				MapHeight:        2,
				MapXoro:          3,
				MapPath:          "4",
				HostName:         "5",
			},
			GameFlags:  w3gs.GameFlagCustomGame,
			NumSlots:   12,
			LanguageID: 0x0012F824,
		},
		&w3g.PlayerInfo{},
		&w3g.PlayerInfo{
			ID:          2,
			Name:        "Moon",
			Race:        w3gs.RaceNightElf,
			JoinCounter: 456,
		},
		&w3g.PlayerLeft{},
		&w3g.PlayerLeft{
			Local:    true,
			PlayerID: 3,
			Reason:   w3gs.LeaveLost,
			Counter:  777,
		},
		&w3g.SlotInfo{},
		&w3g.SlotInfo{
			SlotInfo: w3gs.SlotInfo{
				Slots: []w3gs.SlotData{
					w3gs.SlotData{
						PlayerID:       1,
						DownloadStatus: 2,
						SlotStatus:     3,
						Computer:       true,
						Team:           5,
						Color:          6,
						Race:           7,
						ComputerType:   8,
						Handicap:       9,
					},
					w3gs.SlotData{
						PlayerID:       9,
						DownloadStatus: 8,
						SlotStatus:     7,
						Computer:       false,
						Team:           5,
						Color:          4,
						Race:           3,
						ComputerType:   2,
						Handicap:       1,
					},
				},
				RandomSeed: 10,
				SlotLayout: w3gs.LayoutMelee,
				NumPlayers: 12,
			},
		},
		&w3g.CountDownStart{},
		&w3g.CountDownEnd{},
		&w3g.GameStart{},
		&w3g.TimeSlot{},
		&w3g.TimeSlot{
			TimeSlot: w3gs.TimeSlot{
				Fragment:        false,
				TimeIncrementMS: 50,
				Actions: []w3gs.PlayerAction{
					w3gs.PlayerAction{PlayerID: 1, Data: make([]byte, 23)},
					w3gs.PlayerAction{PlayerID: 12, Data: make([]byte, 3)},
				},
			},
		},
		&w3g.ChatMessage{},
		&w3g.ChatMessage{
			Message: w3gs.Message{
				SenderID: 4,
				Type:     w3gs.MsgChatExtra,
				Scope:    w3gs.ScopeAllies,
				Content:  "Pitiful",
			},
		},
		&w3g.TimeSlotAck{},
		&w3g.TimeSlotAck{
			Checksum: []byte{4, 5, 6},
		},
		&w3g.Desync{},
		&w3g.Desync{
			Desync: w3gs.Desync{
				Unknown1:       234,
				Checksum:       567,
				PlayersInState: []uint8{1, 2, 3},
			},
		},
		&w3g.EndTimer{},
		&w3g.EndTimer{
			GameOver:     true,
			CountDownSec: 5,
		},
	}

	for _, rec := range types {
		var err error
		var buf = protocol.Buffer{Bytes: make([]byte, 0, 2048)}

		if err = rec.Serialize(&buf); err != nil {
			t.Log(reflect.TypeOf(rec))
			t.Fatal(err)
		}

		var buf2 = protocol.Buffer{Bytes: make([]byte, 0, 2048)}
		if _, err = w3g.SerializeRecord(&buf2, rec); err != nil {
			t.Log(reflect.TypeOf(rec))
			t.Fatal(err)
		}

		if bytes.Compare(buf.Bytes, buf2.Bytes) != 0 {
			t.Fatalf("SerializeRecord != Record.Serialize %v", reflect.TypeOf(rec))
		}

		var rec2, n, e = w3g.DeserializeRecordRaw(buf.Bytes)
		if e != nil {
			t.Log(reflect.TypeOf(rec))
			t.Fatal(e)
		}
		if n != buf.Size() {
			t.Fatalf("DeserializeRecord size mismatch for %v", reflect.TypeOf(rec))
		}
		if reflect.TypeOf(rec2) != reflect.TypeOf(rec) {
			t.Fatalf("DeserializeRecord type mismatch %v != %v", reflect.TypeOf(rec2), reflect.TypeOf(rec))
		}
		if !reflect.DeepEqual(rec, rec2) {
			t.Logf("I: %+v", rec)
			t.Logf("O: %+v", rec2)
			t.Errorf("DeserializeRecord value mismatch for %v", reflect.TypeOf(rec))
		}

		err = rec.Deserialize(&protocol.Buffer{Bytes: make([]byte, 0)})
		if err != w3g.ErrBufferTooShort {
			t.Fatalf("ErrBufferTooShort expected for %v", reflect.TypeOf(rec))
		}
	}
}
