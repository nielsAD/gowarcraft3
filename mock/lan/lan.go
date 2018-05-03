// Author:  Niels A.D.
// Project: gowarcraft3 (https://github.com/nielsAD/gowarcraft3)
// License: Mozilla Public License, v2.0

// Package lan implements a mocked Warcraft 3 LAN client that can be used to discover local games.
package lan

import (
	"context"
	"net"
	"sync"

	"github.com/nielsAD/gowarcraft3/protocol"

	"github.com/nielsAD/gowarcraft3/mock"
	"github.com/nielsAD/gowarcraft3/protocol/w3gs"
)

// BroadcastAddr for LAN games
var BroadcastAddr = net.UDPAddr{IP: net.IPv4bcast, Port: 6112}

// FindGame returns the an arbitrary game hosted in LAN
func FindGame(ctx context.Context, gv w3gs.GameVersion) (addr string, hostCounter uint32, entryKey uint32, err error) {
	var g *GameList
	g, err = NewGameList(gv, 6112)
	if err != nil {
		g, err = NewGameList(gv, 0)
	}
	if err != nil {
		return
	}

	defer g.Close()

	var stop = make(chan struct{})
	g.On(Update{}, func(ev *mock.Event) {
		for k, v := range g.Games() {
			addr = k
			hostCounter = v.HostCounter
			entryKey = v.EntryKey
			stop <- struct{}{}
			break
		}
	})
	g.On(Stop{}, func(ev *mock.Event) {
		stop <- struct{}{}
	})

	go g.Run()

	select {
	case <-ctx.Done():
		err = ctx.Err()
	case <-stop:
	}

	return
}

var bcmut sync.Mutex
var bccon *net.UDPConn
var bcbuf protocol.Buffer

// Broadcast hosted game information to LAN
func Broadcast(game *w3gs.GameInfo) (err error) {
	bcmut.Lock()

	if bccon == nil {
		bccon, err = net.ListenUDP("udp4", &net.UDPAddr{})
	}

	if err == nil {
		if err = game.Serialize(&bcbuf); err == nil {
			_, err = bccon.WriteTo(bcbuf.Bytes, &BroadcastAddr)
		}
		bcbuf.Truncate()
	}

	bcmut.Unlock()
	return err
}
