package w3gs

import (
	"bytes"
	"net"
	"testing"
)

const iterations = 3

func reverse(numbers []byte) []byte {
	numbers = append([]byte(nil), numbers...)
	for i := 0; i < len(numbers)/2; i++ {
		j := len(numbers) - i - 1
		numbers[i], numbers[j] = numbers[j], numbers[i]
	}
	return numbers
}

func TestBlob(t *testing.T) {
	var blob = []byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
	var buf = packetBuffer{bytes: make([]byte, 0)}

	for i := 1; i <= iterations; i++ {
		buf.writeBlob(blob)
		if buf.size() != i*len(blob) {
			t.Fatalf("Write(%v): %v != %v", i, buf.size(), i*len(blob))
		}
	}

	var rev = reverse(blob)
	buf.writeBlobAt(len(blob), rev)
	if buf.size() != iterations*len(blob) {
		t.Fatalf("WriteAt: %v != %v", buf.size(), iterations*len(blob))
	}

	for i := iterations - 1; i >= 0; i-- {
		var read = buf.readBlob(len(blob))

		if i == 1 {
			read = reverse(read)
		}
		if bytes.Compare(read, blob) != 0 {
			t.Fatalf("Read(%v): %v != %v", i, read, blob)
		}

		if buf.size() != i*len(blob) {
			t.Fatalf("Leftover(%v): %v != %v", i, buf.size(), i*len(blob))
		}
	}
}

func TestUInt8(t *testing.T) {
	var val = uint8(127)
	var buf = packetBuffer{bytes: make([]byte, 0)}

	for i := 1; i <= iterations; i++ {
		buf.writeUInt8(val)
		if buf.size() != i {
			t.Fatalf("Write(%v): %v != %v", i, buf.size(), i)
		}
	}

	var alt = ^val
	buf.writeUInt8At(1, alt)
	if buf.size() != iterations {
		t.Fatalf("WriteAt: %v != %v", buf.size(), iterations)
	}

	for i := iterations - 1; i >= 0; i-- {
		var read = buf.readUInt8()

		if i == 1 {
			read = ^read
		}
		if read != val {
			t.Fatalf("Read(%v): %v != %v", i, read, val)
		}

		if buf.size() != i {
			t.Fatalf("Leftover(%v): %v != %v", i, buf.size(), i)
		}
	}
}

func TestUInt16(t *testing.T) {
	var val = uint16(65534)
	var buf = packetBuffer{bytes: make([]byte, 0)}

	for i := 1; i <= iterations; i++ {
		buf.writeUInt16(val)
		if buf.size() != i*2 {
			t.Fatalf("Write(%v): %v != %v", i, buf.size(), i*2)
		}
	}

	var alt = ^val
	buf.writeUInt16At(2, alt)
	if buf.size() != iterations*2 {
		t.Fatalf("WriteAt: %v != %v", buf.size(), iterations*2)
	}

	for i := iterations - 1; i >= 0; i-- {
		var read = buf.readUInt16()

		if i == 1 {
			read = ^read
		}
		if read != val {
			t.Fatalf("Read(%v): %v != %v", i, read, val)
		}

		if buf.size() != i*2 {
			t.Fatalf("Leftover(%v): %v != %v", i, buf.size(), i*2)
		}
	}
}

func TestUInt32(t *testing.T) {
	var val = uint32(4294967294)
	var buf = packetBuffer{bytes: make([]byte, 0)}

	for i := 1; i <= iterations; i++ {
		buf.writeUInt32(val)
		if buf.size() != i*4 {
			t.Fatalf("Write(%v): %v != %v", i, buf.size(), i*4)
		}
	}

	var alt = ^val
	buf.writeUInt32At(4, alt)
	if buf.size() != iterations*4 {
		t.Fatalf("WriteAt: %v != %v", buf.size(), iterations*4)
	}

	for i := iterations - 1; i >= 0; i-- {
		var read = buf.readUInt32()

		if i == 1 {
			read = ^read
		}
		if read != val {
			t.Fatalf("Read(%v): %v != %v", i, read, val)
		}

		if buf.size() != i*4 {
			t.Fatalf("Leftover(%v): %v != %v", i, buf.size(), i*4)
		}
	}
}

