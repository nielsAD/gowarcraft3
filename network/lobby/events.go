// Author:  Niels A.D.
// Project: gowarcraft3 (https://github.com/nielsAD/gowarcraft3)
// License: Mozilla Public License, v2.0

package lobby

import "github.com/nielsAD/gowarcraft3/protocol/w3gs"

// Ready event
type Ready struct{}

// StartLag event
type StartLag struct{}

// StopLag event
type StopLag struct{}

// PlayerJoined event
type PlayerJoined struct {
	*Player
}

// PlayerLeft event
type PlayerLeft struct {
	*Player
}

// PlayerChat event
type PlayerChat struct {
	*Player
	*w3gs.Message
}

// StageChanged event
type StageChanged struct {
	Old Stage
	New Stage
}
