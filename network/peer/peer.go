// Author:  Niels A.D.
// Project: gowarcraft3 (https://github.com/nielsAD/gowarcraft3)
// License: Mozilla Public License, v2.0

// Package peer implements a mocked Warcraft III client that can be used to manage peer connections in lobbies.
package peer

import (
	"sync/atomic"
	"time"

	"github.com/nielsAD/gowarcraft3/network"
	"github.com/nielsAD/gowarcraft3/protocol"
	"github.com/nielsAD/gowarcraft3/protocol/w3gs"
)

// Player represents a (real) player in game
// Public methods/fields are thread-safe unless explicitly stated otherwise
type Player struct {
	network.EventEmitter
	network.W3GSConn

	// Atomic
	rtt     uint32
	peerset uint32

	// Set once before Run(), read-only after that
	PlayerInfo w3gs.PlayerInfo
	StartTime  time.Time
}

// NewPlayer initializes a new Player struct
func NewPlayer(info *w3gs.PlayerInfo) *Player {
	var p = Player{
		PlayerInfo: *info,
		StartTime:  time.Now(),
	}

	p.InitDefaultHandlers()
	p.SetWriteTimeout(time.Second)

	return &p
}

// RTT to host
func (p *Player) RTT() uint32 {
	return atomic.LoadUint32(&p.rtt)
}

// PeerSet of connected peers
func (p *Player) PeerSet() protocol.BitSet32 {
	return protocol.BitSet32(atomic.LoadUint32(&p.peerset))
}

// SendOrClose sends pkt to player, closes connection on failure
func (p *Player) SendOrClose(pkt w3gs.Packet) (int, error) {
	n, err := p.W3GSConn.Send(pkt)
	if err == nil || network.IsCloseError(err) {
		return n, nil
	}

	p.Close()
	return n, err
}

// Run reads packets and emits an event for each received packet
// Not safe for concurrent invocation
func (p *Player) Run() error {
	return p.W3GSConn.Run(&p.EventEmitter, 15*time.Second)
}

// InitDefaultHandlers adds the default callbacks for relevant packets
func (p *Player) InitDefaultHandlers() {
	p.On(&w3gs.PeerPing{}, p.onPing)
	p.On(&w3gs.PeerPong{}, p.onPong)
}

func (p *Player) onPing(ev *network.Event) {
	var pkt = ev.Arg.(*w3gs.PeerPing)

	atomic.StoreUint32(&p.peerset, uint32(pkt.PeerSet))

	if _, err := p.SendOrClose(&w3gs.PeerPong{Ping: w3gs.Ping{Payload: pkt.Payload}}); err != nil {
		p.Fire(&network.AsyncError{Src: "onPing[Send]", Err: err})
	}
}

func (p *Player) onPong(ev *network.Event) {
	var pkt = ev.Arg.(*w3gs.PeerPong)
	var rtt = uint32(time.Since(p.StartTime).Milliseconds()) - pkt.Payload

	atomic.StoreUint32(&p.rtt, rtt)
}
