// Author:  Niels A.D.
// Project: gowarcraft3 (https://github.com/nielsAD/gowarcraft3)
// License: Mozilla Public License, v2.0

// Package w3gs implements the game protocol for Warcraft III.
//
// Based on protocol documentation by https://bnetdocs.org/
//
// Each packet type is mapped to a struct type that implements the Packet
// interface. To deserialize from a binary stream, use w3gs.Read().
//
// This package tries to keep ammortized heap memory allocations to 0.
//
// General serialization format:
//
//    (UINT8)  Protocol signature (0xF7)
//    (UINT8)  Packet type ID
//    (UINT16) Packet size
//    [Packet Data]
//
package w3gs

import "github.com/nielsAD/gowarcraft3/protocol"

// Packet interface.
type Packet interface {
	Serialize(buf *protocol.Buffer, enc *Encoding) error
	Deserialize(buf *protocol.Buffer, enc *Encoding) error
}

// Encoding options for (de)serialization
type Encoding struct {
	GameVersion uint32
}

// DefaultFactory maps packet ID to matching type
var DefaultFactory = MapFactory{
	PidPingFromHost:      func(_ *Encoding) Packet { return &Ping{} },
	PidSlotInfoJoin:      func(_ *Encoding) Packet { return &SlotInfoJoin{} },
	PidRejectJoin:        func(_ *Encoding) Packet { return &RejectJoin{} },
	PidPlayerInfo:        func(_ *Encoding) Packet { return &PlayerInfo{} },
	PidPlayerLeft:        func(_ *Encoding) Packet { return &PlayerLeft{} },
	PidPlayerLoaded:      func(_ *Encoding) Packet { return &PlayerLoaded{} },
	PidSlotInfo:          func(_ *Encoding) Packet { return &SlotInfo{} },
	PidCountDownStart:    func(_ *Encoding) Packet { return &CountDownStart{} },
	PidCountDownEnd:      func(_ *Encoding) Packet { return &CountDownEnd{} },
	PidIncomingAction:    func(_ *Encoding) Packet { return &TimeSlot{} },
	PidDesync:            func(_ *Encoding) Packet { return &Desync{} },
	PidChatFromHost:      func(_ *Encoding) Packet { return &MessageRelay{} },
	PidStartLag:          func(_ *Encoding) Packet { return &StartLag{} },
	PidStopLag:           func(_ *Encoding) Packet { return &StopLag{} },
	PidGameOver:          func(_ *Encoding) Packet { return &GameOver{} },
	PidPlayerKicked:      func(_ *Encoding) Packet { return &PlayerKicked{} },
	PidLeaveAck:          func(_ *Encoding) Packet { return &LeaveAck{} },
	PidReqJoin:           func(_ *Encoding) Packet { return &Join{} },
	PidLeaveReq:          func(_ *Encoding) Packet { return &Leave{} },
	PidGameLoadedSelf:    func(_ *Encoding) Packet { return &GameLoaded{} },
	PidOutgoingAction:    func(_ *Encoding) Packet { return &GameAction{} },
	PidOutgoingKeepAlive: func(_ *Encoding) Packet { return &TimeSlotAck{} },
	PidChatToHost:        func(_ *Encoding) Packet { return &Message{} },
	PidDropReq:           func(_ *Encoding) Packet { return &DropLaggers{} },
	PidSearchGame:        func(_ *Encoding) Packet { return &SearchGame{} },
	PidGameInfo:          func(_ *Encoding) Packet { return &GameInfo{} },
	PidCreateGame:        func(_ *Encoding) Packet { return &CreateGame{} },
	PidRefreshGame:       func(_ *Encoding) Packet { return &RefreshGame{} },
	PidDecreateGame:      func(_ *Encoding) Packet { return &DecreateGame{} },
	PidChatFromOthers:    func(_ *Encoding) Packet { return &PeerMessage{} },
	PidPingFromOthers:    func(_ *Encoding) Packet { return &PeerPing{} },
	PidPongToOthers:      func(_ *Encoding) Packet { return &PeerPong{} },
	PidClientInfo:        func(_ *Encoding) Packet { return &PeerConnect{} },
	PidPeerSet:           func(_ *Encoding) Packet { return &PeerSet{} },
	PidMapCheck:          func(_ *Encoding) Packet { return &MapCheck{} },
	PidStartDownload:     func(_ *Encoding) Packet { return &StartDownload{} },
	PidMapSize:           func(_ *Encoding) Packet { return &MapState{} },
	PidMapPart:           func(_ *Encoding) Packet { return &MapPart{} },
	PidMapPartOK:         func(_ *Encoding) Packet { return &MapPartOK{} },
	PidMapPartError:      func(_ *Encoding) Packet { return &MapPartError{} },
	PidPongToHost:        func(_ *Encoding) Packet { return &Pong{} },
	PidIncomingAction2:   func(_ *Encoding) Packet { return &TimeSlot{} },
}
