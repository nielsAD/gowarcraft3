// Author:  Niels A.D.
// Project: gowarcraft3 (https://github.com/nielsAD/gowarcraft3)
// License: Mozilla Public License, v2.0

package protocol

// DWordString is a string of size dword (4 bytes / characters)
type DWordString [4]byte

// DString converts str to DWordString
// panic if input invalid
func DString(str string) DWordString {
	if len(str) != 4 {
		panic("Invalid input for DString")
	}

	return DWordString{str[0], str[1], str[2], str[3]}
}

func (s *DWordString) String() string {
	return string(s[:])
}
