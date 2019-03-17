// Author:  Niels A.D.
// Project: gowarcraft3 (https://github.com/nielsAD/gowarcraft3)
// License: Mozilla Public License, v2.0

package chat

// UserJoined event
type UserJoined struct {
	User
}

// UserLeft event
type UserLeft struct {
	User
}

// UserUpdate event
type UserUpdate struct {
	User
}
