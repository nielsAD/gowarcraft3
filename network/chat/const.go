// Author:  Niels A.D.
// Project: gowarcraft3 (https://github.com/nielsAD/gowarcraft3)
// License: Mozilla Public License, v2.0

package chat

import (
	"errors"
	"fmt"

	"github.com/nielsAD/gowarcraft3/protocol/capi"
)

// Errors
var (
	ErrUnexpectedPacket = errors.New("chat: Received unexpected packet")
)

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

// Marshal to its JSON representation
func (f UserFlags) Marshal() []string {
	var arr []string
	if f&UserFlagAdmin != 0 {
		arr = append(arr, capi.UserFlagAdmin)
	}
	if f&UserFlagModerator != 0 {
		arr = append(arr, capi.UserFlagModerator)
	}
	if f&UserFlagSpeaker != 0 {
		arr = append(arr, capi.UserFlagSpeaker)
	}
	if f&UserFlagMuteGlobal != 0 {
		arr = append(arr, capi.UserFlagMuteGlobal)
	}
	if f&UserFlagMuteWhisper != 0 {
		arr = append(arr, capi.UserFlagMuteWhisper)
	}
	return arr
}

// UnmarshalUserFlags from its JSON representation
func UnmarshalUserFlags(arr []string) UserFlags {
	var f UserFlags
	for _, v := range arr {
		switch v {
		case capi.UserFlagAdmin:
			f |= UserFlagAdmin
		case capi.UserFlagModerator:
			f |= UserFlagModerator
		case capi.UserFlagSpeaker:
			f |= UserFlagSpeaker
		case capi.UserFlagMuteGlobal:
			f |= UserFlagMuteGlobal
		case capi.UserFlagMuteWhisper:
			f |= UserFlagMuteWhisper
		}
	}
	return f
}
