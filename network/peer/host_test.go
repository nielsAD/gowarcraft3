// Author:  Niels A.D.
// Project: gowarcraft3 (https://github.com/nielsAD/gowarcraft3)
// License: Mozilla Public License, v2.0

package peer_test

import (
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/nielsAD/gowarcraft3/network"
	"github.com/nielsAD/gowarcraft3/network/peer"
	"github.com/nielsAD/gowarcraft3/protocol/w3gs"
)

func makeHosts(t *testing.T, n int) []*peer.Host {
	var res []*peer.Host
	for i := 1; i <= n; i++ {
		var p = peer.Host{
			PlayerInfo: w3gs.PlayerInfo{
				JoinCounter: uint32(i + 100),
				PlayerID:    uint8(i),
			},
			PingInterval: 5 * time.Millisecond,
		}

		var idx = i
		p.On(&network.AsyncError{}, func(ev *network.Event) {
			var err = ev.Arg.(*network.AsyncError)
			t.Logf("[ERROR][HOST%d] %s\n", idx, err.Error())
		})
		p.On(&peer.Registered{}, func(ev *network.Event) {
			var reg = ev.Arg.(*peer.Registered)
			t.Logf("[HOST%d] %d registered\n", idx, reg.Peer.PlayerInfo.PlayerID)
		})
		p.On(&peer.Deregistered{}, func(ev *network.Event) {
			var reg = ev.Arg.(*peer.Deregistered)
			t.Logf("[HOST%d] %d deregistered\n", idx, reg.Peer.PlayerInfo.PlayerID)
		})
		p.On(&peer.Connected{}, func(ev *network.Event) {
			var conn = ev.Arg.(*peer.Connected)
			t.Logf("[HOST%d] %d connected (dial: %v)\n", idx, conn.Peer.PlayerInfo.PlayerID, conn.Dial)
		})
		p.On(&peer.Disconnected{}, func(ev *network.Event) {
			var conn = ev.Arg.(*peer.Disconnected)
			t.Logf("[HOST%d] %d disconnected\n", idx, conn.Peer.PlayerInfo.PlayerID)
		})

		res = append(res, &p)
	}

	return res
}

func registerHosts(t *testing.T, hosts []*peer.Host) {
	var n = len(hosts)
	for i := 0; i < n; i++ {
		for j := 0; j < n; j++ {
			if i == j {
				continue
			}

			var idxi = i
			var idxj = j

			var p, err = hosts[i].Register(&hosts[j].PlayerInfo)
			if err != nil {
				t.Fatal(idxi, err)
			}

			p.On(&network.AsyncError{}, func(ev *network.Event) {
				var err = ev.Arg.(*network.AsyncError)
				t.Logf("[ERROR][HOST%d][PEER%d] %s\n", idxi, idxj, err.Error())
			})
		}
	}
}

func closeAll(hosts []*peer.Host) {
	var n = len(hosts)
	for i := 0; i < n; i++ {
		hosts[i].Close()
	}
}

