// Package w3gs implements the game protocol for Warcraft III.
//
// Based on the protocol documentation from https://bnetdocs.org/
//
// General format:
//
//    (UINT8)  Protocol signature (0xF7)
//    (UINT8)  Packet type ID
//    (UINT16) Packet size
//    [Packet Data]
//
package w3gs

import (
	"encoding"
	"encoding/binary"
)

// Packet interface.
type Packet interface {
	encoding.BinaryMarshaler
	encoding.BinaryUnmarshaler
}

// W3GS magic number
const protocolsig = 0xF7

// AF_INET
const connAddressFamily uint16 = 2

// W3GS packet type identifiers
const (
	pidUnknownPacket     = 0x00
	pidPingFromHost      = 0x01
	pidSlotInfoJoin      = 0x04
	pidRejectJoin        = 0x05
	pidPlayerInfo        = 0x06
	pidPlayerLeft        = 0x07
	pidPlayerLoaded      = 0x08
	pidSlotInfo          = 0x09
	pidCountDownStart    = 0x0A
	pidCountDownEnd      = 0x0B
	pidIncomingAction    = 0x0C
	pidChatFromHost      = 0x0F
	pidStartLag          = 0x10
	pidStopLag           = 0x11
	pidPlayerKicked      = 0x1C
	pidLeaveAck          = 0x1B
	pidReqJoin           = 0x1E
	pidLeaveReq          = 0x21
	pidGameLoadedSelf    = 0x23
	pidOutgoingAction    = 0x26
	pidOutgoingKeepAlive = 0x27
	pidChatToHost        = 0x28
	pidDropReq           = 0x29
	pidSearchGame        = 0x2F
	pidGameInfo          = 0x30
	pidCreateGame        = 0x31
	pidRefreshGame       = 0x32
	pidDecreateGame      = 0x33
	pidPingFromOthers    = 0x35
	pidPongToOthers      = 0x36
	pidClientInfo        = 0x37
	pidMapCheck          = 0x3D
	pidStartDownload     = 0x3F
	pidMapSize           = 0x42
	pidMapPart           = 0x43
	pidMapPartOK         = 0x44
	pidMapPartError      = 0x45
	pidPongToHost        = 0x46
	pidIncomingAction2   = 0x48
)

//pidChatOthers = 0x34 (?)
//pidGameOver   = 0x14 (?) [Payload is a single byte (PlayerID?) after game is over]

// Game versions
var (
	gameWAR3 = "3RAW"
	gameW3XP = "PX3W"
)

// Slot layout
const (
	LayoutMelee               = 0x00
	LayoutCustomForces        = 0x01
	LayoutFixedPlayerSettings = 0x02
)

// Slot status
const (
	SlotOpen     = 0x00
	SlotClosed   = 0x01
	SlotOccupied = 0x02
)

// Race preference
const (
	RaceHuman      = 0x01
	RaceOrc        = 0x02
	RaceNightElf   = 0x04
	RaceUndead     = 0x08
	RaceDemon      = 0x10
	RaceRandom     = 0x20
	RaceSelectable = 0x40
)

// AI difficulty
const (
	ComputerEasy   = 0x00
	ComputerNormal = 0x01
	ComputerInsane = 0x02
)

// RejectJoin reason
const (
	RejectJoinInvalid   = 0x07
	RejectJoinFull      = 0x09
	RejectJoinStarted   = 0x0A
	RejectJoinWrongPass = 0x1B
)

// PlayerLeft reason
const (
	LeaveDisconnect    = 0x01
	LeaveLost          = 0x07
	LeaveLostBuildings = 0x08
	LeaveWon           = 0x09
	LeaveDraw          = 0x0A
	LabeObserver       = 0x0B
	LeaveLobby         = 0x0D
)

// Chat type
const (
	ChatMessage        = 0x10
	ChatTeamChange     = 0x11
	ChatColorChange    = 0x12
	ChatRaceChange     = 0x13
	ChatHandicapChange = 0x14
	ChatMessageExtra   = 0x20
)

