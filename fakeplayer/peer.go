// Author:  Niels A.D.
// Project: gowarcraft3 (https://github.com/nielsAD/gowarcraft3)
// License: Mozilla Public License, v2.0

package fakeplayer

import (
	"io"
	"net"
	"sync"
	"time"

	"github.com/nielsAD/gowarcraft3/protocol"
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

	StartTime time.Time
	RTT       uint32
	PeerSet   protocol.BitSet32

	//goroutine event handlers
	OnError  func(err error)
	OnPacket func(pkt w3gs.Packet) bool
}

// NextPacket waits for the next packet from Peer (with given timeout) and returns its deserialized representation
func (p *Peer) NextPacket(timeout time.Duration) (w3gs.Packet, error) {
	var c = p.conn
	if c == nil {
		return nil, io.EOF
	}

	if timeout != 0 {
		if err := c.SetReadDeadline(time.Now().Add(timeout)); err != nil {
			return nil, err
		}
	}

	pkt, _, err := w3gs.DeserializePacketWithBuffer(c, &p.rbuf)
	return pkt, err
}

// Send serializes packet and sends it to Peer
func (p *Peer) Send(pkt w3gs.Packet) (int, error) {
	p.smutex.Lock()
	defer p.smutex.Unlock()

	var c = p.conn
	if c == nil {
		return 0, io.EOF
	}
	if err := c.SetWriteDeadline(time.Now().Add(5 * time.Millisecond)); err != nil {
		return 0, err
	}

	return w3gs.SerializePacketWithBuffer(c, &p.sbuf, pkt)
}

func (p *Peer) onError(err error) {
	if p.OnError != nil {
		p.OnError(err)
	}
}

// onPacket must return true if packet is handled, false to continue with default handler
func (p *Peer) onPacket(pkt w3gs.Packet) bool {
	if p.OnPacket == nil {
		return false
	}
	return p.OnPacket(pkt)
}