func TestEvents(t *testing.T) {
	var hosts = makeHosts(t, 2)
	if err := hosts[0].ListenAndServe(); err != nil {
		t.Fatal(err)
	}

	var registerCount int32
	var connectCount int32
	for i := 0; i < len(hosts); i++ {
		hosts[i].On(&peer.Registered{}, func(ev *network.Event) { atomic.AddInt32(&registerCount, 1) })
		hosts[i].On(&peer.Deregistered{}, func(ev *network.Event) { atomic.AddInt32(&registerCount, -1) })
		hosts[i].On(&peer.Connected{}, func(ev *network.Event) { atomic.AddInt32(&connectCount, 1) })
		hosts[i].On(&peer.Disconnected{}, func(ev *network.Event) { atomic.AddInt32(&connectCount, -1) })
	}
	registerHosts(t, hosts)
	if atomic.LoadInt32(&registerCount) != 2 {
		t.Fatal("Expected registerCount to be 2")
	}
	if atomic.LoadInt32(&connectCount) != 0 {
		t.Fatal("Expected connectCount to be 0")
	}

	hosts[0].On(&peer.Connected{}, func(ev *network.Event) {
		var conn = ev.Arg.(*peer.Connected)
		if conn.Dial {
			t.Fatal("Expected host[0].dial to be false")
		}
		if conn.Peer.PlayerInfo.PlayerID != hosts[1].PlayerInfo.PlayerID {
			t.Fatal("Unexpected host[0].ID")
		}
	})

	hosts[1].On(&peer.Connected{}, func(ev *network.Event) {
		var conn = ev.Arg.(*peer.Connected)
		if !conn.Dial {
			t.Fatal("Expected host[1].dial to be true")
		}
		if conn.Peer.PlayerInfo.PlayerID != hosts[0].PlayerInfo.PlayerID {
			t.Fatal("Unexpected host[1].ID")
		}
	})

	var p, err = hosts[1].Dial(hosts[0].PlayerInfo.PlayerID)
	if err != nil {
		t.Fatal(err)
	}
	if p == nil {
		t.Fatal("Expected p != nil")
	}
	if p.PlayerInfo.PlayerID != hosts[0].PlayerInfo.PlayerID {
		t.Fatal("Expected p.ID")
	}

	time.Sleep(10 * time.Millisecond)
	if atomic.LoadInt32(&connectCount) != 2 {
		t.Fatal("Expected connectCount to be 2")
	}

	if _, err := hosts[0].Dial(hosts[1].PlayerInfo.PlayerID); err != peer.ErrAlreadyConnected {
		t.Fatal("Expected ErrAlreadyConnected")
	}

	hosts[0].Close()
	time.Sleep(10 * time.Millisecond)

	if atomic.LoadInt32(&registerCount) != 2 {
		t.Fatal("Expected registerCount to be 2 again")
	}
	if atomic.LoadInt32(&connectCount) != 0 {
		t.Fatal("Expected connectCount to be 0 again")
	}

	closeAll(hosts)
}

func TestMass(t *testing.T) {
	var chat int32
	var wg sync.WaitGroup

	// test 12 * 11 / 2 = 56 simultaneous connections
	var hosts = makeHosts(t, 12)
	for i := 0; i < len(hosts); i++ {
		var h = hosts[i]
		if err := h.ListenAndServe(); err != nil {
			t.Fatal(i, err)
		}
		h.On(&peer.Registered{}, func(ev *network.Event) {
			var reg = ev.Arg.(*peer.Registered)
			wg.Add(1)
			go func() {
				time.Sleep(time.Millisecond)
				if _, err := h.Dial(reg.Peer.PlayerInfo.PlayerID); err != nil {
					t.Logf("[ERROR][DIAL] %d -> %d %s\n", h.PlayerInfo.PlayerID, reg.Peer.PlayerInfo.PlayerID, err)
				}
				wg.Done()
			}()
		})
		h.On(&peer.Chat{}, func(ev *network.Event) {
			var reg = ev.Arg.(*peer.Chat)
			if reg.Content == "Hello world!" {
				atomic.AddInt32(&chat, 1)
			}
		})
	}

	registerHosts(t, hosts)
	wg.Wait()

	for i := 0; i < len(hosts); i++ {
		for j := 0; j < len(hosts); j++ {
			if i == j {
				continue
			}
			peer := hosts[i].Peer(uint8(j + 1))
			if peer == nil || peer.Conn() == nil {
				t.Fatalf("%d <-> %d not connected\n", i+1, j+1)
			}
		}
	}

	t.Logf("Sending chat to %v\n", len(hosts)-1)
	if fail := hosts[0].Say("Hello world!"); len(fail) > 0 {
		t.Fatal("Failed to send to ", fail)
	}

	// Wait for chat
	for try := 0; try < 100 && atomic.LoadInt32(&chat) != int32(len(hosts)-1); try++ {
		time.Sleep(10 * time.Millisecond)
	}

	if atomic.LoadInt32(&chat) != int32(len(hosts)-1) {
		t.Fatalf("Expected chat from %v peers, got %v\n", len(hosts)-1, atomic.LoadInt32(&chat))
	}

	atomic.StoreInt32(&chat, 0)
	hosts[0].Deregister(uint8(len(hosts)))

	t.Logf("Sending chat to %v\n", len(hosts)-2)
	if fail := hosts[0].Say("Hello world!"); len(fail) > 0 {
		t.Fatal("Failed to send to ", fail)
	}

	// Wait for chat
	for try := 0; try < 100 && atomic.LoadInt32(&chat) != int32(len(hosts)-2); try++ {
		time.Sleep(10 * time.Millisecond)
	}

	if atomic.LoadInt32(&chat) != int32(len(hosts)-2) {
		t.Fatalf("Expected chat from %v peers, got %v\n", len(hosts)-2, atomic.LoadInt32(&chat))
	}

	closeAll(hosts)
}
