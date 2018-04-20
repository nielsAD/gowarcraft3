// Author:  Niels A.D.
// Project: gowarcraft3 (https://github.com/nielsAD/gowarcraft3)
// License: Mozilla Public License, v2.0

package w3m_test

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/nielsAD/gowarcraft3/file/w3m"
)

func TestLoadMap(t *testing.T) {
	var cROC = w3m.Map{
		FileName:         "./test_roc.w3m",
		FileFormat:       18,
		SaveCount:        2,
		EditorVersion:    6059,
		Name:             "",
		Author:           "DragonX",
		Description:      "Smallest map in W3",
		SuggestedPlayers: "Any",
		CamBounds:        [8]float32{-768, -1280, 768, 768, -768, 768, 768, -1280},
		CamBoundsCompl:   [4]uint32{6, 6, 4, 8},
		Width:            20,
		Height:           20,
		Flags:            w3m.MapFlagRevealTerrain | w3m.MapFlagWaterWavesOnCliffShoes | w3m.MapFlagWaterWavesOnSlopeShores | w3m.MapFlags(0x8400),
		Tileset:          w3m.TileLordaeronSummer,
		LsBackground:     0xFFFFFFFF,
		Players: []w3m.Player{
			w3m.Player{
				Type:      w3m.PlayerHuman,
				Race:      w3m.RaceHuman,
				Name:      "Player 1",
				StartPosX: -640,
				StartPosY: 320,
			},
		},
		Forces: []w3m.Force{
			w3m.Force{
				PlayerSet: 0xFFFFFFFF,
				Name:      "Force 1",
			},
		},
		CustomUpgradeAvailabilities: []w3m.CustomUpgradeAvailability{},
		CustomTechAvailabilities:    []w3m.CustomTechAvailability{},
	}
	var cTFT = w3m.Map{
		FileName:         "./test_tft.w3x",
		FileFormat:       25,
		SaveCount:        14,
		EditorVersion:    6059,
		Name:             "Small Wars",
		Author:           "Rorslae",
		Description:      "Needs 2 people to play, both teams should be evenly balanced.",
		SuggestedPlayers: "2",
		CamBounds:        [8]float32{-1408, -1664, 1408, 1152, -1408, 1152, 1408, -1664},
		CamBoundsCompl:   [4]uint32{1, 1, 1, 5},
		Width:            30,
		Height:           26,
		Flags:            w3m.MapFlagRevealTerrain | w3m.MapFlagWaterWavesOnCliffShoes | w3m.MapFlagWaterWavesOnSlopeShores | w3m.MapFlags(0xC400),
		Tileset:          w3m.TileLordaeronSummer,
		LsBackground:     0xFFFFFFFF,
		FogStart:         3000,
		FogEnd:           5000,
		FogDensity:       0.5,
		FogColor:         0x0FF000000,
		WaterColor:       0xFFFFFFFF,
		Players: []w3m.Player{
			w3m.Player{
				Type:      w3m.PlayerHuman,
				Race:      w3m.RaceHuman,
				Flags:     w3m.PlayerFlagFixedPos,
				StartPosX: -1664,
				StartPosY: 1152,
			},
			w3m.Player{
				ID:        1,
				Type:      w3m.PlayerComputer,
				Race:      w3m.RaceNightElf,
				Flags:     w3m.PlayerFlagFixedPos,
				Name:      "Player 2",
				StartPosX: 1280,
				StartPosY: -1664,
			},
		},
		Forces: []w3m.Force{
			w3m.Force{
				PlayerSet: 0xFFFFFFFF,
				Name:      "Force 1",
			},
		},
		CustomUpgradeAvailabilities: []w3m.CustomUpgradeAvailability{},
		CustomTechAvailabilities:    []w3m.CustomTechAvailability{},
	}

	roc, err := w3m.Load("./test_roc.w3m")
	if err != nil {
		t.Fatal("test_roc.w3m", err)
	}

	if !reflect.DeepEqual(&cROC, roc) {
		t.Log(fmt.Sprintf("%+v\n", *roc))
		t.Fatal("Load return value not deep equal (ROC)")
	}

	if roc.Size() != w3m.SizeTiny {
		t.Fatal("SizeTiny expected")
	}

	tft, err := w3m.Load("./test_tft.w3x")
	if err != nil {
		t.Fatal("test_tft.w3x", err)
	}

	if !reflect.DeepEqual(&cTFT, tft) {
		t.Log(fmt.Sprintf("%+v\n", *tft))
		t.Fatal("Load return value not deep equal (TFT)")
	}

	if tft.Size() != w3m.SizeTiny {
		t.Fatal("SizeTiny expected")
	}
}
