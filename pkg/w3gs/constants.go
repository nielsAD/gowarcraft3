package w3gs

import (
	"errors"
	"fmt"

	"github.com/nielsAD/noot/pkg/util"
)

// Errors
var (
	ErrNoProtocolSig     = errors.New("w3gs: Invalid w3gs packet (no signature found)")
	ErrInvalidPacketSize = errors.New("w3gs: Invalid packet size")
	ErrInvalidChecksum   = errors.New("w3gs: Checksum invalid")
	ErrUnexpectedConst   = errors.New("w3gs: Unexpected constant value")
)

// ProtocolSig is the W3GS magic number used in the packet header.
const ProtocolSig = 0xF7

// W3GS packet type identifiers
const (
	PidUnknownPacket     = 0x00
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
	PidPeerMask          = 0x3B
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

// GameProduct identifier
type GameProduct util.DWordString

// Game product
var (
	ProductDemo = GameProduct(util.DString("W3DM")) // Demo
	ProductROC  = GameProduct(util.DString("WAR3")) // ROC
	ProductTFT  = GameProduct(util.DString("W3XP")) // TFT
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
	LeaveDisconnect    LeaveReason = 0x01
	LeaveLost          LeaveReason = 0x07
	LeaveLostBuildings LeaveReason = 0x08
	LeaveWon           LeaveReason = 0x09
	LeaveDraw          LeaveReason = 0x0A
	LabeObserver       LeaveReason = 0x0B
	LeaveLobby         LeaveReason = 0x0D
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
	case LabeObserver:
		return "Observer"
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

// GameType enum
type GameType uint32

// Game type flags
const (
	GameTypeNewGame     GameType = 0x000001
	GameTypeLadder      GameType = 0x000020
	GameTypeSavedGame   GameType = 0x000200
	GameTypePrivateGame GameType = 0x000800
	GameTypeLoadMask             = GameTypeNewGame | GameTypeLadder | GameTypeSavedGame | GameTypePrivateGame

	GameTypeMapBlizzardSigned GameType = 0x000008
	GameTypeMapCustom         GameType = 0x002000
	GameTypeMapBlizzard       GameType = 0x004000
	GameTypeMapMask                    = GameTypeMapBlizzardSigned | GameTypeMapCustom | GameTypeMapBlizzard

	GameTypeTypeMelee    GameType = 0x008000
	GameTypeTypeScenario GameType = 0x010000
	GameTypeTypeMask              = GameTypeTypeMelee | GameTypeTypeScenario

	GameTypeSizeSmall  GameType = 0x020000
	GameTypeSizeMedium GameType = 0x040000
	GameTypeSizeLarge  GameType = 0x080000
	GameTypeSizeMask            = GameTypeSizeSmall | GameTypeSizeMedium | GameTypeSizeLarge

	GameTypeObsFull    GameType = 0x100000
	GameTypeObsOnDeath GameType = 0x200000
	GameTypeObsNone    GameType = 0x400000
	GameTypeObsMask             = GameTypeObsFull | GameTypeObsFull | GameTypeObsNone
)

func (f GameType) String() string {
	if f&(GameTypeLoadMask|GameTypeMapMask|GameTypeTypeMask|GameTypeSizeMask|GameTypeObsMask) != f {
		return fmt.Sprintf("GameType(0x%06X)", uint32(f))
	}

	var res string
	switch f & GameTypeMapMask {
	case GameTypeTypeMelee:
		res = "CustomMap"
	case GameTypeMapBlizzardSigned, GameTypeMapBlizzard:
		res = "BlizzardMap"
	case 0:
		// No map maker
	default:
		return fmt.Sprintf("GameType(0x%06X)", uint32(f))
	}

	switch f & GameTypeSizeMask {
	case GameTypeSizeSmall:
		res += "|Small"
	case GameTypeSizeMedium:
		res += "|Medium"
	case GameTypeSizeLarge:
		res += "|Large"
	case 0:
		// No map size
	default:
		return fmt.Sprintf("GameType(0x%06X)", uint32(f))
	}

	switch f & GameTypeTypeMask {
	case GameTypeMapCustom:
		res += "|Melee"
	case GameTypeTypeScenario:
		res += "|Scenario"
	case 0:
		// No map type
	default:
		return fmt.Sprintf("GameType(0x%06X)", uint32(f))
	}

	switch f & GameTypeObsMask {
	case GameTypeObsFull:
		res += "|FullObs"
	case GameTypeObsOnDeath:
		res += "|ObsOnDeath"
	case GameTypeObsNone:
		res += "|NoObs"
	case 0:
		// No obs info
	default:
		return fmt.Sprintf("GameType(0x%06X)", uint32(f))
	}

	if f&GameTypeNewGame != 0 {
		res += "|NewGame"
	}
	if f&GameTypeSavedGame != 0 {
		res += "|SavedGame"
	}
	if f&GameTypeLadder != 0 {
		res += "|Ladder"
	}
	if f&GameTypePrivateGame != 0 {
		res += "|Private"
	}

	return res
}
