// Author:  Niels A.D.
// Project: gowarcraft3 (https://github.com/nielsAD/gowarcraft3)
// License: Mozilla Public License, v2.0

package peer

import (
	"net"
	"sync"
	"sync/atomic"
	"time"

	"github.com/nielsAD/gowarcraft3/network"
	"github.com/nielsAD/gowarcraft3/protocol"
	"github.com/nielsAD/gowarcraft3/protocol/w3gs"
)

// Host manages (incoming/outgoing) peer connections in a game lobby
// Public methods/fields are thread-safe unless explicitly stated otherwise
type Host struct {
	network.EventEmitter

	wg       sync.WaitGroup
	listener *net.TCPListener

	pmut    sync.Mutex
	peers   map[uint8]*Player
	peerset protocol.BitSet32

	// Atomic
	gameticks uint32

	// Set once before ListenAndServe(), read-only after that
	Encoding     w3gs.Encoding
	PlayerInfo   w3gs.PlayerInfo
	EntryKey     uint32
	PingInterval time.Duration
}

// GameTicks state sent to peers
func (h *Host) GameTicks() uint32 {
	return atomic.LoadUint32(&h.gameticks)
}

// IncGameTicks increments the GameTicks state with t
func (h *Host) IncGameTicks(t uint32) {
	atomic.AddUint32(&h.gameticks, t)
}

// PeerSet of connected peers
func (h *Host) PeerSet() protocol.BitSet32 {
	h.pmut.Lock()
	peerset := h.peerset
	h.pmut.Unlock()

	return peerset
}

// Peer returns registered Player for playerID
func (h *Host) Peer(playerID uint8) *Player {
	h.pmut.Lock()
	var peer = h.peers[playerID]
	h.pmut.Unlock()

	return peer
}

// ListenAndServe opens a new TCP listener on InternalAddr and serves incoming peer connections
// On success, listening address overrides InternalAddr/ExternalAddr
// Not safe for concurrent invocation
func (h *Host) ListenAndServe() error {
	if h.listener != nil {
		h.listener.Close()
	}

	var l, err = net.ListenTCP("tcp4", h.PlayerInfo.InternalAddr.TCPAddr())
	if err != nil {
		return err
	}

	h.listener = l
	h.PlayerInfo.InternalAddr = protocol.Addr(l.Addr())
	h.PlayerInfo.ExternalAddr = h.PlayerInfo.InternalAddr

	h.wg.Add(1)
	go func() {
		h.acceptAndServe(l)
		h.wg.Done()
	}()

	return nil
}

// Register new player info
func (h *Host) Register(info *w3gs.PlayerInfo) (*Player, error) {
	var player = NewPlayer(info)

	player.On(&w3gs.PeerMessage{}, func(ev *network.Event) {
		var pkt = ev.Arg.(*w3gs.PeerMessage)
		if pkt.Content == "" {
			return
		}

		h.Fire(&Chat{
			Event:   Event{Peer: player},
			Content: pkt.Content,
		})
	})

	h.pmut.Lock()
	if h.peers[info.PlayerID] != nil {
		h.pmut.Unlock()
		return nil, ErrDupPeerID
	}
	if h.peers == nil {
		h.peers = make(map[uint8]*Player)
	}
	h.peers[info.PlayerID] = player
	h.pmut.Unlock()

	h.Fire(&Registered{Peer: player})
	return player, nil
}

// Deregister (and disconnect) player
func (h *Host) Deregister(playerID uint8) {
	h.pmut.Lock()
	var p = h.peers[playerID]
	if p != nil {
		p.Close()
		delete(h.peers, playerID)
	}
	h.pmut.Unlock()

	if p != nil {
		h.Fire(&Deregistered{Peer: p})
	}
}

