// Author:  Niels A.D.
// Project: gowarcraft3 (https://github.com/nielsAD/gowarcraft3)
// License: Mozilla Public License, v2.0

// Package bncs implements the old Battle.net chat protocol for Warcraft III.
// The protocol is used for multiple classic games, but this package only
// implements the small set of packets required for Warcraft III to log
// on and enter chat.
//
// Based on protocol documentation by https://bnetdocs.org/
//
// Each packet type is mapped to a struct type that implements the Packet
// interface. To deserialize from a binary stream, use bncs.Read().
//
// This package tries to keep ammortized heap memory allocations to 0.
//
// General serialization format:
//
//    (UINT8)  Protocol signature (0xFF)
//    (UINT8)  Packet type ID
//    (UINT16) Packet size
//    [Packet Data]
//
package bncs

import (
	"github.com/nielsAD/gowarcraft3/protocol"
	"github.com/nielsAD/gowarcraft3/protocol/w3gs"
)

// Packet interface.
type Packet interface {
	Serialize(buf *protocol.Buffer, enc *Encoding) error
	Deserialize(buf *protocol.Buffer, enc *Encoding) error
}

// Encoding options for (de)serialization
type Encoding struct {
	w3gs.Encoding

	// Assume request when deserializing ambiguous packet IDs
	Request bool
}

// DefaultFactory maps packet IDs to matching type
var DefaultFactory = MapFactory{
	PidNull:          func(_ *Encoding) Packet { return &KeepAlive{} },
	PidStopAdv:       func(_ *Encoding) Packet { return &StopAdv{} },
	PidJoinChannel:   func(_ *Encoding) Packet { return &JoinChannel{} },
	PidChatCommand:   func(_ *Encoding) Packet { return &ChatCommand{} },
	PidChatEvent:     func(_ *Encoding) Packet { return &ChatEvent{} },
	PidFloodDetected: func(_ *Encoding) Packet { return &FloodDetected{} },
	PidMessageBox:    func(_ *Encoding) Packet { return &MessageBox{} },
	PidNotifyJoin:    func(_ *Encoding) Packet { return &NotifyJoin{} },
	PidPing:          func(_ *Encoding) Packet { return &Ping{} },
	PidNetGamePort:   func(_ *Encoding) Packet { return &NetGamePort{} },
	PidSetEmail:      func(_ *Encoding) Packet { return &SetEmail{} },
	PidClanInfo:      func(_ *Encoding) Packet { return &ClanInfo{} },

	PidGetAdvListEx: ReqResp(
		func(_ *Encoding) Packet { return &GetAdvListReq{} },
		func(_ *Encoding) Packet { return &GetAdvListResp{} },
	),
	PidEnterChat: ReqResp(
		func(_ *Encoding) Packet { return &EnterChatReq{} },
		func(_ *Encoding) Packet { return &EnterChatResp{} },
	),
	PidStartAdvex3: ReqResp(
		func(_ *Encoding) Packet { return &StartAdvex3Req{} },
		func(_ *Encoding) Packet { return &StartAdvex3Resp{} },
	),
	PidAuthInfo: ReqResp(
		func(_ *Encoding) Packet { return &AuthInfoReq{} },
		func(_ *Encoding) Packet { return &AuthInfoResp{} },
	),
	PidAuthCheck: ReqResp(
		func(_ *Encoding) Packet { return &AuthCheckReq{} },
		func(_ *Encoding) Packet { return &AuthCheckResp{} },
	),
	PidAuthAccountCreate: ReqResp(
		func(_ *Encoding) Packet { return &AuthAccountCreateReq{} },
		func(_ *Encoding) Packet { return &AuthAccountCreateResp{} },
	),
	PidAuthAccountLogon: ReqResp(
		func(_ *Encoding) Packet { return &AuthAccountLogonReq{} },
		func(_ *Encoding) Packet { return &AuthAccountLogonResp{} },
	),
	PidAuthAccountLogonProof: ReqResp(
		func(_ *Encoding) Packet { return &AuthAccountLogonProofReq{} },
		func(_ *Encoding) Packet { return &AuthAccountLogonProofResp{} },
	),
	PidAuthAccountChange: ReqResp(
		func(_ *Encoding) Packet { return &AuthAccountChangePassReq{} },
		func(_ *Encoding) Packet { return &AuthAccountChangePassResp{} },
	),
	PidAuthAccountChangeProof: ReqResp(
		func(_ *Encoding) Packet { return &AuthAccountChangePassProofReq{} },
		func(_ *Encoding) Packet { return &AuthAccountChangePassProofResp{} },
	),
}
