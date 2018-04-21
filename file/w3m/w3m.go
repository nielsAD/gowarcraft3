// Author:  Niels A.D.
// Project: gowarcraft3 (https://github.com/nielsAD/gowarcraft3)
// License: Mozilla Public License, v2.0

// Package w3m implements basic information extraction functions for w3m/w3x files.
package w3m

import (
	"bufio"
	"io"
	"os"
	"regexp"
	"strconv"
	"strings"

	"github.com/nielsAD/gowarcraft3/file/mpq"
	"github.com/nielsAD/gowarcraft3/protocol"
)

// Map information for Warcraft III maps (.w3i file)
type Map struct {
	FileName string
	Signed   bool

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

// Player structure in .w3i file
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

// Force structure in .w3i file
type Force struct {
	Flags     ForceFlags
	PlayerSet protocol.BitSet32
	Name      string
}

// CustomUpgradeAvailability in .w3i file
type CustomUpgradeAvailability struct {
	PlayerSet    protocol.BitSet32
	UpgradeID    protocol.DWordString
	Level        uint32
	Availability UpgradeAvailability
}

// CustomTechAvailability in .w3i file
type CustomTechAvailability struct {
	PlayerSet protocol.BitSet32
	TechID    protocol.DWordString
}

// TriggerString recognition
var reWTS = regexp.MustCompile("^STRING (\\d+)$")
var reTS = regexp.MustCompile("^TRIGSTR_(\\d+)$")

// Load TriggerStrings
func readWTS(archive *mpq.Archive) (map[int]string, error) {
	wts, err := archive.Open("war3map.wts")
	if err != nil {
		return nil, err
	}
	defer wts.Close()

	buf := bufio.NewReader(wts)

	if _, err := buf.Discard(1); err != nil && err != io.EOF {
		return nil, err
	}

	var m = make(map[int]string)
	for true {
		l, err := buf.ReadString('\n')
		if err == io.EOF {
			break
		} else if err != nil {
			return nil, err
		}

		match := reWTS.FindStringSubmatch(strings.TrimSpace(l))
		if len(match) < 2 {
			continue
		}

		id, err := strconv.Atoi(match[1])
		if err != nil {
			continue
		}

		for true {
			p1, err := buf.ReadString('\n')
			if err != nil {
				return nil, err
			}
			if strings.TrimSpace(p1) == "{" {
				break
			} else if !strings.HasPrefix(p1, "//") {
				return nil, ErrBadFormat
			}
		}

		var sb strings.Builder
		for true {
			l, err := buf.ReadString('\n')
			if err != nil {
				return nil, err
			}
			if strings.TrimSpace(l) == "}" {
				break
			}

			if sb.Len() > 0 {
				sb.WriteByte('\n')
			}
			sb.WriteString(strings.TrimRight(l, "\r\n"))
		}

		m[id] = sb.String()
	}

	return m, nil
}

// Load a Warcraft III map file
func Load(fileName string) (*Map, error) {
	a, err := mpq.OpenArchive(fileName)
	if err != nil {
		return nil, err
	}
	defer a.Close()

	f, err := a.Open("war3map.w3i")
	if err != nil {
		return nil, err
	}
	defer f.Close()

	if f.Size() < 96 {
		return nil, ErrBadFormat
	}

	wts, err := readWTS(a)
	if err != nil && err != os.ErrNotExist {
		return nil, err
	}

	var b protocol.Buffer
	if _, err := io.Copy(&b, f); err != nil {
		return nil, err
	}

	var readTS = func() (string, error) {
		s, err := b.ReadCString()
		if err != nil {
			return "", err
		}

		match := reTS.FindStringSubmatch(s)
		if wts == nil || len(match) == 0 {
			return s, nil
		}

		id, err := strconv.Atoi(match[1])
		if err != nil {
			return "", err
		}

		return wts[id], nil
	}

	var m = Map{
		FileName:   fileName,
		FileFormat: b.ReadUInt32(),
		Signed:     a.StrongSigned(),
	}

	switch m.FileFormat {
	case 18: //.w3m
	case 25: //.w3x
	default:
		return nil, ErrBadFormat
	}

	m.SaveCount = b.ReadUInt32()
	m.EditorVersion = b.ReadUInt32()

	m.Name, _ = readTS()
	m.Author, _ = readTS()
	m.Description, _ = readTS()
	m.SuggestedPlayers, err = readTS()
	if err != nil {
		return nil, err
	} else if b.Size() < 80 {
		return nil, ErrBadFormat
	}

	for i := 0; i < len(m.CamBounds); i++ {
		m.CamBounds[i] = b.ReadFloat32()
	}
	for i := 0; i < len(m.CamBoundsCompl); i++ {
		m.CamBoundsCompl[i] = b.ReadUInt32()
	}

	m.Width = b.ReadUInt32()
	m.Height = b.ReadUInt32()
	m.Flags = MapFlags(b.ReadUInt32())

	m.Tileset = Tileset(b.ReadUInt8())
	m.LsBackground = b.ReadUInt32()

	if m.FileFormat == 25 {
		m.LsPath, _ = readTS()
	}
	m.LsText, _ = readTS()
	m.LsTitle, _ = readTS()
	m.LsSubTitle, err = readTS()
	if err != nil {
		return nil, err
	} else if b.Size() < 13 {
		return nil, ErrBadFormat
	}

	m.DataSet = b.ReadUInt32()

	if m.FileFormat == 25 {
		m.PsPath, _ = readTS()
	}
	m.PsText, _ = readTS()
	m.PsTitle, _ = readTS()
	m.PsSubTitle, err = readTS()
	if err != nil {
		return nil, err
	}

	if m.FileFormat == 25 {
		if b.Size() < 54 {
			return nil, ErrBadFormat
		}
		m.Fog = b.ReadUInt32()
		m.FogStart = b.ReadFloat32()
		m.FogEnd = b.ReadFloat32()
		m.FogDensity = b.ReadFloat32()
		m.FogColor = b.ReadUInt32()
		m.WeatherID = b.ReadDString()
		m.SoundEnv, err = readTS()
		if err != nil {
			return nil, err
		} else if b.Size() < 12 {
			return nil, ErrBadFormat
		}
		m.LightEnv = Tileset(b.ReadUInt8())
		m.WaterColor = b.ReadUInt32()
	}

	if b.Size() < 8 {
		return nil, ErrBadFormat
	}

	var numPlayers = b.ReadUInt32()
	m.Players = make([]Player, numPlayers)

	for p := uint32(0); p < numPlayers; p++ {
		if b.Size() < 37 {
			return nil, ErrBadFormat
		}
		m.Players[p].ID = b.ReadUInt32()
		m.Players[p].Type = PlayerType(b.ReadUInt32())
		m.Players[p].Race = Race(b.ReadUInt32())
		m.Players[p].Flags = PlayerFlags(b.ReadUInt32())
		m.Players[p].Name, err = readTS()
		if err != nil {
			return nil, err
		} else if b.Size() < 20 {
			return nil, ErrBadFormat
		}
		m.Players[p].StartPosX = b.ReadFloat32()
		m.Players[p].StartPosY = b.ReadFloat32()
		m.Players[p].AllyPrioLow = protocol.BitSet32(b.ReadUInt32())
		m.Players[p].AllyPrioHigh = protocol.BitSet32(b.ReadUInt32())
	}

	var numForces = b.ReadUInt32()
	m.Forces = make([]Force, numForces)

	for f := uint32(0); f < numForces; f++ {
		if b.Size() < 9 {
			return nil, ErrBadFormat
		}
		m.Forces[f].Flags = ForceFlags(b.ReadUInt32())
		m.Forces[f].PlayerSet = protocol.BitSet32(b.ReadUInt32())
		m.Forces[f].Name, err = readTS()
		if err != nil {
			return nil, err
		}
	}

	if b.Size() >= 8 {
		var numUpgrades = b.ReadUInt32()
		m.CustomUpgradeAvailabilities = make([]CustomUpgradeAvailability, numUpgrades)

		for u := uint32(0); u < numUpgrades; u++ {
			if b.Size() < 24 {
				return nil, ErrBadFormat
			}
			m.CustomUpgradeAvailabilities[u].PlayerSet = protocol.BitSet32(b.ReadUInt32())
			m.CustomUpgradeAvailabilities[u].UpgradeID = b.ReadDString()
			m.CustomUpgradeAvailabilities[u].Level = b.ReadUInt32()
			m.CustomUpgradeAvailabilities[u].Availability = UpgradeAvailability(b.ReadUInt32())
		}

		var numTechs = b.ReadUInt32()
		m.CustomTechAvailabilities = make([]CustomTechAvailability, numTechs)

		for u := uint32(0); u < numTechs; u++ {
			if b.Size() < 12 {
				return nil, ErrBadFormat
			}
			m.CustomTechAvailabilities[u].PlayerSet = protocol.BitSet32(b.ReadUInt32())
			m.CustomTechAvailabilities[u].TechID = b.ReadDString()
		}

		// TODO: RandomUnitTable and RandomItemTable
	}

	return &m, nil
}

// Size returns the map size category
func (m *Map) Size() Size {
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
