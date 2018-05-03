// Author:  Niels A.D.
// Project: gowarcraft3 (https://github.com/nielsAD/gowarcraft3)
// License: Mozilla Public License, v2.0

// Package player implements a mocked Warcraft 3 game client that can be used to add dummy players to games.
package player

import (
	"fmt"
	"net"
	"sync"
	"time"

	"github.com/nielsAD/gowarcraft3/protocol"
	"github.com/nielsAD/gowarcraft3/protocol/w3gs"
)

// FakePlayer represents a mocked player that can join a game lobby
type FakePlayer struct {
	Peer

	wg        sync.WaitGroup
	listener  *net.TCPListener
	peers     map[uint8]*Peer
	peerMutex sync.Mutex

	DialPeers bool

	GameTicks   uint32
	HostCounter uint32
	EntryKey    uint32

	OnPeerError      func(err error)
	OnPeerAccept     func(conn *net.TCPConn) bool
	OnPeerConnect    func(peer *Peer) bool
	OnPeerDisconnect func(peer *Peer)
}

// PeerName returns the name for given player ID
func (f *FakePlayer) PeerName(id uint8) string {
	f.peerMutex.Lock()
	var peer = f.peers[id]
	f.peerMutex.Unlock()

	if peer == nil {
		return fmt.Sprintf("PID(%d)", id)
	}

	return peer.Name
}

// Say sends a chat message
func (f *FakePlayer) Say(s string) error {
	f.peerMutex.Lock()
	defer f.peerMutex.Unlock()

	var nonPeers = []uint8{}
	for _, p := range f.peers {
		if _, err := p.Send(&w3gs.PeerMessage{Message: w3gs.Message{
			RecipientIDs: []uint8{p.ID},
			SenderID:     f.ID,
			Type:         w3gs.MsgChat,
			Content:      s,
		}}); err != nil {
			nonPeers = append(nonPeers, p.ID)
		}
	}

	var _, err = f.Send(&w3gs.Message{
		RecipientIDs: nonPeers,
		SenderID:     f.ID,
		Type:         w3gs.MsgChat,
		Content:      s,
	})
	return err
}

func (f *FakePlayer) changeVal(t w3gs.MessageType, v uint8) error {
	var _, err = f.Send(&w3gs.Message{
		RecipientIDs: []uint8{1},
		SenderID:     f.ID,
		Type:         t,
		NewVal:       v,
	})
	return err
}

// ChangeRace to r
func (f *FakePlayer) ChangeRace(r w3gs.RacePref) error {
	return f.changeVal(w3gs.MsgRaceChange, uint8(r&w3gs.RaceMask))
}

// ChangeTeam to t
func (f *FakePlayer) ChangeTeam(t uint8) error {
	return f.changeVal(w3gs.MsgTeamChange, t)
}

// ChangeColor to c
func (f *FakePlayer) ChangeColor(c uint8) error {
	return f.changeVal(w3gs.MsgColorChange, c)
}

// ChangeHandicap to h
func (f *FakePlayer) ChangeHandicap(h uint8) error {
	return f.changeVal(w3gs.MsgHandicapChange, h)
}

// Leave game
func (f *FakePlayer) Leave(reason w3gs.LeaveReason) error {
	defer f.conn.Close()
	defer f.disconnectPeers()

	var _, err = f.Send(&w3gs.Leave{Reason: reason})
	return err
}

func (f *FakePlayer) updatePeerSet(id uint8, set bool) error {
	var old = f.PeerSet
	if set {
		f.PeerSet.Set(uint(id))
	} else {
		f.PeerSet.Clear(uint(id))
	}

	if old != f.PeerSet {
		if _, err := f.Send(&w3gs.PeerSet{PeerSet: protocol.BitSet16(f.PeerSet)}); err != nil {
			return err
		}
	}

	return nil
}

func (f *FakePlayer) onPeerError(err error) {
	if f.OnPeerError != nil {
		f.OnPeerError(err)
	}
}

// onPeerAccept must return true if peer is allowed to connect
func (f *FakePlayer) onPeerAccept(conn *net.TCPConn) bool {
	if f.OnPeerAccept == nil {
		return true
	}
	return f.OnPeerAccept(conn)
}

// onPeerConnect must return true if peer is allowed to connect
func (f *FakePlayer) onPeerConnect(peer *Peer) bool {
	if f.OnPeerConnect == nil {
		return true
	}
	return f.OnPeerConnect(peer)
}