// Game type flags
const (
	GameTypeNewGame        = 0x000001
	GameTypeBlizzardSigned = 0x000008
	GameTypeLadder         = 0x000020
	GameTypeSavedGame      = 0x000200
	GameTypePrivateGame    = 0x000800
	GameTypeUserMap        = 0x002000
	GameTypeBlizzardMap    = 0x004000
	GameTypeTypeMelee      = 0x008000
	GameTypeTypeScenario   = 0x010000
	GameTypeSizeSmall      = 0x020000
	GameTypeSizeMedium     = 0x040000
	GameTypeSizeLarge      = 0x080000
	GameTypeObsFull        = 0x100000
	GameTypeObsOnDeath     = 0x200000
	GameTypeObsNone        = 0x400000

	GameTypeMaskObs   = GameTypeObsFull | GameTypeObsFull | GameTypeObsNone
	GameTypeMaskMaker = GameTypeUserMap | GameTypeBlizzardMap
	GameTypeMaskSize  = GameTypeSizeSmall | GameTypeSizeMedium | GameTypeSizeLarge
)

// UnmarshalPacket decodes binary data and returns it in the proper (unmarshalled) packet type.
func UnmarshalPacket(data []byte) (Packet, int, error) {
	if len(data) < 4 || (data)[0] != protocolsig {
		return nil, 0, errMalformedData
	}
	var size = int(binary.LittleEndian.Uint16(data[2:]))
	if size > len(data) {
		return nil, 0, errMalformedData
	}

	var pkt Packet

	switch data[1] {
	case pidPingFromHost:
		pkt = &PingFromHost{}
	case pidSlotInfoJoin:
		pkt = &SlotInfoJoin{}
	case pidRejectJoin:
		pkt = &RejectJoin{}
	case pidPlayerInfo:
		pkt = &PlayerInfo{}
	case pidPlayerLeft:
		pkt = &PlayerLeft{}
	case pidPlayerLoaded:
		pkt = &PlayerLoaded{}
	case pidSlotInfo:
		pkt = &SlotInfo{}
	case pidCountDownStart:
		pkt = &CountDownStart{}
	case pidCountDownEnd:
		pkt = &CountDownEnd{}
	case pidIncomingAction:
		pkt = &IncomingAction{}
	case pidChatFromHost:
		pkt = &ChatFromHost{}
	case pidStartLag:
		pkt = &StartLag{}
	case pidStopLag:
		pkt = &StopLag{}
	case pidPlayerKicked:
		pkt = &LeaveReq{}
	case pidLeaveAck:
		pkt = &LeaveAck{}
	case pidReqJoin:
		pkt = &ReqJoin{}
	case pidLeaveReq:
		pkt = &LeaveReq{}
	case pidGameLoadedSelf:
		pkt = &GameLoadedSelf{}
	case pidOutgoingAction:
		pkt = &OutgoingAction{}
	case pidOutgoingKeepAlive:
		pkt = &OutgoingKeepAlive{}
	case pidChatToHost:
		pkt = &ChatToHost{}
	case pidDropReq:
		pkt = &DropReq{}
	case pidSearchGame:
		pkt = &SearchGame{}
	case pidGameInfo:
		pkt = &GameInfo{}
	case pidCreateGame:
		pkt = &CreateGame{}
	case pidRefreshGame:
		pkt = &RefreshGame{}
	case pidDecreateGame:
		pkt = &DecreateGame{}
	case pidPingFromOthers:
		pkt = &PingFromOthers{}
	case pidPongToOthers:
		pkt = &PongToOthers{}
	case pidClientInfo:
		pkt = &ClientInfo{}
	case pidMapCheck:
		pkt = &MapCheck{}
	case pidStartDownload:
		pkt = &StartDownload{}
	case pidMapSize:
		pkt = &MapSize{}
	case pidMapPart:
		pkt = &MapPart{}
	case pidMapPartOK:
		pkt = &MapPartOK{}
	case pidMapPartError:
		pkt = &MapPartError{}
	case pidPongToHost:
		pkt = &PongToHost{}
	case pidIncomingAction2:
		pkt = &IncomingAction{}
	default:
		pkt = &UnknownPacket{}
	}

	if err := pkt.UnmarshalBinary(data[:size]); err != nil {
		return nil, size, err
	}

	return pkt, size, nil
}
