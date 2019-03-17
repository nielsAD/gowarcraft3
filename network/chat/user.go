// Author:  Niels A.D.
// Project: gowarcraft3 (https://github.com/nielsAD/gowarcraft3)
// License: Mozilla Public License, v2.0

package chat

import (
	"time"

	"github.com/nielsAD/gowarcraft3/protocol/capi"
)

// User in chat
type User struct {
	UserID   int64
	Username string
	Flags    UserFlags

	ProgramID string
	Rate      string
	Rank      string
	Wins      string

	Joined   time.Time
	LastSeen time.Time
}

// Update u with UserUpdateEvent information
func (u *User) Update(ev *capi.UserUpdateEvent) {
	u.UserID = ev.UserID
	if ev.Username != "" {
		u.Username = ev.Username
	}
	if ev.Flags != nil {
		u.Flags = UnmarshalUserFlags(ev.Flags)
	}

	for _, v := range ev.Attributes {
		switch v.Key {
		case capi.UserAttrProgramID:
			u.ProgramID = v.Value
		case capi.UserAttrRate:
			u.Rate = v.Value
		case capi.UserAttrRank:
			u.Rank = v.Value
		case capi.UserAttrWins:
			u.Wins = v.Value
		}
	}
}

// Operator in channel
func (u *User) Operator() bool {
	return u.Flags&(UserFlagAdmin|UserFlagModerator) != 0
}
