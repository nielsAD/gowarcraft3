// Author:  Niels A.D.
// Project: gowarcraft3 (https://github.com/nielsAD/gowarcraft3)
// License: Mozilla Public License, v2.0

// Package lobby implements a mocked Warcraft III game client that can be used to host lobbies.
package lobby

import (
	"math"
	"math/bits"
	"math/rand"
	"net"
	"sync"
	"time"

	"github.com/nielsAD/gowarcraft3/network"
	"github.com/nielsAD/gowarcraft3/protocol"
	"github.com/nielsAD/gowarcraft3/protocol/w3gs"
)

// Lobby represents a mocked game lobby
// Public methods/fields are thread-safe unless explicitly stated otherwise
type Lobby struct {
	network.EventEmitter

	wg sync.WaitGroup

	slotmut  sync.Mutex
	slotBase w3gs.SlotInfo
	slots    []w3gs.SlotData
	players  map[uint8]*Player
	locked   bool

	// Set once before Run(), read-only after that
	w3gs.Encoder
	w3gs.MapCheck
	ObsTeam      uint8
	ColorSet     protocol.BitSet32
	ReadyTimeout time.Duration
	ShareAddr    bool
}

// NewLobby initializes a new Lobby struct
func NewLobby(encoding w3gs.Encoding, slotInfo w3gs.SlotInfo, mapInfo w3gs.MapCheck) *Lobby {
	var (
		obsteam uint8             = 12
		colors  protocol.BitSet32 = 0xFFF // First 12 bits
	)

	if encoding.GameVersion == 0 || encoding.GameVersion >= 29 {
		obsteam = 24
		colors = 0xFFFFFF // First 24 bits
	}

	return &Lobby{
		Encoder:      w3gs.Encoder{Encoding: encoding},
		MapCheck:     mapInfo,
		ObsTeam:      obsteam,
		ColorSet:     colors,
		ReadyTimeout: 10 * time.Second,

		slotBase: slotInfo,
		slots:    append([]w3gs.SlotData{}, slotInfo.Slots...),

		players: make(map[uint8]*Player),
	}
}

// slotmut should be locked
func (l *Lobby) pidToSID(pid uint8) int {
	for i, s := range l.slots {
		if s.PlayerID == pid {
			return i
		}
	}
	panic("lobby: Cannot not find slot with PlayerID")
}

// slotmut should be locked
func (l *Lobby) findEmptySlot() int {
	for i, s := range l.slots {
		if s.SlotStatus == w3gs.SlotOpen {
			return i
		}
	}
	return -1
}

// slotmut should be locked
func (l *Lobby) findEmptyTeamSlot(team uint8) int {
	for i, s := range l.slots {
		if s.SlotStatus == w3gs.SlotOpen && s.Team == team {
			return i
		}
	}
	return -1
}

// slotmut should be locked
func (l *Lobby) findEmptyPID() uint8 {
	var players protocol.BitSet32
	for uid := range l.players {
		players.Set(uint(uid))
	}
	players = ^players
	return (uint8)(bits.TrailingZeros32(uint32(players)) + 1)
}

// slotmut should be locked
func (l *Lobby) findEmptyTeam() uint8 {
	var teams protocol.BitSet32
	for _, s := range l.slots {
		if s.SlotStatus != w3gs.SlotOccupied || s.Team == l.ObsTeam {
			continue
		}
		teams.Set(uint(s.Team) + 1)
	}
	teams = ^teams
	return (uint8)(bits.TrailingZeros32(uint32(teams)))
}

func (l *Lobby) countPlayers() int {
	var c = 0
	for _, s := range l.slots {
		if s.SlotStatus == w3gs.SlotOccupied && s.Team != l.ObsTeam {
			c++
		}
	}
	return c
}

// slotmut should be locked
func (l *Lobby) availableColors() protocol.BitSet32 {
	var colors = l.ColorSet
	for _, s := range l.slots {
		if s.SlotStatus != w3gs.SlotOccupied || s.Team == l.ObsTeam {
			continue
		}
		colors.Clear(uint(s.Color) + 1)
	}
	return colors
}

