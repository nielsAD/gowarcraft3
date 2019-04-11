// Author:  Niels A.D.
// Project: gowarcraft3 (https://github.com/nielsAD/gowarcraft3)
// License: Mozilla Public License, v2.0

package w3gs

import (
	"io"

	"github.com/nielsAD/gowarcraft3/protocol"
)

// Encoder keeps amortized allocs at 0 for repeated Packet.Serialize calls.
// Byte slices are valid until the next Serialize() call.
type Encoder struct {
	Encoding
	buf protocol.Buffer
}

// NewEncoder initialization
func NewEncoder(e Encoding) *Encoder {
	return &Encoder{
		Encoding: e,
	}
}

// Serialize packet and returns its byte representation.
// Result is valid until the next Serialize() call.
func (enc *Encoder) Serialize(p Packet) ([]byte, error) {
	enc.buf.Truncate()
	if err := p.Serialize(&enc.buf, &enc.Encoding); err != nil {
		return nil, err
	}
	return enc.buf.Bytes, nil
}

// Write serializes p and writes it to w.
func (enc *Encoder) Write(w io.Writer, p Packet) (int, error) {
	b, err := enc.Serialize(p)
	if err != nil {
		return 0, err
	}

	return w.Write(b)
}

// Serialize serializes p and returns its byte representation.
func Serialize(p Packet, e Encoding) ([]byte, error) {
	return NewEncoder(e).Serialize(p)
}

// Write serializes p and writes it to w.
func Write(w io.Writer, p Packet, e Encoding) (int, error) {
	return NewEncoder(e).Write(w, p)
}

// Decoder keeps amortized allocs at 0 for repeated Packet.Deserialize calls.
// Packets are valid until the next Deserialize() call.
type Decoder struct {
	Encoding
	bufRaw protocol.Buffer
	bufDes protocol.Buffer

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
	peerSet        PeerSet
	mapCheck       MapCheck
	startDownload  StartDownload
	mapState       MapState
	mapPart        MapPart
	mapPartOK      MapPartOK
	mapPartError   MapPartError
	pong           Pong
	unknownPacket  UnknownPacket
}

// NewDecoder initialization
func NewDecoder(e Encoding) *Decoder {
	return &Decoder{
		Encoding: e,
	}
}

