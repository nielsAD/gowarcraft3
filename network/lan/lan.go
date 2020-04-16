// Author:  Niels A.D.
// Project: gowarcraft3 (https://github.com/nielsAD/gowarcraft3)
// License: Mozilla Public License, v2.0

// Package lan implements a mocked Warcraft III LAN client that can be used to discover local games.
package lan

import (
	"context"

	"github.com/nielsAD/gowarcraft3/network"
	"github.com/nielsAD/gowarcraft3/protocol/w3gs"
)

// Update event for GameList changes
type Update struct{}

// GameList keeps track of all the hosted games in the Local Area Network
// Emits events for every received packet and Update{} when the output of Games() changes
type GameList interface {
	network.Listener
	Games() map[string]w3gs.GameInfo
	Run() error
	Close() error
}

// NewGameList initializes proper GameList type for game version
func NewGameList(gv w3gs.GameVersion) (GameList, error) {
	if gv.Version > 0 && gv.Version < 30 {
		// Use random port to not occupy port 6112 by default
		return NewUDPGameList(gv, 0)
	}

	return NewMDNSGameList(gv)
}

// Advertiser broadcasts available game information to the Local Area Network
// Emits events for every received packet, responds to search queries
type Advertiser interface {
	network.Listener

	Create() error
	Refresh(slotsUsed uint32, slotsAvailable uint32) error
	Decreate() error

	Run() error
	Close() error
}

// NewAdvertiser initializes proper Advertiser type for game version
func NewAdvertiser(info *w3gs.GameInfo) (Advertiser, error) {
	if info.GameVersion.Version > 0 && info.GameVersion.Version < 30 {
		// Use random port to not occupy port 6112 by default
		return NewUDPAdvertiser(info, 0)
	}

	return NewMDNSAdvertiser(info)
}

// FindGame returns entry information for an arbitrary game hosted in LAN
func FindGame(ctx context.Context, gv w3gs.GameVersion) (addr string, hostCounter uint32, entryKey uint32, err error) {
	var g GameList
	g, err = NewGameList(gv)
	if err != nil {
		return
	}

	var stop = make(chan error)
	g.On(Update{}, func(ev *network.Event) {
		for k, v := range g.Games() {
			addr = k
			hostCounter = v.HostCounter
			entryKey = v.EntryKey
			stop <- nil
			return
		}
	})
	g.On(&network.AsyncError{}, func(ev *network.Event) {
		var err = ev.Arg.(*network.AsyncError)
		stop <- err
	})

	go func() {
		var err = g.Run()
		stop <- err
	}()

	select {
	case <-ctx.Done():
		err = ctx.Err()
	case e := <-stop:
		err = e
	}

	g.Close()
	return
}
