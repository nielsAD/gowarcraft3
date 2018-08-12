// Author:  Niels A.D.
// Project: gowarcraft3 (https://github.com/nielsAD/gowarcraft3)
// License: Mozilla Public License, v2.0

package w3gs

import (
	"errors"
	"fmt"

	"github.com/nielsAD/gowarcraft3/protocol"
)

// Errors
var (
	ErrNoProtocolSig     = errors.New("w3gs: Invalid w3gs packet (no signature found)")
	ErrInvalidPacketSize = errors.New("w3gs: Invalid packet size")
	ErrInvalidChecksum   = errors.New("w3gs: Checksum invalid")
	ErrUnexpectedConst   = errors.New("w3gs: Unexpected constant value")
	ErrBufferTooSmall    = errors.New("w3gs: Packet exceeds buffer size")
)

// CurrentGameVersion used by stable release
const CurrentGameVersion = uint32(30)

// ProtocolSig is the W3GS magic number used in the packet header.
const ProtocolSig = 0xF7

// W3GS packet type identifiers
const (
	PidPingFromHost      = 0x01
	PidSlotInfoJoin      = 0x04
	PidRejectJoin        = 0x05
	PidPlayerInfo        = 0x06
	PidPlayerLeft        = 0x07
	PidPlayerLoaded      = 0x08
	PidSlotInfo          = 0x09
	PidCountDownStart    = 0x0A
	PidCountDownEnd      = 0x0B
	PidIncomingAction    = 0x0C
	PidDesync            = 0x0D
	PidChatFromHost      = 0x0F
	PidStartLag          = 0x10
	PidStopLag           = 0x11
	PidGameOver          = 0x14
	PidPlayerKicked      = 0x1C
	PidLeaveAck          = 0x1B
	PidReqJoin           = 0x1E
	PidLeaveReq          = 0x21
	PidGameLoadedSelf    = 0x23
	PidOutgoingAction    = 0x26
	PidOutgoingKeepAlive = 0x27
	PidChatToHost        = 0x28
	PidDropReq           = 0x29
	PidSearchGame        = 0x2F
	PidGameInfo          = 0x30
	PidCreateGame        = 0x31
	PidRefreshGame       = 0x32
	PidDecreateGame      = 0x33
	PidChatFromOthers    = 0x34
	PidPingFromOthers    = 0x35
	PidPongToOthers      = 0x36
	PidClientInfo        = 0x37
	PidPeerSet           = 0x3B
	PidMapCheck          = 0x3D
	PidStartDownload     = 0x3F
	PidMapSize           = 0x42
	PidMapPart           = 0x43
	PidMapPartOK         = 0x44
	PidMapPartError      = 0x45
	PidPongToHost        = 0x46
	PidIncomingAction2   = 0x48
)

// Failover related: 0x15 0x16 0x2B 0x2C 0x39

// Game product
var (
	ProductDemo = protocol.DString("W3DM") // Demo
	ProductROC  = protocol.DString("WAR3") // ROC
	ProductTFT  = protocol.DString("W3XP") // TFT
)

// SlotLayout enum
type SlotLayout uint8

// Slot layout
const (
	LayoutMelee               SlotLayout = 0x00
	LayoutCustomForces        SlotLayout = 0x01
	LayoutFixedPlayerSettings SlotLayout = 0x02
)

func (s SlotLayout) String() string {
	if s&(LayoutMelee|LayoutCustomForces|LayoutFixedPlayerSettings) != s {
		return fmt.Sprintf("SlotLayout(0x%02X)", uint8(s))
	}

	var res string
	if s&LayoutCustomForces != 0 {
		res = "CustomForces"
	} else {
		res = "Melee"
	}
	if s&LayoutFixedPlayerSettings != 0 {
		res += "(Fixed)"
	}
	return res
}

// SlotStatus enum
type SlotStatus uint8

// Slot status
const (
	SlotOpen     SlotStatus = 0x00
	SlotClosed   SlotStatus = 0x01
	SlotOccupied SlotStatus = 0x02
)

