// Author:  Niels A.D.
// Project: gowarcraft3 (https://github.com/nielsAD/gowarcraft3)
// License: Mozilla Public License, v2.0

// Package bnet implements a mocked BNCS client that can be used to interact with BNCS servers.
package bnet

import (
	"github.com/nielsAD/gowarcraft3/protocol/bncs"
)

// JoinError event
type JoinError struct {
	Channel string
	Error   bncs.ChatEventType
}

// Channel joined event
type Channel struct {
	Name  string
	Flags bncs.ChatChannelFlags
}

// UserJoined event
type UserJoined struct {
	User
	AlreadyInChannel bool
}

// UserLeft event
type UserLeft struct {
	User
}

// UserUpdate event
type UserUpdate struct {
	User
}

// Say event
type Say struct {
	Content string
}

// Chat event
type Chat struct {
	User
	Content string
	Type    bncs.ChatEventType
}

// Whisper event
type Whisper struct {
	Username string
	Content  string
	Flags    bncs.ChatUserFlags
	Ping     uint32
}

// SystemMessage event
type SystemMessage struct {
	Content string
	Type    bncs.ChatEventType
}