// slotmut should be locked
func (l *Lobby) findEmptyColor() uint8 {
	return (uint8)(bits.TrailingZeros32((uint32)(l.availableColors())))
}

// slotmut should be locked
func (l *Lobby) initSlot(slot int) error {
	var sd = w3gs.SlotData{
		PlayerID:       255,
		DownloadStatus: 255,
		SlotStatus:     w3gs.SlotOccupied,
		Color:          l.slotBase.Slots[slot].Color,
		Team:           l.slotBase.Slots[slot].Team,
		Race:           l.slotBase.Slots[slot].Race,
		Handicap:       l.slotBase.Slots[slot].Handicap,
	}
	if l.slotBase.SlotLayout&w3gs.LayoutFixedPlayerSettings == 0 {
		sd.Color = l.findEmptyColor()
	}
	if l.slotBase.SlotLayout&w3gs.LayoutCustomForces == 0 {
		var t = l.ObsTeam
		if uint8(l.countPlayers()) < l.slotBase.NumPlayers {
			t = l.findEmptyTeam()
		}
		if t < l.slotBase.NumPlayers {
			sd.Team = t
		} else if l.ObsTeam == ObsDisabled {
			return ErrFull
		} else {
			sd.Team = l.ObsTeam
		}
	}

	l.slots[slot] = sd
	return nil
}

// slotmut should be locked
func (l *Lobby) sendToAll(pkt w3gs.Packet) {
	var b, err = l.Encoder.Serialize(pkt)
	if err != nil {
		l.Fire(&network.AsyncError{Src: "sendToAll[Serialize]", Err: err})
		return
	}

	for _, p := range l.players {
		if _, err := p.Write(b); err != nil && !network.IsCloseError(err) {
			p.Fire(&network.AsyncError{Src: "Lobby.sendToAll[Write]", Err: err})
			p.Close()
		}
	}
}

// slotmut should be locked
func (l *Lobby) slotInfo() *w3gs.SlotInfo {
	var slotInfo = l.slotBase
	slotInfo.Slots = l.slots

	return &slotInfo
}

// slotmut should be locked
func (l *Lobby) refreshSlots() {
	var s = l.slotInfo()

	if l.Fire(s) {
		// Do not relay if event.PreventNext()
		return
	}

	l.sendToAll(s)
}

