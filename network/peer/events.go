// Author:  Niels A.D.
// Project: gowarcraft3 (https://github.com/nielsAD/gowarcraft3)
// License: Mozilla Public License, v2.0

package peer

// Event root struct
type Event struct {
	Peer *Player
}

// Registered event
type Registered Event

// Deregistered event
type Deregistered Event

// Disconnected event
type Disconnected Event

// Connected event
type Connected struct {
	Event
	Dial bool
}

// Chat event
type Chat struct {
	Event
	Content string
}
