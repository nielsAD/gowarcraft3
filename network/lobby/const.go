// Author:  Niels A.D.
// Project: gowarcraft3 (https://github.com/nielsAD/gowarcraft3)
// License: Mozilla Public License, v2.0

package lobby

import (
	"errors"
	"fmt"
	"time"
)

// Errors
var (
	ErrFull            = errors.New("lobby: Lobby is full")
	ErrLocked          = errors.New("lobby: Lobby is locked")
	ErrInvalidArgument = errors.New("lobby: Invalid argument")
	ErrInvalidSlot     = errors.New("lobby: Invalid slot")
	ErrInvalidPacket   = errors.New("lobby: Invalid packet")
	ErrMapUnavailable  = errors.New("lobby: Map unavailable")
	ErrNotReady        = errors.New("lobby: Player was not ready")
	ErrPlayersOccupied = errors.New("lobby: No player slots left")
	ErrSlotOccupied    = errors.New("lobby: Slot occupied")
	ErrColorOccupied   = errors.New("lobby: Color occupied")
	ErrHighPing        = errors.New("lobby: Ping exceeds lag recovery delay")
	ErrStraggling      = errors.New("lobby: Player was straggling")
	ErrDesync          = errors.New("lobby: Timeslot checksum mismatch")
)

// ObsDisabled constant
const ObsDisabled uint8 = 0xFF

// Maximum transmission unit
const mtu = 1200

// LagDelay timeout before showing lag screen
const LagDelay = 2 * time.Second

// LagRecoverDelay timeout before ending lag screen
const LagRecoverDelay = 1 * time.Second

// Tick counter
type Tick uint32

// Stage enum
type Stage uint32

// Stage enums
const (
	StageLobby Stage = iota
	StageLoading
	StagePlaying
	StageDone
)

func (s Stage) String() string {
	switch s {
	case StageLobby:
		return "Lobby"
	case StageLoading:
		return "Loading"
	case StagePlaying:
		return "Playing"
	case StageDone:
		return "Done"
	default:
		return fmt.Sprintf("Stage(%d)", uint32(s))
	}
}
