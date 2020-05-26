// Author:  Niels A.D.
// Project: gowarcraft3 (https://github.com/nielsAD/gowarcraft3)
// License: Mozilla Public License, v2.0

package lobby

import (
	"math"
	"sync"
	"sync/atomic"
	"time"

	"github.com/nielsAD/gowarcraft3/network"
	"github.com/nielsAD/gowarcraft3/protocol/w3gs"
)

// Player represents a (real) player in game
// Public methods/fields are thread-safe unless explicitly stated otherwise
type Player struct {
	network.EventEmitter
	network.W3GSConn

	// Atomic
	tick  uint32
	rtt   uint32
	ready uint32
	leave uint32
	lag   uint32
	tag   atomic.Value //string

	ackmut sync.Mutex
	ackarr [2048]uint32
	ackidx int
	acklen int

	// Set once before Run(), read-only after that
	PlayerInfo   w3gs.PlayerInfo
	StartTime    time.Time
	PingInterval time.Duration
}

// NewPlayer initializes a new Player struct
func NewPlayer(info *w3gs.PlayerInfo) *Player {
	var p = Player{
		PlayerInfo:   *info,
		StartTime:    time.Now(),
		PingInterval: 5 * time.Second,

		rtt: math.MaxUint32,
	}

	p.InitDefaultHandlers()
	p.SetWriteTimeout(time.Second)

	return &p
}

// Tick counter
func (p *Player) Tick() Tick {
	return Tick(atomic.LoadUint32(&p.tick))
}

// RTT to host
func (p *Player) RTT() uint32 {
	return atomic.LoadUint32(&p.rtt)
}

// Ready to start the game (ping and map packets received)
func (p *Player) Ready() bool {
	return atomic.LoadUint32(&p.ready) != 0 && p.RTT() != math.MaxUint32
}

// LeaveReason from lobby
func (p *Player) LeaveReason() w3gs.LeaveReason {
	var reason = (w3gs.LeaveReason)(atomic.LoadUint32(&p.leave))
	if reason == 0 {
		reason = w3gs.LeaveDisconnect
	}
	return reason
}

func (p *Player) setLeaveReason(reason w3gs.LeaveReason) {
	atomic.CompareAndSwapUint32(&p.leave, 0, (uint32)(reason))
}

// Lag in receiving packets
func (p *Player) Lag() bool {
	return atomic.LoadUint32(&p.lag) != 0
}

func (p *Player) setLag(lag bool) bool {
	var l uint32
	if lag {
		l = ^l
	}
	if !atomic.CompareAndSwapUint32(&p.lag, ^l, l) {
		return false
	}
	if lag {
		p.Fire(&StartLag{})
	} else {
		p.Fire(&StopLag{})
	}
	return true
}