// slotmut should be locked
func (l *Lobby) join(conn net.Conn, join *w3gs.Join) (*Player, error) {
	var p = NewPlayer(&w3gs.PlayerInfo{
		JoinCounter: join.JoinCounter,
		PlayerName:  join.PlayerName,
	})
	p.SetConn(conn, w3gs.NewFactoryCache(w3gs.DefaultFactory), l.Encoding)

	if l.locked {
		p.Send(&w3gs.RejectJoin{Reason: w3gs.RejectJoinStarted})
		return nil, ErrLocked
	}

	var sid = l.findEmptySlot()
	if sid < 0 {
		p.Send(&w3gs.RejectJoin{Reason: w3gs.RejectJoinFull})
		return nil, ErrFull
	}
	if err := l.initSlot(sid); err != nil {
		p.Send(&w3gs.RejectJoin{Reason: w3gs.RejectJoinFull})
		return nil, err
	}

	var pid = l.findEmptyPID()
	p.PlayerInfo.PlayerID = pid
	l.slots[sid].PlayerID = pid

	var slotInfo = w3gs.SlotInfoJoin{
		SlotInfo:     *l.slotInfo(),
		PlayerID:     pid,
		ExternalAddr: protocol.Addr(conn.RemoteAddr()),
	}
	slotInfo.Slots = l.slots

	if _, err := p.Send(&slotInfo); err != nil {
		p.Send(&w3gs.RejectJoin{Reason: w3gs.RejectJoinInvalid})
		return nil, err
	}
	if _, err := p.SendOrClose(&w3gs.Ping{}); err != nil {
		return nil, err
	}
	if _, err := p.SendOrClose(&l.MapCheck); err != nil {
		return nil, err
	}

	if l.ShareAddr {
		p.PlayerInfo.ExternalAddr = protocol.Addr(conn.RemoteAddr())
		p.PlayerInfo.InternalAddr = join.InternalAddr
		p.PlayerInfo.ExternalAddr.Port = join.ListenPort
		p.PlayerInfo.InternalAddr.Port = join.ListenPort
	}

	l.sendToAll(&p.PlayerInfo)
	l.sendToAll(&slotInfo.SlotInfo)

	for _, player := range l.players {
		// Use SendOrClose, but continue after error to avoid inconsistent slot states among players.
		// PlayerLeft event will be sent quickly after Join because of the closed socket.
		p.SendOrClose(&player.PlayerInfo)

		if l.Encoding.GameVersion >= 10032 {
			if tag := player.BattleTag(); tag != "" {
				p.SendOrClose(&w3gs.PlayerExtra{
					Type: w3gs.PlayerProfile,
					Profiles: []w3gs.PlayerDataProfile{w3gs.PlayerDataProfile{
						PlayerID:  uint32(player.PlayerInfo.PlayerID),
						BattleTag: player.BattleTag(),
					}},
				})
			}
			p.SendOrClose(&w3gs.PlayerExtra{
				Type: w3gs.PlayerSkins,
				Skins: []w3gs.PlayerDataSkins{w3gs.PlayerDataSkins{
					PlayerID: uint32(p.PlayerInfo.PlayerID),
				}},
			})
			p.SendOrClose(&w3gs.PlayerExtra{
				Type: w3gs.PlayerExtra5,
				Unknown5: []w3gs.PlayerData5{w3gs.PlayerData5{
					PlayerID: uint32(p.PlayerInfo.PlayerID),
				}},
			})
		}
	}

	l.players[pid] = p
	l.Fire(&slotInfo.SlotInfo)

	return p, nil
}

// slotmut should be locked
func (l *Lobby) swapSlots(slotA int, slotB int, swapTeams bool) {
	l.slots[slotA], l.slots[slotB] = l.slots[slotB], l.slots[slotA]

	// Swap back teams if CustomForces
	if !swapTeams {
		l.slots[slotA].Team, l.slots[slotB].Team = l.slots[slotB].Team, l.slots[slotA].Team
	}
	// Swap back color and handicap if FixedPlayerSettings
	if l.slotBase.SlotLayout&w3gs.LayoutFixedPlayerSettings != 0 {
		l.slots[slotA].Color, l.slots[slotB].Color = l.slots[slotB].Color, l.slots[slotA].Color
		l.slots[slotA].Handicap, l.slots[slotB].Handicap = l.slots[slotB].Handicap, l.slots[slotA].Handicap
	}
	// Swap back races if either slot did not have selectable race
	if l.slots[slotA].Race&w3gs.RaceSelectable == 0 || l.slots[slotB].Race&w3gs.RaceSelectable == 0 {
		l.slots[slotA].Race, l.slots[slotB].Race = l.slots[slotB].Race, l.slots[slotA].Race
	}
}

// slotmut should be locked
func (l *Lobby) changeSlotStatus(slot int, status w3gs.SlotStatus, kick bool) error {
	if status == w3gs.SlotOccupied {
		return ErrInvalidArgument
	}

	l.slotBase.Slots[slot].SlotStatus = status
	if l.slots[slot].SlotStatus != w3gs.SlotOccupied {
		l.slots[slot].SlotStatus = status
		return nil
	}

	if p, ok := l.players[l.slots[slot].PlayerID]; ok {
		if !kick {
			return ErrSlotOccupied
		}

		p.Kick(w3gs.LeaveLobby)
		return nil
	}

	// Reset computer slot
	l.slots[slot] = l.slotBase.Slots[slot]
	return nil
}

// slotmut should be locked
func (l *Lobby) changeRace(slot int, r w3gs.RacePref) error {
	if l.slots[slot].Race&w3gs.RaceSelectable == 0 || bits.OnesCount(uint(r&w3gs.RaceMask)) != 1 || r&w3gs.RaceDemon != 0 {
		return ErrInvalidArgument
	}

	l.slots[slot].Race = r | w3gs.RaceSelectable
	return nil
}

