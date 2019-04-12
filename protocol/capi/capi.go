// Author:  Niels A.D.
// Project: gowarcraft3 (https://github.com/nielsAD/gowarcraft3)
// License: Mozilla Public License, v2.0

// Package capi implements the datastructures for the official classic Battle.net chat API.
//
// The Chat API uses JSON with UTF8 encoding as its protocol with secure websockets as the transport.
package capi

// Endpoint for websocket connection
//
// It is recommended that the certificate is checked to ensure that the common name matches *.classic.blizzard.com
const Endpoint = "wss://connect-bot.classic.blizzard.com/v1/rpc/chat"

// Packet structure
type Packet struct {
	Command   string      `json:"command"`
	RequestID int64       `json:"request_id"`
	Status    *Status     `json:"status,omitempty"`
	Payload   interface{} `json:"payload"`
}

// DefaultFactory maps command to matching payload type
var DefaultFactory = MapFactory{
	CmdAuthenticate + CmdRequestSuffix:    func() interface{} { return &Authenticate{} },
	CmdConnect + CmdRequestSuffix:         func() interface{} { return &Connect{} },
	CmdDisconnect + CmdRequestSuffix:      func() interface{} { return &Disconnect{} },
	CmdSendMessage + CmdRequestSuffix:     func() interface{} { return &SendMessage{} },
	CmdSendEmote + CmdRequestSuffix:       func() interface{} { return &SendEmote{} },
	CmdSendWhisper + CmdRequestSuffix:     func() interface{} { return &SendWhisper{} },
	CmdKickUser + CmdRequestSuffix:        func() interface{} { return &KickUser{} },
	CmdBanUser + CmdRequestSuffix:         func() interface{} { return &BanUser{} },
	CmdUnbanUser + CmdRequestSuffix:       func() interface{} { return &UnbanUser{} },
	CmdSetModerator + CmdRequestSuffix:    func() interface{} { return &SetModerator{} },
	CmdConnectEvent + CmdRequestSuffix:    func() interface{} { return &ConnectEvent{} },
	CmdDisconnectEvent + CmdRequestSuffix: func() interface{} { return &DisconnectEvent{} },
	CmdMessageEvent + CmdRequestSuffix:    func() interface{} { return &MessageEvent{} },
	CmdUserUpdateEvent + CmdRequestSuffix: func() interface{} { return &UserUpdateEvent{} },
	CmdUserLeaveEvent + CmdRequestSuffix:  func() interface{} { return &UserLeaveEvent{} },
}
