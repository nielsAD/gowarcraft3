// Author:  Niels A.D.
// Project: gowarcraft3 (https://github.com/nielsAD/gowarcraft3)
// License: Mozilla Public License, v2.0

package w3m_test

import (
	"bytes"
	"crypto/sha512"
	"encoding/base64"
	"fmt"
	"image"
	"image/png"
	"reflect"
	"testing"

	"github.com/nielsAD/gowarcraft3/file/w3m"
)

func shaImage(img image.Image) string {
	if img == nil {
		return ""
	}

	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		return ""
	}

	var sha = sha512.Sum512(buf.Bytes())
	return base64.RawStdEncoding.EncodeToString(sha[:])
}

func TestFiles(t *testing.T) {
	var files = []struct {
		file       string
		signed     bool
		size       w3m.Size
		info       w3m.Info
		previewSHA string
		minimapSHA string
		checksum   string
	}{
		{
			"test_roc.w3m",
			false,
			w3m.SizeTiny,
			w3m.Info{
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
				Flags:            w3m.MapFlagRevealTerrain | w3m.MapFlagWaterWavesOnCliffShores | w3m.MapFlagWaterWavesOnSlopeShores | w3m.MapFlags(0x8400),
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
			},
			"",
			"rEfl+K13/fxgOhjUqXxjPjsoLb7JulvzFvNpMab101cr8V9wKLNZFQcUD+TFSH2j7mgMoSb9bAyBkYA6sZU0Cg",
			"0xDD4E3EBE|P5c/izfa1qstJu5zYYVyc2FD2gE",
		},
		{
			"test_tft.w3x",
			false,
			w3m.SizeTiny,
			w3m.Info{
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
				Flags:            w3m.MapFlagRevealTerrain | w3m.MapFlagWaterWavesOnCliffShores | w3m.MapFlagWaterWavesOnSlopeShores | w3m.MapFlags(0xC400),
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
			},
			"",
			"cF03T1FzQzhwZwm3F/yp0fo8uDbHe/3qqqOQyJLKcg5HEHQTtk5M08L6mbDoRvzdbWd8SgWNQ+Fb3qSaovCuYg",
			"0x7F321A74|/1ndO+WvBCWiQutD9VyCefo3GYM",
		},
	}

	for _, f := range files {
		m, err := w3m.Open("./" + f.file)
		if err != nil {
			t.Fatal(f.file, err)
		}

		if m.Signed() != f.signed {
			t.Fatal(f.file, "signed mismatch")
		}

		inf, err := m.Info()
		if err != nil {
			t.Fatal(f.file, err)
		}

		if !reflect.DeepEqual(&f.info, inf) {
			t.Log(fmt.Sprintf("%+v\n", *inf))
			t.Fatal(f.file, "Info() return value not deep equal")
		}

		if inf.Size() != f.size {
			t.Fatalf("%v size mismatch %v != %v\n", f.file, inf.Size(), f.size)
		}

		prv, _ := m.Preview()
		if sha := shaImage(prv); sha != f.previewSHA {
			t.Fatalf("%v preview mismatch %v != %v\n", f.file, sha, f.previewSHA)
		}

		mmap, _ := m.Minimap()
		if sha := shaImage(mmap); sha != f.minimapSHA {
			t.Fatalf("%v minimap mismatch %v != %v\n", f.file, sha, f.minimapSHA)
		}

		hash, err := m.Checksum(nil)
		if err != nil {
			t.Fatal(err)
		}
		if hash.String() != f.checksum {
			t.Fatalf("%v checksum mismatch %v != %v\n", f.file, hash, f.checksum)
		}

		if err := m.Close(); err != nil {
			t.Fatal(err)
		}
	}
}
