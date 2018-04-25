// Author:  Niels A.D.
// Project: gowarcraft3 (https://github.com/nielsAD/gowarcraft3)
// License: Mozilla Public License, v2.0

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
type Size uint8

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

// MapFlags enum
type MapFlags uint32

// Map Flags
const (
	MapFlagHideMinimap             MapFlags = 0x0001
	MapFlagModifyAllyPriorities    MapFlags = 0x0002
	MapFlagMelee                   MapFlags = 0x0004
	MapFlagRevealTerrain           MapFlags = 0x0010
	MapFlagFixedPlayerSettings     MapFlags = 0x0020
	MapFlagCustomForces            MapFlags = 0x0040
	MapFlagCustomTechTree          MapFlags = 0x0080
	MapFlagCustomAbilities         MapFlags = 0x0100
	MapFlagCustomUpgrades          MapFlags = 0x0200
	MapFlagWaterWavesOnCliffShoes  MapFlags = 0x0800
	MapFlagWaterWavesOnSlopeShores MapFlags = 0x1000
)

func (f MapFlags) String() string {
	var res string
	if f&MapFlagHideMinimap != 0 {
		res += "|HideMinimap"
		f &= ^MapFlagHideMinimap
	}
	if f&MapFlagModifyAllyPriorities != 0 {
		res += "|ModifyAllyPriorities"
		f &= ^MapFlagModifyAllyPriorities
	}
	if f&MapFlagMelee != 0 {
		res += "|Melee"
		f &= ^MapFlagMelee
	}
	if f&MapFlagRevealTerrain != 0 {
		res += "|RevealTerrain"
		f &= ^MapFlagRevealTerrain
	}
	if f&MapFlagFixedPlayerSettings != 0 {
		res += "|FixedPlayerSettings"
		f &= ^MapFlagFixedPlayerSettings
	}
	if f&MapFlagCustomForces != 0 {
		res += "|CustomForces"
		f &= ^MapFlagCustomForces
	}
	if f&MapFlagCustomTechTree != 0 {
		res += "|CustomTechTree"
		f &= ^MapFlagCustomTechTree
	}
	if f&MapFlagCustomAbilities != 0 {
		res += "|CustomAbilities"
		f &= ^MapFlagCustomAbilities
	}
	if f&MapFlagCustomUpgrades != 0 {
		res += "|CustomUpgrades"
		f &= ^MapFlagCustomUpgrades
	}
	if f&MapFlagWaterWavesOnCliffShoes != 0 {
		res += "|WaterWavesOnCliffShoes"
		f &= ^MapFlagWaterWavesOnCliffShoes
	}
	if f&MapFlagWaterWavesOnSlopeShores != 0 {
		res += "|WaterWavesOnSlopeShores"
		f &= ^MapFlagWaterWavesOnSlopeShores
	}
	if f != 0 {
		res += fmt.Sprintf("|MapFlags(0x%02X)", uint32(f))
	}
	if res != "" {
		res = res[1:]
	}
	return res
}

// PlayerFlags enum
type PlayerFlags uint32

// Player Flags
const (
	PlayerFlagFixedPos PlayerFlags = 0x01
)

func (f PlayerFlags) String() string {
	var res string
	if f&PlayerFlagFixedPos != 0 {
		res += "|FixedPos"
		f &= ^PlayerFlagFixedPos
	}
	if f != 0 {
		res += fmt.Sprintf("|PlayerFlags(0x%02X)", uint32(f))
	}
	if res != "" {
		res = res[1:]
	}
	return res
}

// ForceFlags enum
type ForceFlags uint32

// Force Flags
const (
	ForceFlagAllied           ForceFlags = 0x01
	ForceFlagAlliedVictory    ForceFlags = 0x02
	ForceFlagShareVision      ForceFlags = 0x08
	ForceFlagShareUnitControl ForceFlags = 0x10
	ForceFlagShareAdvUnit     ForceFlags = 0x20
)

func (f ForceFlags) String() string {
	var res string
	if f&ForceFlagAllied != 0 {
		res += "|Allied"
		f &= ^ForceFlagAllied
	}
	if f&ForceFlagAlliedVictory != 0 {
		res += "|AlliedVictory"
		f &= ^ForceFlagAlliedVictory
	}
	if f&ForceFlagShareVision != 0 {
		res += "|ShareVision"
		f &= ^ForceFlagShareVision
	}
	if f&ForceFlagShareUnitControl != 0 {
		res += "|ShareUnitControl"
		f &= ^ForceFlagShareUnitControl
	}
	if f&ForceFlagShareAdvUnit != 0 {
		res += "|ShareAdvUnit"
		f &= ^ForceFlagShareAdvUnit
	}
	if f != 0 {
		res += fmt.Sprintf("|ForceFlags(0x%02X)", uint32(f))
	}
	if res != "" {
		res = res[1:]
	}
	return res
}

// PlayerType enum
type PlayerType uint32

// Player types
const (
	PlayerHuman PlayerType = iota + 1
	PlayerComputer
	PlayerNeutral
	PlayerRescuable
)

func (p PlayerType) String() string {
	switch p {
	case PlayerHuman:
		return "Human"
	case PlayerComputer:
		return "Computer"
	case PlayerNeutral:
		return "Neutral"
	case PlayerRescuable:
		return "Rescuable"
	default:
		return fmt.Sprintf("PlayerType(0x%02X)", uint32(p))
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

func (r Race) String() string {
	switch r {
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
		return fmt.Sprintf("Race(0x%02X)", uint32(r))
	}
}

// UpgradeAvailability enum
type UpgradeAvailability uint32

// Upgrade availabilities
const (
	UpgradeUnavailable UpgradeAvailability = iota
	UpgradeAvailable
	UpgradeResearched
)

func (u UpgradeAvailability) String() string {
	switch u {
	case UpgradeUnavailable:
		return "Unavailable"
	case UpgradeAvailable:
		return "Available"
	case UpgradeResearched:
		return "Researched"
	default:
		return fmt.Sprintf("UpgradeAvailability(0x%02X)", uint32(u))
	}
}
