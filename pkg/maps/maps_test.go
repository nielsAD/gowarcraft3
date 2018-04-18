package maps_test

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/nielsAD/noot/pkg/maps"
)

func TestLoadMap(t *testing.T) {
	var cW3M = maps.Map{
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
		Flags:            maps.FlagRevealTerrain | maps.FlagWaterWavesOnCliffShoes | maps.FlagWaterWavesOnSlopeShores | maps.Flags(0x8400),
		Tileset:          maps.TileLordaeronSummer,
		LsBackground:     0xFFFFFFFF,
		Slots: []maps.SlotData{
			maps.SlotData{
				Type:      maps.SlotHuman,
				Race:      maps.RaceHuman,
				Name:      "Player 1",
				StartPosX: -640,
				StartPosY: 320,
			},
		},
		Teams: []maps.TeamData{
			maps.TeamData{
				PlayerMask: 0xFFFFFFFF,
				Name:       "Force 1",
			},
		},
	}
	var cW3X = maps.Map{
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
		Flags:            maps.FlagRevealTerrain | maps.FlagWaterWavesOnCliffShoes | maps.FlagWaterWavesOnSlopeShores | maps.Flags(0xC400),
		Tileset:          maps.TileLordaeronSummer,
		LsBackground:     0xFFFFFFFF,
		FogStart:         3000,
		FogEnd:           5000,
		FogDensity:       0.5,
		FogColor:         0x0FF000000,
		WaterColor:       0xFFFFFFFF,
		Slots: []maps.SlotData{
			maps.SlotData{
				Type:      maps.SlotHuman,
				Race:      maps.RaceHuman,
				StartPos:  1,
				StartPosX: -1664,
				StartPosY: 1152,
			},
			maps.SlotData{
				ID:        1,
				Type:      maps.SlotComputer,
				Race:      maps.RaceNightElf,
				StartPos:  1,
				Name:      "Player 2",
				StartPosX: 1280,
				StartPosY: -1664,
			},
		},
		Teams: []maps.TeamData{
			maps.TeamData{
				PlayerMask: 0xFFFFFFFF,
				Name:       "Force 1",
			},
		},
	}

	w3m, err := maps.Load("./test_roc.w3m")
	if err != nil {
		t.Fatal("test_roc.w3m", err)
	}

	if !reflect.DeepEqual(&cW3M, w3m) {
		t.Log(fmt.Sprintf("%+v\n", *w3m))
		t.Fatal("Load return value not deep equal (w3m)")
	}

	if w3m.Size() != maps.SizeTiny {
		t.Fatal("SizeTiny expected")
	}

	w3x, err := maps.Load("./test_tft.w3x")
	if err != nil {
		t.Fatal("test_tft.w3x", err)
	}

	if !reflect.DeepEqual(&cW3X, w3x) {
		t.Log(fmt.Sprintf("%+v\n", *w3x))
		t.Fatal("Load return value not deep equal (w3x)")
	}

	if w3m.Size() != maps.SizeTiny {
		t.Fatal("SizeTiny expected")
	}
}