// slotmut should be locked
func (l *Lobby) changeTeam(slot int, t uint8) error {
	if l.slotBase.SlotLayout&w3gs.LayoutCustomForces != 0 {
		return ErrInvalidArgument
	}
	if (t >= l.slotBase.NumPlayers) && ((l.ObsTeam == ObsDisabled) || (t != l.ObsTeam)) {
		return ErrInvalidArgument
	}

	t, l.slots[slot].Team = l.slots[slot].Team, t
	if uint8(l.countPlayers()) > l.slotBase.NumPlayers {
		l.slots[slot].Team = t
		return ErrPlayersOccupied
	}

	return nil
}

// slotmut should be locked
func (l *Lobby) changeColor(slot int, c uint8) error {
	if c >= 32 {
		return ErrInvalidArgument
	} else if !l.availableColors().Test(uint(c) + 1) {
		return ErrColorOccupied
	}

	l.slots[slot].Color = c
	return nil
}

// slotmut should be locked
func (l *Lobby) changeHandicap(slot int, h uint8) error {
	if h > 100 {
		return ErrInvalidArgument
	}

	l.slots[slot].Handicap = h
	return nil
}

// slotmut should be locked
func (l *Lobby) changeComputer(slot int, ai w3gs.AI) error {
	if l.slots[slot].SlotStatus == w3gs.SlotOccupied {
		if !l.slots[slot].Computer {
			return ErrSlotOccupied
		}

		l.slots[slot].ComputerType = ai
		return nil
	}

	if err := l.initSlot(slot); err != nil {
		return err
	}
	if l.ObsTeam != ObsDisabled && l.slotBase.Slots[slot].Team == l.ObsTeam {
		l.slots[slot] = l.slotBase.Slots[slot]
		return ErrPlayersOccupied
	}

	l.slots[slot].DownloadStatus = 100
	l.slots[slot].Computer = true
	l.slots[slot].ComputerType = ai

	return nil
}

// Player returns Player for id
func (l *Lobby) Player(id uint8) *Player {
	l.slotmut.Lock()
	var player = l.players[id]
	l.slotmut.Unlock()

	return player
}

// Lock lobby, disabling joins and slot changes
func (l *Lobby) Lock() {
	l.slotmut.Lock()
	l.locked = true
	l.slotmut.Unlock()
}

// Unlock lobby, enabling joins and slot changes
func (l *Lobby) Unlock() {
	l.slotmut.Lock()
	l.locked = false
	l.slotmut.Unlock()
}

// Wait for all goroutines to finish
func (l *Lobby) Wait() {
	l.wg.Wait()
}

// Close closes all connections to players
func (l *Lobby) Close() {
	l.slotmut.Lock()
	for idx, p := range l.players {
		p.Close()
		delete(l.players, idx)
	}
	l.slotmut.Unlock()
}

// SlotInfo in current state
func (l *Lobby) SlotInfo() *w3gs.SlotInfo {
	var slotInfo = l.slotBase

	l.slotmut.Lock()
	slotInfo.Slots = append([]w3gs.SlotData{}, l.slots...)
	l.slotmut.Unlock()

	return &slotInfo
}

// CountPlayers that are not observing
func (l *Lobby) CountPlayers() int {
	l.slotmut.Lock()
	var c = l.countPlayers()
	l.slotmut.Unlock()
	return c
}

// SlotsUsed counts occupied+closed slots
func (l *Lobby) SlotsUsed() int {
	var c = 0
	l.slotmut.Lock()
	for _, s := range l.slots {
		if s.SlotStatus != w3gs.SlotOpen {
			c++
		}
	}
	l.slotmut.Unlock()
	return c
}

// SlotsAvailable counts open slots
func (l *Lobby) SlotsAvailable() int {
	var c = 0
	l.slotmut.Lock()
	for _, s := range l.slots {
		if s.SlotStatus == w3gs.SlotOpen {
			c++
		}
	}
	l.slotmut.Unlock()
	return c
}

