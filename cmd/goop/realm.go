// Author:  Niels A.D.
// Project: gowarcraft3 (https://github.com/nielsAD/gowarcraft3)
// License: Mozilla Public License, v2.0

package main

import (
	"context"

	"github.com/nielsAD/gowarcraft3/network"
)

// RealmDelimiter between main/sub realm name (i.e. discord.{CHANNELID})
const RealmDelimiter = "."

// Rank for user
type Rank int32

// Rank constants
const (
	RankOwner     Rank = 1000000
	RankWhitelist Rank = 1
	RankDefault   Rank = 0
	RankIgnore    Rank = -1
	RankBan       Rank = -1000000
)

// Realm interface
type Realm interface {
	network.Emitter
	Run(ctx context.Context) error
	Relay(ev *network.Event, sender string)
}

// Connected event
type Connected struct{}

// Disconnected event
type Disconnected struct{}

// SystemMessage event
type SystemMessage struct {
	Content string
}

// User event
type User struct {
	ID        string
	Name      string
	Rank      Rank
	AvatarURL string
}

// Channel event
type Channel struct {
	ID   string
	Name string
}

// Chat event
type Chat struct {
	User
	Channel
	Content string
	Self    bool
}

// PrivateChat event
type PrivateChat struct {
	User
	Content string
}

// Join event
type Join struct {
	User
	Channel
}

// Leave event
type Leave struct {
	User
	Channel
}
