// Author:  Niels A.D.
// Project: gowarcraft3 (https://github.com/nielsAD/gowarcraft3)
// License: Mozilla Public License, v2.0

package bnet

import (
	"strconv"
	"strings"
	"time"

	"github.com/nielsAD/gowarcraft3/protocol"
	"github.com/nielsAD/gowarcraft3/protocol/bncs"
)

// User in chat
type User struct {
	Name       string
	StatString string
	Flags      bncs.ChatUserFlags
	Ping       uint32
	Joined     time.Time
	LastSeen   time.Time
}

// Operator in channel
func (u *User) Operator() bool {
	return u.Flags&(bncs.ChatUserFlagBlizzard|bncs.ChatUserFlagOperator|bncs.ChatUserFlagAdmin) != 0
}

func reverse(s string) string {
	r := []rune(s)
	for i, j := 0, len(r)-1; i < len(r)/2; i, j = i+1, j-1 {
		r[i], r[j] = r[j], r[i]
	}
	return string(r)
}

// Stat split into (icon, level, clan)
func (u User) Stat() (product protocol.DWordString, icon protocol.DWordString, lvl int, tag protocol.DWordString) {
	lvl = -1

	var s = strings.Split(u.StatString, " ")
	if len(s) < 1 || len(s[0]) > 4 {
		return
	}

	product = protocol.DString(reverse(s[0]))

	if len(s) >= 2 {
		icon = protocol.DString(reverse(s[1]))
	}

	if len(s) >= 3 {
		if level, err := strconv.Atoi(s[2]); err == nil {
			lvl = level
		}
	}
	if len(s) >= 4 {
		tag = protocol.DString(reverse(s[3]))
	}

	return
}
