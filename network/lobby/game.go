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
	"github.com/nielsAD/gowarcraft3/protocol"
	"github.com/nielsAD/gowarcraft3/protocol/w3gs"
)

// Game represents a game host
type Game struct {
	Lobby

	actmut   sync.Mutex
	actions  []w3gs.PlayerAction
	laggers  w3gs.StartLag
	dropmask protocol.BitSet32

	ackmut  sync.Mutex
	ackmask protocol.BitSet32
	ackarr  []plack

	// Atomic
	stage uint32
	tick  uint32

	// Set once before Run(), read-only after that
	LoadTimeout time.Duration
	TurnRate    int
}

type plack struct {
	p *Player
	a uint32
}

// NewGame initializes a new Game struct
func NewGame(encoding w3gs.Encoding, slotInfo w3gs.SlotInfo, mapInfo w3gs.MapCheck) *Game {
	var g = Game{
		Lobby:       *NewLobby(encoding, slotInfo, mapInfo),
		LoadTimeout: time.Minute * 2,
		TurnRate:    40,
	}

	g.InitDefaultHandlers()
	return &g
}

// Stage of game
func (g *Game) Stage() Stage {
	return Stage(atomic.LoadUint32(&g.stage))
}

// Tick counter
func (g *Game) Tick() Tick {
	return Tick(atomic.LoadUint32(&g.tick))
}

func (g *Game) swapStage(old Stage, new Stage) bool {
	if !atomic.CompareAndSwapUint32(&g.stage, uint32(old), uint32(new)) {
		return false
	}

	g.Fire(&StageChanged{
		Old: old,
		New: new,
	})
	return true
}

// actmut should be locked
func (g *Game) splitActions() ([]w3gs.PlayerAction, bool) {
	var total = 0
	for i := range g.actions {
		var size = len(g.actions[i].Data) + 3
		if total+size+7 > mtu {
			var res = g.actions[:i]
			g.actions = g.actions[i:]
			return res, true
		}
		total += size
	}

	var res = g.actions
	g.actions = g.actions[:0]
	return res, false
}

// EnqueueAction for next TimeSlot
func (g *Game) EnqueueAction(a *w3gs.PlayerAction) {
	g.actmut.Lock()

	var i = len(g.actions)
	if cap(g.actions) < i+1 {
		g.actions = append(g.actions, w3gs.PlayerAction{})
	} else {
		g.actions = g.actions[:i+1]
	}

	g.actions[i].PlayerID = a.PlayerID
	g.actions[i].Data = append(g.actions[i].Data[:0], a.Data...)

	g.actmut.Unlock()
}

// Start game
func (g *Game) Start() error {
	if !g.swapStage(StageLobby, StageLoading) {
		return ErrLocked
	}

	var wg sync.WaitGroup

	g.slotmut.Lock()
	g.locked = true

	for pid := range g.players {
		// capture player
		var p = g.players[pid]

		wg.Add(1)
		var timeout = time.AfterFunc(g.LoadTimeout, func() {
			p.Kick(w3gs.LeaveDisconnect)
		})
		p.Once(&w3gs.GameLoaded{}, func(ev *network.Event) {
			timeout.Stop()
			wg.Done()

			g.SendToAll(&w3gs.PlayerLoaded{
				PlayerID: p.PlayerInfo.PlayerID,
			})
		})
		// Do not leave the game until everyone is done loading to prevent desync
		p.Once(network.RunStop{}, func(ev *network.Event) {
			p.Fire(&w3gs.GameLoaded{})
			p.Fire(&w3gs.DropLaggers{})
			wg.Wait()
		})
	}

	g.sendToAll(&w3gs.CountDownStart{})
	g.sendToAll(&w3gs.CountDownEnd{})
	g.slotmut.Unlock()

	go func() {
		wg.Wait()
		if !g.swapStage(StageLoading, StagePlaying) {
			panic("lobby: Could not switch stage to Playing")
		}

		if g.TurnRate > 0 {
			g.gameloop()
		}

		if !g.swapStage(StagePlaying, StageDone) {
			panic("lobby: Could not switch stage to Done")
		}
	}()

	return nil
}

