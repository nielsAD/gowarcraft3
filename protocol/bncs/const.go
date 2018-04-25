// Author:  Niels A.D.
// Project: gowarcraft3 (https://github.com/nielsAD/gowarcraft3)
// License: Mozilla Public License, v2.0

package bncs

import "errors"

// Errors
var (
	ErrNoProtocolSig     = errors.New("bncs: Invalid bncs packet (no signature found)")
	ErrInvalidPacketSize = errors.New("bncs: Invalid packet size")
	ErrInvalidChecksum   = errors.New("bncs: Checksum invalid")
	ErrUnexpectedConst   = errors.New("bncs: Unexpected constant value")
	ErrBufferTooSmall    = errors.New("bncs: Packet exceeds buffer size")
)

// ProtocolSig is the BNCS magic number used in the packet header.
const ProtocolSig = 0xFF

// BNCS packet type identifiers
const (
	PidNull                  = 0x00 // C -> S | S -> C
	PidStopAdv               = 0x02 // C -> S |
	PidGetAdvListEx          = 0x09 // C -> S | S -> C
	PidEnterChat             = 0x0A // C -> S | S -> C
	PidJoinChannel           = 0x0C // C -> S |
	PidChatCommand           = 0x0E // C -> S |
	PidChatEvent             = 0x0F //        | S -> C
	PidFloodDetected         = 0x13 //        | S -> C
	PidMessageBox            = 0x19 //        | S -> C
	PidStartAdvex3           = 0x1C // C -> S | S -> C
	PidNotifyJoin            = 0x22 // C -> S |
	PidPing                  = 0x25 // C -> S | S -> C
	PidNetGamePort           = 0x45 // C -> S |
	PidAuthInfo              = 0x50 // C -> S | S -> C
	PidAuthCheck             = 0x51 // C -> S | S -> C
	PidAuthAccountLogon      = 0x53 // C -> S | S -> C
	PidAuthAccountLogonProof = 0x54 // C -> S | S -> C
)