// SendToAll players
func (l *Lobby) SendToAll(pkt w3gs.Packet) {
	l.slotmut.Lock()
	l.sendToAll(pkt)
	l.slotmut.Unlock()
}

// OpenAllSlots opens all closed slots
func (l *Lobby) OpenAllSlots() error {
	var refresh = false

	l.slotmut.Lock()
	if l.locked {
		l.slotmut.Unlock()
		return ErrLocked
	}

	for s := range l.slots {
		if l.slots[s].SlotStatus == w3gs.SlotClosed {
			l.slots[s].SlotStatus = w3gs.SlotOpen
			refresh = true
		}
	}
	if refresh {
		l.refreshSlots()
	}

	l.slotmut.Unlock()
	return nil
}

// CloseAllSlots closes all open slots
func (l *Lobby) CloseAllSlots() error {
	var refresh = false

	l.slotmut.Lock()
	if l.locked {
		l.slotmut.Unlock()
		return ErrLocked
	}

	for s := range l.slots {
		if l.slots[s].SlotStatus == w3gs.SlotOpen {
			l.slots[s].SlotStatus = w3gs.SlotClosed
			refresh = true
		}
	}
	if refresh {
		l.refreshSlots()
	}

	l.slotmut.Unlock()
	return nil
}

// OpenSlot by sid, optionally kick if occupied
func (l *Lobby) OpenSlot(sid int, kick bool) error {
	if sid < 0 || sid >= len(l.slotBase.Slots) {
		return ErrInvalidSlot
	}

	var err error

	l.slotmut.Lock()
	if l.locked {
		err = ErrLocked
	} else if err = l.changeSlotStatus(sid, w3gs.SlotOpen, kick); err == nil {
		l.refreshSlots()
	}
	l.slotmut.Unlock()

	return err
}

// CloseSlot by sid, optionally kick if occupied
func (l *Lobby) CloseSlot(sid int, kick bool) error {
	if sid < 0 || sid >= len(l.slotBase.Slots) {
		return ErrInvalidSlot
	}

	var err error

	l.slotmut.Lock()
	if l.locked {
		err = ErrLocked
	} else if err = l.changeSlotStatus(sid, w3gs.SlotClosed, kick); err == nil {
		l.refreshSlots()
	}
	l.slotmut.Unlock()

	return err
}

// SwapSlots slotA and slotB
func (l *Lobby) SwapSlots(slotA int, slotB int) error {
	if slotA < 0 || slotB < 0 || slotA >= len(l.slotBase.Slots) || slotB >= len(l.slotBase.Slots) {
		return ErrInvalidSlot
	}

	var err error

	l.slotmut.Lock()
	if l.locked {
		err = ErrLocked
	} else {
		l.swapSlots(slotA, slotB, l.slotBase.SlotLayout&w3gs.LayoutCustomForces == 0)
		l.refreshSlots()
	}
	l.slotmut.Unlock()

	return err
}

// ShuffleSlots randomizes slots and teams
func (l *Lobby) ShuffleSlots(shuffleTeams bool) error {
	l.slotmut.Lock()
	if l.locked {
		l.slotmut.Unlock()
		return ErrLocked
	}

	var customForces = l.slotBase.SlotLayout&w3gs.LayoutCustomForces != 0
	rand.Shuffle(len(l.slots), func(i, j int) {
		var occupied = l.slots[i].SlotStatus == w3gs.SlotOccupied && l.slots[j].SlotStatus == w3gs.SlotOccupied
		l.swapSlots(i, j, !customForces && occupied && !shuffleTeams)
	})
	l.refreshSlots()

	l.slotmut.Unlock()
	return nil
}

// ChangeRace to r for sid
func (l *Lobby) ChangeRace(sid int, r w3gs.RacePref) error {
	if sid < 0 || sid >= len(l.slotBase.Slots) {
		return ErrInvalidSlot
	}

	var err error

	l.slotmut.Lock()
	if l.locked {
		err = ErrLocked
	} else if l.slots[sid].SlotStatus != w3gs.SlotOccupied {
		err = ErrInvalidSlot
	} else if l.slots[sid].Race == r {
		// no action required
	} else if err = l.changeRace(sid, r); err == nil {
		l.refreshSlots()
	}
	l.slotmut.Unlock()

	return err
}

