// Author:  Niels A.D.
// Project: gowarcraft3 (https://github.com/nielsAD/gowarcraft3)
// License: Mozilla Public License, v2.0

package dummy

import (
	"errors"

	"github.com/nielsAD/gowarcraft3/protocol/w3gs"
)

// Errors
var (
	ErrJoinRejected       = errors.New("dummy: Join rejected")
	ErrGameFull           = errors.New("dummy: Join rejected (game full)")
	ErrGameStarted        = errors.New("dummy: Join rejected (game started)")
	ErrInvalidFirstPacket = errors.New("dummy: Invalid first packet")
)

// RejectReasonToError converts w3gs.RejectReason to an appropriate error
func RejectReasonToError(r w3gs.RejectReason) error {
	switch r {
	case w3gs.RejectJoinFull:
		return ErrGameFull
	case w3gs.RejectJoinStarted:
		return ErrGameStarted
	default:
		return ErrJoinRejected
	}
}