func (s SlotStatus) String() string {
	switch s {
	case SlotOpen:
		return "Open"
	case SlotClosed:
		return "Closed"
	case SlotOccupied:
		return "Occupied"
	default:
		return fmt.Sprintf("SlotStatus(0x%02X)", uint8(s))
	}
}

// RacePref enum
type RacePref uint8

// Race preference
const (
	RaceHuman    RacePref = 0x01
	RaceOrc      RacePref = 0x02
	RaceNightElf RacePref = 0x04
	RaceUndead   RacePref = 0x08
	RaceDemon    RacePref = 0x10
	RaceRandom   RacePref = 0x20
	RaceMask              = RaceHuman | RaceOrc | RaceNightElf | RaceUndead | RaceDemon | RaceRandom

	RaceSelectable RacePref = 0x40
)

func (r RacePref) String() string {
	if r&(RaceMask|RaceSelectable) != r {
		return fmt.Sprintf("RacePref(0x%02X)", uint8(r))
	}

	var res string
	switch r & RaceMask {
	case RaceHuman:
		res = "Human"
	case RaceOrc:
		res = "Orc"
	case RaceNightElf:
		res = "Nightelf"
	case RaceUndead:
		res = "Undead"
	case RaceDemon:
		res = "Demon"
	case RaceRandom:
		res = "Random"
	default:
		return fmt.Sprintf("RacePref(0x%02X)", uint8(r))
	}
	if r&RaceSelectable != 0 {
		res += "(Selectable)"
	}
	return res
}

// AI difficulty enum
type AI uint8

// AI difficulty
const (
	ComputerEasy   AI = 0x00
	ComputerNormal AI = 0x01
	ComputerInsane AI = 0x02
)

func (a AI) String() string {
	switch a {
	case ComputerEasy:
		return "Easy"
	case ComputerNormal:
		return "Normal"
	case ComputerInsane:
		return "Insane"
	default:
		return fmt.Sprintf("AI(0x%02X)", uint8(a))
	}
}

// RejectReason enum
type RejectReason uint32

// RejectJoin reason
const (
	RejectJoinInvalid  RejectReason = 0x07
	RejectJoinFull     RejectReason = 0x09
	RejectJoinStarted  RejectReason = 0x0A
	RejectJoinWrongKey RejectReason = 0x1B
)

func (r RejectReason) String() string {
	switch r {
	case RejectJoinInvalid:
		return "JoinInvalid"
	case RejectJoinFull:
		return "GameFull"
	case RejectJoinStarted:
		return "GameStarted"
	case RejectJoinWrongKey:
		return "WrongKey"
	default:
		return fmt.Sprintf("RejectReason(0x%02X)", uint32(r))
	}
}

// LeaveReason enum
type LeaveReason uint32

// PlayerLeft reason
const (
	LeaveDisconnect      LeaveReason = 0x01
	LeaveLost            LeaveReason = 0x07
	LeaveLostBuildings   LeaveReason = 0x08
	LeaveWon             LeaveReason = 0x09
	LeaveDraw            LeaveReason = 0x0A
	LeaveObserver        LeaveReason = 0x0B
	LeaveInvalidSaveGame LeaveReason = 0x0C // (?)
	LeaveLobby           LeaveReason = 0x0D
)

func (l LeaveReason) String() string {
	switch l {
	case LeaveDisconnect:
		return "Disconnect"
	case LeaveLost:
		return "Lost"
	case LeaveLostBuildings:
		return "LostBuildings"
	case LeaveWon:
		return "Won"
	case LeaveDraw:
		return "Draw"
	case LeaveObserver:
		return "Observer"
	case LeaveInvalidSaveGame:
		return "InvalidSaveGame"
	case LeaveLobby:
		return "Lobby"
	default:
		return fmt.Sprintf("LeaveReason(0x%02X)", uint32(l))
	}
}

// MessageType enum
type MessageType uint8

// Chat type
const (
	MsgChat           MessageType = 0x10
	MsgTeamChange     MessageType = 0x11
	MsgColorChange    MessageType = 0x12
	MsgRaceChange     MessageType = 0x13
	MsgHandicapChange MessageType = 0x14
	MsgChatExtra      MessageType = 0x20
)