// ChangeTeam to t for sid
func (l *Lobby) ChangeTeam(sid int, t uint8) error {
	if sid < 0 || sid >= len(l.slotBase.Slots) {
		return ErrInvalidSlot
	}

	var err error

	l.slotmut.Lock()
	if l.locked {
		err = ErrLocked
	} else if l.slots[sid].SlotStatus != w3gs.SlotOccupied {
		err = ErrInvalidSlot
	} else if l.slots[sid].Team == t {
		// no action required
	} else if err = l.changeTeam(sid, t); err == nil {
		l.refreshSlots()
	}
	l.slotmut.Unlock()

	return err
}

// ChangeColor to c for sid
func (l *Lobby) ChangeColor(sid int, c uint8) error {
	if sid < 0 || sid >= len(l.slotBase.Slots) {
		return ErrInvalidSlot
	}

	var err error

	l.slotmut.Lock()
	if l.locked {
		err = ErrLocked
	} else if l.slots[sid].SlotStatus != w3gs.SlotOccupied {
		err = ErrInvalidSlot
	} else if l.slots[sid].Color == c {
		// no action required
	} else if err = l.changeColor(sid, c); err == nil {
		l.refreshSlots()
	}
	l.slotmut.Unlock()

	return err
}

// ChangeHandicap to h for sid
func (l *Lobby) ChangeHandicap(sid int, h uint8) error {
	if sid < 0 || sid >= len(l.slotBase.Slots) {
		return ErrInvalidSlot
	}

	var err error

	l.slotmut.Lock()
	if l.locked {
		err = ErrLocked
	} else if l.slots[sid].SlotStatus != w3gs.SlotOccupied {
		err = ErrInvalidSlot
	} else if l.slots[sid].Color == h {
		// no action required
	} else if err = l.changeHandicap(sid, h); err == nil {
		l.refreshSlots()
	}
	l.slotmut.Unlock()

	return err
}

// ChangeComputer to ai for sid
func (l *Lobby) ChangeComputer(sid int, ai w3gs.AI) error {
	if sid < 0 || sid >= len(l.slotBase.Slots) {
		return ErrInvalidSlot
	}

	var err error

	l.slotmut.Lock()
	if l.locked {
		err = ErrLocked
	} else if l.slots[sid].SlotStatus == w3gs.SlotOccupied && l.slots[sid].Computer && l.slots[sid].ComputerType == ai {
		// no action required
	} else if err = l.changeComputer(sid, ai); err == nil {
		l.refreshSlots()
	}
	l.slotmut.Unlock()

	return err
}

// JoinAndServe player connection
func (l *Lobby) JoinAndServe(conn net.Conn, join *w3gs.Join) (*Player, error) {
	l.slotmut.Lock()
	p, err := l.join(conn, join)
	l.slotmut.Unlock()

	if err != nil {
		conn.Close()
		return nil, err
	}

	var timeout = time.AfterFunc(l.ReadyTimeout, func() {
		p.Fire(&network.AsyncError{Src: "Lobby.JoinAndServe[ReadyTimeout]", Err: ErrNotReady})
		p.Kick(w3gs.LeaveLobby)
	})
	p.Once(&Ready{}, func(ev *network.Event) {
		timeout.Stop()
	})
	p.On(&w3gs.MapState{}, func(ev *network.Event) {
		l.onMapState(p, ev.Arg.(*w3gs.MapState))
	})
	p.On(&w3gs.Message{}, func(ev *network.Event) {
		l.onMessage(p, ev.Arg.(*w3gs.Message))
	})
	p.On(&w3gs.PlayerExtra{}, func(ev *network.Event) {
		l.onPlayerExtra(p, ev.Arg.(*w3gs.PlayerExtra))
	})

	l.wg.Add(1)
	go func() {
		l.Fire(&PlayerJoined{p})
		if err := p.Run(); err != nil && !network.IsCloseError(err) {
			p.Fire(&network.AsyncError{Src: "Lobby.JoinAndServe[Run]", Err: err})
		}

		timeout.Stop()
		l.onLeave(p)
	}()

	return p, nil
}

