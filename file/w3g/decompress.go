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
	SizeRead  uint32 // Compressed size read in total
	SizeTotal uint32 // Decompressed size left to read in total
	SizeBlock uint16 // Decompressed size left to read current block
	NumBlocks uint32 // Blocks left to read

	r   io.Reader
	dec io.ReadCloser
	lim io.LimitedReader

	crc     hash.Hash32
	crcData uint16
}

// NewDecompressor for compressed w3g data
func NewDecompressor(r io.Reader, numBlocks uint32, sizeTotal uint32) *Decompressor {
	return &Decompressor{
		r:         r,
		SizeTotal: sizeTotal,
		NumBlocks: numBlocks,
	}
}

func (d *Decompressor) nextBlock() error {
	if d.NumBlocks == 0 {
		return io.EOF
	}
	if err := d.closeBlock(); err != nil {
		return err
	}

	d.NumBlocks--

	var buf [8]byte
	n, err := io.ReadFull(d.r, buf[:8])

	d.SizeRead += uint32(n)
	if err != nil {
		return err
	}

	var pbuf = protocol.Buffer{Bytes: buf[:]}
	var lenDeflate = pbuf.ReadUInt16()
	d.SizeBlock = pbuf.ReadUInt16()

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
	d.SizeRead += uint32(lenDeflate - uint16(d.lim.N))

	return err
}

func (d *Decompressor) closeBlock() error {
	if d.dec == nil {
		return nil
	}
	d.dec = nil

	if d.SizeBlock > 0 || d.lim.N > 0 {
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
	if d.SizeTotal == 0 {
		return 0, io.EOF
	}

	var n = 0
	var l = len(b)
	if uint32(l) > d.SizeTotal {
		b = b[:d.SizeTotal]
		l = len(b)
	}

	for n != l {
		if d.dec == nil {
			if err := d.nextBlock(); err != nil {
				return n, err
			}
		}

		var r = d.lim.N
		nn, err := io.ReadFull(d.dec, b[n:])
		d.SizeRead += uint32(r - d.lim.N)
		d.SizeTotal -= uint32(nn)
		d.SizeBlock -= uint16(nn)
		n += nn

		switch err {
		case nil:
			if d.SizeTotal == 0 && d.SizeBlock > 0 {
				nn, _ := io.Copy(ioutil.Discard, d.dec)
				d.SizeBlock -= uint16(nn)
			}
			if d.SizeBlock > 0 {
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
