// Author:  Niels A.D.
// Project: gowarcraft3 (https://github.com/nielsAD/gowarcraft3)
// License: Mozilla Public License, v2.0

package protocol

import (
	"strings"
)

// DWordString is a string of size dword (4 bytes / characters)
type DWordString uint32

// DString converts str to DWordString
// panic if input invalid
func DString(str string) DWordString {
	if len(str) != 4 {
		panic("Invalid input for DString")
	}

	return DWordString(uint32(str[0]) | uint32(str[1])<<8 | uint32(str[2])<<16 | uint32(str[3])<<24)
}

func (s DWordString) String() string {
	if s == 0 {
		return ""
	}

	var b strings.Builder
	b.WriteByte(byte(s))
	if s&0xFFFFFF00 != 0 {
		b.WriteByte(byte(uint32(s) >> 8))
	}
	if s&0xFFFF0000 != 0 {
		b.WriteByte(byte(uint32(s) >> 16))
	}
	if s&0xFF000000 != 0 {
		b.WriteByte(byte(uint32(s) >> 24))
	}
	return b.String()
}
