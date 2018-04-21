// Author:  Niels A.D.
// Project: gowarcraft3 (https://github.com/nielsAD/gowarcraft3)
// License: Mozilla Public License, v2.0

// Package w3m implements basic information extraction functions for w3m/w3x files.
package w3m

import (
	"io"

	"github.com/nielsAD/gowarcraft3/protocol"
)

// Info for Warcraft III maps as found in the war3map.w3i file
type Info struct {
	FileFormat    uint32
	SaveCount     uint32
	EditorVersion uint32

	Name             string
	Author           string
	Description      string
	SuggestedPlayers string

	CamBounds      [8]float32
	CamBoundsCompl [4]uint32

	Width  uint32
	Height uint32
	Flags  MapFlags

	Tileset      Tileset
	LsBackground uint32
	LsPath       string
	LsText       string
	LsTitle      string
	LsSubTitle   string

	DataSet    uint32
	PsPath     string
	PsText     string
	PsTitle    string
	PsSubTitle string

	Fog        uint32
	FogStart   float32
	FogEnd     float32
	FogDensity float32
	FogColor   uint32
	WeatherID  protocol.DWordString

	SoundEnv   string
	LightEnv   Tileset
	WaterColor uint32

	Players                     []Player
	Forces                      []Force
	CustomUpgradeAvailabilities []CustomUpgradeAvailability
	CustomTechAvailabilities    []CustomTechAvailability
}

// Player structure in war3map.w3i file
type Player struct {
	ID           uint32
	Type         PlayerType
	Race         Race
	Flags        PlayerFlags
	Name         string
	StartPosX    float32
	StartPosY    float32
	AllyPrioLow  protocol.BitSet32
	AllyPrioHigh protocol.BitSet32
}

// Force structure in war3map.w3i file
type Force struct {
	Flags     ForceFlags
	PlayerSet protocol.BitSet32
	Name      string
}

// CustomUpgradeAvailability in war3map.w3i file
type CustomUpgradeAvailability struct {
	PlayerSet    protocol.BitSet32
	UpgradeID    protocol.DWordString
	Level        uint32
	Availability UpgradeAvailability
}

// CustomTechAvailability in war3map.w3i file
type CustomTechAvailability struct {
	PlayerSet protocol.BitSet32
	TechID    protocol.DWordString
}

