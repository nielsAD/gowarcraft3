// Author:  Niels A.D.
// Project: gowarcraft3 (https://github.com/nielsAD/gowarcraft3)
// License: Mozilla Public License, v2.0

package w3g

import (
	"errors"
)

// Errors
var (
	ErrBadFormat       = errors.New("w3g: Invalid file format")
	ErrInvalidChecksum = errors.New("w3g: Checksum invalid")
	ErrUnexpectedConst = errors.New("w3g: Unexpected constant value")
	ErrUnknownRecord   = errors.New("w3g: Unknown record ID")
)

// Signature constant for w3g files
var Signature = "Warcraft III recorded game\x1A"

// Record type identifiers
const (
	RidGameInfo       = 0x10
	RidPlayerInfo     = 0x16 // w3gs.PlayerInfo
	RidPlayerLeft     = 0x17 // w3gs.Leave
	RidSlotInfo       = 0x19 // w3gs.SlotInfo
	RidCountDownStart = 0x1A // w3gs.CountDownStart
	RidCountDownEnd   = 0x1B // w3gs.CountDownEnd
	RidGameStart      = 0x1C
	RidTimeSlot2      = 0x1E // w3gs.TimeSlot2 (fragment)
	RidTimeSlot       = 0x1F // w3gs.TimeSlot2
	RidChatMessage    = 0x20 // w3gs.MessageRelay
	RidTimeSlotAck    = 0x22 // w3gs.TimeSlotAck
	RidDesync         = 0x23
	RidEndTimer       = 0x2F
)