// Accept a new player connection
func (l *Lobby) Accept(conn net.Conn) (*Player, error) {
	var c = network.NewW3GSConn(conn, nil, l.Encoding)

	pkt, err := c.NextPacket(10 * time.Second)
	if err != nil {
		return nil, err
	}

	join, ok := pkt.(*w3gs.Join)
	if !ok {
		return nil, ErrInvalidPacket
	}

	return l.JoinAndServe(conn, join)
}

func (l *Lobby) onLeave(p *Player) {
	l.slotmut.Lock()

	var pid = p.PlayerInfo.PlayerID
	delete(l.players, pid)

	var sid = l.pidToSID(pid)
	l.slots[sid] = l.slotBase.Slots[sid]

	l.sendToAll(&w3gs.PlayerLeft{
		PlayerID: pid,
		Reason:   p.LeaveReason(),
	})
	l.refreshSlots()

	l.slotmut.Unlock()
	l.Fire(&PlayerLeft{p})
	l.wg.Done()

	p.Close()
}

func (l *Lobby) onPlayerExtra(p *Player, msg *w3gs.PlayerExtra) {
	if msg.Type != w3gs.PlayerProfile {
		return
	}

	if len(msg.Profiles) != 1 || msg.Profiles[0].PlayerID != uint32(p.PlayerInfo.PlayerID) {
		p.Fire(&network.AsyncError{Src: "Lobby.onPlayerExtra[Profiles]", Err: ErrInvalidPacket})
		p.Kick(w3gs.LeaveLobby)
	} else if p.BattleTag() != msg.Profiles[0].BattleTag {
		l.slotmut.Lock()
		if l.locked {
			p.Fire(&network.AsyncError{Src: "Lobby.onPlayerExtra[Locked]", Err: ErrLocked})
		} else {
			l.sendToAll(msg)
		}
		l.slotmut.Unlock()
	}
}

func (l *Lobby) onMapState(p *Player, s *w3gs.MapState) {
	var progress uint8 = 100
	if !s.Ready {
		progress = uint8(math.Min(100.0, math.Round(float64(s.FileSize)/float64(l.MapCheck.FileSize))*100.0))
	}

	l.slotmut.Lock()
	if l.slots[l.pidToSID(p.PlayerInfo.PlayerID)].DownloadStatus != progress {
		l.slots[l.pidToSID(p.PlayerInfo.PlayerID)].DownloadStatus = progress
		l.refreshSlots()
	}
	l.slotmut.Unlock()
}

func (l *Lobby) onMessage(p *Player, msg *w3gs.Message) {
	if msg.SenderID != p.PlayerInfo.PlayerID {
		p.Fire(&network.AsyncError{Src: "Lobby.onMessage[SenderID]", Err: ErrInvalidPacket})
		p.Kick(w3gs.LeaveLobby)
		return
	}

	switch msg.Type {
	case w3gs.MsgChat, w3gs.MsgChatExtra:
		l.onPlayerChat(p, msg)
	case w3gs.MsgTeamChange:
		l.onTeamChange(p, msg)
	case w3gs.MsgColorChange:
		l.onColorChange(p, msg)
	case w3gs.MsgRaceChange:
		l.onRaceChange(p, msg)
	case w3gs.MsgHandicapChange:
		l.onHandicapChange(p, msg)
	default:
		p.Fire(&network.AsyncError{Src: "Lobby.onMessage[Type]", Err: ErrInvalidPacket})
		p.Kick(w3gs.LeaveLobby)
	}
}

