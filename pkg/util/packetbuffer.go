package util

import (
	"bytes"
	"errors"
	"io"
	"net"
)

// Errors
var (
	ErrInvalidIP4              = errors.New("pbuf: Invalid IP4 address")
	ErrInvalidSockAddr         = errors.New("pbuf: Invalid SockAddr structure")
	ErrNoStringTerminatorFound = errors.New("pbuf: No null terminator for string found in buffer")
)

// AF_INET
const connAddressFamily uint16 = 2

// PacketBuffer wraps a []byte slice and adds helper functions for binary (de)serialization
type PacketBuffer struct {
	Bytes []byte
}

// Size returns the total size of the buffer
func (b *PacketBuffer) Size() int {
	return len(b.Bytes)
}

// Skip consumes len bytes and throws away the result
func (b *PacketBuffer) Skip(len int) {
	b.Bytes = b.Bytes[len:]
}

// Truncate resets the buffer to size 0
func (b *PacketBuffer) Truncate() {
	b.Bytes = b.Bytes[:0]
}

// Write implements io.Writer interface
func (b *PacketBuffer) Write(p []byte) (int, error) {
	b.WriteBlob(p)
	return len(p), nil
}

// WriteBlob appends blob v to the buffer
func (b *PacketBuffer) WriteBlob(v []byte) {
	b.Bytes = append(b.Bytes, v...)
}

// WriteUInt8 appends uint8 v to the buffer
func (b *PacketBuffer) WriteUInt8(v byte) {
	b.Bytes = append(b.Bytes, v)
}

// WriteUInt16 appends uint16 v to the buffer
func (b *PacketBuffer) WriteUInt16(v uint16) {
	b.Bytes = append(b.Bytes, byte(v), byte(v>>8))
}

// WriteUInt32 appends uint32 v to the buffer
func (b *PacketBuffer) WriteUInt32(v uint32) {
	b.Bytes = append(b.Bytes, byte(v), byte(v>>8), byte(v>>16), byte(v>>24))
}

// WriteBool appends bool v to the buffer
func (b *PacketBuffer) WriteBool(v bool) {
	var i uint8
	if v {
		i = 1
	}
	b.Bytes = append(b.Bytes, i)
}

// WritePort appends port v to the buffer
func (b *PacketBuffer) WritePort(v uint16) {
	b.Bytes = append(b.Bytes, byte(v>>8), byte(v))
}

// WriteIP appends ip v to the buffer
func (b *PacketBuffer) WriteIP(v net.IP) error {
	if ip4 := v.To4(); ip4 != nil {
		b.WriteBlob(ip4)
		return nil
	} else if v != nil {
		return ErrInvalidIP4
	}

	b.WriteUInt32(0)
	return nil
}

// WriteSockAddr appends SockAddr v to the buffer
func (b *PacketBuffer) WriteSockAddr(v *SockAddr) error {
	if v.Port == 0 && v.IP == nil {
		b.WriteUInt16(0)
	} else {
		b.WriteUInt16(connAddressFamily)
	}

	b.WritePort(v.Port)
	if err := b.WriteIP(v.IP); err != nil {
		return err
	}

	b.WriteUInt32(0)
	b.WriteUInt32(0)
	return nil
}

// WriteString appends string v to the buffer
func (b *PacketBuffer) WriteString(s string) {
	b.WriteBlob([]byte(s))
	b.WriteUInt8(0)
}

// WriteBlobAt overwrites position p in the buffer with blob v
func (b *PacketBuffer) WriteBlobAt(p int, v []byte) {
	copy(b.Bytes[p:], v)
}

// WriteUInt8At overwrites position p in the buffer with uint8 v
func (b *PacketBuffer) WriteUInt8At(p int, v byte) {
	b.Bytes[p] = v
}

// WriteUInt16At overwrites position p in the buffer with uint16 v
func (b *PacketBuffer) WriteUInt16At(p int, v uint16) {
	b.Bytes[p+1], b.Bytes[p] = byte(v>>8), byte(v)
}

// WriteUInt32At overwrites position p in the buffer with uint32 v
func (b *PacketBuffer) WriteUInt32At(p int, v uint32) {
	b.Bytes[p+3], b.Bytes[p+2], b.Bytes[p+1], b.Bytes[p] = byte(v>>24), byte(v>>16), byte(v>>8), byte(v)
}

// WriteBoolAt overwrites position p in the buffer with bool v
func (b *PacketBuffer) WriteBoolAt(p int, v bool) {
	var i uint8
	if v {
		i = 1
	}
	b.Bytes[p] = i
}

// WritePortAt overwrites position p in the buffer with port v
func (b *PacketBuffer) WritePortAt(p int, v uint16) {
	b.Bytes[p+1], b.Bytes[p] = byte(v), byte(v>>8)
}

