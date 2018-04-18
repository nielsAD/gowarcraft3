// Package maps implements basic information extraction functions for w3m/w3x files.
package maps

import (
	"bufio"
	"io"
	"os"
	"regexp"
	"strconv"
	"strings"

	"github.com/nielsAD/noot/pkg/mpq"
	"github.com/nielsAD/noot/pkg/util"
)

// Map information for Warcraft III maps
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
	Flags  Flags

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
	WeatherID  util.DWordString

	SoundEnv   string
	LightEnv   Tileset
	WaterColor uint32

	Slots []SlotData
	Teams []TeamData
}

// SlotData for a single map slot
type SlotData struct {
	ID           uint32
	Type         SlotType
	Race         Race
	StartPos     uint32
	Name         string
	StartPosX    float32
	StartPosY    float32
	AllyPrioLow  uint32
	AllyPrioHigh uint32
}

// TeamData for a single map team
type TeamData struct {
	Flags      TeamFlags
	PlayerMask uint32
	Name       string
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
	if _, err := buf.Discard(1); err != nil {
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

		p1, err := buf.ReadString('\n')
		if err != nil {
			return nil, err
		}
		if strings.TrimSpace(p1) != "{" {
			return nil, ErrBadFormat
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
	var m = Map{
		FileName: fileName,
	}

	a, err := mpq.OpenArchive(fileName)
	if err != nil {
		return nil, err
	}
	defer a.Close()

	m.Signed = a.StrongSigned()

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

	var b util.PacketBuffer
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

	m.FileFormat = b.ReadUInt32()
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
	m.Flags = Flags(b.ReadUInt32())

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
	} else if b.Size() < 12 {
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
		} else if b.Size() < 13 {
			return nil, ErrBadFormat
		}
		m.LightEnv = Tileset(b.ReadUInt8())
		m.WaterColor = b.ReadUInt32()
	}

	if b.Size() < 8 {
		return nil, ErrBadFormat
	}

	var numSlots = b.ReadUInt32()
	m.Slots = make([]SlotData, numSlots)

	for p := uint32(0); p < numSlots; p++ {
		if b.Size() < 37 {
			return nil, ErrBadFormat
		}
		m.Slots[p].ID = b.ReadUInt32()
		m.Slots[p].Type = SlotType(b.ReadUInt32())
		m.Slots[p].Race = Race(b.ReadUInt32())
		m.Slots[p].StartPos = b.ReadUInt32()
		m.Slots[p].Name, err = readTS()
		if err != nil {
			return nil, err
		} else if b.Size() < 20 {
			return nil, ErrBadFormat
		}
		m.Slots[p].StartPosX = b.ReadFloat32()
		m.Slots[p].StartPosY = b.ReadFloat32()
		m.Slots[p].AllyPrioLow = b.ReadUInt32()
		m.Slots[p].AllyPrioHigh = b.ReadUInt32()
	}

	var numTeams = b.ReadUInt32()
	m.Teams = make([]TeamData, numTeams)

	for p := uint32(0); p < numTeams; p++ {
		if b.Size() < 9 {
			return nil, ErrBadFormat
		}
		m.Teams[p].Flags = TeamFlags(b.ReadUInt32())
		m.Teams[p].PlayerMask = b.ReadUInt32()
		m.Teams[p].Name, err = readTS()
		if err != nil {
			return nil, err
		}
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
