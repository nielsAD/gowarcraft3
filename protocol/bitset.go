// Author:  Niels A.D.
// Project: gowarcraft3 (https://github.com/nielsAD/gowarcraft3)
// License: Mozilla Public License, v2.0

package protocol

import (
	"fmt"
	"strconv"
)

// BitSet8 is a set of 8 bits
type BitSet8 uint8

// BS8 is a BitSet8 contructor
func BS8(b ...bool) BitSet8 {
	var res BitSet8
	for i := uint(0); i < uint(len(b)); i++ {
		if b[i] {
			res.Set(i + 1)
		}
	}
	return res
}

// Test if bit i is in set
func (b BitSet8) Test(i uint) bool {
	if i < 1 || i > 8 {
		panic("bitset: Index out of range")
	}
	return b&(1<<(i-1)) != 0
}

// Set bit i
func (b *BitSet8) Set(i uint) *BitSet8 {
	if i < 1 || i > 8 {
		panic("bitset: Index out of range")
	}
	*b |= (1 << (i - 1))
	return b
}

// Clear bit i
func (b *BitSet8) Clear(i uint) *BitSet8 {
	if i < 1 || i > 8 {
		panic("bitset: Index out of range")
	}
	*b &= ^(1 << (i - 1))
	return b
}

func (b BitSet8) String() string {
	return fmt.Sprintf("0b%v", strconv.FormatUint(uint64(b), 2))
}

// BitSet16 is a set of 16 bits
type BitSet16 uint16

// BS16 is a BitSet16 contructor
func BS16(b ...bool) BitSet16 {
	var res BitSet16
	for i := uint(0); i < uint(len(b)); i++ {
		if b[i] {
			res.Set(i + 1)
		}
	}
	return res
}

// Test if bit i is in set
func (b BitSet16) Test(i uint) bool {
	if i < 1 || i > 16 {
		panic("bitset: Index out of range")
	}
	return b&(1<<(i-1)) != 0
}

// Set bit i
func (b *BitSet16) Set(i uint) *BitSet16 {
	if i < 1 || i > 16 {
		panic("bitset: Index out of range")
	}
	*b |= (1 << (i - 1))
	return b
}

// Clear bit i
func (b *BitSet16) Clear(i uint) *BitSet16 {
	if i < 1 || i > 16 {
		panic("bitset: Index out of range")
	}
	*b &= ^(1 << (i - 1))
	return b
}

func (b BitSet16) String() string {
	return fmt.Sprintf("0b%v", strconv.FormatUint(uint64(b), 2))
}

// BitSet32 is a set of 32 bits
type BitSet32 uint32

// BS32 is a BitSet32 contructor
func BS32(b ...bool) BitSet32 {
	var res BitSet32
	for i := uint(0); i < uint(len(b)); i++ {
		if b[i] {
			res.Set(i + 1)
		}
	}
	return res
}

// Test if bit i is in set
func (b BitSet32) Test(i uint) bool {
	if i < 1 || i > 32 {
		panic("bitset: Index out of range")
	}
	return b&(1<<(i-1)) != 0
}

// Set bit i
func (b *BitSet32) Set(i uint) *BitSet32 {
	if i < 1 || i > 32 {
		panic("bitset: Index out of range")
	}
	*b |= (1 << (i - 1))
	return b
}

// Clear bit i
func (b *BitSet32) Clear(i uint) *BitSet32 {
	if i < 1 || i > 32 {
		panic("bitset: Index out of range")
	}
	*b &= ^(1 << (i - 1))
	return b
}

func (b BitSet32) String() string {
	return fmt.Sprintf("0b%v", strconv.FormatUint(uint64(b), 2))
}