func TestBool(t *testing.T) {
	var buf = packetBuffer{bytes: make([]byte, 0)}

	for i := 1; i <= iterations; i++ {
		buf.writeBool(i%2 != 0)
		if buf.size() != i {
			t.Fatalf("Write(%v): %v != %v", i, buf.size(), i)
		}
	}

	buf.writeBoolAt(1, true)
	if buf.size() != iterations {
		t.Fatalf("WriteAt: %v != %v", buf.size(), iterations)
	}

	for i := iterations - 1; i >= 0; i-- {
		var val = i%2 == 0
		var read = buf.readBool()

		if i == 1 {
			read = !read
		}
		if read != val {
			t.Fatalf("Read(%v): %v != %v", i, read, val)
		}

		if buf.size() != i {
			t.Fatalf("Leftover(%v): %v != %v", i, buf.size(), i)
		}
	}
}

func TestIP(t *testing.T) {
	var ip4 = net.IPv4(192, 168, 1, 101).To4()
	var buf = packetBuffer{bytes: make([]byte, 0)}

	for i := 1; i <= iterations; i++ {
		if err := buf.writeIP(ip4); err != nil {
			t.Fatal(err)
		}
		if buf.size() != i*4 {
			t.Fatalf("Write(%v): %v != %v", i, buf.size(), i*4)
		}
	}

	var rev = net.IP(reverse([]byte(ip4)))
	if err := buf.writeIPAt(4, net.IP([]byte{0, 0})); err != errInvalidIP4 {
		t.Fatal("errInvalidIP expected")
	}
	if err := buf.writeIPAt(4, rev); err != nil {
		t.Fatal(err)
	}
	if buf.size() != iterations*4 {
		t.Fatalf("WriteAt: %v != %v", buf.size(), iterations*4)
	}

	for i := iterations - 1; i >= 0; i-- {
		var read = buf.readIP()

		if i == 1 {
			read = net.IP(reverse([]byte(read)))
		}
		if !read.Equal(ip4) {
			t.Fatalf("Read(%v): %v != %v", i, read, ip4)
		}

		if buf.size() != i*4 {
			t.Fatalf("Leftover(%v): %v != %v", i, buf.size(), i*4)
		}
	}

	buf.writeIP(net.IPv4(127, 0, 0, 1))
	if buf.readUInt32() != 2130706433 {
		t.Fatal("Wrong integer format")
	}

	if buf.writeIP(net.IP([]byte{0, 0})) != errInvalidIP4 {
		t.Fatal("errInvalidIP expected")
	}
	if buf.readUInt32() != 0 {
		t.Fatal("Default value expected")
	}
}

func TestString(t *testing.T) {
	var str = "helloworld"
	var buf = packetBuffer{bytes: make([]byte, 0)}

	for i := 1; i <= iterations; i++ {
		buf.writeString(str)
		if buf.size() != i*(len(str)+1) {
			t.Fatalf("Write(%v): %v != %v", i, buf.size(), i*len(str))
		}
	}

	var rev = string(reverse([]byte(str)))
	buf.writeStringAt(len(str)+1, rev)
	if buf.size() != iterations*(len(str)+1) {
		t.Fatalf("WriteAt: %v != %v", buf.size(), iterations*len(str))
	}

	for i := iterations - 1; i >= 0; i-- {
		var read, err = buf.readString()
		if err != nil {
			t.Fatal(err)
		}

		if i == 1 {
			read = string(reverse([]byte(read)))
		}
		if read != str {
			t.Fatalf("Read(%v): %v != %v", i, read, str)
		}

		if buf.size() != i*(len(str)+1) {
			t.Fatalf("Leftover(%v): %v != %v", i, buf.size(), i*len(str))
		}
	}

	buf.writeUInt32(4294967294)
	if _, err := buf.readString(); err != errNoStringTerminatorFound {
		t.Fatal("errNoStringTerminatorFound expected")
	}
	if buf.size() > 0 {
		t.Fatal("Leftover after invalid string")
	}
}

func BenchmarkWriteUInt32(b *testing.B) {
	var buf = packetBuffer{bytes: make([]byte, 0)}
	for n := 0; n < b.N; n++ {
		buf.writeUInt32(4294967294)
	}
}

func BenchmarkReadUInt32(b *testing.B) {
	var buf = packetBuffer{bytes: make([]byte, 0)}
	for n := 0; n < b.N; n++ {
		buf.writeUInt32(4294967294)
	}
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		buf.readUInt32()
	}
}
