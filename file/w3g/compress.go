// Author:  Niels A.D.
// Project: gowarcraft3 (https://github.com/nielsAD/gowarcraft3)
// License: Mozilla Public License, v2.0

package w3g

import (
	"bufio"
	"compress/zlib"
	"hash/crc32"
	"io"
	"math"

	"github.com/nielsAD/gowarcraft3/protocol"
	"github.com/nielsAD/gowarcraft3/protocol/w3gs"
)

const defaultBufSize = 8192

// BlockCompressor is an io.Writer that compresses data blocks
type BlockCompressor struct {
	Encoding

	SizeWritten uint32 // Compressed size written in total
	SizeTotal   uint32 // Decompressed size written in total
	NumBlocks   uint32 // Blocks written in total

	w io.Writer
	b protocol.Buffer
	z *zlib.Writer
}

// NewBlockCompressor for compressed w3g data
func NewBlockCompressor(w io.Writer, e Encoding) *BlockCompressor {
	z, _ := zlib.NewWriterLevelDict(nil, zlib.BestCompression, nil)
	return &BlockCompressor{
		Encoding: e,
		w:        w,
		z:        z,
	}
}

// Write implements the io.Writer interface.
func (d *BlockCompressor) Write(b []byte) (int, error) {
	var n = 0
	for len(b) > 0 {
		var lenBuf = len(b)
		var lenHdr = 12

		// Header with placeholders for size
		d.b.Truncate()
		if d.GameVersion >= w3gs.ReforgedGameVersion {
			d.b.WriteUInt32(0)
			d.b.WriteUInt32(uint32(lenBuf))
		} else {
			if lenBuf > math.MaxUint16 {
				lenBuf = math.MaxUint16
			}
			lenHdr = 8
			d.b.WriteUInt16(0)
			d.b.WriteUInt16(uint16(lenBuf))
		}

		d.b.WriteUInt32(0)

		d.z.Reset(&d.b)
		zn, err := d.z.Write(b[:lenBuf])
		n += zn

		if err != nil {
			return n, err
		}
		if err := d.z.Flush(); err != nil {
			return n, err
		}

		// Update header
		if d.GameVersion >= w3gs.ReforgedGameVersion {
			d.b.WriteUInt32At(0, uint32(d.b.Size()-lenHdr))
		} else {
			d.b.WriteUInt16At(0, uint16(d.b.Size()-lenHdr))
		}

		var crcHead = crc32.ChecksumIEEE(d.b.Bytes[:lenHdr])
		d.b.WriteUInt16At(lenHdr-4, uint16(crcHead^crcHead>>16))

		var crcData = crc32.ChecksumIEEE(d.b.Bytes[lenHdr:])
		d.b.WriteUInt16At(lenHdr-2, uint16(crcData^crcData>>16))

		wn, err := d.w.Write(d.b.Bytes)
		d.SizeWritten += uint32(wn)
		d.SizeTotal += uint32(zn)
		d.NumBlocks++

		if err != nil {
			return n, err
		}

		b = b[lenBuf:]
	}

	return n, nil
}

// Compressor is an io.Writer that compresses buffered data blocks
type Compressor struct {
	RecordEncoder
	*BlockCompressor
	*bufio.Writer
}

// NewCompressorSize for compressed w3g with specified buffer size
func NewCompressorSize(w io.Writer, e Encoding, size int) *Compressor {
	var c = NewBlockCompressor(w, e)
	var b = bufio.NewWriterSize(c, size)

	return &Compressor{
		RecordEncoder: RecordEncoder{
			Encoding: e,
		},
		BlockCompressor: c,
		Writer:          b,
	}
}

// NewCompressor for compressed w3g with default buffer size
func NewCompressor(w io.Writer, e Encoding) *Compressor {
	return NewCompressorSize(w, e, defaultBufSize)
}

// Write implements the io.Writer interface.
func (d *Compressor) Write(p []byte) (int, error) {
	return d.Writer.Write(p)
}

// WriteRecord serializes r and writes it to d
func (d *Compressor) WriteRecord(r Record) (int, error) {
	return d.RecordEncoder.Write(d.Writer, r)
}

// WriteRecords serializes r and writes to d
func (d *Compressor) WriteRecords(r ...Record) (int, error) {
	var n = 0
	for _, v := range r {
		nn, err := d.WriteRecord(v)
		n += nn

		if err != nil {
			return n, err
		}
	}
	return n, nil
}

// Close adds padding to fill last block and flushes any buffered data
func (d *Compressor) Close() error {
	var a = d.Writer.Available()
	if a > 0 && d.Writer.Buffered() > 0 {
		n, _ := d.Writer.Write(make([]byte, a))
		d.SizeTotal -= uint32(n)
	}
	return d.Writer.Flush()
}
