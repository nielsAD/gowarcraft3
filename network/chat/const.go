// Author:  Niels A.D.
// Project: gowarcraft3 (https://github.com/nielsAD/gowarcraft3)
// License: Mozilla Public License, v2.0

package chat

import (
	"errors"
)

// Errors
var (
	ErrUnexpectedPacket = errors.New("chat: Received unexpected packet")
)
