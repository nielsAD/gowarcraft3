// Package fakeplayer implements a mocked Warcraft 3 game client that can be used to add dummy players to games.
package fakeplayer

import (
	"net"
)

// JoinLobby joins a game as a mocked player
func JoinLobby(addr *net.TCPAddr, name string, hostCounter uint32, entryKey uint32, listenPort int) (*FakePlayer, error) {
	var f = FakePlayer{
		Peer: Peer{
			Name: name,
		},
		peers:       make(map[uint8]*Peer),
		DialPeers:   true,
		HostCounter: hostCounter,
		EntryKey:    entryKey,
	}

	if listenPort >= 0 {
		var err error
		f.listener, err = net.ListenTCP("tcp4", &net.TCPAddr{Port: listenPort})
		if err != nil {
			return nil, err
		}
	}

	if err := f.connectToHost(addr); err != nil {
		if f.listener != nil {
			f.listener.Close()
		}
		return nil, err
	}

	f.wg.Add(1)
	go f.acceptPeers()

	return &f, nil
}