// Deserialize reads exactly one packet from b and returns it in the proper (deserialized) packet type.
// Result is valid until the next Deserialize() call.
func (dec *Decoder) Deserialize(b []byte) (Packet, int, error) {
	dec.bufDes.Reset(b)

	var size = dec.bufDes.Size()
	if size < 4 || b[0] != ProtocolSig {
		return nil, 0, ErrNoProtocolSig
	}

	var pkt Packet
	var err error

	// Explicitly call deserialize on type instead of interface for compiler optimizations
	switch b[1] {
	case PidPingFromHost:
		err = dec.ping.Deserialize(&dec.bufDes, &dec.Encoding)
		pkt = &dec.ping
	case PidSlotInfoJoin:
		err = dec.slotInfoJoin.Deserialize(&dec.bufDes, &dec.Encoding)
		pkt = &dec.slotInfoJoin
	case PidRejectJoin:
		err = dec.rejectJoin.Deserialize(&dec.bufDes, &dec.Encoding)
		pkt = &dec.rejectJoin
	case PidPlayerInfo:
		err = dec.playerInfo.Deserialize(&dec.bufDes, &dec.Encoding)
		pkt = &dec.playerInfo
	case PidPlayerLeft:
		err = dec.playerLeft.Deserialize(&dec.bufDes, &dec.Encoding)
		pkt = &dec.playerLeft
	case PidPlayerLoaded:
		err = dec.playerLoaded.Deserialize(&dec.bufDes, &dec.Encoding)
		pkt = &dec.playerLoaded
	case PidSlotInfo:
		err = dec.slotInfo.Deserialize(&dec.bufDes, &dec.Encoding)
		pkt = &dec.slotInfo
	case PidCountDownStart:
		err = dec.countDownStart.Deserialize(&dec.bufDes, &dec.Encoding)
		pkt = &dec.countDownStart
	case PidCountDownEnd:
		err = dec.countDownEnd.Deserialize(&dec.bufDes, &dec.Encoding)
		pkt = &dec.countDownEnd
	case PidIncomingAction:
		err = dec.timeSlot.Deserialize(&dec.bufDes, &dec.Encoding)
		pkt = &dec.timeSlot
	case PidDesync:
		err = dec.desync.Deserialize(&dec.bufDes, &dec.Encoding)
		pkt = &dec.desync
	case PidChatFromHost:
		err = dec.messageRelay.Deserialize(&dec.bufDes, &dec.Encoding)
		pkt = &dec.messageRelay
	case PidStartLag:
		err = dec.startLag.Deserialize(&dec.bufDes, &dec.Encoding)
		pkt = &dec.startLag
	case PidStopLag:
		err = dec.stopLag.Deserialize(&dec.bufDes, &dec.Encoding)
		pkt = &dec.stopLag
	case PidGameOver:
		err = dec.gameOver.Deserialize(&dec.bufDes, &dec.Encoding)
		pkt = &dec.gameOver
	case PidPlayerKicked:
		err = dec.playerKicked.Deserialize(&dec.bufDes, &dec.Encoding)
		pkt = &dec.playerKicked
	case PidLeaveAck:
		err = dec.leaveAck.Deserialize(&dec.bufDes, &dec.Encoding)
		pkt = &dec.leaveAck
	case PidReqJoin:
		err = dec.join.Deserialize(&dec.bufDes, &dec.Encoding)
		pkt = &dec.join
	case PidLeaveReq:
		err = dec.leave.Deserialize(&dec.bufDes, &dec.Encoding)
		pkt = &dec.leave
	case PidGameLoadedSelf:
		err = dec.gameLoaded.Deserialize(&dec.bufDes, &dec.Encoding)
		pkt = &dec.gameLoaded
	case PidOutgoingAction:
		err = dec.gameAction.Deserialize(&dec.bufDes, &dec.Encoding)
		pkt = &dec.gameAction
	case PidOutgoingKeepAlive:
		err = dec.timeSlotAck.Deserialize(&dec.bufDes, &dec.Encoding)
		pkt = &dec.timeSlotAck
	case PidChatToHost:
		err = dec.message.Deserialize(&dec.bufDes, &dec.Encoding)
		pkt = &dec.message
	case PidDropReq:
		err = dec.dropLaggers.Deserialize(&dec.bufDes, &dec.Encoding)
		pkt = &dec.dropLaggers
	case PidSearchGame:
		err = dec.searchGame.Deserialize(&dec.bufDes, &dec.Encoding)
		pkt = &dec.searchGame
	case PidGameInfo:
		err = dec.gameInfo.Deserialize(&dec.bufDes, &dec.Encoding)
		pkt = &dec.gameInfo
	case PidCreateGame:
		err = dec.createGame.Deserialize(&dec.bufDes, &dec.Encoding)
		pkt = &dec.createGame
	case PidRefreshGame:
		err = dec.refreshGame.Deserialize(&dec.bufDes, &dec.Encoding)
		pkt = &dec.refreshGame
	case PidDecreateGame:
		err = dec.decreateGame.Deserialize(&dec.bufDes, &dec.Encoding)
		pkt = &dec.decreateGame
	case PidChatFromOthers:
		err = dec.peerMessage.Deserialize(&dec.bufDes, &dec.Encoding)
		pkt = &dec.peerMessage
	case PidPingFromOthers:
		err = dec.peerPing.Deserialize(&dec.bufDes, &dec.Encoding)
		pkt = &dec.peerPing
	case PidPongToOthers:
		err = dec.peerPong.Deserialize(&dec.bufDes, &dec.Encoding)
		pkt = &dec.peerPong
	case PidClientInfo:
		err = dec.peerConnect.Deserialize(&dec.bufDes, &dec.Encoding)
		pkt = &dec.peerConnect
	case PidPeerSet:
		err = dec.peerSet.Deserialize(&dec.bufDes, &dec.Encoding)
		pkt = &dec.peerSet
	case PidMapCheck:
		err = dec.mapCheck.Deserialize(&dec.bufDes, &dec.Encoding)
		pkt = &dec.mapCheck
	case PidStartDownload:
		err = dec.startDownload.Deserialize(&dec.bufDes, &dec.Encoding)
		pkt = &dec.startDownload
	case PidMapSize:
		err = dec.mapState.Deserialize(&dec.bufDes, &dec.Encoding)
		pkt = &dec.mapState
	case PidMapPart:
		err = dec.mapPart.Deserialize(&dec.bufDes, &dec.Encoding)
		pkt = &dec.mapPart
	case PidMapPartOK:
		err = dec.mapPartOK.Deserialize(&dec.bufDes, &dec.Encoding)
		pkt = &dec.mapPartOK
	case PidMapPartError:
		err = dec.mapPartError.Deserialize(&dec.bufDes, &dec.Encoding)
		pkt = &dec.mapPartError
	case PidPongToHost:
		err = dec.pong.Deserialize(&dec.bufDes, &dec.Encoding)
		pkt = &dec.pong
	case PidIncomingAction2:
		err = dec.timeSlot.Deserialize(&dec.bufDes, &dec.Encoding)
		pkt = &dec.timeSlot
	default:
		err = dec.unknownPacket.Deserialize(&dec.bufDes, &dec.Encoding)
		pkt = &dec.unknownPacket
	}

	var n = size - dec.bufDes.Size()
	if err != nil {
		return nil, n, err
	}

	return pkt, n, nil
}

