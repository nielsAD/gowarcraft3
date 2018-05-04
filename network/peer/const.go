// Author:  Niels A.D.
// Project: gowarcraft3 (https://github.com/nielsAD/gowarcraft3)
// License: Mozilla Public License, v2.0

package peer

import (
	"errors"
)

// Errors
var (
	ErrInvalidFirstPacket = errors.New("peer: Invalid first packet")
	ErrUnknownPeerID      = errors.New("peer: Unknown ID")
	ErrDupPeerID          = errors.New("peer: Duplicate peer ID")
	ErrAlreadyConnected   = errors.New("peer: Already connected")
	ErrInvalidEntryKey    = errors.New("peer: Wrong entry key")
	ErrInvalidJoinCounter = errors.New("peer: Wrong join counter")
)