func (f *FakePlayer) onPeerDisconnect(peer *Peer) {
	f.peerMutex.Lock()
	defer f.peerMutex.Unlock()

	if f.OnPeerDisconnect != nil {
		f.OnPeerDisconnect(peer)
	}

	peer.conn = nil
	peer.RTT = 0
	peer.PeerSet = 0

	f.updatePeerSet(peer.ID, false)
}

func (f *FakePlayer) connectToPeer(conn *net.TCPConn, accepted bool) (*Peer, error) {
	if err := conn.SetReadDeadline(time.Now().Add(5 * time.Second)); err != nil {
		return nil, err
	}

	pkt, _, err := w3gs.DeserializePacket(conn)
	if err != nil {
		return nil, err
	}

	f.peerMutex.Lock()
	defer f.peerMutex.Unlock()

	switch p := pkt.(type) {
	case *w3gs.PeerConnect:
		var peer = f.peers[p.PlayerID]

		if peer == nil || peer.ID < 1 || peer.ID > 32 {
			return nil, ErrUnknownPeerID
		}
		if peer.conn != nil && peer.conn != conn {
			if peer.ID > f.ID && accepted {
				return nil, ErrAlreadyConnected
			}
			peer.conn.Close()
		}
		if p.EntryKey != f.EntryKey {
			return nil, ErrInvalidEntryKey
		}
		if p.JoinCounter != f.JoinCounter {
			return nil, ErrInvalidJoinCounter
		}

		f.updatePeerSet(peer.ID, true)
		peer.PeerSet = p.PeerSet
		peer.conn = conn

		if accepted {
			if _, err := peer.Send(&w3gs.PeerConnect{
				JoinCounter: peer.JoinCounter,
				EntryKey:    f.EntryKey,
				PlayerID:    f.ID,
				PeerSet:     f.PeerSet,
			}); err != nil {
				return nil, err
			}
		}

		peer.StartTime = time.Now()
		return peer, nil

	default:
		return nil, ErrInvalidFirstPacket
	}
}

func (f *FakePlayer) processPeerPacket(peer *Peer, pkt w3gs.Packet) error {
	switch p := pkt.(type) {
	case *w3gs.PeerPing:
		peer.PeerSet = p.PeerSet
		peer.Send(&w3gs.PeerPong{Ping: w3gs.Ping{Payload: p.Payload}})
	case *w3gs.PeerPong:
		peer.RTT = uint32(time.Now().Sub(peer.StartTime).Nanoseconds()/1e6) - p.Payload
	}

	return nil
}

func (f *FakePlayer) servePeer(conn *net.TCPConn, accepted bool) {
	defer f.wg.Done()

	conn.SetNoDelay(true)
	conn.SetLinger(3)
	defer conn.Close()

	var peer *Peer
	var err error
	if peer, err = f.connectToPeer(conn, accepted); err != nil {
		f.onPeerError(err)
		return
	}

	defer f.onPeerDisconnect(peer)
	if !f.onPeerConnect(peer) {
		return
	}

	var pingTicker = time.NewTicker(10 * time.Second)
	defer pingTicker.Stop()

	go func() {
		for range pingTicker.C {
			if _, err := peer.Send(&w3gs.PeerPing{
				Payload:   uint32(time.Now().Sub(peer.StartTime).Nanoseconds() / 1e6),
				PeerSet:   f.PeerSet,
				GameTicks: f.GameTicks,
			}); err != nil {
				peer.onError(err)
			}
		}
	}()

	for {
		// Expecting a ping every 10 seconds
		pkt, err := peer.NextPacket(30 * time.Second)
		if err != nil {
			peer.onError(err)
			break
		}

		if peer.onPacket(pkt) {
			continue
		}

		if err := f.processPeerPacket(peer, pkt); err != nil {
			peer.onError(err)
			break
		}
	}
}

func (f *FakePlayer) acceptPeers() {
	defer f.wg.Done()
	if f.listener == nil {
		return
	}

	defer f.listener.Close()
	for {
		conn, err := f.listener.AcceptTCP()
		if err != nil {
			f.onPeerError(err)
			break
		}

		if !f.onPeerAccept(conn) {
			continue
		}

		f.wg.Add(1)
		go f.servePeer(conn, true)
	}
}

func (f *FakePlayer) disconnectPeers() {
	if f.listener != nil {
		f.listener.Close()
	}

	f.peerMutex.Lock()
	defer f.peerMutex.Unlock()

	for idx, p := range f.peers {
		if p.conn != nil {
			p.conn.Close()
		}
		delete(f.peers, idx)
	}
}

