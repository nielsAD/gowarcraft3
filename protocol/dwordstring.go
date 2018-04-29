// Author:  Niels A.D.
// Project: gowarcraft3 (https://github.com/nielsAD/gowarcraft3)
// License: Mozilla Public License, v2.0

package protocol

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

	var r = string([]byte{byte(uint32(s)), byte(uint32(s) >> 8), byte(uint32(s) >> 16), byte(uint32(s) >> 24)})
	if s&0xFF000000 != 0 {
		return string(r)
	}
	if s&0x00FF0000 != 0 {
		return string(r[:3])
	}
	if s&0x0000FF00 != 0 {
		return string(r[:2])
	}
	return string(r[:1])
}
