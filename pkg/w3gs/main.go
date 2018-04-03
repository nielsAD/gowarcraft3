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
	"encoding/binary"
	"errors"
	"io"

	"github.com/nielsAD/noot/pkg/util"
)

// Packet interface.
type Packet interface {
	Serialize(buf *util.PacketBuffer) error
	Deserialize(buf *util.PacketBuffer) error
}

// Errors
var (
	ErrNoProtocolSig   = errors.New("w3gs: Invalid w3gs packet (no signature found)")
	ErrWrongSize       = errors.New("w3gs: Wrong size input data")
	ErrMalformedData   = errors.New("w3gs: Malformed input data")
	ErrInvalidChecksum = errors.New("w3gs: Checksum invalid")
	ErrBufferTooSmall  = errors.New("w3gs: Buffer too small")
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

// DeserializePacketWithBuffer reads exactly one packet from r (using b as buffer) and
// returns it in the proper (deserialized) packet type. Buffer should be large enough
// to hold the entire packet (in general, wc3 doesn't send packets larger than ~1500b)
func DeserializePacketWithBuffer(r io.Reader, b []byte) (Packet, int, error) {
	if len(b) < 4 {
		return nil, 0, ErrBufferTooSmall
	}
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
	if len(b) < size {
		return nil, 4, ErrBufferTooSmall
	}

	if n, err := io.ReadFull(r, b[4:size]); err != nil {
		if err == io.EOF {
			err = io.ErrUnexpectedEOF
		}
		return nil, n + 4, err
	}

	var buf = util.PacketBuffer{Bytes: b[:size]}

	var pkt Packet
	var err error

	// Explicitly call deserialize on type instead of interface for compiler optimizations
	switch b[1] {
	case PidPingFromHost:
		var tmp Ping
		err = tmp.Deserialize(&buf)
		pkt = &tmp
	case PidSlotInfoJoin:
		var tmp SlotInfoJoin
		err = tmp.Deserialize(&buf)
		pkt = &tmp
	case PidRejectJoin:
		var tmp RejectJoin
		err = tmp.Deserialize(&buf)
		pkt = &tmp
	case PidPlayerInfo:
		var tmp PlayerInfo
		err = tmp.Deserialize(&buf)
		pkt = &tmp
	case PidPlayerLeft:
		var tmp PlayerLeft
		err = tmp.Deserialize(&buf)
		pkt = &tmp
	case PidPlayerLoaded:
		var tmp PlayerLoaded
		err = tmp.Deserialize(&buf)
		pkt = &tmp
	case PidSlotInfo:
		var tmp SlotInfo
		err = tmp.Deserialize(&buf)
		pkt = &tmp
	case PidCountDownStart:
		var tmp CountDownStart
		err = tmp.Deserialize(&buf)
		pkt = &tmp
	case PidCountDownEnd:
		var tmp CountDownEnd
		err = tmp.Deserialize(&buf)
		pkt = &tmp
	case PidIncomingAction:
		var tmp TimeSlot
		err = tmp.Deserialize(&buf)
		pkt = &tmp
	case PidChatFromHost:
		var tmp MessageRelay
		err = tmp.Deserialize(&buf)
		pkt = &tmp
	case PidStartLag:
		var tmp StartLag
		err = tmp.Deserialize(&buf)
		pkt = &tmp
	case PidStopLag:
		var tmp StopLag
		err = tmp.Deserialize(&buf)
		pkt = &tmp
	case PidPlayerKicked:
		var tmp PlayerKicked
		err = tmp.Deserialize(&buf)
		pkt = &tmp
	case PidLeaveAck:
		var tmp LeaveAck
		err = tmp.Deserialize(&buf)
		pkt = &tmp
	case PidReqJoin:
		var tmp Join
		err = tmp.Deserialize(&buf)
		pkt = &tmp
	case PidLeaveReq:
		var tmp Leave
		err = tmp.Deserialize(&buf)
		pkt = &tmp
	case PidGameLoadedSelf:
		var tmp GameLoaded
		err = tmp.Deserialize(&buf)
		pkt = &tmp
	case PidOutgoingAction:
		var tmp GameAction
		err = tmp.Deserialize(&buf)
		pkt = &tmp
	case PidOutgoingKeepAlive:
		var tmp TimeSlotAck
		err = tmp.Deserialize(&buf)
		pkt = &tmp
	case PidChatToHost:
		var tmp Message
		err = tmp.Deserialize(&buf)
		pkt = &tmp
	case PidDropReq:
		var tmp DropLaggers
		err = tmp.Deserialize(&buf)
		pkt = &tmp
	case PidSearchGame:
		var tmp SearchGame
		err = tmp.Deserialize(&buf)
		pkt = &tmp
	case PidGameInfo:
		var tmp GameInfo
		err = tmp.Deserialize(&buf)
		pkt = &tmp
	case PidCreateGame:
		var tmp CreateGame
		err = tmp.Deserialize(&buf)
		pkt = &tmp
	case PidRefreshGame:
		var tmp RefreshGame
		err = tmp.Deserialize(&buf)
		pkt = &tmp
	case PidDecreateGame:
		var tmp DecreateGame
		err = tmp.Deserialize(&buf)
		pkt = &tmp
	case PidPingFromOthers:
		var tmp PeerPing
		err = tmp.Deserialize(&buf)
		pkt = &tmp
	case PidPongToOthers:
		var tmp PeerPong
		err = tmp.Deserialize(&buf)
		pkt = &tmp
	case PidClientInfo:
		var tmp ClientInfo
		err = tmp.Deserialize(&buf)
		pkt = &tmp
	case PidMapCheck:
		var tmp MapCheck
		err = tmp.Deserialize(&buf)
		pkt = &tmp
	case PidStartDownload:
		var tmp StartDownload
		err = tmp.Deserialize(&buf)
		pkt = &tmp
	case PidMapSize:
		var tmp MapSize
		err = tmp.Deserialize(&buf)
		pkt = &tmp
	case PidMapPart:
		var tmp MapPart
		err = tmp.Deserialize(&buf)
		pkt = &tmp
	case PidMapPartOK:
		var tmp MapPartOK
		err = tmp.Deserialize(&buf)
		pkt = &tmp
	case PidMapPartError:
		var tmp MapPartError
		err = tmp.Deserialize(&buf)
		pkt = &tmp
	case PidPongToHost:
		var tmp Pong
		err = tmp.Deserialize(&buf)
		pkt = &tmp
	case PidIncomingAction2:
		var tmp TimeSlot
		err = tmp.Deserialize(&buf)
		pkt = &tmp
	default:
		var tmp UnknownPacket
		err = tmp.Deserialize(&buf)
		pkt = &tmp
	}

	if err != nil {
		return nil, size, err
	}

	return pkt, size, nil
}

// DeserializePacket reads exactly one packet from r andreturns it in the proper (deserialized) packet type.
func DeserializePacket(r io.Reader) (Packet, int, error) {
	var buf [2048]byte
	return DeserializePacketWithBuffer(r, buf[:])
}
