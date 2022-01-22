// Author:  Niels A.D.
// Project: gowarcraft3 (https://github.com/nielsAD/gowarcraft3)
// License: Mozilla Public License, v2.0

package protocol

import (
	"bytes"
	"errors"
	"io"
	"math"
	"math/bits"
	"net"
)

// Errors
var (
	ErrInvalidIP4               = errors.New("pbuf: Invalid IP4 address")
	ErrInvalidSockAddr          = errors.New("pbuf: Invalid SockAddr structure")
	ErrNoCStringTerminatorFound = errors.New("pbuf: No null terminator for string found in buffer")
)

// AF_INET
const connAddressFamily uint16 = 2

// Buffer wraps a []byte slice and adds helper functions for binary (de)serialization
type Buffer struct {
	Bytes []byte
}

// Size returns the total size of the buffer
func (b *Buffer) Size() int {
	return len(b.Bytes)
}

// Skip consumes len bytes and throws away the result
func (b *Buffer) Skip(len int) {
	b.Reset(b.Bytes[len:])
}

// Truncate resets the buffer to size 0
func (b *Buffer) Truncate() {
	b.Reset(b.Bytes[:0])
}

// Reset buffer to p
func (b *Buffer) Reset(p []byte) {
	b.Bytes = p
}

// Write implements io.Writer interface
func (b *Buffer) Write(p []byte) (int, error) {
	b.WriteBlob(p)
	return len(p), nil
}

// WriteByte implements io.ByteWriter interface
func (b *Buffer) WriteByte(c byte) error {
	b.WriteUInt8(c)
	return nil
}

// WriteBlob appends blob v to the buffer
func (b *Buffer) WriteBlob(v []byte) {
	b.Reset(append(b.Bytes, v...))
}

// WriteUInt8 appends uint8 v to the buffer
func (b *Buffer) WriteUInt8(v byte) {
	b.Reset(append(b.Bytes, v))
}

// WriteUInt16 appends uint16 v to the buffer
func (b *Buffer) WriteUInt16(v uint16) {
	b.Reset(append(b.Bytes, byte(v), byte(v>>8)))
}

// WriteUInt32 appends uint32 v to the buffer
func (b *Buffer) WriteUInt32(v uint32) {
	b.Reset(append(b.Bytes, byte(v), byte(v>>8), byte(v>>16), byte(v>>24)))
}

// WriteUInt64 appends uint64 v to the buffer
func (b *Buffer) WriteUInt64(v uint64) {
	b.Reset(append(b.Bytes, byte(v), byte(v>>8), byte(v>>16), byte(v>>24), byte(v>>32), byte(v>>40), byte(v>>48), byte(v>>56)))
}

// WriteFloat32 appends float32 v to the buffer
func (b *Buffer) WriteFloat32(v float32) {
	b.WriteUInt32(math.Float32bits(v))
}

// WriteBool8 appends bool v to the buffer
func (b *Buffer) WriteBool8(v bool) {
	var i uint8
	if v {
		i = 1
	}
	b.WriteUInt8(i)
}

// WriteBool32 appends bool v to the buffer
func (b *Buffer) WriteBool32(v bool) {
	var i uint32
	if v {
		i = 1
	}
	b.WriteUInt32(i)
}

