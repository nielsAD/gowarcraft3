package w3gs

import (
	"io"

	"github.com/nielsAD/gowarcraft3/protocol"
)

// SerializationBuffer is used by SerializePacketWithBuffer to bring amortized allocs to 0 for repeated calls
type SerializationBuffer = protocol.Buffer

// SerializePacketWithBuffer serializes p and writes it to w.
func SerializePacketWithBuffer(w io.Writer, b *SerializationBuffer, p Packet) (int, error) {
	b.Truncate()
	if err := p.Serialize(b); err != nil {
		return 0, err
	}
	return w.Write(b.Bytes)
}

// SerializePacket serializes p and writes it to w.
func SerializePacket(w io.Writer, p Packet) (int, error) {
	return SerializePacketWithBuffer(w, &SerializationBuffer{}, p)
}

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
	desync         Desync
	messageRelay   MessageRelay
	startLag       StartLag
	stopLag        StopLag
	gameOver       GameOver
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
	peerMessage    PeerMessage
	peerPing       PeerPing
	peerPong       PeerPong
	peerConnect    PeerConnect
	peerMask       PeerMask
	mapCheck       MapCheck
	startDownload  StartDownload
	mapState       MapState
	mapPart        MapPart
	mapPartOK      MapPartOK
	mapPartError   MapPartError
	pong           Pong
	unknownPacket  UnknownPacket
}

// ReadPacketWithBuffer reads exactly one packet from r and returns its raw bytes.
func ReadPacketWithBuffer(r io.Reader, b *DeserializationBuffer) ([]byte, int, error) {
	if n, err := io.ReadFull(r, b.Buffer[:4]); err != nil {
		if err == io.ErrUnexpectedEOF {
			err = ErrNoProtocolSig
		}

		return nil, n, err
	}

	if b.Buffer[0] != ProtocolSig {
		return nil, 4, ErrNoProtocolSig
	}

	var size = int(uint16(b.Buffer[3])<<8 | uint16(b.Buffer[2]))
	if size < 4 || size > len(b.Buffer) {
		return nil, 4, ErrInvalidPacketSize
	}

	if n, err := io.ReadFull(r, b.Buffer[4:size]); err != nil {
		if err == io.EOF {
			err = io.ErrUnexpectedEOF
		}
		return nil, n + 4, err
	}

	return b.Buffer[:size], size, nil
}

// DeserializePacketWithBuffer reads exactly one packet from r and returns it in the proper (deserialized) packet type.
func DeserializePacketWithBuffer(r io.Reader, b *DeserializationBuffer) (Packet, int, error) {
	var bytes, n, err = ReadPacketWithBuffer(r, b)
	if err != nil {
		return nil, n, err
	}

	var pbuf = protocol.Buffer{Bytes: bytes}

	var pkt Packet

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
	case PidDesync:
		err = b.desync.Deserialize(&pbuf)
		pkt = &b.desync
	case PidChatFromHost:
		err = b.messageRelay.Deserialize(&pbuf)
		pkt = &b.messageRelay
	case PidStartLag:
		err = b.startLag.Deserialize(&pbuf)
		pkt = &b.startLag
	case PidStopLag:
		err = b.stopLag.Deserialize(&pbuf)
		pkt = &b.stopLag
	case PidGameOver:
		err = b.gameOver.Deserialize(&pbuf)
		pkt = &b.gameOver
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
	case PidChatFromOthers:
		err = b.peerMessage.Deserialize(&pbuf)
		pkt = &b.peerMessage
	case PidPingFromOthers:
		err = b.peerPing.Deserialize(&pbuf)
		pkt = &b.peerPing
	case PidPongToOthers:
		err = b.peerPong.Deserialize(&pbuf)
		pkt = &b.peerPong
	case PidClientInfo:
		err = b.peerConnect.Deserialize(&pbuf)
		pkt = &b.peerConnect
	case PidPeerMask:
		err = b.peerMask.Deserialize(&pbuf)
		pkt = &b.peerMask
	case PidMapCheck:
		err = b.mapCheck.Deserialize(&pbuf)
		pkt = &b.mapCheck
	case PidStartDownload:
		err = b.startDownload.Deserialize(&pbuf)
		pkt = &b.startDownload
	case PidMapSize:
		err = b.mapState.Deserialize(&pbuf)
		pkt = &b.mapState
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
		return nil, n, err
	}

	return pkt, n, nil
}

// ReadPacket reads exactly one packet from r and returns it in the proper (deserialized) packet type.
func ReadPacket(r io.Reader, b *DeserializationBuffer) ([]byte, int, error) {
	return ReadPacketWithBuffer(r, &DeserializationBuffer{})
}

// DeserializePacket reads exactly one packet from r and returns it in the proper (deserialized) packet type.
func DeserializePacket(r io.Reader) (Packet, int, error) {
	return DeserializePacketWithBuffer(r, &DeserializationBuffer{})
}
