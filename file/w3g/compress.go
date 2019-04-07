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
)

const defaultBufSize = 8192

// Compressor is an io.Writer that compresses data blocks
type Compressor struct {
	SizeWritten uint32 // Compressed size written in total
	SizeTotal   uint32 // Decompressed size written in total
	NumBlocks   uint32 // Blocks written in total

	w io.Writer
	b protocol.Buffer
}

// NewCompressor for compressed w3g data
func NewCompressor(w io.Writer) *Compressor {
	return &Compressor{
		w: w,
	}
}

// Write implements the io.Writer interface.
func (d *Compressor) Write(b []byte) (int, error) {
	var n = 0
	for len(b) > 0 {
		var l = len(b)
		if l > math.MaxUint16 {
			l = math.MaxUint16
		}

		// Header with placeholders for size
		d.b.Truncate()
		d.b.WriteUInt16(0)
		d.b.WriteUInt16(uint16(l))
		d.b.WriteUInt32(0)

		var zw = zlib.NewWriter(&d.b)
		zn, err := zw.Write(b[:l])
		n += zn

		if err != nil {
			return n, err
		}
		if err := zw.Flush(); err != nil {
			return n, err
		}

		// Update header
		d.b.WriteUInt16At(0, uint16(d.b.Size()-8))

		var crcHead = crc32.ChecksumIEEE(d.b.Bytes[:8])
		d.b.WriteUInt16At(4, uint16(crcHead^crcHead>>16))

		var crcData = crc32.ChecksumIEEE(d.b.Bytes[8:])
		d.b.WriteUInt16At(6, uint16(crcData^crcData>>16))

		wn, err := d.w.Write(d.b.Bytes)
		d.SizeWritten += uint32(wn)
		d.SizeTotal += uint32(zn)
		d.NumBlocks++

		if err != nil {
			return n, err
		}

		b = b[l:]
	}

	return n, nil
}

// BufferedCompressor is an io.Writer that compresses data blocks
type BufferedCompressor struct {
	*bufio.Writer
	*Compressor
	b SerializationBuffer
}

// NewBufferedCompressorSize for compressed w3g with specified buffer size
func NewBufferedCompressorSize(w io.Writer, size int) *BufferedCompressor {
	var c = NewCompressor(w)
	return &BufferedCompressor{
		Compressor: c,
		Writer:     bufio.NewWriterSize(c, size),
	}
}

// NewBufferedCompressor for compressed w3g with default buffer size
func NewBufferedCompressor(w io.Writer) *BufferedCompressor {
	return NewBufferedCompressorSize(w, defaultBufSize)
}

// Write implements the io.Writer interface.
func (d *BufferedCompressor) Write(p []byte) (int, error) {
	return d.Writer.Write(p)
}

// WriteRecord serializes r and writes it to d
func (d *BufferedCompressor) WriteRecord(r Record) (int, error) {
	return SerializeRecordWithBuffer(d, &d.b, r)
}

// WriteRecords serializes r and writes to d
func (d *BufferedCompressor) WriteRecords(r ...Record) (int, error) {
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
func (d *BufferedCompressor) Close() error {
	var a = d.Available()
	if a > 0 && d.Buffered() > 0 {
		n, _ := d.Writer.Write(make([]byte, a))
		d.SizeTotal -= uint32(n)
	}
	return d.Writer.Flush()
}
