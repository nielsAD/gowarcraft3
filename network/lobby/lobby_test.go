// Author:  Niels A.D.
// Project: gowarcraft3 (https://github.com/nielsAD/gowarcraft3)
// License: Mozilla Public License, v2.0

package lobby_test

import (
	"fmt"
	"net"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/nielsAD/gowarcraft3/network"
	"github.com/nielsAD/gowarcraft3/network/dummy"
	"github.com/nielsAD/gowarcraft3/network/lobby"
	"github.com/nielsAD/gowarcraft3/network/peer"
	"github.com/nielsAD/gowarcraft3/protocol/w3gs"
)

// netPipe is analogous to net.Pipe, but it uses a real net.Conn, and
// therefore is buffered (net.Pipe deadlocks if both sides start with
// a write.)
// source: crypto/ssh/handshake_test.go
func netPipe() (net.Conn, net.Conn, error) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		listener, err = net.Listen("tcp", "[::1]:0")
		if err != nil {
			return nil, nil, err
		}
	}
	defer listener.Close()
	c1, err := net.Dial("tcp", listener.Addr().String())
	if err != nil {
		return nil, nil, err
	}

	c2, err := listener.Accept()
	if err != nil {
		c1.Close()
		return nil, nil, err
	}

	return c1, c2, nil
}

func makeSlots(n int) w3gs.SlotInfo {
	var s = w3gs.SlotInfo{
		RandomSeed: 123,
		SlotLayout: w3gs.LayoutMelee,
		NumPlayers: uint8(n),
	}
	for i := 0; i < n; i++ {
		s.Slots = append(s.Slots, w3gs.SlotData{
			SlotStatus: w3gs.SlotOpen,
			Race:       w3gs.RaceRandom | w3gs.RaceSelectable,
			Handicap:   100,
		})
	}
	return s
}

func makeGame(t *testing.T, n int) *lobby.Game {
	var g = lobby.NewGame(w3gs.Encoding{GameVersion: w3gs.CurrentGameVersion}, makeSlots(n), w3gs.MapCheck{})

	if t != nil {
		g.On(&network.AsyncError{}, func(ev *network.Event) {
			var err = ev.Arg.(*network.AsyncError)
			t.Logf("[ERROR][HOST] %s\n", err.Error())
		})
		g.On(&lobby.StageChanged{}, func(ev *network.Event) {
			var s = ev.Arg.(*lobby.StageChanged)
			t.Logf("[HOST] stage changed from '%s' to '%s'\n", s.Old.String(), s.New.String())
		})
		g.On(&lobby.PlayerJoined{}, func(ev *network.Event) {
			var p = ev.Arg.(*lobby.PlayerJoined).Player
			t.Logf("[HOST] player%d ('%s') joined\n", p.PlayerInfo.PlayerID, p.PlayerInfo.PlayerName)

			p.PingInterval = 5 * time.Millisecond

			p.On(&network.AsyncError{}, func(ev *network.Event) {
				var err = ev.Arg.(*network.AsyncError)
				t.Logf("[ERROR][HOST][PLAYER%d] %s\n", p.PlayerInfo.PlayerID, err.Error())
			})
		})
		g.On(&lobby.PlayerLeft{}, func(ev *network.Event) {
			var p = ev.Arg.(*lobby.PlayerLeft).Player
			t.Logf("[HOST] player%d ('%s') left\n", p.PlayerInfo.PlayerID, p.PlayerInfo.PlayerName)
		})
	}

	return g
}

func joinDummy(t *testing.T, g *lobby.Game, name string) (*dummy.Player, error) {
	var p = dummy.Player{
		Host: peer.Host{
			PlayerInfo: w3gs.PlayerInfo{PlayerName: name},
			Encoding:   g.Encoding,
		},
	}
	p.InitDefaultHandlers()

	p.On(&network.AsyncError{}, func(ev *network.Event) {
		var err = ev.Arg.(*network.AsyncError)
		if network.IsCloseError(err) || network.IsUnexpectedCloseError(err) {
			return
		}
		if t != nil {
			t.Logf("[ERROR][DUMMY][%s] %s\n", name, err.Error())
		}
	})

	c1, c2, err := netPipe()
	if err != nil {
		return nil, err
	}

	ch := make(chan error)
	go func() {
		defer p.Close()

		if err := p.JoinWithConn(c1); err != nil {
			ch <- err
			return
		}

		ch <- nil
		p.Run()
	}()

	if _, err := g.Accept(c2); err != nil {
		return nil, err
	}

	return &p, <-ch
}

