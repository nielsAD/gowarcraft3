// Author:  Niels A.D.
// Project: gowarcraft3 (https://github.com/nielsAD/gowarcraft3)
// License: Mozilla Public License, v2.0

package protocol

// DWordString is a string of size dword (4 bytes / characters)
type DWordString uint32

// DString converts str to DWordString
// panic if input invalid
func DString(str string) DWordString {
	switch len(str) {
	case 0:
		return DWordString(0)
	case 1:
		return DWordString(uint32(str[0]))
	case 2:
		return DWordString(uint32(str[0]) | uint32(str[1])<<8)
	case 3:
		return DWordString(uint32(str[0]) | uint32(str[1])<<8 | uint32(str[2])<<16)
	case 4:
		return DWordString(uint32(str[0]) | uint32(str[1])<<8 | uint32(str[2])<<16 | uint32(str[3])<<24)
	default:
		panic("dwstr: Length of input for DString() exceeds 4")
	}
}

func (s DWordString) String() string {
	if s&0xFF000000 != 0 {
		return string([]byte{byte(uint32(s)), byte(uint32(s) >> 8), byte(uint32(s) >> 16), byte(uint32(s) >> 24)})
	}
	if s&0xFF0000 != 0 {
		return string([]byte{byte(uint32(s)), byte(uint32(s) >> 8), byte(uint32(s) >> 16)})
	}
	if s&0xFF00 != 0 {
		return string([]byte{byte(uint32(s)), byte(uint32(s) >> 8)})
	}
	return string([]byte{byte(uint32(s))})
}
