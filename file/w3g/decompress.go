// Author:  Niels A.D.
// Project: gowarcraft3 (https://github.com/nielsAD/gowarcraft3)
// License: Mozilla Public License, v2.0

package w3g

import (
	"bufio"
	"compress/zlib"
	"hash"
	"hash/crc32"
	"io"
	"io/ioutil"

	"github.com/nielsAD/gowarcraft3/protocol"
)

// Decompressor is an io.Reader that decompresses data blocks
type Decompressor struct {
	Size      uint32
	BlockSize uint16

	r   io.Reader
	idx uint32
	len uint32

	dec io.ReadCloser
	lim io.LimitedReader

	crc     hash.Hash32
	crcData uint16
}

// NewDecompressor for compressed w3g data
func NewDecompressor(r io.Reader, numBlocks uint32, size uint32) *Decompressor {
	return &Decompressor{
		r:    r,
		len:  numBlocks,
		Size: size,
	}
}

func (d *Decompressor) nextBlock() error {
	if d.idx >= d.len {
		return io.EOF
	}
	if d.idx > 0 {
		if err := d.closeBlock(); err != nil {
			return err
		}
	}
	d.idx++

	var buf [8]byte
	_, err := io.ReadFull(d.r, buf[:8])
	if err != nil {
		return err
	}

	var pbuf = protocol.Buffer{Bytes: buf[:]}
	var lenDeflate = pbuf.ReadUInt16()
	d.BlockSize = pbuf.ReadUInt16()

	var crcHead = pbuf.ReadUInt16()
	d.crcData = pbuf.ReadUInt16()

	buf[4], buf[5], buf[6], buf[7] = 0, 0, 0, 0
	var crc = crc32.ChecksumIEEE(buf[:8])
	if crcHead != uint16(crc^crc>>16) {
		return ErrInvalidChecksum
	}

	// Use limr to keep track of how many compressed bytes are read
	d.lim.R = d.r
	d.lim.N = int64(lenDeflate)

	if d.crc == nil {
		d.crc = crc32.NewIEEE()
	} else {
		d.crc.Reset()
	}

	// Tee to hash to calculate crc while decompressing
	d.dec, err = zlib.NewReader(io.TeeReader(&d.lim, d.crc))
	return err
}

func (d *Decompressor) closeBlock() error {
	if d.dec == nil {
		return nil
	}

	d.dec.Close()
	d.dec = nil

	if d.BlockSize > 0 || d.lim.N > 0 {
		return io.ErrUnexpectedEOF
	}

	var sum = d.crc.Sum32()
	if d.crcData != uint16(sum^sum>>16) {
		return ErrInvalidChecksum
	}

	return nil
}

// Read implements the io.Reader interface.
func (d *Decompressor) Read(b []byte) (int, error) {
	if d.Size == 0 {
		return 0, io.EOF
	}

	var n = 0
	var l = len(b)
	if uint32(l) > d.Size {
		b = b[:d.Size]
		l = len(b)
	}

	for n != l {
		if d.dec == nil {
			if err := d.nextBlock(); err != nil {
				return n, err
			}
		}

		nn, err := io.ReadFull(d.dec, b[n:])
		d.Size -= uint32(nn)
		d.BlockSize -= uint16(nn)
		n += nn

		switch err {
		case nil:
			if d.Size == 0 && d.BlockSize > 0 {
				nn, _ := io.Copy(ioutil.Discard, d.dec)
				d.BlockSize -= uint16(nn)
			}
			if d.BlockSize > 0 {
				continue
			}
			fallthrough
		case io.ErrUnexpectedEOF:
			if err := d.closeBlock(); err != nil {
				return n, err
			}
		default:
			return n, err
		}
	}

	return n, nil
}

// ForEach record call f
func (d *Decompressor) ForEach(f func(r Record) error) error {
	var b = bufio.NewReaderSize(d, 8192)
	var buf DeserializationBuffer

	for {
		r, _, err := DeserializeRecordWithBuffer(b, &buf)
		switch err {
		case nil:
			if err := f(r); err != nil {
				return err
			}
		case io.EOF:
			return nil
		default:
			return err
		}
	}
}

// Close DataDecoder
func (d *Decompressor) Close() {
	if d.dec != nil {
		d.dec.Close()
	}
}
