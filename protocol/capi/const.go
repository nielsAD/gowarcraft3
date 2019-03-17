// Author:  Niels A.D.
// Project: gowarcraft3 (https://github.com/nielsAD/gowarcraft3)
// License: Mozilla Public License, v2.0

package capi

import (
	"fmt"
)

// CAPI command identifiers
const (
	CmdRequestSuffix  = "Request"
	CmdResponseSuffix = "Response"

	CmdAuthenticate    = "Botapiauth.Authenticate"
	CmdConnect         = "Botapichat.Connect"
	CmdDisconnect      = "Botapichat.Disconnect"
	CmdSendMessage     = "Botapichat.SendMessage"
	CmdSendEmote       = "Botapichat.SendEmote"
	CmdSendWhisper     = "Botapichat.SendWhisper"
	CmdKickUser        = "Botapichat.KickUser"
	CmdBanUser         = "Botapichat.BanUser"
	CmdUnbanUser       = "Botapichat.UnbanUser"
	CmdSetModerator    = "Botapichat.SendSetModerator"
	CmdConnectEvent    = "Botapichat.ConnectEvent"
	CmdDisconnectEvent = "Botapichat.DisconnectEvent"
	CmdMessageEvent    = "Botapichat.MessageEvent"
	CmdUserUpdateEvent = "Botapichat.UserUpdateEvent"
	CmdUserLeaveEvent  = "Botapichat.UserLeaveEvent"
)

// MessageEventType enum
type MessageEventType uint32

// Message event types
const (
	Unknown MessageEventType = iota
	MessageWhisper
	MessageChannel
	MessageServerInfo
	MessageServerError
	MessageEmote
)

func (m MessageEventType) String() string {
	switch m {
	case MessageWhisper:
		return "Whisper"
	case MessageChannel:
		return "Channel"
	case MessageServerInfo:
		return "ServerInfo"
	case MessageServerError:
		return "ServerError"
	case MessageEmote:
		return "Emote"
	default:
		return fmt.Sprintf("MessageEventType(0x%02X)", uint32(m))
	}
}

// MarshalText implements TextMarshaler
func (m MessageEventType) MarshalText() ([]byte, error) {
	return []byte(m.String()), nil
}

// UnmarshalText implements TextUnmarshaler
func (m *MessageEventType) UnmarshalText(txt []byte) error {
	switch string(txt) {
	case "Whisper":
		*m = MessageWhisper
	case "Channel":
		*m = MessageChannel
	case "ServerInfo":
		*m = MessageServerInfo
	case "ServerError":
		*m = MessageServerError
	case "Emote":
		*m = MessageEmote
	default:
		*m = Unknown
	}

	return nil
}

// User flags
const (
	UserFlagAdmin       = "Admin"
	UserFlagModerator   = "Moderator"
	UserFlagSpeaker     = "Speaker"
	UserFlagMuteGlobal  = "MuteGlobal"
	UserFlagMuteWhisper = "MuteWhisper"
)

// User attribute keys
const (
	UserAttrProgramID = "ProgramID"
	UserAttrRate      = "Rate"
	UserAttrRank      = "Rank"
	UserAttrWins      = "Wins"
)
