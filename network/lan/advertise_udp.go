// Author:  Niels A.D.
// Project: gowarcraft3 (https://github.com/nielsAD/gowarcraft3)
// License: Mozilla Public License, v2.0

package lan

import (
	"net"
	"sync"
	"time"

	"github.com/nielsAD/gowarcraft3/network"
	"github.com/nielsAD/gowarcraft3/protocol/w3gs"
)

// UDPAdvertiser advertises a hosted game in the Local Area Network using UDP broadcast
type UDPAdvertiser struct {
	network.EventEmitter
	network.W3GSPacketConn

	imut sync.Mutex
	info w3gs.GameInfo

	created time.Time

	// Set once before Run(), read-only after that
	BroadcastInterval time.Duration
}

// NewUDPAdvertiser initializes UDPAdvertiser struct
func NewUDPAdvertiser(info *w3gs.GameInfo, port int) (*UDPAdvertiser, error) {
	conn, err := net.ListenUDP("udp4", &net.UDPAddr{Port: port})
	if err != nil {
		return nil, err
	}

	var a = UDPAdvertiser{
		info:              *info,
		created:           time.Now().Add(time.Duration(info.UptimeSec) * -time.Second),
		BroadcastInterval: 3 * time.Second,
	}

	a.InitDefaultHandlers()
	a.SetConn(conn, w3gs.NewFactoryCache(w3gs.DefaultFactory), w3gs.Encoding{GameVersion: info.GameVersion.Version})
	return &a, nil
}

// Create local game
func (a *UDPAdvertiser) Create() error {
	a.imut.Lock()
	var pkt = w3gs.CreateGame{
		GameVersion: a.info.GameVersion,
		HostCounter: a.info.HostCounter,
	}
	a.imut.Unlock()

	_, err := a.Broadcast(&pkt)
	return err
}

func (a *UDPAdvertiser) refresh() error {
	a.imut.Lock()
	var pkt = w3gs.RefreshGame{
		HostCounter:    a.info.HostCounter,
		SlotsUsed:      a.info.SlotsUsed,
		SlotsAvailable: a.info.SlotsAvailable,
	}
	a.imut.Unlock()

	_, err := a.Broadcast(&pkt)
	return err
}

// Refresh game info
func (a *UDPAdvertiser) Refresh(slotsUsed uint32, slotsAvailable uint32) error {
	a.imut.Lock()
	a.info.SlotsUsed = slotsUsed
	a.info.SlotsAvailable = slotsAvailable
	a.imut.Unlock()

	return a.refresh()
}

// Decreate game
func (a *UDPAdvertiser) Decreate() error {
	a.imut.Lock()
	var pkt = w3gs.DecreateGame{
		HostCounter: a.info.HostCounter,
	}
	a.imut.Unlock()

	_, err := a.Broadcast(&pkt)
	return err
}

// Run broadcasts gameinfo in Local Area Network
func (a *UDPAdvertiser) Run() error {
	if err := a.Create(); err != nil {
		return err
	}

	if a.BroadcastInterval > 0 {
		var ticker = time.NewTicker(a.BroadcastInterval)
		defer ticker.Stop()

		go func() {
			for range ticker.C {
				if err := a.refresh(); err != nil && !network.IsConnClosedError(err) {
					a.Fire(&network.AsyncError{Src: "Run[refresh]", Err: err})
				}
			}
		}()
	}

	return a.W3GSPacketConn.Run(&a.EventEmitter, 0)
}

// Close the connection
func (a *UDPAdvertiser) Close() error {
	if err := a.Decreate(); err != nil && !network.IsConnClosedError(err) {
		a.Fire(&network.AsyncError{Src: "Close[Decreate]", Err: err})
	}
	return a.W3GSPacketConn.Close()
}

// InitDefaultHandlers adds the default callbacks for relevant packets
func (a *UDPAdvertiser) InitDefaultHandlers() {
	a.On(&w3gs.SearchGame{}, a.onSearchGame)
}

func (a *UDPAdvertiser) onSearchGame(ev *network.Event) {
	var pkt = ev.Arg.(*w3gs.SearchGame)

	a.imut.Lock()
	if pkt.Product != a.info.Product {
		a.imut.Unlock()
		return
	}

	a.info.UptimeSec = (uint32)(time.Now().Sub(a.created).Seconds())
	if _, err := a.Send(ev.Opt[0].(net.Addr), &a.info); err != nil && !network.IsConnClosedError(err) {
		a.Fire(&network.AsyncError{Src: "onSearchGame[Send]", Err: err})
	}
	a.imut.Unlock()
}
