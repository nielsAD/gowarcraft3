// Author:  Niels A.D.
// Project: gowarcraft3 (https://github.com/nielsAD/gowarcraft3)
// License: Mozilla Public License, v2.0

package protocol_test

import (
	"testing"

	"github.com/nielsAD/gowarcraft3/protocol"
)

func TestBitSet8(t *testing.T) {
	var b = protocol.BS8(false, false, true)
	if b != 4 || !b.Test(3) {
		t.Fatal("b != 4")
	}

	b.Set(1)
	if b != 5 || !b.Test(1) {
		t.Fatal("b != 5")
	}

	b.Set(2)
	if b != 7 || !b.Test(2) {
		t.Fatal("b != 7")
	}

	b.Clear(1)
	if b != 6 || b.Test(1) {
		t.Fatal("b != 6")
	}

	if b.String() != "0b110" {
		t.Fatal("b != 0b110")
	}

	b.Set(8)
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("Panic expected when out of index")
		}
	}()
	b.Set(9)
}

func TestBitSet16(t *testing.T) {
	var b = protocol.BS16(false, false, true)
	if b != 4 || !b.Test(3) {
		t.Fatal("b != 4")
	}

	b.Set(1)
	if b != 5 || !b.Test(1) {
		t.Fatal("b != 5")
	}

	b.Set(2)
	if b != 7 || !b.Test(2) {
		t.Fatal("b != 7")
	}

	b.Clear(1)
	if b != 6 || b.Test(1) {
		t.Fatal("b != 6")
	}

	if b.String() != "0b110" {
		t.Fatal("b != 0b110")
	}

	b.Set(16)
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("Panic expected when out of index")
		}
	}()
	b.Set(17)
}

func TestBitSet32(t *testing.T) {
	var b = protocol.BS32(false, false, true)
	if b != 4 || !b.Test(3) {
		t.Fatal("b != 4")
	}

	b.Set(1)
	if b != 5 || !b.Test(1) {
		t.Fatal("b != 5")
	}

	b.Set(2)
	if b != 7 || !b.Test(2) {
		t.Fatal("b != 7")
	}

	b.Clear(1)
	if b != 6 || b.Test(1) {
		t.Fatal("b != 6")
	}

	if b.String() != "0b110" {
		t.Fatal("b != 0b110")
	}

	b.Set(32)
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("Panic expected when out of index")
		}
	}()
	b.Set(33)
}