func (f *FakePlayer) connectToHost(addr *net.TCPAddr) error {
	if f.conn != nil {
		return ErrAlreadyConnected
	}

	d, err := net.DialTCP("tcp4", nil, addr)
	if err != nil {
		return err
	}

	f.conn = d
	f.conn.SetNoDelay(true)
	f.conn.SetLinger(3)
	f.JoinCounter++

	var listenPort uint16
	if f.listener != nil {
		listenPort = uint16(f.listener.Addr().(*net.TCPAddr).Port)
	}

	if _, err := f.Send(&w3gs.Join{
		HostCounter:  f.HostCounter,
		EntryKey:     f.EntryKey,
		ListenPort:   listenPort,
		JoinCounter:  f.JoinCounter,
		PlayerName:   f.Name,
		InternalAddr: protocol.Addr(f.conn.LocalAddr()),
	}); err != nil {
		return err
	}

	pkt, err := f.NextPacket(5 * time.Second)
	if err != nil {
		return err
	}

	switch p := pkt.(type) {
	case *w3gs.SlotInfoJoin:
		f.ID = p.PlayerID
	case *w3gs.RejectJoin:
		return RejectReasonToError(p.Reason)
	default:
		return ErrInvalidFirstPacket
	}

	f.StartTime = time.Now()
	return nil
}

// ProcessPacket processes a w3gs packet from host
func (f *FakePlayer) ProcessPacket(pkt w3gs.Packet) error {
	switch p := pkt.(type) {
	case *w3gs.Ping:
		_, err := f.Send(&w3gs.Pong{Ping: w3gs.Ping{Payload: p.Payload}})
		return err

	case *w3gs.MapCheck:
		_, err := f.Send(&w3gs.MapState{Ready: true, FileSize: p.FileSize})
		return err

	case *w3gs.PlayerInfo:
		var peer = Peer{
			Name:        p.PlayerName,
			ID:          p.PlayerID,
			JoinCounter: p.JoinCounter,
		}

		var conn *net.TCPConn
		if f.DialPeers {
			if c, e := net.DialTCP("tcp4", nil, p.InternalAddr.TCPAddr()); e == nil {
				conn = c
			} else if c, e := net.DialTCP("tcp4", nil, p.ExternalAddr.TCPAddr()); e == nil {
				conn = c
			}

			if conn != nil {
				if _, e := w3gs.SerializePacketWithBuffer(conn, &peer.sbuf, &w3gs.PeerConnect{
					JoinCounter: p.JoinCounter,
					EntryKey:    f.EntryKey,
					PlayerID:    f.ID,
					PeerSet:     f.PeerSet,
				}); e != nil {
					conn = nil
				}
			}
		}

		f.peerMutex.Lock()
		f.peers[p.PlayerID] = &peer
		f.peerMutex.Unlock()

		if conn != nil {
			f.wg.Add(1)
			go f.servePeer(conn, false)
		}

	case *w3gs.PlayerLeft:
		f.peerMutex.Lock()

		var peer = f.peers[p.PlayerID]
		delete(f.peers, p.PlayerID)

		if peer != nil && peer.conn != nil {
			peer.conn.Close()
		}

		f.peerMutex.Unlock()

	case *w3gs.CountDownEnd:
		f.Send(&w3gs.GameLoaded{})

	case *w3gs.TimeSlot:
		// Cannot reply to this as we don't know the correct checksum for this round
		// replying with wrong info will result in a desync
		// not replying will result in lagscreen and drop
		//w3gs.SerializePacketWithBuffer(conn, &sbuf, &w3gs.TimeSlotAck{})

		f.GameTicks += uint32(p.TimeIncrementMS)

	}

	return nil
}

// Wait for all goroutines to finish
func (f *FakePlayer) Wait() {
	f.wg.Wait()
}

// Run processes all host packets in a goroutine. Use Wait() to wait for the goroutine to end.
func (f *FakePlayer) Run() {
	f.wg.Add(1)

	go func() {
		defer f.Leave(w3gs.LeaveLost)
		defer f.wg.Done()

		for {
			pkt, err := f.NextPacket(60 * time.Second)
			if err != nil {
				f.onError(err)
				break
			}

			if f.onPacket(pkt) {
				continue
			}

			if err := f.ProcessPacket(pkt); err != nil {
				f.onError(err)
				break
			}
		}
	}()
}

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
