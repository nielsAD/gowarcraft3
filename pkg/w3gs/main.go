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

// DeserializationBuffer is used by DeserializePacketWithBuffer to bring amortized allocs to 0 for repeated calls
type DeserializationBuffer struct {
	Buffer         [2048]byte
	ping           Ping
	slotInfoJoin   SlotInfoJoin
	rejectJoin     RejectJoin
	playerInfo     PlayerInfo
	playerLeft     PlayerLeft
	playerLoaded   PlayerLoaded
	slotInfo       SlotInfo
	countDownStart CountDownStart
	countDownEnd   CountDownEnd
	timeSlot       TimeSlot
	messageRelay   MessageRelay
	startLag       StartLag
	stopLag        StopLag
	playerKicked   PlayerKicked
	leaveAck       LeaveAck
	join           Join
	leave          Leave
	gameLoaded     GameLoaded
	gameAction     GameAction
	timeSlotAck    TimeSlotAck
	message        Message
	dropLaggers    DropLaggers
	searchGame     SearchGame
	gameInfo       GameInfo
	createGame     CreateGame
	refreshGame    RefreshGame
	decreateGame   DecreateGame
	peerPing       PeerPing
	peerPong       PeerPong
	clientInfo     ClientInfo
	mapCheck       MapCheck
	startDownload  StartDownload
	mapSize        MapSize
	mapPart        MapPart
	mapPartOK      MapPartOK
	mapPartError   MapPartError
	pong           Pong
	unknownPacket  UnknownPacket
}

