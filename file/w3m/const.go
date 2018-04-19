package w3m

import (
	"errors"
	"fmt"
)

// Errors
var (
	ErrBadFormat = errors.New("w3m: Invalid file format")
)

// Size enum
type Size byte

// Map size categories
const (
	SizeTiny Size = iota
	SizeExtraSmall
	SizeSmall
	SizeNormal
	SizeLarge
	SizeExtraLarge
	SizeHuge
	SizeEpic
)

func (s Size) String() string {
	switch s {
	case SizeTiny:
		return "Tiny"
	case SizeExtraSmall:
		return "ExtraSmall"
	case SizeSmall:
		return "Small"
	case SizeNormal:
		return "Normal"
	case SizeLarge:
		return "Large"
	case SizeExtraLarge:
		return "ExtraLarge"
	case SizeHuge:
		return "Huge"
	case SizeEpic:
		return "Epic"
	default:
		return fmt.Sprintf("Size(0x%02X)", uint8(s))
	}
}

// Tileset enum
type Tileset byte

// Tileset
const (
	TileAshenvale       Tileset = 'A'
	TileBarrens         Tileset = 'B'
	TileFelwood         Tileset = 'C'
	TileDungeon         Tileset = 'D'
	TileLordaeronFall   Tileset = 'F'
	TileUnderground     Tileset = 'G'
	TileLordaeronSummer Tileset = 'L'
	TileNorthrend       Tileset = 'N'
	TileVillageFall     Tileset = 'Q'
	TileVillage         Tileset = 'V'
	TileLordaeronWinter Tileset = 'W'
	TileDalaran         Tileset = 'X'
	TileCityscape       Tileset = 'Y'
	TileSunkenRuins     Tileset = 'Z'
	TileIcecrown        Tileset = 'I'
	TileDalaranRuins    Tileset = 'J'
	TileOutland         Tileset = 'O'
	TileBlackCitadel    Tileset = 'K'
)

func (t Tileset) String() string {
	switch t {
	case TileAshenvale:
		return "Ashenvale"
	case TileBarrens:
		return "Barrens"
	case TileFelwood:
		return "Felwood"
	case TileDungeon:
		return "Dungeon"
	case TileLordaeronFall:
		return "LordaeronFall"
	case TileUnderground:
		return "Underground"
	case TileLordaeronSummer:
		return "LordaeronSummer"
	case TileNorthrend:
		return "Northrend"
	case TileVillageFall:
		return "VillageFall"
	case TileVillage:
		return "Village"
	case TileLordaeronWinter:
		return "LordaeronWinter"
	case TileDalaran:
		return "Dalaran"
	case TileCityscape:
		return "Cityscape"
	case TileSunkenRuins:
		return "SunkenRuins"
	case TileIcecrown:
		return "Icecrown"
	case TileDalaranRuins:
		return "DalaranRuins"
	case TileOutland:
		return "Outland"
	case TileBlackCitadel:
		return "BlackCitadel"
	default:
		return fmt.Sprintf("Tileset(0x%02X)", uint8(t))
	}
}

// Flags enum
type Flags uint32

// Map Flags
const (
	FlagHideMinimap             Flags = 0x0001
	FlagModifyAllyPriorities    Flags = 0x0002
	FlagMelee                   Flags = 0x0004
	FlagRevealTerrain           Flags = 0x0010
	FlagFixedPlayerSettings     Flags = 0x0020
	FlagCustomForces            Flags = 0x0040
	FlagCustomTechTree          Flags = 0x0080
	FlagCustomAbilities         Flags = 0x0100
	FlagCustomUpgrades          Flags = 0x0200
	FlagWaterWavesOnCliffShoes  Flags = 0x0800
	FlagWaterWavesOnSlopeShores Flags = 0x1000
)

