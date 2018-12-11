// Author:  Niels A.D.
// Project: gowarcraft3 (https://github.com/nielsAD/gowarcraft3)
// License: Mozilla Public License, v2.0

package capi

import (
	"encoding/json"
	"fmt"
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

var messageEventTypeStr = map[string]MessageEventType{
	"Whisper":     MessageWhisper,
	"Channel":     MessageChannel,
	"ServerInfo":  MessageServerInfo,
	"ServerError": MessageServerError,
	"Emote":       MessageEmote,
}

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

// UserFlags enum
type UserFlags uint32

// User flags
const (
	UserFlagAdmin UserFlags = 1 << iota
	UserFlagModerator
	UserFlagSpeaker
	UserFlagMuteGlobal
	UserFlagMuteWhisper
)

func (f UserFlags) String() string {
	var res string
	if f&UserFlagAdmin != 0 {
		res += "|Blizzard"
		f &= ^UserFlagAdmin
	}
	if f&UserFlagModerator != 0 {
		res += "|Operator"
		f &= ^UserFlagModerator
	}
	if f&UserFlagSpeaker != 0 {
		res += "|Speaker"
		f &= ^UserFlagSpeaker
	}
	if f&UserFlagMuteGlobal != 0 {
		res += "|Muted"
		f &= ^UserFlagMuteGlobal
	}
	if f&UserFlagMuteWhisper != 0 {
		res += "|Squelched"
		f &= ^UserFlagMuteWhisper
	}
	if f != 0 {
		res += fmt.Sprintf("|UserFlags(0x%02X)", uint32(f))
	}
	if res != "" {
		res = res[1:]
	}
	return res
}

// MarshalJSON implements JSONMarshaler
func (f UserFlags) MarshalJSON() ([]byte, error) {
	var arr []string
	if f&UserFlagAdmin != 0 {
		arr = append(arr, "Admin")
	}
	if f&UserFlagModerator != 0 {
		arr = append(arr, "Moderator")
	}
	if f&UserFlagSpeaker != 0 {
		arr = append(arr, "Speaker")
	}
	if f&UserFlagMuteGlobal != 0 {
		arr = append(arr, "MuteGlobal")
	}
	if f&UserFlagMuteWhisper != 0 {
		arr = append(arr, "MuteWhisper")
	}
	return json.Marshal(arr)
}

// UnmarshalJSON implements JSONUnmarshaler
func (f *UserFlags) UnmarshalJSON(b []byte) error {
	var arr []string
	if err := json.Unmarshal(b, &arr); err != nil {
		return err
	}

	*f = 0
	for _, v := range arr {
		switch v {
		case "Admin":
			*f |= UserFlagAdmin
		case "Moderator":
			*f |= UserFlagModerator
		case "Speaker":
			*f |= UserFlagSpeaker
		case "MuteGlobal":
			*f |= UserFlagMuteGlobal
		case "MuteWhisper":
			*f |= UserFlagMuteWhisper
		}
	}

	return nil
}
