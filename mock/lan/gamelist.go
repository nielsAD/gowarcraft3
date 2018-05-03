// Author:  Niels A.D.
// Project: gowarcraft3 (https://github.com/nielsAD/gowarcraft3)
// License: Mozilla Public License, v2.0

package lan

import (
	"net"
	"os"
	"sync"
	"time"

	"github.com/nielsAD/gowarcraft3/mock"
	"github.com/nielsAD/gowarcraft3/protocol"
	"github.com/nielsAD/gowarcraft3/protocol/w3gs"
)

// Update event
type Update struct{}

// Stop event
type Stop struct{}

// GameList keeps track of all the hosted games in the Local Area Network
// Emits events for every received packet, Update{} when the output of Games() changes, and Stop{} on shutdown
// Public methods/fields are thread-safe unless explicitly stated otherwise
type GameList struct {
	mock.Emitter

	smut sync.Mutex
	sbuf protocol.Buffer

	gmut  sync.Mutex
	games map[string]w3gs.GameInfo

	// Set these once before Run(), read-only after that
	Conn        net.PacketConn
	ReadTimeout time.Duration
	GameVersion w3gs.GameVersion

	// This is only safe to access during packet handling events
	LastAddr net.Addr
}

// NewGameList opens a new UDP socket to listen for LAN GameList updates
func NewGameList(gv w3gs.GameVersion, port int) (*GameList, error) {
	conn, err := net.ListenUDP("udp4", &net.UDPAddr{Port: port})
	if err != nil {
		return nil, err
	}

	var g = GameList{
		Conn:        conn,
		ReadTimeout: 30 * time.Second,
		GameVersion: gv,
	}

	g.InitDefaultHandlers()
	return &g, nil
}

// Games returns the current list of LAN games. Map key is the remote address.
func (g *GameList) Games() map[string]w3gs.GameInfo {
	var res = make(map[string]w3gs.GameInfo)

	g.smut.Lock()
	for k, v := range g.games {
		if v.GameVersion == g.GameVersion {
			res[k] = v
		}
	}
	g.smut.Unlock()

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

// Close Conn and stop updating the GameList
func (g *GameList) Close() error {
	return g.Conn.Close()
}

func (g *GameList) send(addr net.Addr, pkt w3gs.Packet) (int, error) {
	var n int
	var e error

	g.smut.Lock()
	if e = pkt.Serialize(&g.sbuf); e == nil {
		n, e = g.Conn.WriteTo(g.sbuf.Bytes, addr)
	}
	g.sbuf.Truncate()
	g.smut.Unlock()

	return n, e
}

// Broadcast a packet over LAN
func (g *GameList) Broadcast(pkt w3gs.Packet) (int, error) {
	return g.send(&BroadcastAddr, pkt)
}

// Respond to the sender of the last received packet
// Only to be called during packet handling events
func (g *GameList) Respond(pkt w3gs.Packet) (int, error) {
	return g.send(g.LastAddr, pkt)
}

// Run reads packets from Conn and emits an event for each received packet
func (g *GameList) Run() {
	defer g.Fire(Stop{})
	if g.Conn == nil {
		return
	}

	var sg = w3gs.SearchGame{
		GameVersion: g.GameVersion,
	}

	var buf [2048]byte
	for {
		g.Broadcast(&sg)

		if g.ReadTimeout > 0 {
			g.Conn.SetReadDeadline(time.Now().Add(g.ReadTimeout))
		}

		for {
			g.LastAddr = nil

			size, addr, err := g.Conn.ReadFrom(buf[:])
			if err != nil {
				if os.IsTimeout(err) {
					break
				}
				if !mock.IsConnClosedError(err) {
					g.Fire(&mock.AsyncError{Src: "ReadAndFire[Read]", Err: err})
				}
				return
			}

			g.LastAddr = addr

			pkt, _, err := w3gs.DeserializePacket(&protocol.Buffer{Bytes: buf[:size]})
			if err != nil {
				g.Fire(&mock.AsyncError{Src: "ReadAndFire[Deserialize]", Err: err})
				continue
			}

			g.Fire(pkt)
		}
	}
}

func (g *GameList) onRefreshGame(ev *mock.Event) {
	var pkt = ev.Arg.(*w3gs.RefreshGame)
	var idx = g.LastAddr.String()

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

	if _, err := g.Respond(&sg); err != nil {
		g.Fire(&mock.AsyncError{Src: "onRefreshGame[Respond]", Err: err})
	}
}

func (g *GameList) onCreateGame(ev *mock.Event) {
	var pkt = ev.Arg.(*w3gs.CreateGame)
	var idx = g.LastAddr.String()

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

	if _, err := g.Respond(&sg); err != nil {
		g.Fire(&mock.AsyncError{Src: "onCreateGame[Respond]", Err: err})
	}
}

func (g *GameList) onDecreateGame(ev *mock.Event) {
	var pkt = ev.Arg.(*w3gs.RefreshGame)
	var idx = g.LastAddr.String()

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

func (g *GameList) onGameInfo(ev *mock.Event) {
	var pkt = ev.Arg.(*w3gs.GameInfo)
	var idx = g.LastAddr.String()
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