// Info read from war3map.w3i
func (m *Map) Info() (*Info, error) {
	w3i, err := m.Archive.Open("war3map.w3i")
	if err != nil {
		return nil, err
	}
	defer w3i.Close()

	if w3i.Size() < 96 {
		return nil, ErrBadFormat
	}

	var b protocol.Buffer
	if _, err := io.Copy(&b, w3i); err != nil {
		return nil, err
	}

	var readTS = func() (string, error) {
		s, err := b.ReadCString()
		if err != nil {
			return "", err
		}
		return m.ExpandString(s)
	}

	var i = Info{
		FileFormat: b.ReadUInt32(),
	}

	switch i.FileFormat {
	case 18: //.w3m
	case 25: //.w3x
	default:
		return nil, ErrBadFormat
	}

	i.SaveCount = b.ReadUInt32()
	i.EditorVersion = b.ReadUInt32()

	i.Name, _ = readTS()
	i.Author, _ = readTS()
	i.Description, _ = readTS()
	i.SuggestedPlayers, err = readTS()
	if err != nil {
		return nil, err
	} else if b.Size() < 80 {
		return nil, ErrBadFormat
	}

	for c := 0; c < len(i.CamBounds); c++ {
		i.CamBounds[c] = b.ReadFloat32()
	}
	for c := 0; c < len(i.CamBoundsCompl); c++ {
		i.CamBoundsCompl[c] = b.ReadUInt32()
	}

	i.Width = b.ReadUInt32()
	i.Height = b.ReadUInt32()
	i.Flags = MapFlags(b.ReadUInt32())

	i.Tileset = Tileset(b.ReadUInt8())
	i.LsBackground = b.ReadUInt32()

	if i.FileFormat == 25 {
		i.LsPath, _ = readTS()
	}
	i.LsText, _ = readTS()
	i.LsTitle, _ = readTS()
	i.LsSubTitle, err = readTS()
	if err != nil {
		return nil, err
	} else if b.Size() < 13 {
		return nil, ErrBadFormat
	}

	i.DataSet = b.ReadUInt32()

	if i.FileFormat == 25 {
		i.PsPath, _ = readTS()
	}
	i.PsText, _ = readTS()
	i.PsTitle, _ = readTS()
	i.PsSubTitle, err = readTS()
	if err != nil {
		return nil, err
	}

	if i.FileFormat == 25 {
		if b.Size() < 54 {
			return nil, ErrBadFormat
		}
		i.Fog = b.ReadUInt32()
		i.FogStart = b.ReadFloat32()
		i.FogEnd = b.ReadFloat32()
		i.FogDensity = b.ReadFloat32()
		i.FogColor = b.ReadUInt32()
		i.WeatherID = b.ReadDString()
		i.SoundEnv, err = readTS()
		if err != nil {
			return nil, err
		} else if b.Size() < 12 {
			return nil, ErrBadFormat
		}
		i.LightEnv = Tileset(b.ReadUInt8())
		i.WaterColor = b.ReadUInt32()
	}

	if b.Size() < 8 {
		return nil, ErrBadFormat
	}

	var numPlayers = b.ReadUInt32()
	i.Players = make([]Player, numPlayers)

	for p := uint32(0); p < numPlayers; p++ {
		if b.Size() < 37 {
			return nil, ErrBadFormat
		}
		i.Players[p].ID = b.ReadUInt32()
		i.Players[p].Type = PlayerType(b.ReadUInt32())
		i.Players[p].Race = Race(b.ReadUInt32())
		i.Players[p].Flags = PlayerFlags(b.ReadUInt32())
		i.Players[p].Name, err = readTS()
		if err != nil {
			return nil, err
		} else if b.Size() < 20 {
			return nil, ErrBadFormat
		}
		i.Players[p].StartPosX = b.ReadFloat32()
		i.Players[p].StartPosY = b.ReadFloat32()
		i.Players[p].AllyPrioLow = protocol.BitSet32(b.ReadUInt32())
		i.Players[p].AllyPrioHigh = protocol.BitSet32(b.ReadUInt32())
	}

	var numForces = b.ReadUInt32()
	i.Forces = make([]Force, numForces)

	for f := uint32(0); f < numForces; f++ {
		if b.Size() < 9 {
			return nil, ErrBadFormat
		}
		i.Forces[f].Flags = ForceFlags(b.ReadUInt32())
		i.Forces[f].PlayerSet = protocol.BitSet32(b.ReadUInt32())
		i.Forces[f].Name, err = readTS()
		if err != nil {
			return nil, err
		}
	}

	if b.Size() >= 8 {
		var numUpgrades = b.ReadUInt32()
		i.CustomUpgradeAvailabilities = make([]CustomUpgradeAvailability, numUpgrades)

		for u := uint32(0); u < numUpgrades; u++ {
			if b.Size() < 24 {
				return nil, ErrBadFormat
			}
			i.CustomUpgradeAvailabilities[u].PlayerSet = protocol.BitSet32(b.ReadUInt32())
			i.CustomUpgradeAvailabilities[u].UpgradeID = b.ReadDString()
			i.CustomUpgradeAvailabilities[u].Level = b.ReadUInt32()
			i.CustomUpgradeAvailabilities[u].Availability = UpgradeAvailability(b.ReadUInt32())
		}

		var numTechs = b.ReadUInt32()
		i.CustomTechAvailabilities = make([]CustomTechAvailability, numTechs)

		for u := uint32(0); u < numTechs; u++ {
			if b.Size() < 12 {
				return nil, ErrBadFormat
			}
			i.CustomTechAvailabilities[u].PlayerSet = protocol.BitSet32(b.ReadUInt32())
			i.CustomTechAvailabilities[u].TechID = b.ReadDString()
		}

		// TODO: RandomUnitTable and RandomItemTable
	}

	return &i, nil
}

// Size returns the map size category
func (m *Info) Size() Size {
	var s = m.Width * m.Height
	if s < 92*92 {
		return SizeTiny
	} else if s < 128*128 {
		return SizeExtraSmall
	} else if s < 160*160 {
		return SizeSmall
	} else if s < 192*192 {
		return SizeNormal
	} else if s < 224*224 {
		return SizeLarge
	} else if s < 288*288 {
		return SizeExtraLarge
	} else if s < 384*384 {
		return SizeHuge
	} else {
		return SizeEpic
	}
}
