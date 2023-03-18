// Author:  Niels A.D.
// Project: gowarcraft3 (https://github.com/nielsAD/gowarcraft3)
// License: Mozilla Public License, v2.0

package lan

import (
	"fmt"
	"net"
	"sync"
	"time"

	"github.com/nielsAD/gowarcraft3/network"
	"github.com/nielsAD/gowarcraft3/protocol/w3gs"
)

type udpIndex struct {
	source string
	gameID uint32
}

type udpRecord struct {
	w3gs.GameInfo
	created time.Time
	expires time.Time
}

// UDPGameList keeps track of all the hosted games in the Local Area Network using UDP broadcast
// Emits events for every received packet and Update{} when the output of Games() changes
// Public methods/fields are thread-safe unless explicitly stated otherwise
type UDPGameList struct {
	network.EventEmitter
	network.W3GSPacketConn

	gmut  sync.Mutex
	games map[udpIndex]*udpRecord

	// Set once before Run(), read-only after that
	GameVersion       w3gs.GameVersion
	BroadcastInterval time.Duration
}

// NewUDPGameList opens a new UDP socket to listen for LAN GameList updates
func NewUDPGameList(gv w3gs.GameVersion, port int) (*UDPGameList, error) {
	conn, err := net.ListenUDP("udp4", &net.UDPAddr{Port: port})
	if err != nil {
		return nil, err
	}

	var g = UDPGameList{
		GameVersion:       gv,
		BroadcastInterval: 15 * time.Second,
	}

	g.InitDefaultHandlers()
	g.SetWriteTimeout(time.Second)
	g.SetConn(conn, w3gs.NewFactoryCache(w3gs.DefaultFactory), g.Encoding())

	return &g, nil
}

// Encoding for w3gs packets
func (g *UDPGameList) Encoding() w3gs.Encoding {
	return w3gs.Encoding{
		GameVersion: g.GameVersion.Version,
	}
}

// Games returns the current list of LAN games. Map key is the remote address.
func (g *UDPGameList) Games() map[string]w3gs.GameInfo {
	var res = make(map[string]w3gs.GameInfo)
	var now = time.Now()

	g.gmut.Lock()
	for k, v := range g.games {
		if v.GameVersion != g.GameVersion || v.GamePort == 0 {
			continue
		}

		var host = k.source
		if h, _, err := net.SplitHostPort(host); err == nil {
			host = fmt.Sprintf("%s:%d", h, v.GamePort)
		}

		v.UptimeSec = (uint32)(now.Sub(v.created).Seconds())
		if r, ok := res[host]; ok && r.UptimeSec < v.UptimeSec {
			continue
		}

		res[host] = v.GameInfo
	}
	g.gmut.Unlock()

	return res
}

// Make sure gmut is locked before calling
func (g *UDPGameList) initMap() {
	if g.games != nil {
		return
	}
	g.games = make(map[udpIndex]*udpRecord)
}

func (g *UDPGameList) expire() {
	var now = time.Now()
	var update = false

	g.gmut.Lock()
	for idx, game := range g.games {
		if now.After(game.expires) {
			update = true
			delete(g.games, idx)
		}
	}
	g.gmut.Unlock()

	if update {
		g.Fire(Update{})
	}
}

// InitDefaultHandlers adds the default callbacks for relevant packets
func (g *UDPGameList) InitDefaultHandlers() {
	g.On(&w3gs.RefreshGame{}, g.onRefreshGame)
	g.On(&w3gs.CreateGame{}, g.onCreateGame)
	g.On(&w3gs.DecreateGame{}, g.onDecreateGame)
	g.On(&w3gs.GameInfo{}, g.onGameInfo)
}

func (g *UDPGameList) runSearch(sg *w3gs.SearchGame) func() {
	var stop = make(chan struct{})

	go func() {
		var ticker = time.NewTicker(g.BroadcastInterval)
		for {
			select {
			case <-stop:
				ticker.Stop()
				return
			case <-ticker.C:
				if _, err := g.Broadcast(sg); err != nil && !network.IsCloseError(err) {
					g.Fire(&network.AsyncError{Src: "runSearch[Broadcast]", Err: err})
				}

				g.expire()
			}
		}
	}()

	return func() {
		stop <- struct{}{}
	}
}

// Run reads packets from Conn and emits an event for each received packet
// Not safe for concurrent invocation
func (g *UDPGameList) Run() error {
	var sg = w3gs.SearchGame{
		GameVersion: g.GameVersion,
	}

	if _, err := g.Broadcast(&sg); err != nil {
		return err
	}

	if g.BroadcastInterval > 0 {
		var stop = g.runSearch(&sg)
		defer stop()
	}

	return g.W3GSPacketConn.Run(&g.EventEmitter, network.NoTimeout)
}

func (g *UDPGameList) onRefreshGame(ev *network.Event) {
	var pkt = ev.Arg.(*w3gs.RefreshGame)
	var adr = ev.Opt[0].(net.Addr)
	var idx = udpIndex{
		source: adr.String(),
		gameID: pkt.HostCounter,
	}

	g.gmut.Lock()

	game, ok := g.games[idx]
	if ok {
		game.expires = time.Now().Add(g.BroadcastInterval + 5*time.Second)

		var update = game.SlotsUsed != pkt.SlotsUsed || game.SlotsAvailable != pkt.SlotsAvailable
		if update {
			game.SlotsUsed = pkt.SlotsUsed
			game.SlotsAvailable = pkt.SlotsAvailable

			update = game.GameVersion == g.GameVersion
		}

		g.gmut.Unlock()

		if update {
			g.Fire(Update{})
		}

		return
	}

	g.initMap()
	g.games[idx] = &udpRecord{
		expires: time.Now().Add(5 * time.Minute),
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

func (g *UDPGameList) onCreateGame(ev *network.Event) {
	var pkt = ev.Arg.(*w3gs.CreateGame)
	var adr = ev.Opt[0].(net.Addr)
	var idx = udpIndex{
		source: adr.String(),
		gameID: pkt.HostCounter,
	}

	g.gmut.Lock()
	g.initMap()
	g.games[idx] = &udpRecord{
		created: time.Now(),
		expires: time.Now().Add(5 * time.Minute),
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

func (g *UDPGameList) onDecreateGame(ev *network.Event) {
	var pkt = ev.Arg.(*w3gs.DecreateGame)
	var adr = ev.Opt[0].(net.Addr)
	var idx = udpIndex{
		source: adr.String(),
		gameID: pkt.HostCounter,
	}

	g.gmut.Lock()
	game, update := g.games[idx]

	if update {
		update = game.GameVersion == g.GameVersion
		delete(g.games, idx)
	}

	g.gmut.Unlock()

	if update {
		g.Fire(Update{})
	}
}

func (g *UDPGameList) onGameInfo(ev *network.Event) {
	var pkt = ev.Arg.(*w3gs.GameInfo)
	var adr = ev.Opt[0].(net.Addr)
	var idx = udpIndex{
		source: adr.String(),
		gameID: pkt.HostCounter,
	}

	var update = pkt.GameVersion == g.GameVersion

	g.gmut.Lock()

	game, ok := g.games[idx]
	if !ok {
		game = &udpRecord{
			created: time.Now().Add(time.Duration(pkt.UptimeSec) * -time.Second),
		}

		g.initMap()
		g.games[idx] = game
	}

	update = update && game.GameInfo != *pkt

	game.expires = time.Now().Add(g.BroadcastInterval + 5*time.Second)
	game.GameInfo = *pkt

	g.gmut.Unlock()

	if update {
		g.Fire(Update{})
	}
}
