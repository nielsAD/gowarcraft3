// Author:  Niels A.D.
// Project: gowarcraft3 (https://github.com/nielsAD/gowarcraft3)
// License: Mozilla Public License, v2.0

package lan

import (
	"net"
	"os"
	"sync"
	"time"

	"github.com/nielsAD/gowarcraft3/network"
	"github.com/nielsAD/gowarcraft3/protocol/w3gs"
)

// Update event
type Update struct{}

// GameList keeps track of all the hosted games in the Local Area Network
// Emits events for every received packet and Update{} when the output of Games() changes
// Public methods/fields are thread-safe unless explicitly stated otherwise
type GameList struct {
	network.EventEmitter
	network.W3GSPacketConn

	gmut  sync.Mutex
	games map[string]w3gs.GameInfo

	// Set once before Run(), read-only after that
	GameVersion       w3gs.GameVersion
	BroadcastInterval time.Duration
}

// NewGameList opens a new UDP socket to listen for LAN GameList updates
func NewGameList(gv w3gs.GameVersion, port int) (*GameList, error) {
	conn, err := net.ListenUDP("udp4", &net.UDPAddr{Port: port})
	if err != nil {
		return nil, err
	}

	var g = GameList{
		GameVersion:       gv,
		BroadcastInterval: 30 * time.Second,
	}

	g.InitDefaultHandlers()
	g.SetConn(conn)
	return &g, nil
}

// Games returns the current list of LAN games. Map key is the remote address.
func (g *GameList) Games() map[string]w3gs.GameInfo {
	var res = make(map[string]w3gs.GameInfo)

	g.gmut.Lock()
	for k, v := range g.games {
		if v.GameVersion == g.GameVersion {
			res[k] = v
		}
	}
	g.gmut.Unlock()

	return res
}

// Make sure gmut is locked before calling
func (g *GameList) initMap() {
	if g.games != nil {
		return
	}
	g.games = make(map[string]w3gs.GameInfo)
}

// InitDefaultHandlers adds the default callbacks for relevant packets
func (g *GameList) InitDefaultHandlers() {
	g.On(&w3gs.RefreshGame{}, g.onRefreshGame)
	g.On(&w3gs.CreateGame{}, g.onCreateGame)
	g.On(&w3gs.DecreateGame{}, g.onDecreateGame)
	g.On(&w3gs.GameInfo{}, g.onGameInfo)
}

// Run reads packets from Conn and emits an event for each received packet
// Not safe for concurrent invocation
func (g *GameList) Run() error {
	var sg = w3gs.SearchGame{
		GameVersion: g.GameVersion,
	}

	for {
		if _, err := g.Broadcast(&sg); err != nil {
			g.Fire(&network.AsyncError{Src: "Run[Broadcast]", Err: err})
		}

		if err := g.W3GSPacketConn.Run(&g.EventEmitter, g.BroadcastInterval); err != nil {
			if os.IsTimeout(err) {
				continue
			}
			if !network.IsConnClosedError(err) {
				g.Fire(&network.AsyncError{Src: "Run[W3GSUConn]", Err: err})
			}
			return err
		}
	}
}

func (g *GameList) onRefreshGame(ev *network.Event) {
	var pkt = ev.Arg.(*w3gs.RefreshGame)
	var adr = ev.Opt[0].(net.Addr)
	var idx = adr.String()

	g.gmut.Lock()
	if g.games[idx].HostCounter == pkt.HostCounter {
		g.gmut.Unlock()
		return
	}

	g.initMap()
	g.games[idx] = w3gs.GameInfo{
		HostCounter: pkt.HostCounter,
	}

	g.gmut.Unlock()

	var sg = w3gs.SearchGame{
		GameVersion: g.GameVersion,
		HostCounter: pkt.HostCounter,
	}

	if _, err := g.Send(adr, &sg); err != nil {
		g.Fire(&network.AsyncError{Src: "onRefreshGame[Send]", Err: err})
	}
}

func (g *GameList) onCreateGame(ev *network.Event) {
	var pkt = ev.Arg.(*w3gs.CreateGame)
	var adr = ev.Opt[0].(net.Addr)
	var idx = adr.String()

	g.gmut.Lock()
	g.initMap()
	g.games[idx] = w3gs.GameInfo{
		HostCounter: pkt.HostCounter,
	}
	g.gmut.Unlock()

	if pkt.GameVersion != g.GameVersion {
		return
	}

	var sg = w3gs.SearchGame{
		GameVersion: g.GameVersion,
		HostCounter: pkt.HostCounter,
	}

	if _, err := g.Send(adr, &sg); err != nil {
		g.Fire(&network.AsyncError{Src: "onCreateGame[Send]", Err: err})
	}
}

func (g *GameList) onDecreateGame(ev *network.Event) {
	var pkt = ev.Arg.(*w3gs.RefreshGame)
	var adr = ev.Opt[0].(net.Addr)
	var idx = adr.String()

	g.gmut.Lock()
	var update = g.games[idx].HostCounter == pkt.HostCounter
	if update {
		update = g.games[idx].GameVersion == g.GameVersion
		delete(g.games, idx)
	}
	g.gmut.Unlock()

	if update {
		g.Fire(Update{})
	}
}

func (g *GameList) onGameInfo(ev *network.Event) {
	var pkt = ev.Arg.(*w3gs.GameInfo)
	var adr = ev.Opt[0].(net.Addr)
	var idx = adr.String()

	var update = pkt.GameVersion == g.GameVersion

	g.gmut.Lock()
	update = update && g.games[idx] != *pkt
	g.initMap()
	g.games[idx] = *pkt
	g.gmut.Unlock()

	if update {
		g.Fire(Update{})
	}
}