// BattleTag for player
func (p *Player) BattleTag() string {
	if t := p.tag.Load(); t != nil {
		return t.(string)
	}
	return ""
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

// Kick from lobby
func (p *Player) Kick(reason w3gs.LeaveReason) {
	p.setLeaveReason(reason)
	if reason != w3gs.LeaveDisconnect {
		p.Send(&w3gs.PlayerKicked{Leave: w3gs.Leave{
			Reason: reason,
		}})
	}
	p.Close()
}

// DequeueAck from queue
func (p *Player) DequeueAck() (checksum uint32, ok bool, more bool) {
	p.ackmut.Lock()
	if ok = p.acklen > 0; ok {
		checksum = p.ackarr[p.ackidx]
		p.acklen--
		p.ackidx++
		if p.ackidx >= len(p.ackarr) {
			p.ackidx = 0
		}
		more = p.acklen > 0
	}
	p.ackmut.Unlock()

	return checksum, ok, more
}

func (p *Player) runPing() func() {
	var stop = make(chan struct{})

	go func() {
		var pong = make(chan uint32, 8)
		p.On(&w3gs.Pong{}, func(ev *network.Event) {
			select {
			case pong <- ev.Arg.(*w3gs.Pong).Payload:
				// Sent to channel
			default:
				// Ignore full buffer
			}
		})

		var delay = LagDelay
		var timeout = time.NewTimer(time.Hour)
		if !timeout.Stop() {
			<-timeout.C
		}

		var ticker = time.NewTicker(p.PingInterval)

		var ping w3gs.Ping
		for {
			select {
			case <-stop:
				ticker.Stop()
				return
			case <-pong:
				// Leftover message in channel buffer
			case c := <-ticker.C:
				ping.Payload = uint32(c.Sub(p.StartTime).Milliseconds())
				if _, err := p.SendOrClose(&ping); err != nil {
					p.Fire(&network.AsyncError{Src: "runPing[Send]", Err: err})
					break
				}

				timeout.Reset(delay)
				var lagging = false

				for {
					select {
					case <-stop:
						ticker.Stop()
						return
					case payload := <-pong:
						// Check if this is the pong we are waiting for
						if payload != ping.Payload {
							continue
						}
						// Prepare timeout for .Reset()
						if !timeout.Stop() && !lagging {
							<-timeout.C
						}
						// Stop lagging
						if !lagging {
							delay = LagDelay
							p.setLag(false)
						}
					case <-timeout.C:
						// Response timeout, start lagging
						delay = LagRecoverDelay
						lagging = true
						p.setLag(true)
						continue
					}
					break
				}
			}
		}
	}()

	return func() {
		stop <- struct{}{}
		p.setLag(false)
	}
}

// Run reads packets and emits an event for each received packet
// Not safe for concurrent invocation
func (p *Player) Run() error {
	if p.PingInterval != 0 {
		var stop = p.runPing()
		defer stop()
	}

	return p.W3GSConn.Run(&p.EventEmitter, time.Minute)
}

// InitDefaultHandlers adds the default callbacks for relevant packets
func (p *Player) InitDefaultHandlers() {
	p.On(&w3gs.Pong{}, p.onPong)
	p.On(&w3gs.Leave{}, p.onLeave)
	p.On(&w3gs.PlayerExtra{}, p.onPlayerExtra)
	p.On(&w3gs.MapState{}, p.onMapState)
	p.On(&w3gs.StartDownload{}, p.onStartDownload)
	p.On(&w3gs.TimeSlotAck{}, p.onTimeSlotAck)
}

func (p *Player) onPong(ev *network.Event) {
	var pkt = ev.Arg.(*w3gs.Pong)
	var rtt = uint32(time.Now().Sub(p.StartTime).Milliseconds()) - pkt.Payload

	if rtt > uint32(LagRecoverDelay.Milliseconds()) {
		p.Fire(&network.AsyncError{Src: "onPong[rtt]", Err: ErrHighPing})
		return
	}

	if atomic.CompareAndSwapUint32(&p.rtt, math.MaxUint32, rtt) {
		if atomic.LoadUint32(&p.ready) != 0 {
			p.Fire(&Ready{})
		}
	} else {
		atomic.StoreUint32(&p.rtt, rtt)
	}
}

func (p *Player) onLeave(ev *network.Event) {
	var pkt = ev.Arg.(*w3gs.Leave)
	p.setLeaveReason(pkt.Reason)

	p.Send(&w3gs.LeaveAck{})
	p.Close()
}

func (p *Player) onPlayerExtra(ev *network.Event) {
	var pkt = ev.Arg.(*w3gs.PlayerExtra)
	if pkt.Type != w3gs.PlayerProfile {
		return
	}

	for _, pf := range pkt.Profiles {
		if pf.PlayerID != uint32(p.PlayerInfo.PlayerID) {
			continue
		}

		p.tag.Store(pf.BattleTag)
		break
	}
}

func (p *Player) onMapState(ev *network.Event) {
	var s = ev.Arg.(*w3gs.MapState)
	if !s.Ready {
		p.Fire(&network.AsyncError{Src: "onMapState[notReady]", Err: ErrMapUnavailable})
		p.Kick(w3gs.LeaveLobby)
		return
	}

	if atomic.CompareAndSwapUint32(&p.ready, 0, ^uint32(0)) {
		if p.RTT() != math.MaxUint32 {
			p.Fire(&Ready{})
		}
	}
}

func (p *Player) onStartDownload(ev *network.Event) {
	p.Fire(&network.AsyncError{Src: "onStartDownload", Err: ErrMapUnavailable})
	p.Kick(w3gs.LeaveLobby)
}

func (p *Player) onTimeSlotAck(ev *network.Event) {
	var pkt = ev.Arg.(*w3gs.TimeSlotAck)

	p.ackmut.Lock()
	var l = p.acklen
	if l >= len(p.ackarr) {
		p.Fire(&network.AsyncError{Src: "onTimeSlotAck[Len]", Err: ErrInvalidPacket})
		p.Kick(w3gs.LeaveDisconnect)
		l = -1
	} else {
		p.ackarr[(p.ackidx+p.acklen)%len(p.ackarr)] = pkt.Checksum
		p.acklen++
	}
	p.ackmut.Unlock()

	if l >= 0 {
		var t = atomic.AddUint32(&p.tick, 1)
		p.Fire(Tick(t), l+1)
	}
}
