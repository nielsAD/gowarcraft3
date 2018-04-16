package fakeplayer

import (
	"fmt"
	"net"
	"sync"
	"time"

	"github.com/nielsAD/noot/pkg/util"
	"github.com/nielsAD/noot/pkg/w3gs"
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

	OnPeerConnected    func(peer *Peer)
	OnPeerDisconnected func(peer *Peer)
	OnPeerPacket       func(peer *Peer, pkt w3gs.Packet) bool

	OnPacket func(pkt w3gs.Packet) bool
}

// PeerName returns the name for given player ID
func (f *FakePlayer) PeerName(id uint8) string {
	f.peerMutex.Lock()
	var peer = f.peers[id]
	f.peerMutex.Unlock()

	if peer == nil {
		return fmt.Sprintf("PID(%v)", id)
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

func (f *FakePlayer) updatePeerMask(id uint8, set bool) error {
	var old = f.PeerMask
	if set {
		f.PeerMask |= (1 << (id - 1))
	} else {
		f.PeerMask &= ^(1 << (id - 1))
	}

	if old != f.PeerMask {
		if _, err := f.Send(&w3gs.PeerMask{PeerMask: uint16(f.PeerMask)}); err != nil {
			return err
		}
	}

	return nil
}

func (f *FakePlayer) peerConnected(peer *Peer) {
	if f.OnPeerConnected != nil {
		f.OnPeerConnected(peer)
	}
}

func (f *FakePlayer) peerDisconnected(peer *Peer) {
	f.peerMutex.Lock()
	defer f.peerMutex.Unlock()

	if f.OnPeerDisconnected != nil {
		f.OnPeerDisconnected(peer)
	}

	peer.conn = nil
	peer.RTT = 0
	peer.PeerMask = 0

	f.updatePeerMask(peer.ID, false)
}

func (f *FakePlayer) processPeerPacket(peer *Peer, pkt w3gs.Packet) bool {
	if f.OnPeerPacket != nil {
		return f.OnPeerPacket(peer, pkt)
	}
	return false
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

		f.updatePeerMask(peer.ID, true)
		peer.PeerMask = p.PeerMask
		peer.conn = conn

		if accepted {
			if _, err := peer.Send(&w3gs.PeerConnect{
				JoinCounter: peer.JoinCounter,
				EntryKey:    f.EntryKey,
				PlayerID:    f.ID,
				PeerMask:    f.PeerMask,
			}); err != nil {
				return nil, err
			}
		}

		return peer, nil

	default:
		return nil, ErrInvalidFirstPacket
	}
}

func (f *FakePlayer) servePeer(conn *net.TCPConn, accepted bool) {
	defer f.wg.Done()

	conn.SetNoDelay(true)
	conn.SetLinger(3)
	defer conn.Close()

	var peer *Peer
	var err error
	if peer, err = f.connectToPeer(conn, accepted); err != nil {
		return
	}

	f.peerConnected(peer)
	defer f.peerDisconnected(peer)

	var start = time.Now()
	var pingTicker = time.NewTicker(10 * time.Second)
	defer pingTicker.Stop()

	go func() {
		for range pingTicker.C {
			peer.Send(&w3gs.PeerPing{
				Payload:   uint32(time.Now().Sub(start).Nanoseconds() / 1e6),
				PeerMask:  f.PeerMask,
				GameTicks: f.GameTicks,
			})
		}
	}()

	for {
		// Expecting a ping every 10 seconds
		pkt, err := peer.NextRawPacket(13 * time.Second)
		if err != nil {
			break
		}

		if f.processPeerPacket(peer, pkt) {
			continue
		}

		switch p := pkt.(type) {
		case *w3gs.PeerPing:
			peer.PeerMask = p.PeerMask
			peer.Send(&w3gs.PeerPong{Ping: w3gs.Ping{Payload: p.Payload}})
		case *w3gs.PeerPong:
			peer.RTT = uint32(time.Now().Sub(start).Nanoseconds()/1e6) - p.Payload
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
		conn, err := f.listener.Accept()
		if err != nil {
			break
		}

		f.wg.Add(1)
		go f.servePeer(conn.(*net.TCPConn), true)
	}
}

func (f *FakePlayer) disconnectPeers() {
	f.peerMutex.Lock()
	defer f.peerMutex.Unlock()

	if f.listener != nil {
		f.listener.Close()
	}

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
		InternalAddr: util.Addr(f.conn.LocalAddr()),
	}); err != nil {
		return err
	}

	pkt, err := f.NextRawPacket(5 * time.Second)
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

	return nil
}

// NextPacket waits for the next packet from host and processes it
func (f *FakePlayer) NextPacket() (w3gs.Packet, error) {
	pkt, err := f.NextRawPacket(33 * time.Second)
	if err != nil {
		return nil, err
	}

	if f.OnPacket != nil && f.OnPacket(pkt) {
		return pkt, nil
	}

	switch p := pkt.(type) {
	case *w3gs.Ping:
		_, err = f.Send(&w3gs.Pong{Ping: w3gs.Ping{Payload: p.Payload}})

	case *w3gs.MapCheck:
		_, err = f.Send(&w3gs.MapState{Ready: true, FileSize: p.FileSize})

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
					PeerMask:    f.PeerMask,
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

	return pkt, nil
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
			if _, err := f.NextPacket(); err != nil {
				break
			}
		}
	}()
}
