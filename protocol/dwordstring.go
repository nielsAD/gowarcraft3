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
		panic("dwstr: Length of input string for DString() not equal to 4")
	}

	return DWordString(uint32(str[0]) | uint32(str[1])<<8 | uint32(str[2])<<16 | uint32(str[3])<<24)
}

func (s DWordString) String() string {
	if s == 0 {
		return ""
	}

	return string([]byte{byte(uint32(s)), byte(uint32(s) >> 8), byte(uint32(s) >> 16), byte(uint32(s) >> 24)})
}
