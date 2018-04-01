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
	"errors"
	"io"
)

// Packet interface.
type Packet interface {
	encoding.BinaryMarshaler
	encoding.BinaryUnmarshaler
}

// Errors
var (
	ErrNoProtocolSig   = errors.New("w3gs: Invalid w3gs packet (no signature found)")
	ErrWrongSize       = errors.New("w3gs: Wrong size input data")
	ErrMalformedData   = errors.New("w3gs: Malformed input data")
	ErrInvalidChecksum = errors.New("w3gs: Checksum invalid")
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
	PidChatFromHost      = 0x0F
	PidStartLag          = 0x10
	PidStopLag           = 0x11
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
	PidPingFromOthers    = 0x35
	PidPongToOthers      = 0x36
	PidClientInfo        = 0x37
	PidMapCheck          = 0x3D
	PidStartDownload     = 0x3F
	PidMapSize           = 0x42
	PidMapPart           = 0x43
	PidMapPartOK         = 0x44
	PidMapPartError      = 0x45
	PidPongToHost        = 0x46
	PidIncomingAction2   = 0x48
)

//PidChatOthers = 0x34 (?)
//PidGameOver   = 0x14 (?) [Payload is a single byte (PlayerID?) after game is over]

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

// UnmarshalPacketWithBuffer reads exactly one packet from r (using b as buffer) and
// returns it in the proper (unmarshalled) packet type. Buffer should be large enough
// to hold the entire packet (in general, wc3 doesn't send packets larger than ~1500b)
func UnmarshalPacketWithBuffer(r io.Reader, b []byte) (Packet, int, error) {
	if n, err := io.ReadFull(r, b[:4]); err != nil {
		if err == io.ErrUnexpectedEOF {
			err = ErrNoProtocolSig
		}

		return nil, n, err
	}

	if b[0] != ProtocolSig {
		return nil, 0, ErrNoProtocolSig
	}

	var size = int(binary.LittleEndian.Uint16(b[2:]))
	if size < 4 {
		return nil, 0, ErrMalformedData
	}

	if n, err := io.ReadFull(r, b[4:size]); err != nil {
		if err == io.EOF {
			err = io.ErrUnexpectedEOF
		}
		return nil, n, err
	}

	var pkt Packet

	switch b[1] {
	case PidPingFromHost:
		pkt = &Ping{}
	case PidSlotInfoJoin:
		pkt = &SlotInfoJoin{}
	case PidRejectJoin:
		pkt = &RejectJoin{}
	case PidPlayerInfo:
		pkt = &PlayerInfo{}
	case PidPlayerLeft:
		pkt = &PlayerLeft{}
	case PidPlayerLoaded:
		pkt = &PlayerLoaded{}
	case PidSlotInfo:
		pkt = &SlotInfo{}
	case PidCountDownStart:
		pkt = &CountDownStart{}
	case PidCountDownEnd:
		pkt = &CountDownEnd{}
	case PidIncomingAction:
		pkt = &TimeSlot{}
	case PidChatFromHost:
		pkt = &MessageRelay{}
	case PidStartLag:
		pkt = &StartLag{}
	case PidStopLag:
		pkt = &StopLag{}
	case PidPlayerKicked:
		pkt = &PlayerKicked{}
	case PidLeaveAck:
		pkt = &LeaveAck{}
	case PidReqJoin:
		pkt = &Join{}
	case PidLeaveReq:
		pkt = &Leave{}
	case PidGameLoadedSelf:
		pkt = &GameLoaded{}
	case PidOutgoingAction:
		pkt = &GameAction{}
	case PidOutgoingKeepAlive:
		pkt = &TimeSlotAck{}
	case PidChatToHost:
		pkt = &Message{}
	case PidDropReq:
		pkt = &DropLaggers{}
	case PidSearchGame:
		pkt = &SearchGame{}
	case PidGameInfo:
		pkt = &GameInfo{}
	case PidCreateGame:
		pkt = &CreateGame{}
	case PidRefreshGame:
		pkt = &RefreshGame{}
	case PidDecreateGame:
		pkt = &DecreateGame{}
	case PidPingFromOthers:
		pkt = &PeerPing{}
	case PidPongToOthers:
		pkt = &PeerPong{}
	case PidClientInfo:
		pkt = &ClientInfo{}
	case PidMapCheck:
		pkt = &MapCheck{}
	case PidStartDownload:
		pkt = &StartDownload{}
	case PidMapSize:
		pkt = &MapSize{}
	case PidMapPart:
		pkt = &MapPart{}
	case PidMapPartOK:
		pkt = &MapPartOK{}
	case PidMapPartError:
		pkt = &MapPartError{}
	case PidPongToHost:
		pkt = &Pong{}
	case PidIncomingAction2:
		pkt = &TimeSlot{}
	default:
		pkt = &UnknownPacket{}
	}

	if err := pkt.UnmarshalBinary(b[:size]); err != nil {
		return nil, size, err
	}

	return pkt, size, nil
}

// UnmarshalPacket reads exactly one packet from r andreturns it in the proper (unmarshalled) packet type.
func UnmarshalPacket(r io.Reader) (Packet, int, error) {
	var buf [2048]byte
	return UnmarshalPacketWithBuffer(r, buf[:])
}
