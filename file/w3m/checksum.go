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

	"github.com/nielsAD/gowarcraft3/file/fs"
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

func (m *Map) findFile(files []string, stor *fs.Storage) (io.ReadCloser, error) {
	for _, p := range files {
		if file, err := m.Archive.Open(p); err == nil {
			return file, nil
		} else if err != os.ErrNotExist {
			return nil, err
		}

		if stor != nil {
			if file, err := stor.Open(p); err == nil {
				return file, nil
			} else if err != os.ErrNotExist {
				return nil, err
			}
		}
	}

	return nil, nil
}

// Checksum returns the content hash that identifies the map (used in version < 1.32)
func (m *Map) Checksum(stor *fs.Storage) (*Hash, error) {
	var sha = sha1.New()
	var xor = xoro(0)
	var buf = make([]byte, 32*1024)

	for _, file := range hashFiles1 {
		r, err := m.findFile(file, stor)
		if err != nil {
			return nil, err
		}
		if r == nil {
			continue
		}

		var subxor xoro
		io.CopyBuffer(io.MultiWriter(sha, &subxor), r, buf)
		xor ^= subxor

		if r != nil {
			r.Close()
		}
	}

	xor = xoro(bits.RotateLeft32(uint32(xor), 3))

	var magic = []byte{0x9E, 0x37, 0xF1, 0x03}
	xor.Write(magic)
	sha.Write(magic)

	for _, file := range hashFiles2 {
		r, err := m.findFile(file, stor)
		if err != nil {
			return nil, err
		}
		if r == nil {
			continue
		}

		var subxor xoro
		io.CopyBuffer(io.MultiWriter(sha, &subxor), r, buf)
		xor.Write32(uint32(subxor))

		if r != nil {
			r.Close()
		}
	}

	var h = Hash{Xoro: uint32(xor)}
	copy(h.Sha1[:], sha.Sum(nil))

	return &h, nil
}