// DeserializePacketWithBuffer reads exactly one packet from r (using b as buffer) and
// returns it in the proper (deserialized) packet type. Buffer should be large enough
// to hold the entire packet (in general, wc3 doesn't send packets larger than ~1500b)
func DeserializePacketWithBuffer(r io.Reader, b *DeserializationBuffer) (Packet, int, error) {
	if n, err := io.ReadFull(r, b.Buffer[:4]); err != nil {
		if err == io.ErrUnexpectedEOF {
			err = ErrNoProtocolSig
		}

		return nil, n, err
	}

	if b.Buffer[0] != ProtocolSig {
		return nil, 0, ErrNoProtocolSig
	}

	var size = int(uint16(b.Buffer[3])<<8 | uint16(b.Buffer[2]))
	if size < 4 || size > len(b.Buffer) {
		return nil, 0, ErrMalformedData
	}

	if n, err := io.ReadFull(r, b.Buffer[4:size]); err != nil {
		if err == io.EOF {
			err = io.ErrUnexpectedEOF
		}
		return nil, n + 4, err
	}

	var pbuf = util.PacketBuffer{Bytes: b.Buffer[:size]}

	var pkt Packet
	var err error

	// Explicitly call deserialize on type instead of interface for compiler optimizations
	switch b.Buffer[1] {
	case PidPingFromHost:
		err = b.ping.Deserialize(&pbuf)
		pkt = &b.ping
	case PidSlotInfoJoin:
		err = b.slotInfoJoin.Deserialize(&pbuf)
		pkt = &b.slotInfoJoin
	case PidRejectJoin:
		err = b.rejectJoin.Deserialize(&pbuf)
		pkt = &b.rejectJoin
	case PidPlayerInfo:
		err = b.playerInfo.Deserialize(&pbuf)
		pkt = &b.playerInfo
	case PidPlayerLeft:
		err = b.playerLeft.Deserialize(&pbuf)
		pkt = &b.playerLeft
	case PidPlayerLoaded:
		err = b.playerLoaded.Deserialize(&pbuf)
		pkt = &b.playerLoaded
	case PidSlotInfo:
		err = b.slotInfo.Deserialize(&pbuf)
		pkt = &b.slotInfo
	case PidCountDownStart:
		err = b.countDownStart.Deserialize(&pbuf)
		pkt = &b.countDownStart
	case PidCountDownEnd:
		err = b.countDownEnd.Deserialize(&pbuf)
		pkt = &b.countDownEnd
	case PidIncomingAction:
		err = b.timeSlot.Deserialize(&pbuf)
		pkt = &b.timeSlot
	case PidChatFromHost:
		err = b.messageRelay.Deserialize(&pbuf)
		pkt = &b.messageRelay
	case PidStartLag:
		err = b.startLag.Deserialize(&pbuf)
		pkt = &b.startLag
	case PidStopLag:
		err = b.stopLag.Deserialize(&pbuf)
		pkt = &b.stopLag
	case PidPlayerKicked:
		err = b.playerKicked.Deserialize(&pbuf)
		pkt = &b.playerKicked
	case PidLeaveAck:
		err = b.leaveAck.Deserialize(&pbuf)
		pkt = &b.leaveAck
	case PidReqJoin:
		err = b.join.Deserialize(&pbuf)
		pkt = &b.join
	case PidLeaveReq:
		err = b.leave.Deserialize(&pbuf)
		pkt = &b.leave
	case PidGameLoadedSelf:
		err = b.gameLoaded.Deserialize(&pbuf)
		pkt = &b.gameLoaded
	case PidOutgoingAction:
		err = b.gameAction.Deserialize(&pbuf)
		pkt = &b.gameAction
	case PidOutgoingKeepAlive:
		err = b.timeSlotAck.Deserialize(&pbuf)
		pkt = &b.timeSlotAck
	case PidChatToHost:
		err = b.message.Deserialize(&pbuf)
		pkt = &b.message
	case PidDropReq:
		err = b.dropLaggers.Deserialize(&pbuf)
		pkt = &b.dropLaggers
	case PidSearchGame:
		err = b.searchGame.Deserialize(&pbuf)
		pkt = &b.searchGame
	case PidGameInfo:
		err = b.gameInfo.Deserialize(&pbuf)
		pkt = &b.gameInfo
	case PidCreateGame:
		err = b.createGame.Deserialize(&pbuf)
		pkt = &b.createGame
	case PidRefreshGame:
		err = b.refreshGame.Deserialize(&pbuf)
		pkt = &b.refreshGame
	case PidDecreateGame:
		err = b.decreateGame.Deserialize(&pbuf)
		pkt = &b.decreateGame
	case PidPingFromOthers:
		err = b.peerPing.Deserialize(&pbuf)
		pkt = &b.peerPing
	case PidPongToOthers:
		err = b.peerPong.Deserialize(&pbuf)
		pkt = &b.peerPong
	case PidClientInfo:
		err = b.clientInfo.Deserialize(&pbuf)
		pkt = &b.clientInfo
	case PidMapCheck:
		err = b.mapCheck.Deserialize(&pbuf)
		pkt = &b.mapCheck
	case PidStartDownload:
		err = b.startDownload.Deserialize(&pbuf)
		pkt = &b.startDownload
	case PidMapSize:
		err = b.mapSize.Deserialize(&pbuf)
		pkt = &b.mapSize
	case PidMapPart:
		err = b.mapPart.Deserialize(&pbuf)
		pkt = &b.mapPart
	case PidMapPartOK:
		err = b.mapPartOK.Deserialize(&pbuf)
		pkt = &b.mapPartOK
	case PidMapPartError:
		err = b.mapPartError.Deserialize(&pbuf)
		pkt = &b.mapPartError
	case PidPongToHost:
		err = b.pong.Deserialize(&pbuf)
		pkt = &b.pong
	case PidIncomingAction2:
		err = b.timeSlot.Deserialize(&pbuf)
		pkt = &b.timeSlot
	default:
		err = b.unknownPacket.Deserialize(&pbuf)
		pkt = &b.unknownPacket
	}

	if err != nil {
		return nil, size, err
	}

	return pkt, size, nil
}

// DeserializePacket reads exactly one packet from r andreturns it in the proper (deserialized) packet type.
func DeserializePacket(r io.Reader) (Packet, int, error) {
	return DeserializePacketWithBuffer(r, &DeserializationBuffer{})
}
