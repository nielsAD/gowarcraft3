// Author:  Niels A.D.
// Project: gowarcraft3 (https://github.com/nielsAD/gowarcraft3)
// License: Mozilla Public License, v2.0

// Package blp is a BLIzzard Picture image format decoder.
package blp

import (
	"errors"
	"image"
	"image/jpeg"
	"io"
	"strings"

	"github.com/nielsAD/gowarcraft3/protocol"
)

// Errors
var (
	ErrBadFormat          = errors.New("blp: Invalid file format")
	ErrInvalidCompression = errors.New("blp: Compression not supported")
)

// Header constant for BLP files
var Header = protocol.DString("BLP1")

// Decode a BLP image. Only take the first image if it's a mipmap.
func Decode(r io.Reader) (image.Image, error) {
	var b protocol.Buffer
	if _, err := io.Copy(&b, r); err != nil {
		return nil, err
	}

	var size = b.Size()
	if size < 156 {
		return nil, ErrBadFormat
	}

	if b.ReadLEDString() != Header {
		return nil, ErrBadFormat
	}

	var compression = b.ReadUInt32() //compression
	var alphaBits = b.ReadUInt32()   //alpha

	switch alphaBits {
	case 0, 8:
	default:
		return nil, ErrBadFormat
	}

	b.ReadUInt32() //width
	b.ReadUInt32() //height
	b.ReadUInt32() //flags
	b.ReadUInt32() //hasMipmap

	var mmOffset [16]uint32
	for i := 0; i < len(mmOffset); i++ {
		mmOffset[i] = b.ReadUInt32()
	}

	var mmSize [16]uint32
	for i := 0; i < len(mmOffset); i++ {
		mmSize[i] = b.ReadUInt32()
	}

	switch compression {
	// JPEG
	case 0x00:
		var hSize = b.ReadUInt32()
		if b.Size() < int(hSize) || mmOffset[0] == 0 || mmSize[0] == 0 {
			return nil, ErrBadFormat
		}

		var imgBuf = make([]byte, 0, hSize+mmSize[0])
		imgBuf = append(imgBuf, b.ReadBlob(int(hSize))...)

		var offset = int(mmOffset[0]) - size + b.Size()
		if offset < 0 || offset+int(mmSize[0]) > b.Size() {
			return nil, ErrBadFormat
		}

		b.Skip(offset)
		imgBuf = append(imgBuf, b.ReadBlob(int(mmSize[0]))...)

		img, err := jpeg.Decode(&protocol.Buffer{Bytes: imgBuf})

		// Workaround for CMYK image without APP14 marker
		if err != nil && strings.Contains(err.Error(), "Adobe APP14") {
			imgBuf = append([]byte{
				0xFF, 0xD8, //SOIMAGE
				0xFF, 0xEE, 0x00, 0x0E, //App14Marker
				'A', 'd', 'o', 'b', 'e',
				0, 0, 0, 0, 0, 0, 0,
			}, imgBuf[2:]...)
			img, err = jpeg.Decode(&protocol.Buffer{Bytes: imgBuf})
		}

		return img, err

	// case 0x01: PALETTE
	default:
		return nil, ErrInvalidCompression
	}
}