func (f Flags) String() string {
	var res string
	if f&FlagHideMinimap != 0 {
		res += "|HideMinimap"
		f &= ^FlagHideMinimap
	}
	if f&FlagModifyAllyPriorities != 0 {
		res += "|ModifyAllyPriorities"
		f &= ^FlagModifyAllyPriorities
	}
	if f&FlagMelee != 0 {
		res += "|Melee"
		f &= ^FlagMelee
	}
	if f&FlagRevealTerrain != 0 {
		res += "|RevealTerrain"
		f &= ^FlagRevealTerrain
	}
	if f&FlagFixedPlayerSettings != 0 {
		res += "|FixedPlayerSettings"
		f &= ^FlagFixedPlayerSettings
	}
	if f&FlagCustomForces != 0 {
		res += "|CustomForces"
		f &= ^FlagCustomForces
	}
	if f&FlagCustomTechTree != 0 {
		res += "|CustomTechTree"
		f &= ^FlagCustomTechTree
	}
	if f&FlagCustomAbilities != 0 {
		res += "|CustomAbilities"
		f &= ^FlagCustomAbilities
	}
	if f&FlagCustomUpgrades != 0 {
		res += "|CustomUpgrades"
		f &= ^FlagCustomUpgrades
	}
	if f&FlagWaterWavesOnCliffShoes != 0 {
		res += "|WaterWavesOnCliffShoes"
		f &= ^FlagWaterWavesOnCliffShoes
	}
	if f&FlagWaterWavesOnSlopeShores != 0 {
		res += "|WaterWavesOnSlopeShores"
		f &= ^FlagWaterWavesOnSlopeShores
	}
	if f != 0 {
		res += fmt.Sprintf("|Flags(0x%02X)", uint32(f))
	}
	if res != "" {
		res = res[1:]
	}
	return res
}

// TeamFlags enum
type TeamFlags uint32

// Team Flags
const (
	TeamFlagAllied           TeamFlags = 0x01
	TeamFlagAlliedVictory    TeamFlags = 0x02
	TeamFlagShareVision      TeamFlags = 0x08
	TeamFlagShareUnitControl TeamFlags = 0x10
	TeamFlagShareAdvUnit     TeamFlags = 0x20
)

func (f TeamFlags) String() string {
	var res string
	if f&TeamFlagAllied != 0 {
		res += "|Allied"
		f &= ^TeamFlagAllied
	}
	if f&TeamFlagAlliedVictory != 0 {
		res += "|AlliedVictory"
		f &= ^TeamFlagAlliedVictory
	}
	if f&TeamFlagShareVision != 0 {
		res += "|ShareVision"
		f &= ^TeamFlagShareVision
	}
	if f&TeamFlagShareUnitControl != 0 {
		res += "|ShareUnitControl"
		f &= ^TeamFlagShareUnitControl
	}
	if f&TeamFlagShareAdvUnit != 0 {
		res += "|ShareAdvUnit"
		f &= ^TeamFlagShareAdvUnit
	}
	if f != 0 {
		res += fmt.Sprintf("|TeamFlags(0x%02X)", uint32(f))
	}
	if res != "" {
		res = res[1:]
	}
	return res
}

// SlotType enum
type SlotType uint32

// Slot types
const (
	SlotHuman SlotType = iota + 1
	SlotComputer
	SlotNeutral
	SlotRescuable
)

func (s SlotType) String() string {
	switch s {
	case SlotHuman:
		return "Human"
	case SlotComputer:
		return "Computer"
	case SlotNeutral:
		return "Neutral"
	case SlotRescuable:
		return "Rescuable"
	default:
		return fmt.Sprintf("SlotType(0x%02X)", uint32(s))
	}
}

// Race enum
type Race uint32

// Races
const (
	RaceSelectable Race = iota
	RaceHuman
	RaceOrc
	RaceUndead
	RaceNightElf
)

func (s Race) String() string {
	switch s {
	case RaceSelectable:
		return "Selectable"
	case RaceHuman:
		return "Human"
	case RaceOrc:
		return "Orc"
	case RaceUndead:
		return "Undead"
	case RaceNightElf:
		return "NightElf"
	default:
		return fmt.Sprintf("Race(0x%02X)", uint32(s))
	}
}
