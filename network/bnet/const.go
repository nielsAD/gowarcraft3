// Author:  Niels A.D.
// Project: gowarcraft3 (https://github.com/nielsAD/gowarcraft3)
// License: Mozilla Public License, v2.0

package bnet

import (
	"errors"

	"github.com/nielsAD/gowarcraft3/protocol/bncs"
)

// Errors
var (
	ErrCheckRevision      = errors.New("bnet: BNCSUtil call to checkRevision failed")
	ErrExeInfo            = errors.New("bnet: BNCSUtil call to getExeInfo failed")
	ErrKeyDecoder         = errors.New("bnet: BNCSUtil call to keyDecoder failed")
	ErrNLS                = errors.New("bnet: BNCSUtil call to nls initializer failed")
	ErrUnexpectedPacket   = errors.New("bnet: Received unexpected packet")
	ErrAuthFail           = errors.New("bnet: Authentication failed")
	ErrInvalidGameVersion = errors.New("bnet: Authentication failed (game version invalid)")
	ErrCDKeyInvalid       = errors.New("bnet: Authentication failed (CD key invalid)")
	ErrCDKeyInUse         = errors.New("bnet: Authentication failed (CD key in use)")
	ErrCDKeyBanned        = errors.New("bnet: Authentication failed (CD key banned)")
	ErrInvalidAccount     = errors.New("bnet: Authentication failed (account invalid)")
	ErrIncorrectPassword  = errors.New("bnet: Authentication failed (password incorrect)")
)

// AuthResultToError converts bncs.AuthResult to an appropriate error
func AuthResultToError(r bncs.AuthResult) error {
	switch r {
	case bncs.AuthUpgradeRequired, bncs.AuthInvalidVersion, bncs.AuthInvalidVersionMask, bncs.AuthDowngradeRequired, bncs.AuthWrongProduct:
		return ErrInvalidGameVersion
	case bncs.AuthCDKeyInvalid:
		return ErrCDKeyInvalid
	case bncs.AuthCDKeyInUse:
		return ErrCDKeyInUse
	case bncs.AuthCDKeyBanned:
		return ErrCDKeyBanned
	default:
		return ErrAuthFail
	}
}

// LogonProofResultToError converts bncs.LogonProofResult to an appropriate error
func LogonProofResultToError(r bncs.LogonProofResult) error {
	switch r {
	case bncs.LogonProofPasswordIncorrect:
		return ErrIncorrectPassword
	default:
		return ErrInvalidAccount
	}
}
