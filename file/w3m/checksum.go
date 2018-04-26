// Author:  Niels A.D.
// Project: gowarcraft3 (https://github.com/nielsAD/gowarcraft3)
// License: Mozilla Public License, v2.0

package w3m

import (
	"crypto/sha1"
	"encoding/base64"
	"fmt"
	"io"
	"math/bits"
	"os"

	"github.com/nielsAD/gowarcraft3/file/mpq"
	"github.com/nielsAD/gowarcraft3/protocol"
)

// Hash used to identify a loaded w3m/w3x map
type Hash struct {
	Xoro uint32
	Sha1 [20]byte
}

func (h *Hash) String() string {
	return fmt.Sprintf("0x%02X|%s", h.Xoro, base64.RawStdEncoding.EncodeToString(h.Sha1[:]))
}

// Helper for XOR - ROTL hash function
type xoro uint32

func (v *xoro) Write8(b uint8) {
	*v = xoro(bits.RotateLeft32(uint32(*v)^uint32(b), 3))
}

func (v *xoro) Write32(b uint32) {
	*v = xoro(bits.RotateLeft32(uint32(*v)^b, 3))
}

func (v *xoro) Write(b []byte) (int, error) {
	var buf = protocol.Buffer{Bytes: b}
	for buf.Size() >= 4 {
		v.Write32(buf.ReadUInt32())
	}
	for buf.Size() >= 1 {
		v.Write8(buf.ReadUInt8())
	}
	return len(b), nil
}

var hashFiles1 = [][]string{
	[]string{"scripts\\common.j"},
	[]string{"scripts\\blizzard.j"},
}

var hashFiles2 = [][]string{
	[]string{"war3map.j", "scripts\\war3map.j"},
	[]string{"war3map.w3e"},
	[]string{"war3map.wpm"},
	[]string{"war3map.doo"},
	[]string{"war3map.w3u"},
	[]string{"war3map.w3b"},
	[]string{"war3map.w3d"},
	[]string{"war3map.w3a"},
	[]string{"war3map.w3q"},
}

func (m *Map) findFile(file []string, defaultFiles map[string]io.Reader) (io.Reader, *mpq.File, error) {
	for _, p := range file {
		if file, err := m.Archive.Open(p); err == nil {
			return file, file, nil
		} else if err != os.ErrNotExist {
			return nil, nil, err
		}

		if defaultFiles[p] != nil {
			return defaultFiles[p], nil, nil
		}
	}

	return nil, nil, nil
}

// Checksum returns hash that identifies the map
func (m *Map) Checksum(defaultFiles map[string]io.Reader) (*Hash, error) {
	var sha = sha1.New()
	var xor = xoro(0)
	var buf = make([]byte, 32*1024)

	for _, file := range hashFiles1 {
		r, f, err := m.findFile(file, defaultFiles)
		if err != nil {
			return nil, err
		}
		if r == nil {
			continue
		}

		var subxor xoro
		io.CopyBuffer(io.MultiWriter(sha, &subxor), r, buf)
		xor ^= subxor

		if f != nil {
			f.Close()
		}
	}

	xor = xoro(bits.RotateLeft32(uint32(xor), 3))

	var magic = []byte{0x09E, 0x037, 0x0F1, 0x03}
	xor.Write(magic)
	sha.Write(magic)

	for _, file := range hashFiles2 {
		r, f, err := m.findFile(file, defaultFiles)
		if err != nil {
			return nil, err
		}
		if r == nil {
			continue
		}

		var subxor xoro
		io.CopyBuffer(io.MultiWriter(sha, &subxor), r, buf)
		xor.Write32(uint32(subxor))

		if f != nil {
			f.Close()
		}
	}

	var h = Hash{Xoro: uint32(xor)}
	copy(h.Sha1[:], sha.Sum(nil))

	return &h, nil
}