// WriteIP appends ip v to the buffer
func (b *Buffer) WriteIP(v net.IP) error {
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
func (b *Buffer) WriteSockAddr(v *SockAddr) error {
	if v.IP == nil {
		b.WriteUInt32(0)
		b.WriteUInt32(0)
	} else {
		b.WriteUInt16(connAddressFamily)
		b.WriteUInt16(bits.ReverseBytes16(v.Port))
		if err := b.WriteIP(v.IP); err != nil {
			return err
		}
	}

	b.WriteUInt32(0)
	b.WriteUInt32(0)
	return nil
}

// WriteCString appends null terminated string v to the buffer
func (b *Buffer) WriteCString(v string) {
	b.WriteBlob([]byte(v))
	b.WriteUInt8(0)
}

// WriteLEDString appends little-endian dword string v to the buffer
func (b *Buffer) WriteLEDString(v DWordString) {
	b.WriteUInt32(uint32(v))
}

// WriteBEDString appends big-endian dword string v to the buffer
func (b *Buffer) WriteBEDString(v DWordString) {
	b.WriteUInt32(bits.ReverseBytes32(uint32(v)))
}

// WriteBlobAt overwrites position p in the buffer with blob v
func (b *Buffer) WriteBlobAt(p int, v []byte) {
	copy(b.Bytes[p:], v)
}

// WriteUInt8At overwrites position p in the buffer with uint8 v
func (b *Buffer) WriteUInt8At(p int, v byte) {
	b.Bytes[p] = v
}

// WriteUInt16At overwrites position p in the buffer with uint16 v
func (b *Buffer) WriteUInt16At(p int, v uint16) {
	b.Bytes[p+1], b.Bytes[p] = byte(v>>8), byte(v)
}

// WriteUInt32At overwrites position p in the buffer with uint32 v
func (b *Buffer) WriteUInt32At(p int, v uint32) {
	b.Bytes[p+3], b.Bytes[p+2], b.Bytes[p+1], b.Bytes[p] = byte(v>>24), byte(v>>16), byte(v>>8), byte(v)
}

// WriteUInt64At overwrites position p in the buffer with uint64 v
func (b *Buffer) WriteUInt64At(p int, v uint64) {
	b.Bytes[p+7], b.Bytes[p+6], b.Bytes[p+5], b.Bytes[p+4], b.Bytes[p+3], b.Bytes[p+2], b.Bytes[p+1], b.Bytes[p] = byte(v>>56), byte(v>>48), byte(v>>40), byte(v>>32), byte(v>>24), byte(v>>16), byte(v>>8), byte(v)
}

// WriteFloat32At overwrites position p in the buffer with float32 v
func (b *Buffer) WriteFloat32At(p int, v float32) {
	b.WriteUInt32At(p, math.Float32bits(v))
}

// WriteBool8At overwrites position p in the buffer with bool v
func (b *Buffer) WriteBool8At(p int, v bool) {
	var i uint8
	if v {
		i = 1
	}
	b.WriteUInt8At(p, i)
}

// WriteBool32At overwrites position p in the buffer with bool v
func (b *Buffer) WriteBool32At(p int, v bool) {
	var i uint32
	if v {
		i = 1
	}
	b.WriteUInt32At(p, i)
}

// WriteIPAt overwrites position p in the buffer with ip v
func (b *Buffer) WriteIPAt(p int, v net.IP) error {
	if ip4 := v.To4(); ip4 != nil {
		b.WriteBlobAt(p, ip4)
		return nil
	} else if v != nil {
		return ErrInvalidIP4
	}

	b.WriteUInt32At(p, 0)
	return nil
}

// WriteSockAddrAt overwrites position p in the buffer with SockAddr v
func (b *Buffer) WriteSockAddrAt(p int, v *SockAddr) error {
	if v.IP == nil {
		b.WriteUInt32At(p, 0)
		b.WriteUInt32At(p+4, 0)
	} else {
		b.WriteUInt16At(p, connAddressFamily)
		b.WriteUInt16At(p+2, bits.ReverseBytes16(v.Port))
		if err := b.WriteIPAt(p+4, v.IP); err != nil {
			return err
		}
	}

	b.WriteUInt32At(p+8, 0)
	b.WriteUInt32At(p+12, 0)
	return nil
}

// WriteCStringAt overwrites position p in the buffer with null terminated string v
func (b *Buffer) WriteCStringAt(p int, v string) {
	var bv = []byte(v)
	b.WriteBlobAt(p, bv)
	b.WriteUInt8At(p+len(bv), 0)
}

// WriteLEDStringAt overwrites position p in the buffer with little-endian dword string v
func (b *Buffer) WriteLEDStringAt(p int, v DWordString) {
	b.WriteUInt32At(p, uint32(v))
}

// WriteBEDStringAt overwrites position p in the buffer with big-endian dword string v
func (b *Buffer) WriteBEDStringAt(p int, v DWordString) {
	b.WriteUInt32At(p, bits.ReverseBytes32(uint32(v)))
}

// WriteTo implements io.WriterTo interface
func (b *Buffer) WriteTo(w io.Writer) (int64, error) {
	n, err := w.Write(b.Bytes)
	b.Reset(b.Bytes[n:])
	return int64(n), err
}

// ReadSizeFrom reads n bytes from r
func (b *Buffer) ReadSizeFrom(r io.Reader, n int) (int, error) {
	var s = len(b.Bytes)
	if s+n <= cap(b.Bytes) {
		b.Reset(b.Bytes[:s+n])
	} else {
		b.Reset(append(b.Bytes, make([]byte, n)...))
	}

	nn, err := io.ReadFull(r, b.Bytes[s:])
	b.Reset(b.Bytes[:s+nn])

	return nn, err
}

// ReadFrom implements io.ReaderFrom interface
func (b *Buffer) ReadFrom(r io.Reader) (int64, error) {
	var n = int64(0)
	for {
		nn, err := b.ReadSizeFrom(r, 2048)
		n += int64(nn)

		switch err {
		case nil:
			// Success
		case io.EOF, io.ErrUnexpectedEOF:
			return n, nil
		default:
			return n, err
		}
	}
}

// Read implements io.Reader interface
func (b *Buffer) Read(p []byte) (int, error) {
	var size = len(b.Bytes)
	if size == 0 {
		return 0, io.EOF
	}
	if size > len(p) {
		size = len(p)
	}

	copy(p[:size], b.Bytes[:size])
	b.Reset(b.Bytes[size:])

	return size, nil
}

// ReadByte implements io.ByteReader interface
func (b *Buffer) ReadByte() (byte, error) {
	return b.ReadUInt8(), nil
}

// ReadBlob consumes a blob of size len and returns (a slice of) its value
func (b *Buffer) ReadBlob(len int) []byte {
	if len > 0 {
		var res = b.Bytes[:len]
		b.Reset(b.Bytes[len:])
		return res
	}

	return nil
}

// ReadUInt8 consumes a uint8 and returns its value
func (b *Buffer) ReadUInt8() byte {
	var res = byte(b.Bytes[0])
	b.Reset(b.Bytes[1:])
	return res
}

// ReadUInt16 a uint16 and returns its value
func (b *Buffer) ReadUInt16() uint16 {
	var res = uint16(b.Bytes[1])<<8 | uint16(b.Bytes[0])
	b.Reset(b.Bytes[2:])
	return res
}

// ReadUInt32 consumes a uint32 and returns its value
func (b *Buffer) ReadUInt32() uint32 {
	var res = uint32(b.Bytes[3])<<24 | uint32(b.Bytes[2])<<16 | uint32(b.Bytes[1])<<8 | uint32(b.Bytes[0])
	b.Reset(b.Bytes[4:])
	return res
}

// ReadUInt64 consumes a uint32 and returns its value
func (b *Buffer) ReadUInt64() uint64 {
	var res = uint64(b.Bytes[7])<<56 | uint64(b.Bytes[6])<<48 | uint64(b.Bytes[5])<<40 | uint64(b.Bytes[4])<<32 | uint64(b.Bytes[3])<<24 | uint64(b.Bytes[2])<<16 | uint64(b.Bytes[1])<<8 | uint64(b.Bytes[0])
	b.Reset(b.Bytes[8:])
	return res
}

// ReadFloat32 consumes a float32 and returns its value
func (b *Buffer) ReadFloat32() float32 {
	return math.Float32frombits(b.ReadUInt32())
}

// ReadBool8 consumes a bool and returns its value
func (b *Buffer) ReadBool8() bool {
	return b.ReadUInt8() > 0
}

// ReadBool32 consumes a bool and returns its value
func (b *Buffer) ReadBool32() bool {
	return b.ReadUInt32() > 0
}

// ReadIP consumes an ip and returns its value
func (b *Buffer) ReadIP() net.IP {
	var ip = b.ReadUInt32()
	switch ip {
	case 0:
		return nil
	default:
		return net.IP([]byte{byte(ip), byte(ip >> 8), byte(ip >> 16), byte(ip >> 24)})
	}
}

// ReadSockAddr consumes a SockAddr structure and returns its value
func (b *Buffer) ReadSockAddr() (SockAddr, error) {
	var res = SockAddr{}

	switch b.ReadUInt16() {
	case 0:
		if b.ReadUInt16() != 0 || b.ReadUInt32() != 0 {
			return res, ErrInvalidSockAddr
		}
		res.Port = 0
		res.IP = nil
	case connAddressFamily:
		res.Port = bits.ReverseBytes16(b.ReadUInt16())
		res.IP = b.ReadIP()
	default:
		return res, ErrInvalidSockAddr
	}

	//lint:ignore SA4000 Consume two uint32
	if b.ReadUInt32() != 0 || b.ReadUInt32() != 0 {
		return res, ErrInvalidSockAddr
	}

	return res, nil
}

// ReadCString consumes a null terminated string and returns its value
func (b *Buffer) ReadCString() (string, error) {
	var pos = bytes.IndexByte(b.Bytes, 0)
	if pos == -1 {
		b.Reset(b.Bytes[len(b.Bytes):])
		return "", ErrNoCStringTerminatorFound
	}

	var res = string(b.Bytes[:pos])
	b.Reset(b.Bytes[pos+1:])
	return res, nil
}

// ReadLEDString consumes a little-endian dword string and returns its value
func (b *Buffer) ReadLEDString() DWordString {
	return DWordString(b.ReadUInt32())
}

// ReadBEDString consumes a big-endian dword string and returns its value
func (b *Buffer) ReadBEDString() DWordString {
	return DWordString(bits.ReverseBytes32(b.ReadUInt32()))
}