func (g *Game) gameloop() {
	var stop = make(chan struct{})
	go func() {
		g.Wait()
		stop <- struct{}{}
	}()

	var lastTick = time.Now()
	var interval = time.Second / time.Duration(g.TurnRate)
	var ticker = time.NewTicker(interval)

	var pkt w3gs.TimeSlot
	for {
		var inc time.Duration

		select {
		case <-stop:
			ticker.Stop()
			return
		case tick := <-ticker.C:
			inc = tick.Sub(lastTick)
			lastTick = tick
		}

		if inc < time.Millisecond {
			inc = time.Millisecond
		} else {
			inc = inc.Round(time.Millisecond)
		}

		// Do not send timeslot if players are lagging
		g.actmut.Lock()
		if len(g.laggers.Players) > 0 {
			g.incLaggers(uint32(inc.Milliseconds()))
			g.actmut.Unlock()
			continue
		}

		var newTick = atomic.AddUint32(&g.tick, 1)

		pkt.TimeIncrementMS = uint16(inc.Milliseconds())
		for send := true; send; send = len(g.actions) > 0 {
			pkt.Actions, pkt.Fragment = g.splitActions()
			g.SendToAll(&pkt)
		}

		g.actmut.Unlock()
		g.Fire(Tick(newTick))
	}
}

// actmut should be locked
func (g *Game) incLaggers(inc uint32) {
	if len(g.laggers.Players) > 0 && g.laggers.Players[0].LagDurationMS == 0 {
		g.dropmask = 0

		g.slotmut.Lock()
		for _, s := range g.slots {
			if s.SlotStatus != w3gs.SlotOccupied || s.Team == g.ObsTeam {
				continue
			}
			if _, ok := g.players[s.PlayerID]; ok {
				g.dropmask.Set(uint(s.PlayerID))
			}
		}
		g.slotmut.Unlock()

		for i := range g.laggers.Players {
			g.laggers.Players[i].LagDurationMS = 15_000
			g.dropmask.Clear(uint(g.laggers.Players[i].PlayerID))
		}

		g.SendToAll(&g.laggers)
		return
	}

	for i := range g.laggers.Players {
		if g.laggers.Players[i].LagDurationMS == math.MaxUint32 {
			// Already kicked
			continue
		}
		if g.laggers.Players[i].LagDurationMS == 0 {
			// New player requires new lag screen
			break
		}

		g.laggers.Players[i].LagDurationMS += inc

		if g.laggers.Players[i].LagDurationMS >= 180_000 ||
			(g.laggers.Players[i].LagDurationMS >= 25_000 && g.dropmask == 0) {
			g.laggers.Players[i].LagDurationMS = math.MaxUint32

			g.slotmut.Lock()
			if p, ok := g.players[g.laggers.Players[i].PlayerID]; ok {
				p.Fire(&network.AsyncError{Src: "incLaggers[LagDuration]", Err: ErrStraggling})
				p.Kick(w3gs.LeaveDisconnect)
			}
			g.slotmut.Unlock()
		}
	}

}

// ackmut should be locked
func (g *Game) drainAcks(pid uint) {
	// ackmask bit is set if pid is straggling
	if !g.ackmask.Test(pid) {
		return
	}

	g.ackmask.Clear(pid)
	if g.ackmask == 0 {
		g.slotmut.Lock()

		for g.ackmask == 0 {
			g.ackarr = g.ackarr[:0]

			var l = 0
			for id, p := range g.players {
				ack, ok, more := p.DequeueAck()
				if !ok {
					panic("lobby: Could not dequeue ack")
				}
				if !more {
					g.ackmask.Set(uint(id))
				}

				if cap(g.ackarr) < l+1 {
					g.ackarr = append(g.ackarr, plack{p, ack})
				} else {
					g.ackarr = g.ackarr[:l+1]
					g.ackarr[l].a = ack
					g.ackarr[l].p = p
				}
				l++
			}

			if l == 0 {
				break
			}

			for i := 1; i < l; i++ {
				if g.ackarr[0].a != g.ackarr[i].a {
					g.desync()
					break
				}
			}
		}

		g.slotmut.Unlock()
	}
}

// ackmut+slotmut should be locked
func (g *Game) desync() {
	var max uint32
	var m = map[uint32]int{}

	for _, ack := range g.ackarr {
		if g.slots[g.pidToSID(ack.p.PlayerInfo.PlayerID)].Team != g.ObsTeam {
			// Players count extra
			m[ack.a] += 0xFF
		} else {
			m[ack.a]++
		}

		if m[ack.a] > m[max] {
			max = ack.a
		}
	}

	for _, ack := range g.ackarr {
		if ack.a != max {
			ack.p.Fire(&network.AsyncError{Src: "desync", Err: ErrDesync})
			ack.p.Kick(w3gs.LeaveDisconnect)

			// Immediately delete from players to continue draining in drainAcks()
			delete(g.players, ack.p.PlayerInfo.PlayerID)
			g.ackmask.Clear(uint(ack.p.PlayerInfo.PlayerID))
		}
	}
}

