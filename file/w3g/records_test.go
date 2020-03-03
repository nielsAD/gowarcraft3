// Author:  Niels A.D.
// Project: gowarcraft3 (https://github.com/nielsAD/gowarcraft3)
// License: Mozilla Public License, v2.0
package w3g_test

import (
	"bufio"
	"bytes"
	"io"
	"reflect"
	"testing"

	"github.com/nielsAD/gowarcraft3/file/w3g"
	"github.com/nielsAD/gowarcraft3/protocol"
	"github.com/nielsAD/gowarcraft3/protocol/w3gs"
)

var ts = w3g.TimeSlot{TimeSlot: w3gs.TimeSlot{
	TimeIncrementMS: 100,
	Actions: func() []w3gs.PlayerAction {
		var res []w3gs.PlayerAction
		for i := 0; i < 24; i++ {
			res = append(res, w3gs.PlayerAction{
				PlayerID: byte(i),
				Data:     []byte{2, 3, 4, 5, 6},
			})
		}
		return res
	}(),
}}

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
		&ts,
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
		&w3g.PlayerExtra{},
		&w3g.PlayerExtra{
			PlayerExtra: w3gs.PlayerExtra{
				Type: w3gs.PlayerProfile,
				Profiles: []w3gs.PlayerDataProfile{
					w3gs.PlayerDataProfile{
						PlayerID:  1,
						BattleTag: "niels#1234",
						Clan:      "clan",
						Portrait:  "p051",
						Realm:     20,
						Unknown1:  "",
					},
					w3gs.PlayerDataProfile{
						PlayerID:  2,
						BattleTag: "moon#56789",
						Clan:      "clan",
						Portrait:  "p055",
						Realm:     10,
						Unknown1:  "",
					},
				},
			},
		},
		&w3g.PlayerExtra{
			PlayerExtra: w3gs.PlayerExtra{
				Type: w3gs.PlayerSkins,
				Skins: []w3gs.PlayerDataSkins{
					w3gs.PlayerDataSkins{
						PlayerID: 3,
						Skins: []w3gs.PlayerDataSkin{
							w3gs.PlayerDataSkin{
								Unit:       1164207469,
								Skin:       1164207462,
								Collection: "w3-standard",
							},
							w3gs.PlayerDataSkin{
								Unit:       1164666213,
								Skin:       1164665701,
								Collection: "w3-sow-skins",
							},
						},
					},
					w3gs.PlayerDataSkins{
						PlayerID: 4,
						Skins: []w3gs.PlayerDataSkin{
							w3gs.PlayerDataSkin{
								Unit:       1432642913,
								Skin:       1432642918,
								Collection: "w3-standard",
							},
							w3gs.PlayerDataSkin{
								Unit:       1332109682,
								Skin:       1332114536,
								Collection: "w3-sow-skins",
							},
						},
					},
				},
			},
		},
	}

	for _, rec := range types {
		var err error
		var buf = protocol.Buffer{}
		var enc = w3g.Encoding{}

		if err = rec.Serialize(&buf, &enc); err != nil {
			t.Log(reflect.TypeOf(rec))
			t.Fatal(err)
		}

		var buf2 = protocol.Buffer{}
		if _, err = w3g.WriteRecord(&buf2, rec, enc); err != nil {
			t.Log(reflect.TypeOf(rec))
			t.Fatal(err)
		}

		if bytes.Compare(buf.Bytes, buf2.Bytes) != 0 {
			t.Fatalf("encoder.Write != record.Serialize %v", reflect.TypeOf(rec))
		}

		var rec2, _, e = w3g.ReadRecord(bufio.NewReader(&buf), enc)
		if e != nil {
			t.Log(reflect.TypeOf(rec))
			t.Fatal(e)
		}
		if buf.Size() > 0 {
			t.Fatalf("decoder.Read size mismatch for %v", reflect.TypeOf(rec))
		}
		if reflect.TypeOf(rec2) != reflect.TypeOf(rec) {
			t.Fatalf("decoder.Read type mismatch %v != %v", reflect.TypeOf(rec2), reflect.TypeOf(rec))
		}
		if !reflect.DeepEqual(rec, rec2) {
			t.Logf("I: %+v", rec)
			t.Logf("O: %+v", rec2)
			t.Errorf("decoder.Read value mismatch for %v", reflect.TypeOf(rec))
		}

		err = rec.Deserialize(&protocol.Buffer{}, &enc)
		if err != io.ErrShortBuffer {
			t.Fatalf("ErrShortBuffer expected for %v", reflect.TypeOf(rec))
		}
	}
}
