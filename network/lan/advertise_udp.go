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
	a.SetWriteTimeout(time.Second)
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

func (a *UDPAdvertiser) runBroadcast() func() {
	var stop = make(chan struct{})

	go func() {
		var ticker = time.NewTicker(a.BroadcastInterval)
		for {
			select {
			case <-stop:
				ticker.Stop()
				return
			case <-ticker.C:
				if err := a.refresh(); err != nil && !network.IsCloseError(err) {
					a.Fire(&network.AsyncError{Src: "runBroadcast[refresh]", Err: err})
				}
			}
		}
	}()

	return func() {
		stop <- struct{}{}
	}
}

// Run broadcasts gameinfo in Local Area Network
func (a *UDPAdvertiser) Run() error {
	if err := a.Create(); err != nil {
		return err
	}
	defer a.Decreate()

	if a.BroadcastInterval > 0 {
		var stop = a.runBroadcast()
		defer stop()
	}

	return a.W3GSPacketConn.Run(&a.EventEmitter, network.NoTimeout)
}

// Close the connection
func (a *UDPAdvertiser) Close() error {
	if err := a.Decreate(); err != nil && !network.IsCloseError(err) {
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

	a.info.UptimeSec = (uint32)(time.Since(a.created).Seconds())

	var addr = ev.Opt[0].(net.Addr)
	if _, err := a.Send(addr, &a.info); err != nil {
		a.Fire(&network.AsyncError{Src: "onSearchGame[Send]", Err: err})
	}

	a.imut.Unlock()
}
