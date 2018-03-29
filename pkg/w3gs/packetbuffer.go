package w3gs

import (
	"bytes"
	"errors"
	"net"
)

// Errors
var (
	errInvalidIP4              = errors.New("w3gs: Invalid IP4 address")
	errNoStringTerminatorFound = errors.New("w3gs: No null terminator for string found in buffer")
)

type packetBuffer struct {
	bytes []byte
}

func (b *packetBuffer) size() int {
	return len(b.bytes)
}

func (b *packetBuffer) skip(len int) {
	b.bytes = b.bytes[len:]
}

func (b *packetBuffer) writeBlob(v []byte) {
	b.bytes = append(b.bytes, v...)
}

func (b *packetBuffer) writeUInt8(v byte) {
	b.bytes = append(b.bytes, v)
}

func (b *packetBuffer) writeUInt16(v uint16) {
	b.bytes = append(b.bytes, byte(v), byte(v>>8))
}

func (b *packetBuffer) writeUInt32(v uint32) {
	b.bytes = append(b.bytes, byte(v), byte(v>>8), byte(v>>16), byte(v>>24))
}

func (b *packetBuffer) writeBool(v bool) {
	var i uint8
	if v {
		i = 1
	}
	b.bytes = append(b.bytes, i)
}

func (b *packetBuffer) writePort(v uint16) {
	b.bytes = append(b.bytes, byte(v>>8), byte(v))
}

func (b *packetBuffer) writeIP(v net.IP) error {
	if ip4 := v.To4(); ip4 != nil {
		b.writeBlob(ip4)
		return nil
	}

	b.writeUInt32(0)
	return errInvalidIP4
}

func (b *packetBuffer) writeString(s string) {
	b.writeBlob([]byte(s))
	b.writeUInt8(0)
}

func (b *packetBuffer) writeBlobAt(p int, v []byte) {
	copy(b.bytes[p:], v)
}

func (b *packetBuffer) writeUInt8At(p int, v byte) {
	b.bytes[p] = v
}

func (b *packetBuffer) writeUInt16At(p int, v uint16) {
	b.bytes[p+1], b.bytes[p] = byte(v>>8), byte(v)
}

func (b *packetBuffer) writeUInt32At(p int, v uint32) {
	b.bytes[p+3], b.bytes[p+2], b.bytes[p+1], b.bytes[p] = byte(v>>24), byte(v>>16), byte(v>>8), byte(v)
}

func (b *packetBuffer) writeBoolAt(p int, v bool) {
	var i uint8
	if v {
		i = 1
	}
	b.bytes[p] = i
}

func (b *packetBuffer) writePortAt(p int, v uint16) {
	b.bytes[p+1], b.bytes[p] = byte(v), byte(v>>8)
}

func (b *packetBuffer) writeIPAt(p int, v net.IP) error {
	if ip4 := v.To4(); ip4 != nil {
		b.writeBlobAt(p, ip4)
		return nil
	}

	b.writeUInt32At(p, 0)
	return errInvalidIP4
}

func (b *packetBuffer) writeStringAt(p int, s string) {
	var bytes = []byte(s)
	b.writeBlobAt(p, bytes)
	b.writeUInt8At(p+len(bytes), 0)
}

func (b *packetBuffer) readBlob(len int) []byte {
	if len > 0 {
		var res = b.bytes[:len]
		b.bytes = b.bytes[len:]
		return res
	}

	return nil
}

func (b *packetBuffer) readUInt8() byte {
	var res = byte(b.bytes[0])
	b.bytes = b.bytes[1:]
	return res
}

func (b *packetBuffer) readUInt16() uint16 {
	var res = uint16(b.bytes[1])<<8 | uint16(b.bytes[0])
	b.bytes = b.bytes[2:]
	return res
}

func (b *packetBuffer) readUInt32() uint32 {
	var res = uint32(b.bytes[3])<<24 | uint32(b.bytes[2])<<16 | uint32(b.bytes[1])<<8 | uint32(b.bytes[0])
	b.bytes = b.bytes[4:]
	return res
}

func (b *packetBuffer) readBool() bool {
	var res bool
	if b.bytes[0] > 0 {
		res = true
	}
	b.bytes = b.bytes[1:]
	return res
}

func (b *packetBuffer) readPort() uint16 {
	var res = uint16(b.bytes[1]) | uint16(b.bytes[0])<<8
	b.bytes = b.bytes[2:]
	return res
}

func (b *packetBuffer) readIP() net.IP {
	var res = net.IP(b.readBlob(net.IPv4len))
	if res.Equal(net.IPv4zero) {
		return nil
	}
	return res
}

func (b *packetBuffer) readString() (string, error) {
	var pos = bytes.IndexByte(b.bytes, 0)
	if pos == -1 {
		b.bytes = b.bytes[len(b.bytes):]
		return "", errNoStringTerminatorFound
	}

	var res = string(b.bytes[:pos])
	b.bytes = b.bytes[pos+1:]
	return res, nil
}
