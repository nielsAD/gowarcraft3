package fakeplayer

import (
	"errors"

	"github.com/nielsAD/gowarcraft3/protocol/w3gs"
)

// Errors
var (
	ErrJoinRejected = errors.New("fp: Join rejected")
	ErrGameFull     = errors.New("fp: Join rejected (game full)")
	ErrGameStarted  = errors.New("fp: Join rejected (game started)")

	ErrInvalidFirstPacket = errors.New("fp: Invalid first packet")
	ErrUnknownPeerID      = errors.New("fp: Unknown peer ID")
	ErrAlreadyConnected   = errors.New("fp: Already connected")
	ErrInvalidEntryKey    = errors.New("fp: Wrong entry key")
	ErrInvalidJoinCounter = errors.New("fp: Wrong join counter")
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