func TestJoin1(t *testing.T) {
	var g = makeGame(t, 2)

	if g.SlotsUsed() != 0 || g.SlotsAvailable() != 2 {
		t.Fatal("Expected 0 slots to be used")
	}

	d, err := joinDummy(t, g, "DUMMY1")
	if err != nil {
		t.Fatalf("Could not join game with dummy: %s\n", err.Error())
	}

	if g.SlotsUsed() != 1 || g.SlotsAvailable() != 1 {
		t.Fatal("Expected 1 slot to be used")
	}

	var si = g.SlotInfo()
	if si == nil {
		t.Fatal("Expected SlotInfo != nil")
	}
	if si.Slots[0].SlotStatus != w3gs.SlotOccupied {
		t.Fatal("Expected Slots[0] to be occupied")
	}
	if si.Slots[0].PlayerID != d.PlayerInfo.PlayerID {
		t.Fatal("Expected Slots[0].PlayerID to be dummy.PlayerID")
	}
	if si.Slots[0].Race != w3gs.RaceRandom|w3gs.RaceSelectable {
		t.Fatal("Expected Slots[0].Race to be random")
	}
	if si.Slots[0].Team != 0 {
		t.Fatal("Expected Slots[0].Team to be 0")
	}
	if si.Slots[0].Color != 0 {
		t.Fatal("Expected Slots[0].Color to be 0")
	}
	if si.Slots[0].Handicap != 100 {
		t.Fatal("Expected Slots[0].Handicap to be 100")
	}

	if err := d.ChangeRace(w3gs.RaceHuman); err != nil {
		t.Fatalf("Could not change race: %s\n", err.Error())
	}
	if err := d.ChangeTeam(1); err != nil {
		t.Fatalf("Could not change team: %s\n", err.Error())
	}
	if err := d.ChangeColor(10); err != nil {
		t.Fatalf("Could not change color: %s\n", err.Error())
	}
	if err := d.ChangeHandicap(42); err != nil {
		t.Fatalf("Could not change handicap: %s\n", err.Error())
	}

	time.Sleep(50 * time.Millisecond)
	si = g.SlotInfo()

	if si.Slots[0].Race != w3gs.RaceHuman|w3gs.RaceSelectable {
		t.Fatal("Expected Slots[0].Race to be human")
	}
	if si.Slots[0].Team != 1 {
		t.Fatal("Expected Slots[0].Team to be 1")
	}
	if si.Slots[0].Color != 10 {
		t.Fatal("Expected Slots[0].Color to be 10")
	}
	if si.Slots[0].Handicap != 42 {
		t.Fatal("Expected Slots[0].Handicap to be 42")
	}

	d.Leave(w3gs.LeaveLobby)
	g.Wait()

	if g.SlotsUsed() != 0 || g.SlotsAvailable() != 2 {
		t.Fatal("Expected 0 slots to be used again")
	}
}

func TestComputer(t *testing.T) {
	var g = makeGame(t, 24)

	if err := g.ChangeComputer(0, w3gs.ComputerNormal); err != nil {
		t.Fatalf("Could not change computer: %s\n", err.Error())
	}
	if err := g.ChangeRace(0, w3gs.RaceOrc); err != nil {
		t.Fatalf("Could not change race: %s\n", err.Error())
	}
	if err := g.ChangeTeam(0, 1); err != nil {
		t.Fatalf("Could not change team: %s\n", err.Error())
	}
	if err := g.ChangeColor(0, 10); err != nil {
		t.Fatalf("Could not change color: %s\n", err.Error())
	}
	if err := g.ChangeHandicap(0, 42); err != nil {
		t.Fatalf("Could not change handicap: %s\n", err.Error())
	}

	if err := g.ChangeRace(1, w3gs.RaceOrc); err != lobby.ErrInvalidSlot {
		t.Fatalf("Expected ErrInvalidSlot changing race, got %s\n", err.Error())
	}
	if err := g.ChangeTeam(1, 1); err != lobby.ErrInvalidSlot {
		t.Fatalf("Expected ErrInvalidSlot changing team, got %s\n", err.Error())
	}
	if err := g.ChangeColor(1, 10); err != lobby.ErrInvalidSlot {
		t.Fatalf("Expected ErrInvalidSlot changing color, got %s\n", err.Error())
	}
	if err := g.ChangeHandicap(1, 42); err != lobby.ErrInvalidSlot {
		t.Fatalf("Expected ErrInvalidSlot changing handicap, got %s\n", err.Error())
	}

	var si = g.SlotInfo()
	if si.Slots[0].SlotStatus != w3gs.SlotOccupied || !si.Slots[0].Computer || si.Slots[0].ComputerType != w3gs.ComputerNormal {
		t.Fatal("Expected Slots[0].ComputerType to be ComputerNormal")
	}
	if si.Slots[0].Race != w3gs.RaceOrc|w3gs.RaceSelectable {
		t.Fatal("Expected Slots[0].Race to be human")
	}
	if si.Slots[0].Team != 1 {
		t.Fatal("Expected Slots[0].Team to be 1")
	}
	if si.Slots[0].Color != 10 {
		t.Fatal("Expected Slots[0].Color to be 10")
	}
	if si.Slots[0].Handicap != 42 {
		t.Fatal("Expected Slots[0].Handicap to be 42")
	}

	g.Lock()
	if err := g.ChangeComputer(0, w3gs.ComputerNormal); err != lobby.ErrLocked {
		t.Fatalf("Expected ErrLocked changing computer, got %s\n", err.Error())
	}
	if err := g.ChangeRace(0, w3gs.RaceOrc); err != lobby.ErrLocked {
		t.Fatalf("Expected ErrLocked changing race, got %s\n", err.Error())
	}
	if err := g.ChangeTeam(0, 1); err != lobby.ErrLocked {
		t.Fatalf("Expected ErrLocked changing team, got %s\n", err.Error())
	}
	if err := g.ChangeColor(0, 10); err != lobby.ErrLocked {
		t.Fatalf("Expected ErrLocked changing color, got %s\n", err.Error())
	}
	if err := g.ChangeHandicap(0, 42); err != lobby.ErrLocked {
		t.Fatalf("Expected ErrLocked changing handicap, got %s\n", err.Error())
	}

	if _, err := joinDummy(t, g, "DUMMY1"); err != lobby.ErrLocked {
		t.Fatalf("Expected ErrLocked when joining, got %s\n", err.Error())
	}
	g.Unlock()

	g.CloseAllSlots()
	if g.SlotsUsed() != 24 {
		t.Fatal("Expected 24 slots to be used")
	}
	if _, err := joinDummy(t, g, "DUMMY2"); err != lobby.ErrFull {
		t.Fatalf("Expected ErrGameFull when joining, got %s\n", err.Error())
	}

	g.OpenAllSlots()
	if g.SlotsUsed() != 1 {
		t.Fatal("Expected 1 slot to be used")
	}
}