// WriteIPAt overwrites position p in the buffer with ip v
func (b *PacketBuffer) WriteIPAt(p int, v net.IP) error {
	if ip4 := v.To4(); ip4 != nil {
		b.WriteBlobAt(p, ip4)
		return nil
	}

	b.WriteUInt32At(p, 0)
	return ErrInvalidIP4
}

// WriteSockAddrAt overwrites position p in the buffer with SockAddr v
func (b *PacketBuffer) WriteSockAddrAt(p int, v *SockAddr) error {
	if v.Port == 0 && v.IP == nil {
		b.WriteUInt16At(p, 0)
	} else {
		b.WriteUInt16At(p, connAddressFamily)
	}

	b.WritePortAt(p+2, v.Port)
	if err := b.WriteIPAt(p+4, v.IP); err != nil {
		return err
	}

	b.WriteUInt32At(p+8, 0)
	b.WriteUInt32At(p+12, 0)
	return nil
}

// WriteStringAt overwrites position p in the buffer with string v
func (b *PacketBuffer) WriteStringAt(p int, s string) {
	var Bytes = []byte(s)
	b.WriteBlobAt(p, Bytes)
	b.WriteUInt8At(p+len(Bytes), 0)
}

// Read implements io.Reader interface
func (b *PacketBuffer) Read(p []byte) (int, error) {
	var size = len(b.Bytes)
	if size == 0 {
		return 0, io.EOF
	}
	if size > len(p) {
		size = len(p)
	}

	copy(p[:size], b.Bytes[:size])
	b.Bytes = b.Bytes[size:]

	return size, nil
}

// ReadBlob consumes a blob of size len and returns (a slice of) its value
func (b *PacketBuffer) ReadBlob(len int) []byte {
	if len > 0 {
		var res = b.Bytes[:len]
		b.Bytes = b.Bytes[len:]
		return res
	}

	return nil
}

// ReadUInt8 consumes a uint8 and returns its value
func (b *PacketBuffer) ReadUInt8() byte {
	var res = byte(b.Bytes[0])
	b.Bytes = b.Bytes[1:]
	return res
}

// ReadUInt16 a uint16 and returns its value
func (b *PacketBuffer) ReadUInt16() uint16 {
	var res = uint16(b.Bytes[1])<<8 | uint16(b.Bytes[0])
	b.Bytes = b.Bytes[2:]
	return res
}

// ReadUInt32 consumes a uint32 and returns its value
func (b *PacketBuffer) ReadUInt32() uint32 {
	var res = uint32(b.Bytes[3])<<24 | uint32(b.Bytes[2])<<16 | uint32(b.Bytes[1])<<8 | uint32(b.Bytes[0])
	b.Bytes = b.Bytes[4:]
	return res
}

// ReadBool consumes a bool and returns its value
func (b *PacketBuffer) ReadBool() bool {
	var res bool
	if b.Bytes[0] > 0 {
		res = true
	}
	b.Bytes = b.Bytes[1:]
	return res
}

// ReadPort consumes a port and returns its value
func (b *PacketBuffer) ReadPort() uint16 {
	var res = uint16(b.Bytes[1]) | uint16(b.Bytes[0])<<8
	b.Bytes = b.Bytes[2:]
	return res
}

// ReadIP consumes an ip and returns its value
func (b *PacketBuffer) ReadIP() net.IP {
	var res = net.IP(b.ReadBlob(net.IPv4len))
	if res.Equal(net.IPv4zero) {
		return nil
	}
	return res
}

// ReadSockAddr consumes a SockAddr structure and returns its value
func (b *PacketBuffer) ReadSockAddr() (SockAddr, error) {
	var res = SockAddr{}

	switch b.ReadUInt16() {
	case 0:
		if b.ReadPort() != 0 || b.ReadIP() != nil {
			return res, ErrInvalidSockAddr
		}
	case connAddressFamily:
		res.Port = b.ReadPort()
		res.IP = b.ReadIP()
	default:
		return res, ErrInvalidSockAddr
	}

	if b.ReadUInt32() != 0 || b.ReadUInt32() != 0 {
		return res, ErrInvalidSockAddr
	}

	return res, nil
}

// ReadString consumes a null terminated string and returns its value
func (b *PacketBuffer) ReadString() (string, error) {
	var pos = bytes.IndexByte(b.Bytes, 0)
	if pos == -1 {
		b.Bytes = b.Bytes[len(b.Bytes):]
		return "", ErrNoStringTerminatorFound
	}

	var res = string(b.Bytes[:pos])
	b.Bytes = b.Bytes[pos+1:]
	return res, nil
}
