// Author:  Niels A.D.
// Project: gowarcraft3 (https://github.com/nielsAD/gowarcraft3)
// License: Mozilla Public License, v2.0

package fakeplayer

import (
	"io"
	"net"
	"sync"
	"time"

	"github.com/nielsAD/gowarcraft3/protocol/w3gs"
)

// Peer represents a (real) player in game
type Peer struct {
	smutex sync.Mutex
	sbuf   w3gs.SerializationBuffer
	rbuf   w3gs.DeserializationBuffer
	conn   *net.TCPConn

	Name        string
	ID          uint8
	JoinCounter uint32

	RTT      uint32
	PeerMask uint32
}

// NextRawPacket waits for the next packet from Peer (with given timeout) and returns its deserialized representation
func (p *Peer) NextRawPacket(timeout time.Duration) (w3gs.Packet, error) {
	if p.conn == nil {
		return nil, io.EOF
	}

	if timeout != 0 {
		if err := p.conn.SetReadDeadline(time.Now().Add(timeout)); err != nil {
			return nil, err
		}
	}

	pkt, _, err := w3gs.DeserializePacketWithBuffer(p.conn, &p.rbuf)
	return pkt, err
}

// Send serializes packet and sends it to Peer
func (p *Peer) Send(pkt w3gs.Packet) (int, error) {
	p.smutex.Lock()
	defer p.smutex.Unlock()

	if p.conn == nil {
		return 0, io.EOF
	}
	if err := p.conn.SetWriteDeadline(time.Now().Add(5 * time.Millisecond)); err != nil {
		return 0, err
	}

	return w3gs.SerializePacketWithBuffer(p.conn, &p.sbuf, pkt)
}