// Accept a new connection from known player
func (h *Host) Accept(conn net.Conn) (*Player, error) {
	pc, err := h.connectPlayer(conn)
	if err != nil {
		conn.Close()
		return nil, err
	}

	h.pmut.Lock()
	peer := h.peers[pc.PlayerID]
	if peer == nil {
		h.pmut.Unlock()
		conn.Close()
		return nil, ErrUnknownPeerID
	}

	var old = peer.W3GSConn.Conn()
	if old != nil && peer.PlayerInfo.PlayerID > h.PlayerInfo.PlayerID {
		h.pmut.Unlock()
		conn.Close()
		return nil, ErrAlreadyConnected
	}

	if _, err := network.NewW3GSConn(conn, nil, h.Encoding).Send(&w3gs.PeerConnect{
		JoinCounter: peer.PlayerInfo.JoinCounter,
		EntryKey:    h.EntryKey,
		PlayerID:    h.PlayerInfo.PlayerID,
		PeerSet:     h.peerset,
	}); err != nil {
		h.pmut.Unlock()
		conn.Close()
		return nil, err
	}

	var done = make(chan struct{})
	peer.Once(network.RunStart{}, func(ev *network.Event) {
		atomic.StoreUint32(&peer.peerset, uint32(pc.PeerSet))
		h.peerset.Set(uint(peer.PlayerInfo.PlayerID))

		// Unlock only once serve() is running
		// This ensures RunStart is only called once and serve() is actually using conn
		h.pmut.Unlock()

		h.Fire(&Connected{
			Event: Event{Peer: peer},
			Dial:  false,
		})

		done <- struct{}{}
	})

	peer.SetConn(conn, w3gs.NewFactoryCache(w3gs.DefaultFactory), h.Encoding)

	h.wg.Add(1)
	go func() {
		if err := h.serve(peer); err != nil && !network.IsCloseError(err) {
			peer.Fire(&network.AsyncError{Src: "Host.Accept[Serve]", Err: err})
		}

		h.disconnectPlayer(conn, peer)
		h.wg.Done()
	}()

	<-done

	return peer, nil
}

// Dial opens a new connection to player
func (h *Host) Dial(playerID uint8) (*Player, error) {
	h.pmut.Lock()
	peer := h.peers[playerID]
	if peer == nil {
		h.pmut.Unlock()
		return nil, ErrUnknownPeerID
	}
	if peer.W3GSConn.Conn() != nil {
		h.pmut.Unlock()
		return nil, ErrAlreadyConnected
	}

	conn, err := net.DialTCP("tcp", nil, peer.PlayerInfo.InternalAddr.TCPAddr())
	if err != nil {
		conn, err = net.DialTCP("tcp", nil, peer.PlayerInfo.ExternalAddr.TCPAddr())
	}
	if err != nil {
		h.pmut.Unlock()
		return nil, err
	}

	conn.SetKeepAlive(false)
	conn.SetNoDelay(true)
	conn.SetLinger(3)

	if _, err := network.NewW3GSConn(conn, nil, h.Encoding).Send(&w3gs.PeerConnect{
		JoinCounter: peer.PlayerInfo.JoinCounter,
		EntryKey:    h.EntryKey,
		PlayerID:    h.PlayerInfo.PlayerID,
		PeerSet:     h.peerset,
	}); err != nil {
		conn.Close()
		h.pmut.Unlock()
		return nil, err
	}
	h.pmut.Unlock()

	// Release mutex while waiting for a response
	pc, err := h.connectPlayer(conn)
	if err != nil {
		conn.Close()
		return nil, err
	}

	h.pmut.Lock()
	var old = peer.W3GSConn.Conn()
	if old != nil && peer.PlayerInfo.PlayerID <= h.PlayerInfo.PlayerID {
		h.pmut.Unlock()
		conn.Close()
		return nil, ErrAlreadyConnected
	}

	var done = make(chan struct{})
	peer.Once(network.RunStart{}, func(ev *network.Event) {
		atomic.StoreUint32(&peer.peerset, uint32(pc.PeerSet))
		h.peerset.Set(uint(peer.PlayerInfo.PlayerID))

		// Unlock only once serve() is running
		// This ensures RunStart is only called once and serve() is actually using conn
		h.pmut.Unlock()

		h.Fire(&Connected{
			Event: Event{Peer: peer},
			Dial:  true,
		})

		done <- struct{}{}

		ev.PreventNext()
	})

	peer.SetConn(conn, w3gs.NewFactoryCache(w3gs.DefaultFactory), h.Encoding)

	h.wg.Add(1)
	go func() {
		if err := h.serve(peer); err != nil && !network.IsCloseError(err) {
			peer.Fire(&network.AsyncError{Src: "Host.Dial[Serve]", Err: err})
		}

		h.disconnectPlayer(conn, peer)
		h.wg.Done()
	}()

	<-done

	return peer, nil
}

