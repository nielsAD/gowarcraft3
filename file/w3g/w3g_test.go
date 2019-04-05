// Author:  Niels A.D.
// Project: gowarcraft3 (https://github.com/nielsAD/gowarcraft3)
// License: Mozilla Public License, v2.0

package w3g_test

import (
	"encoding/base64"
	"fmt"
	"reflect"
	"testing"

	"github.com/nielsAD/gowarcraft3/protocol/w3gs"

	"github.com/nielsAD/gowarcraft3/file/w3g"
)

func b64(s string) []byte {
	b, _ := base64.StdEncoding.DecodeString(s)
	return b
}

func TestFiles(t *testing.T) {
	var files = []struct {
		file   string
		replay w3g.Replay
	}{
		{
			"test_102.w3g",
			w3g.Replay{
				Header: w3g.Header{
					GameVersion: w3gs.GameVersion{Product: w3gs.ProductROC, Version: 2},
					BuildNumber: 4531,
					DurationMS:  441925,
				},
				GameInfo: w3g.GameInfo{
					HostPlayer: w3g.PlayerInfo{
						ID:   1,
						Name: "Go4WC3.Sapor",
					},
					GameName: "final",
					GameSettings: w3gs.GameSettings{
						GameSettingFlags: w3gs.SettingSpeedFast | w3gs.SettingTerrainDefault | w3gs.SettingObsFull | w3gs.SettingTeamsTogether | w3gs.SettingTeamsFixed,
						MapWidth:         124,
						MapHeight:        124,
						MapXoro:          325121041,
						MapPath:          "Maps\\(4)LostTemple.w3m",
						HostName:         "Go4WC3.Sapor",
					},
					GameFlags:  w3gs.GameFlagOfficialGame | w3gs.GameFlagPrivateGame,
					NumSlots:   12,
					LanguageID: 7206592,
				},
				SlotInfo: w3g.SlotInfo{
					SlotInfo: w3gs.SlotInfo{
						Slots: []w3gs.SlotData{
							w3gs.SlotData{
								PlayerID:       2,
								DownloadStatus: 100,
								SlotStatus:     w3gs.SlotOccupied,
								Team:           12,
								Color:          12,
								Race:           w3gs.RaceRandom,
								ComputerType:   w3gs.ComputerNormal,
								Handicap:       100,
							}},
						RandomSeed: 45792916,
						SlotLayout: w3gs.LayoutMelee,
						NumPlayers: 4,
					},
				},
				Players: []w3g.PlayerInfo{
					w3g.PlayerInfo{
						ID:   2,
						Name: "Go4WC3.Desann",
					},
				},
				Records: []w3g.Record{
					&w3g.TimeSlot{w3gs.TimeSlot{Fragment: true, TimeIncrementMS: 100}},
					&w3g.TimeSlot{w3gs.TimeSlot{Fragment: true, TimeIncrementMS: 100, Actions: []w3gs.PlayerAction{w3gs.PlayerAction{PlayerID: 10, Data: b64("AAAAFgAAAAAWAAAAAAAWAAAAABYAAAAAABYAAAAAFgAAAAAAFgAAAAAWAAAAAAAWAAAAABYAAAA=")}}}},
					&w3g.TimeSlot{w3gs.TimeSlot{Fragment: true, TimeIncrementMS: 100, Actions: []w3gs.PlayerAction{w3gs.PlayerAction{PlayerID: 9, Data: b64("AAAAABYAAAAAABYAAAAAFgAAAAAAFgAAAAAWAAAAAAAWAAAAABYAAAAAABYAAAAAFgAAAAAAFgAAAA==")}, w3gs.PlayerAction{PlayerID: 10, Data: b64("AAAAABYAAAAAFgAAAAAAFgAAAAAWAA==")}}}},
				},
			},
		},
		{
			"test_126.w3g",
			w3g.Replay{
				Header: w3g.Header{
					GameVersion: w3gs.GameVersion{Product: w3gs.ProductTFT, Version: 26},
					BuildNumber: 6059,
					DurationMS:  237175,
				},
				GameInfo: w3g.GameInfo{
					HostPlayer: w3g.PlayerInfo{
						ID:          1,
						Name:        "ForFunyo",
						Race:        w3gs.RaceUndead,
						JoinCounter: 3052346,
					},
					GameName: "BNet",
					GameSettings: w3gs.GameSettings{
						GameSettingFlags: w3gs.SettingSpeedFast | w3gs.SettingTerrainExplored | w3gs.SettingObsNone | w3gs.SettingTeamsTogether | w3gs.SettingTeamsFixed,
						MapWidth:         0,
						MapHeight:        0,
						MapXoro:          4294967295,
						MapPath:          "Maps\\FrozenThrone\\(4)TurtleRock.w3x",
						HostName:         "Battle.net",
						MapSha1:          [20]byte{0, 145, 213, 254, 70, 60, 116, 93, 232, 133, 235, 135, 140, 210, 168, 35, 212, 189, 5, 97},
					},
					GameFlags:  w3gs.GameFlagCustomGame,
					NumSlots:   2,
					LanguageID: 1636528,
				},
				SlotInfo: w3g.SlotInfo{
					SlotInfo: w3gs.SlotInfo{
						Slots: []w3gs.SlotData{
							w3gs.SlotData{
								PlayerID:       2,
								DownloadStatus: 255,
								SlotStatus:     w3gs.SlotOccupied,
								Team:           1,
								Color:          0,
								Race:           w3gs.RaceUndead,
								ComputerType:   w3gs.ComputerNormal,
								Handicap:       100,
							}},
						RandomSeed: 77005536,
						SlotLayout: w3gs.LayoutLadder,
						NumPlayers: 4,
					},
				},
				Players: []w3g.PlayerInfo{
					w3g.PlayerInfo{
						ID:          2,
						Name:        "Fighting-",
						Race:        w3gs.RaceUndead,
						JoinCounter: 31881612,
					},
				},
				Records: []w3g.Record{
					&w3g.TimeSlot{w3gs.TimeSlot{TimeIncrementMS: 251, Actions: []w3gs.PlayerAction{w3gs.PlayerAction{PlayerID: 1, Data: b64("AAEHAAAABQAAAAAABQAAAAAAhTcAAOiWAAAiBCabVw==")}}}},
					&w3g.TimeSlot{w3gs.TimeSlot{TimeIncrementMS: 250, Actions: []w3gs.PlayerAction{w3gs.PlayerAction{PlayerID: 1, Data: b64("AAkBIgSvhXaPHwIA+gAiBGJpFT4fAgD5ACIEqiwIwQ==")}, w3gs.PlayerAction{PlayerID: 1, Data: b64("+gAiBExiuDsfAgAJASIEIDYoaR8CAAkBIgQECteeHw==")}}}},
					&w3g.TimeSlot{w3gs.TimeSlot{TimeIncrementMS: 250}},
				},
			},
		},
		{
			"test_130.w3g",
			w3g.Replay{
				Header: w3g.Header{
					GameVersion:  w3gs.GameVersion{Product: w3gs.ProductTFT, Version: 10030},
					BuildNumber:  6061,
					DurationMS:   640650,
					SinglePlayer: true,
				},
				GameInfo: w3g.GameInfo{
					HostPlayer: w3g.PlayerInfo{
						ID:   1,
						Name: "niels",
					},
					GameName: "Local Game",
					GameSettings: w3gs.GameSettings{
						GameSettingFlags: w3gs.SettingSpeedFast | w3gs.SettingTerrainDefault | w3gs.SettingObsNone | w3gs.SettingTeamsTogether | w3gs.SettingTeamsFixed,
						MapWidth:         116,
						MapHeight:        84,
						MapXoro:          2599102717,
						MapPath:          "Maps/FrozenThrone//(2)EchoIsles.w3x",
						HostName:         "niels",
						MapSha1:          [20]byte{107, 111, 100, 67, 248, 197, 26, 44, 89, 111, 217, 78, 123, 106, 91, 101, 208, 6, 70, 129},
					},
					GameFlags:  w3gs.GameFlagOfficialGame,
					NumSlots:   24,
					LanguageID: 0,
				},
				SlotInfo: w3g.SlotInfo{
					SlotInfo: w3gs.SlotInfo{
						Slots: []w3gs.SlotData{
							w3gs.SlotData{
								PlayerID:       0,
								DownloadStatus: 100,
								SlotStatus:     w3gs.SlotOccupied,
								Computer:       true,
								Team:           1,
								Color:          1,
								Race:           w3gs.RaceRandom | w3gs.RaceSelectable,
								ComputerType:   w3gs.ComputerNormal,
								Handicap:       100,
							}},
						RandomSeed: 40053178,
						SlotLayout: w3gs.LayoutMelee,
						NumPlayers: 2,
					},
				},
				Players: []w3g.PlayerInfo{
					w3g.PlayerInfo{
						ID:   1,
						Name: "niels",
					},
				},
				Records: []w3g.Record{
					&w3g.TimeSlot{w3gs.TimeSlot{TimeIncrementMS: 0, Actions: []w3gs.PlayerAction{w3gs.PlayerAction{PlayerID: 1, Data: b64("AAQAAAAAAAQAAAAABAAAAAAABAAAAAAEAAAAAAAEAA==")}}}},
					&w3g.TimeSlot{w3gs.TimeSlot{TimeIncrementMS: 100}},
					&w3g.TimeSlot{w3gs.TimeSlot{TimeIncrementMS: 100}},
				},
			},
		},
	}

	for _, f := range files {
		rep, err := w3g.Open("./" + f.file)
		if err != nil {
			t.Fatal("Loading file", err)
		}

		if len(rep.Players) > 1 {
			rep.Players = rep.Players[1:2]
		}
		if len(rep.Slots) > 1 {
			rep.Slots = rep.Slots[1:2]
		}
		rep.Records = rep.Records[20:23]

		if !reflect.DeepEqual(&f.replay, rep) {
			t.Log(fmt.Sprintf("%+v\n", *rep))
			t.Fatal(f.file, "Replay is not deep equal")
		}
	}
}
