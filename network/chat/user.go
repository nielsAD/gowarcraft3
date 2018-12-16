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
	capi.UserUpdateEvent
	Joined   time.Time
	LastSeen time.Time
}

// Operator in channel
func (u *User) Operator() bool {
	return u.Flags&(capi.UserFlagAdmin|capi.UserFlagModerator) != 0
}
