// Author:  Niels A.D.
// Project: gowarcraft3 (https://github.com/nielsAD/gowarcraft3)
// License: Mozilla Public License, v2.0

package w3g

import (
	"bytes"
	"hash/crc32"
	"io"
	"io/ioutil"

	"github.com/nielsAD/gowarcraft3/protocol"
	"github.com/nielsAD/gowarcraft3/protocol/w3gs"
)

// Header for a Warcraft III recorded game file
type Header struct {
	GameVersion  w3gs.GameVersion
	BuildNumber  uint16
	DurationMS   uint32
	SinglePlayer bool
}

// FindHeader in r
func FindHeader(r Peeker) (int, error) {
	var n = 0

	for {
		b, err := r.Peek(2048)
		if len(b) == 0 {
			return n, err
		}

		var idx = bytes.Index(b, []byte(Signature))
		var del = idx
		if del < 0 {
			del = 2048 - len(Signature) + 1
		}
		nn, err := r.Discard(del)
		n += nn

		if idx >= 0 || err != nil {
			return n, err
		}
	}
}

// DecodeHeader a w3g file, returns header and a Decompressor to read compressed records
func DecodeHeader(r io.Reader, f RecordFactory) (*Header, *Decompressor, int, error) {
	var buf [68]byte
	var hdr Header

	n, err := io.ReadFull(r, buf[:64])
	if err != nil {
		return nil, nil, n, err
	}

	var pbuf = protocol.Buffer{Bytes: buf[:]}
	if s, err := pbuf.ReadCString(); err != nil {
		return nil, nil, n, err
	} else if s != Signature {
		return nil, nil, n, ErrBadFormat
	}

	var sizeHeader = pbuf.ReadUInt32()
	var sizeFile = pbuf.ReadUInt32()

	var headerVersion = pbuf.ReadUInt32()
	switch headerVersion {
	case 0:
	case 1:
		nn, err := io.ReadFull(r, buf[64:68])
		n += nn
		if err != nil {
			return nil, nil, n, err
		}

	default:
		return nil, nil, n, ErrUnexpectedConst
	}

	var sizeBlocks = pbuf.ReadUInt32()
	var numBlocks = pbuf.ReadUInt32()

	switch headerVersion {
	case 0:
		if pbuf.ReadUInt16() != 0 {
			return nil, nil, n, ErrUnexpectedConst
		}
		hdr.GameVersion.Product = w3gs.ProductROC
		hdr.GameVersion.Version = uint32(pbuf.ReadUInt16())
	case 1:
		hdr.GameVersion.DeserializeContent(&pbuf, &w3gs.Encoding{})
	}

	hdr.BuildNumber = pbuf.ReadUInt16()
	hdr.SinglePlayer = pbuf.ReadUInt16() == 0
	hdr.DurationMS = pbuf.ReadUInt32()

	var crc = pbuf.ReadUInt32()
	buf[n-4], buf[n-3], buf[n-2], buf[n-1] = 0, 0, 0, 0
	if crc != uint32(crc32.ChecksumIEEE(buf[0:n])) {
		return nil, nil, n, ErrInvalidChecksum
	}

	if uint32(n) > sizeHeader || uint32(n) > sizeFile {
		return nil, nil, n, ErrBadFormat
	}

	// Skip to start of data section
	nn, err := io.CopyN(ioutil.Discard, r, int64(sizeHeader-uint32(n)))
	n += int(nn)
	if err != nil {
		return nil, nil, n, err
	}

	return &hdr, NewDecompressor(r, hdr.Encoding(), f, numBlocks, sizeBlocks), n, err
}

// Encoding for (de)serialization
func (h *Header) Encoding() Encoding {
	return Encoding{
		Encoding: w3gs.Encoding{
			GameVersion: h.GameVersion.Version,
		},
	}
}

// Encoder compresses records and updates header on close
type Encoder struct {
	Header
	*Compressor

	b protocol.Buffer
	w io.Writer
}

// NewEncoder for replay file
func NewEncoder(w io.Writer, e Encoding) (*Encoder, error) {
	var res = Encoder{
		w: w,
	}

	if _, ok := w.(io.Seeker); ok {
		// Write placeholder for header
		var h [68]byte
		if _, err := w.Write(h[:]); err != nil {
			return nil, err
		}

		res.Compressor = NewCompressor(w, e)
	} else {
		res.Compressor = NewCompressor(&res.b, e)
	}

	return &res, nil
}

// Close writer, flush data, and update header.
// Does not close underlying writer.
func (e *Encoder) Close() error {
	if err := e.Compressor.Close(); err != nil {
		return err
	}

	var buf [68]byte
	var pbuf = protocol.Buffer{Bytes: buf[:0]}
	pbuf.WriteCString(Signature)
	pbuf.WriteUInt32(68)
	pbuf.WriteUInt32(e.SizeWritten + 68)
	pbuf.WriteUInt32(1)
	pbuf.WriteUInt32(e.SizeTotal)
	pbuf.WriteUInt32(e.NumBlocks)
	e.GameVersion.SerializeContent(&pbuf, &w3gs.Encoding{})
	pbuf.WriteUInt16(e.BuildNumber)
	if e.SinglePlayer {
		pbuf.WriteUInt16(0x0000)
	} else {
		pbuf.WriteUInt16(0x8000)
	}
	pbuf.WriteUInt32(e.DurationMS)
	pbuf.WriteUInt32(0)
	pbuf.WriteUInt32At(64, crc32.ChecksumIEEE(pbuf.Bytes))

	s, seeker := e.w.(io.Seeker)
	if seeker {
		// Seek to beginning
		if _, err := s.Seek(-int64(e.Compressor.SizeWritten+68), io.SeekCurrent); err != nil {
			return err
		}
		// Overwrite header
		if n, err := e.w.Write(pbuf.Bytes); err != nil {
			s.Seek(-int64(e.Compressor.SizeWritten+68-uint32(n)), io.SeekCurrent)
			return err
		}
		// Seek to end
		if _, err := s.Seek(int64(e.Compressor.SizeWritten), io.SeekCurrent); err != nil {
			return err
		}
		return nil
	}

	if _, err := e.w.Write(pbuf.Bytes); err != nil {
		return err
	}
	if _, err := e.w.Write(e.b.Bytes); err != nil {
		return err
	}

	return nil
}