// Say sends a chat message to peers, returns failed PIDs
func (h *Host) Say(s string) []uint8 {
	var fail = []uint8{}

	h.pmut.Lock()
	for _, p := range h.peers {
		if _, err := p.Send(&w3gs.PeerMessage{Message: w3gs.Message{
			RecipientIDs: []uint8{p.PlayerInfo.PlayerID},
			SenderID:     h.PlayerInfo.PlayerID,
			Type:         w3gs.MsgChat,
			Content:      s,
		}}); err != nil {
			fail = append(fail, p.PlayerInfo.PlayerID)
		}
	}
	h.pmut.Unlock()

	return fail
}

// Wait for all goroutines to finish
func (h *Host) Wait() {
	h.wg.Wait()
}

// Close closes all connections to peers
func (h *Host) Close() {
	if h.listener != nil {
		h.listener.Close()
	}

	h.pmut.Lock()
	for idx, p := range h.peers {
		p.Close()
		delete(h.peers, idx)
	}
	h.pmut.Unlock()
}

func (h *Host) connectPlayer(conn net.Conn) (*w3gs.PeerConnect, error) {
	pkt, err := network.NewW3GSConn(conn, nil, h.Encoding).NextPacket(10 * time.Second)
	if err != nil {
		return nil, err
	}

	switch p := pkt.(type) {
	case *w3gs.PeerConnect:
		if p.EntryKey != h.EntryKey {
			return nil, ErrInvalidEntryKey
		}
		if p.JoinCounter != h.PlayerInfo.JoinCounter {
			return nil, ErrInvalidJoinCounter
		}

		return p, nil

	default:
		return nil, ErrInvalidFirstPacket
	}
}

func (h *Host) disconnectPlayer(conn net.Conn, peer *Player) {
	h.pmut.Lock()
	conn.Close()
	var dc = peer.W3GSConn.Conn() == conn
	if dc {
		peer.W3GSConn.SetConn(nil, nil, h.Encoding)
		atomic.StoreUint32(&peer.peerset, 0)
		atomic.StoreUint32(&peer.rtt, 0)
		h.peerset.Clear(uint(peer.PlayerInfo.PlayerID))
	}
	h.pmut.Unlock()

	if dc {
		h.Fire(&Disconnected{Peer: peer})
	}
}

func (h *Host) servePeerPing(peer *Player) func() {
	var stop = make(chan struct{})

	go func() {
		var ticker = time.NewTicker(h.PingInterval)

		var pkt w3gs.PeerPing
		for {
			select {
			case <-stop:
				ticker.Stop()
				return
			case c := <-ticker.C:
				pkt.Payload = uint32(c.Sub(peer.StartTime).Milliseconds())
				pkt.PeerSet = h.PeerSet()
				pkt.GameTicks = h.GameTicks()
				if _, err := peer.SendOrClose(&pkt); err != nil {
					peer.Fire(&network.AsyncError{Src: "Host.serve[PeerPing]", Err: err})
				}
			}
		}
	}()

	return func() {
		stop <- struct{}{}
	}
}

func (h *Host) serve(peer *Player) error {
	if h.PingInterval != 0 {
		var stop = h.servePeerPing(peer)
		defer stop()
	}

	return peer.Run()
}

func (h *Host) acceptAndServe(listener *net.TCPListener) {
	for {
		conn, err := listener.AcceptTCP()
		if err != nil {
			if !network.IsCloseError(err) {
				h.Fire(&network.AsyncError{Src: "acceptAndServe[Listen]", Err: err})
			}
			break
		}

		conn.SetNoDelay(true)
		conn.SetLinger(3)

		_, err = h.Accept(conn)
		if err != nil {
			h.Fire(&network.AsyncError{Src: "acceptAndServe[Accept]", Err: err})
		}
	}
}