func (m MessageType) String() string {
	switch m {
	case MsgChat:
		return "Chat"
	case MsgTeamChange:
		return "TeamChange"
	case MsgColorChange:
		return "ColorChange"
	case MsgRaceChange:
		return "RaceChange"
	case MsgHandicapChange:
		return "HandicapChange"
	case MsgChatExtra:
		return "ChatExtra"
	default:
		return fmt.Sprintf("MessageType(0x%02X)", uint8(m))
	}
}

// GameSettingFlags enum
type GameSettingFlags uint32

// Game setting flags
const (
	SettingSpeedSlow   GameSettingFlags = 0x00000000
	SettingSpeedNormal GameSettingFlags = 0x00000001
	SettingSpeedFast   GameSettingFlags = 0x00000002
	SettingSpeedMask   GameSettingFlags = 0x0000000F

	SettingTerrainHidden   GameSettingFlags = 0x00000100
	SettingTerrainExplored GameSettingFlags = 0x00000200
	SettingTerrainVisible  GameSettingFlags = 0x00000400
	SettingTerrainDefault  GameSettingFlags = 0x00000800
	SettingTerrainMask     GameSettingFlags = 0x00000F00

	SettingObsNone     GameSettingFlags = 0x00000000
	SettingObsOnDefeat GameSettingFlags = 0x00002000
	SettingObsFull     GameSettingFlags = 0x00003000
	SettingObsReferees GameSettingFlags = 0x40000000
	SettingObsMask     GameSettingFlags = 0x40003000

	SettingTeamsTogether GameSettingFlags = 0x00004000
	SettingTeamsFixed    GameSettingFlags = 0x00060000
	SettingSharedControl GameSettingFlags = 0x01000000
	SettingRandomHero    GameSettingFlags = 0x02000000
	SettingRandomRace    GameSettingFlags = 0x04000000
)

func (f GameSettingFlags) String() string {
	var res string
	switch f & SettingSpeedMask {
	case SettingSpeedSlow:
		res = "SpeedSlow"
	case SettingSpeedNormal:
		res = "SpeedNormal"
	case SettingSpeedFast:
		res = "SpeedFast"
	default:
		return fmt.Sprintf("GameSettingFlags(0x%07X)", uint32(f))
	}

	switch f & SettingTerrainMask {
	case SettingTerrainHidden:
		res += "|TerrainHidden"
	case SettingTerrainExplored:
		res += "|TerrainExplored"
	case SettingTerrainVisible:
		res += "|TerrainVisible"
	case SettingTerrainDefault:
		res += "|TerrainDefault"
	case 0:
		// No terrain setting
	default:
		return fmt.Sprintf("GameSettingFlags(0x%07X)", uint32(f))
	}

	switch f & SettingObsMask {
	case SettingObsNone:
		res += "|ObsNone"
	case SettingObsOnDefeat:
		res += "|ObsOnDefeat"
	case SettingObsFull:
		res += "|ObsFull"
	case SettingObsReferees:
		res += "|ObsReferees"
	default:
		return fmt.Sprintf("GameSettingFlags(0x%07X)", uint32(f))
	}

	f &= ^(SettingSpeedMask | SettingTerrainMask | SettingObsMask)

	if f&SettingTeamsTogether != 0 {
		res += "|TeamsTogether"
		f &= ^SettingTeamsTogether
	}
	if f&SettingTeamsFixed != 0 {
		res += "|TeamsFixed"
		f &= ^SettingTeamsFixed
	}
	if f&SettingSharedControl != 0 {
		res += "|SharedControl"
		f &= ^SettingSharedControl
	}
	if f&SettingRandomHero != 0 {
		res += "|RandomHero"
		f &= ^SettingRandomHero
	}
	if f&SettingRandomRace != 0 {
		res += "|RandomRace"
		f &= ^SettingRandomRace
	}

	if f != 0 {
		res += fmt.Sprintf("|GameSettingFlags(0x%02X)", uint32(f))
	}

	return res
}

// GameFlags enum
type GameFlags uint32