// InitDefaultHandlers adds the default callbacks for relevant packets
func (g *Game) InitDefaultHandlers() {
	g.On(&PlayerJoined{}, g.onPlayerJoined)
	g.On(&PlayerLeft{}, g.onPlayerLeft)
}

func (g *Game) onPlayerJoined(ev *network.Event) {
	var p = ev.Arg.(*PlayerJoined).Player

	g.ackmut.Lock()
	g.ackmask.Set(uint(p.PlayerInfo.PlayerID))
	g.ackmut.Unlock()

	p.On(&w3gs.GameAction{}, func(ev *network.Event) {
		g.onGameAction(p, ev.Arg.(*w3gs.GameAction))
	})
	p.On(Tick(0), func(ev *network.Event) {
		g.onGameTick(p, ev.Arg.(Tick), ev.Opt[0].(int))
	})
	p.On(&StartLag{}, func(ev *network.Event) {
		g.onStartLag(p)
	})
	p.On(&StopLag{}, func(ev *network.Event) {
		g.onStopLag(p)
	})
	p.On(&w3gs.DropLaggers{}, func(ev *network.Event) {
		g.onDropLaggers(p)
	})
}

func (g *Game) onPlayerLeft(ev *network.Event) {
	var p = ev.Arg.(*PlayerLeft).Player

	g.ackmut.Lock()
	g.drainAcks(uint(p.PlayerInfo.PlayerID))
	g.ackmut.Unlock()
}

func (g *Game) onGameAction(p *Player, pkt *w3gs.GameAction) {
	if len(pkt.Data) > mtu-10 {
		p.Fire(&network.AsyncError{Src: "onGameAction[notReady]", Err: ErrInvalidPacket})
		p.Kick(w3gs.LeaveDisconnect)
		return
	}
	g.EnqueueAction(&w3gs.PlayerAction{
		PlayerID: p.PlayerInfo.PlayerID,
		Data:     pkt.Data,
	})
}

func (g *Game) onGameTick(p *Player, tick Tick, queue int) {
	// Player is sending an acknowledgement from the future
	if tick > g.Tick() {
		p.Fire(&network.AsyncError{Src: "onGameTick[Tick]", Err: ErrInvalidPacket})
		p.Kick(w3gs.LeaveDisconnect)
		return
	}

	g.ackmut.Lock()
	if queue >= 2000 || (g.TurnRate > 0 && time.Duration(queue)*time.Second/time.Duration(g.TurnRate) > 30*time.Second) {
		// Drop all stragglers, we are more than 30s ahead
		g.slotmut.Lock()
		for pid := uint8(1); pid <= 32; pid++ {
			if !g.ackmask.Test(uint(pid)) {
				continue
			}
			if pl, ok := g.players[pid]; ok {
				pl.Fire(&network.AsyncError{Src: "onGameTick[Pending]", Err: ErrStraggling})
				pl.Kick(w3gs.LeaveDisconnect)

				// Immediately delete from players to allow draining in drainAcks()
				delete(g.players, pid)
			}
		}
		g.slotmut.Unlock()
		g.ackmask = 0
	}

	g.drainAcks(uint(p.PlayerInfo.PlayerID))
	g.ackmut.Unlock()
}

func (g *Game) onStartLag(p *Player) {
	g.slotmut.Lock()
	var obs = g.slots[g.pidToSID(p.PlayerInfo.PlayerID)].Team == g.ObsTeam
	g.slotmut.Unlock()

	// Do not show lag screen for observers
	if obs {
		return
	}

	g.actmut.Lock()
	g.laggers.Players = append(g.laggers.Players, w3gs.LagPlayer{
		PlayerID:      p.PlayerInfo.PlayerID,
		LagDurationMS: 0,
	})
	g.actmut.Unlock()
}

func (g *Game) onStopLag(p *Player) {
	g.actmut.Lock()
	for i, lp := range g.laggers.Players {
		if lp.PlayerID != p.PlayerInfo.PlayerID {
			continue
		}

		// Lag screen is up if LagDurationMS > 0
		if lp.LagDurationMS > 0 {
			g.SendToAll(&w3gs.StopLag{
				LagPlayer: lp,
			})
		}

		g.laggers.Players = append(g.laggers.Players[:i], g.laggers.Players[i+1:]...)
	}
	g.actmut.Unlock()
}

func (g *Game) onDropLaggers(p *Player) {
	g.actmut.Lock()
	g.dropmask.Clear(uint(p.PlayerInfo.PlayerID))
	g.actmut.Unlock()
}