// ReadRaw reads exactly one packet from r and returns its raw bytes.
// Result is valid until the next ReadRaw() call.
func (dec *Decoder) ReadRaw(r io.Reader) ([]byte, int, error) {
	dec.bufRaw.Truncate()

	if n, err := dec.bufRaw.ReadSizeFrom(r, 4); err != nil {
		if err == io.ErrUnexpectedEOF {
			err = ErrNoProtocolSig
		}

		return nil, int(n), err
	}

	if dec.bufRaw.Bytes[0] != ProtocolSig {
		return nil, 4, ErrNoProtocolSig
	}

	var size = int(uint16(dec.bufRaw.Bytes[3])<<8 | uint16(dec.bufRaw.Bytes[2]))
	if size < 4 {
		return nil, 4, ErrNoProtocolSig
	}

	if n, err := dec.bufRaw.ReadSizeFrom(r, size-4); err != nil {
		if err == io.EOF {
			err = io.ErrUnexpectedEOF
		}
		return nil, int(n) + 4, err
	}

	return dec.bufRaw.Bytes, size, nil
}

// Read exactly one packet from r and returns it in the proper (deserialized) packet type.
func (dec *Decoder) Read(r io.Reader) (Packet, int, error) {
	b, n, err := dec.ReadRaw(r)
	if err != nil {
		return nil, n, err
	}

	p, m, err := dec.Deserialize(b)
	if err != nil {
		return nil, n, err
	}
	if m != n {
		return nil, n, ErrInvalidPacketSize
	}

	return p, n, nil
}

// Deserialize reads exactly one packet from b and returns it in the proper (deserialized) packet type.
func Deserialize(b []byte, e Encoding) (Packet, int, error) {
	return NewDecoder(e).Deserialize(b)
}

// Read exactly one packet from r and returns it in the proper (deserialized) packet type.
func Read(r io.Reader, e Encoding) (Packet, int, error) {
	return NewDecoder(e).Read(r)
}