func TestJoin24(t *testing.T) {
	var g = makeGame(t, 24)

	var wg sync.WaitGroup
	g.On(&lobby.PlayerChat{}, func(ev *network.Event) {
		var chat = ev.Arg.(*lobby.PlayerChat)
		if chat.Content == "Hi" {
			wg.Done()
		}
	})

	wg.Add(24)
	for i := 1; i <= 24; i++ {
		var idx = i
		go func() {
			p, err := joinDummy(t, g, fmt.Sprintf("DUMMY%d", idx))
			if err != nil {
				t.Logf("Could not join game with dummy%d: %s\n", idx, err.Error())
				return
			}

			// Process incoming packets
			time.Sleep(5 * time.Millisecond)

			p.ChangeHandicap(75)
			p.Say("Hi")
		}()
	}

	wg.Wait()
	if g.SlotsUsed() != 24 {
		t.Fatal("Expected 24 slots to be used")
	}

	var id = map[uint8]bool{}
	var col = map[uint8]bool{}
	var team = map[uint8]bool{}
	for i, s := range g.SlotInfo().Slots {
		if !g.Player(s.PlayerID).Ready() {
			t.Fatal("Expected player to be ready")
		}
		if s.Handicap != 75 {
			t.Fatalf("Expected Slots[%d].Handicap to be 75\n", i)
		}
		if id[s.PlayerID] {
			t.Fatal("Duplicate player ID")
		}
		if col[s.Color] {
			t.Fatal("Duplicate color")
		}
		if team[s.Team] {
			t.Fatal("Duplicate team")
		}
		id[s.PlayerID] = true
		col[s.Color] = true
		team[s.Team] = true
	}

	g.Close()
	g.Wait()
}

func TestInvalidPackets(t *testing.T) {
	var g = makeGame(t, 9)
	var d [9]*dummy.Player

	var wg sync.WaitGroup
	wg.Add(9)
	g.On(&lobby.PlayerJoined{}, func(ev *network.Event) { wg.Done() })

	for i := 1; i <= 9; i++ {
		p, err := joinDummy(t, g, fmt.Sprintf("DUMMY%d", i))
		if err != nil {
			t.Fatalf("Could not join game with dummy%d: %s\n", i, err.Error())
			return
		}
		d[i-1] = p
	}

	wg.Wait()
	if g.SlotsUsed() != 9 {
		t.Fatal("Expected 9 slots to be used")
	}

	wg.Add(9)
	g.On(&lobby.PlayerLeft{}, func(ev *network.Event) { wg.Done() })

	// All these actions should get the player kicked
	d[0].ChangeColor(200)
	d[1].ChangeHandicap(200)
	d[2].ChangeTeam(200)
	d[3].ChangeRace(w3gs.RaceHuman | w3gs.RaceOrc)
	d[4].ChangeRace(w3gs.RaceDemon)
	d[5].Say(strings.Repeat("LongString", 50))
	g.OpenSlot(6, true)
	g.CloseSlot(7, true)
	g.Player(9).Kick(w3gs.LeaveLost)
	wg.Wait()

	// Closed slot counts as used
	if g.SlotsUsed() != 1 {
		t.Fatal("Expected all players to be kicked")
	}
}