// Game flags
const (
	GameFlagCustomGame   GameFlags = 0x000001
	GameFlagOfficialGame GameFlags = 0x000009 // Blizzard signed map
	GameFlagSinglePlayer GameFlags = 0x00001D
	GameFlagTeamLadder   GameFlags = 0x000020
	GameFlagSavedGame    GameFlags = 0x000200
	GameFlagTypeMask     GameFlags = 0x0002FF

	GameFlagPrivateGame GameFlags = 0x000800

	GameFlagCreatorUser     GameFlags = 0x002000
	GameFlagCreatorBlizzard GameFlags = 0x004000
	GameFlagCreatorMask     GameFlags = 0x006000

	GameFlagMapTypeMelee    GameFlags = 0x008000
	GameFlagMapTypeScenario GameFlags = 0x010000
	GameFlagMapTypeMask     GameFlags = 0x018000

	GameFlagSizeSmall  GameFlags = 0x020000
	GameFlagSizeMedium GameFlags = 0x040000
	GameFlagSizeLarge  GameFlags = 0x080000
	GameFlagSizeMask   GameFlags = 0x0E0000

	GameFlagObsFull     GameFlags = 0x100000
	GameFlagObsOnDefeat GameFlags = 0x200000
	GameFlagObsNone     GameFlags = 0x400000
	GameFlagObsMask     GameFlags = 0x700000

	// Used for filtering game list
	GameFlagFilterMask GameFlags = 0x7FE000
)

func (f GameFlags) String() string {
	var res string

	switch f & GameFlagTypeMask {
	case GameFlagCustomGame:
		res = "|Custom"
	case GameFlagOfficialGame:
		res = "|Official"
	case GameFlagSinglePlayer:
		res = "|SinglePlayer"
	case GameFlagTeamLadder:
		res = "|TeamLadder"
	case GameFlagSavedGame:
		res += "|SavedGame"
	case 0:
		// No game type
	default:
		return fmt.Sprintf("GameFlags(0x%06X)", uint32(f))
	}

	if f&GameFlagPrivateGame != 0 {
		res += "|Private"
	}

	switch f & GameFlagCreatorMask {
	case GameFlagCreatorUser:
		res += "|CreatorUser"
	case GameFlagCreatorBlizzard:
		res += "|CreatorBlizzard"
	case GameFlagCreatorMask:
		res += "|CreatorAny"
	case 0:
		// No map maker
	default:
		return fmt.Sprintf("GameFlags(0x%06X)", uint32(f))
	}

	switch f & GameFlagSizeMask {
	case GameFlagSizeSmall:
		res += "|SizeSmall"
	case GameFlagSizeMedium:
		res += "|SizeMedium"
	case GameFlagSizeLarge:
		res += "|SizeLarge"
	case GameFlagSizeMask:
		res += "|SizeAny"
	case 0:
		// No map size
	default:
		return fmt.Sprintf("GameFlags(0x%06X)", uint32(f))
	}

	switch f & GameFlagMapTypeMask {
	case GameFlagMapTypeMelee:
		res += "|MapTypeMelee"
	case GameFlagMapTypeScenario:
		res += "|MapTypeScenario"
	case GameFlagMapTypeMask:
		res += "|MapTypeAny"
	case 0:
		// No map type
	default:
		return fmt.Sprintf("GameFlags(0x%06X)", uint32(f))
	}

	switch f & GameFlagObsMask {
	case GameFlagObsFull:
		res += "|ObsFull"
	case GameFlagObsOnDefeat:
		res += "|ObsOnDefeat"
	case GameFlagObsNone:
		res += "|ObsNone"
	case GameFlagObsMask:
		res += "|ObsAny"
	case 0:
		// No obs info
	default:
		return fmt.Sprintf("GameFlags(0x%06X)", uint32(f))
	}

	f &= ^(GameFlagTypeMask | GameFlagPrivateGame | GameFlagCreatorMask | GameFlagSizeMask | GameFlagMapTypeMask | GameFlagObsMask)

	if f != 0 {
		res += fmt.Sprintf("|GameFlags(0x%02X)", uint32(f))
	}
	if res != "" {
		res = res[1:]
	}

	return res
}