func (l *Lobby) onPlayerChat(p *Player, msg *w3gs.Message) {
	if len(msg.Content) > 254 {
		p.Fire(&network.AsyncError{Src: "Lobby.onPlayerChat[Content]", Err: ErrInvalidPacket})
		p.Kick(w3gs.LeaveLobby)
	}
	if l.Fire(&PlayerChat{p, msg}) {
		// Do not relay if event.PreventNext()
		return
	}

	var relay = w3gs.MessageRelay{
		Message: *msg,
	}

	l.slotmut.Lock()

	for _, rid := range msg.RecipientIDs {
		var recipient, ok = l.players[rid]
		if !ok {
			continue
		}

		if _, err := recipient.SendOrClose(&relay); err != nil {
			recipient.Fire(&network.AsyncError{Src: "Lobby.onPlayerChat[Relay]", Err: err})
		}
	}

	l.slotmut.Unlock()
}

func (l *Lobby) onTeamChange(p *Player, msg *w3gs.Message) {
	l.slotmut.Lock()

	if l.locked {
		p.Fire(&network.AsyncError{Src: "Lobby.onTeamChange[Locked]", Err: ErrLocked})
	} else if l.slotBase.SlotLayout&w3gs.LayoutCustomForces != 0 {
		var newSlot = l.findEmptyTeamSlot(msg.NewVal)
		if newSlot >= 0 {
			l.swapSlots(l.pidToSID(p.PlayerInfo.PlayerID), newSlot, false)
			l.refreshSlots()
		}
	} else if err := l.changeTeam(l.pidToSID(p.PlayerInfo.PlayerID), msg.NewVal); err == nil {
		l.refreshSlots()
	} else if err != ErrPlayersOccupied {
		p.Fire(&network.AsyncError{Src: "Lobby.onTeamChange[Change]", Err: err})
		p.Kick(w3gs.LeaveLobby)
	}

	l.slotmut.Unlock()
}

func (l *Lobby) onColorChange(p *Player, msg *w3gs.Message) {
	if l.slotBase.SlotLayout&w3gs.LayoutFixedPlayerSettings != 0 {
		p.Fire(&network.AsyncError{Src: "Lobby.onColorChange[FixedPlayerSettings]", Err: ErrInvalidPacket})
		p.Kick(w3gs.LeaveLobby)
		return
	}

	l.slotmut.Lock()
	if l.locked {
		p.Fire(&network.AsyncError{Src: "Lobby.onColorChange[Locked]", Err: ErrLocked})
	} else if err := l.changeColor(l.pidToSID(p.PlayerInfo.PlayerID), msg.NewVal); err == nil {
		l.refreshSlots()
	} else if err != ErrColorOccupied {
		p.Fire(&network.AsyncError{Src: "Lobby.onColorChange[Change]", Err: err})
		p.Kick(w3gs.LeaveLobby)
	}
	l.slotmut.Unlock()
}

func (l *Lobby) onRaceChange(p *Player, msg *w3gs.Message) {
	l.slotmut.Lock()
	if l.locked {
		p.Fire(&network.AsyncError{Src: "Lobby.onRaceChange[Locked]", Err: ErrLocked})
	} else if err := l.changeRace(l.pidToSID(p.PlayerInfo.PlayerID), (w3gs.RacePref)(msg.NewVal)); err == nil {
		l.refreshSlots()
	} else {
		p.Fire(&network.AsyncError{Src: "Lobby.onRaceChange[Change]", Err: err})
		p.Kick(w3gs.LeaveLobby)
	}
	l.slotmut.Unlock()
}

func (l *Lobby) onHandicapChange(p *Player, msg *w3gs.Message) {
	if l.slotBase.SlotLayout&w3gs.LayoutFixedPlayerSettings != 0 {
		p.Fire(&network.AsyncError{Src: "Lobby.onHandicapChange[FixedPlayerSettings]", Err: ErrInvalidPacket})
		p.Kick(w3gs.LeaveLobby)
		return
	}

	l.slotmut.Lock()
	if l.locked {
		p.Fire(&network.AsyncError{Src: "Lobby.onHandicapChange[Locked]", Err: ErrLocked})
	} else if err := l.changeHandicap(l.pidToSID(p.PlayerInfo.PlayerID), msg.NewVal); err == nil {
		l.refreshSlots()
	} else {
		p.Fire(&network.AsyncError{Src: "Lobby.onHandicapChange[Change]", Err: err})
		p.Kick(w3gs.LeaveLobby)
	}
	l.slotmut.Unlock()
}
