package util

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
