package main

import (
	"errors"
	"fmt"
	"io"
	"net"
	"reflect"
	"sync"
	"time"

	"github.com/nielsAD/noot/pkg/util"
	"github.com/nielsAD/noot/pkg/w3gs"
)

// Errors
var (
	ErrJoinRejected       = errors.New("fp: Join rejected")
	ErrInvalidFirstPacket = errors.New("fp: Invalid first packet")
	ErrUnknownPeerID      = errors.New("fp: Unknown peer ID")
	ErrAlreadyConnected   = errors.New("fp: Already connected")
	ErrInvalidEntryKey    = errors.New("fp: Wrong entry key")
	ErrInvalidJoinCounter = errors.New("fp: Wrong join counter")
)

func errUseClosedConn(err error) bool {
	if err == nil {
		return false
	}
	if operr, ok := err.(*net.OpError); ok {
		err = operr.Err
	}
	return err.Error() == "use of closed network connection"
}

func errResetByPeer(err error) bool {
	if err == nil {
		return false
	}
	if operr, ok := err.(*net.OpError); ok {
		err = operr.Err
	}
	return err.Error() == "connection reset by peer"
}

func errClosed(err error) bool {
	return err == io.EOF || errUseClosedConn(err) || errResetByPeer(err)
}

type player struct {
	smutex sync.Mutex
	sbuf   w3gs.SerializationBuffer
	conn   *net.TCPConn

	peerMask uint32

	Name        string
	ID          uint8
	JoinCounter uint32
}

func (p *player) send(pkt w3gs.Packet) error {
	p.smutex.Lock()
	defer p.smutex.Unlock()

	if p.conn == nil {
		return io.EOF
	}

	if err := p.conn.SetWriteDeadline(time.Now().Add(5 * time.Millisecond)); err != nil {
		return err
	}
	if _, err := w3gs.SerializePacketWithBuffer(p.conn, &p.sbuf, pkt); err != nil {
		return err
	}

	return nil
}

// FakePlayer info
type FakePlayer struct {
	player

	wait      sync.WaitGroup
	listener  *net.TCPListener
	peers     map[uint8]*player
	peerMutex sync.Mutex

	gameTicks uint32

	HostCounter uint32
	EntryKey    uint32
}

func (f *FakePlayer) playerName(id uint8) string {
	f.peerMutex.Lock()
	var peer = f.peers[id]
	f.peerMutex.Unlock()

	if peer == nil {
		return fmt.Sprintf("PID(%v)", id)
	}

	return peer.Name
}

func (f *FakePlayer) updatePeerMask(id uint8, set bool) error {
	var old = f.peerMask
	if set {
		f.peerMask |= (1 << (id - 1))
	} else {
		f.peerMask &= ^(1 << (id - 1))
	}
	if old == f.peerMask {
		return nil
	}

	return f.send(&w3gs.PeerMask{PeerMask: uint16(f.peerMask)})
}

func (f *FakePlayer) connectToPeer(conn *net.TCPConn) (*player, error) {
	if err := conn.SetReadDeadline(time.Now().Add(5 * time.Second)); err != nil {
		return nil, err
	}

	pkt, _, err := w3gs.DeserializePacket(conn)
	if err != nil {
		return nil, err
	}

	f.peerMutex.Lock()
	defer f.peerMutex.Unlock()

	var accepted = f.listener != nil && conn.LocalAddr() != nil && f.listener.Addr() != nil &&
		conn.LocalAddr().(*net.TCPAddr).Port == f.listener.Addr().(*net.TCPAddr).Port

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
		peer.peerMask = p.PeerMask
		peer.conn = conn

		if accepted {
			if err := peer.send(&w3gs.PeerConnect{
				JoinCounter: peer.JoinCounter,
				EntryKey:    f.EntryKey,
				PlayerID:    f.ID,
				PeerMask:    f.peerMask,
			}); err != nil {
				return nil, err
			}
		}

		return peer, nil

	default:
		return nil, ErrInvalidFirstPacket
	}
}

func (f *FakePlayer) peerDisconnected(peer *player) {
	f.peerMutex.Lock()
	defer f.peerMutex.Unlock()

	peer.conn = nil
	peer.peerMask = 0

	f.updatePeerMask(peer.ID, false)
}

func (f *FakePlayer) servePeer(conn *net.TCPConn) {
	defer f.wait.Done()

	conn.SetNoDelay(true)
	defer conn.Close()

	var peer *player
	var err error
	if peer, err = f.connectToPeer(conn); err != nil {
		if !errClosed(err) && err != ErrUnknownPeerID {
			logger.Println("[TCP] [PEER] CONNECT ERROR", err)
		}
		return
	}

	defer f.peerDisconnected(peer)
	logger.Printf("[TCP] [PEER] Connected to %v\n", peer.Name)

	var start = time.Now()
	var pingTicker = time.NewTicker(10 * time.Second)
	defer pingTicker.Stop()

	go func() {
		for range pingTicker.C {
			peer.send(&w3gs.PeerPing{
				Payload:   uint32(time.Now().Sub(start).Nanoseconds() / 1000000),
				PeerMask:  f.peerMask,
				GameTicks: f.gameTicks,
			})
		}
	}()

	var rbuf w3gs.DeserializationBuffer

	for {

		// Expecting a ping every 10 seconds
		if err := conn.SetReadDeadline(time.Now().Add(13 * time.Second)); err != nil {
			if !errClosed(err) {
				logger.Println("[TCP] [PEER] DEADLINE ERROR", err)
			}
			break
		}

		pkt, _, err := w3gs.DeserializePacketWithBuffer(conn, &rbuf)
		if errClosed(err) {
			break
		} else if err != nil {
			logger.Println("[TCP] [PEER] DESERIALIZE ERROR", err)
			break
		}

		if *verbose {
			logger.Printf("[TCP] [PEER] Packet %v from peer %v: %v\n", reflect.TypeOf(pkt).String()[6:], peer.Name, pkt)
		}

		switch p := pkt.(type) {
		case *w3gs.PeerPing:
			peer.peerMask = p.PeerMask
			if err := peer.send(&w3gs.PeerPong{Ping: w3gs.Ping{Payload: p.Payload}}); err != nil {
				if !errClosed(err) {
					logger.Println("[TCP] [PEER] SEND ERROR", err)
				}
				break
			}
		case *w3gs.PeerPong:
			logger.Printf("[PEER] RTT to %v is %vms\n", peer.Name, uint32(time.Now().Sub(start).Nanoseconds()/1000000)-p.Payload)

		case *w3gs.PeerMessage:
			if p.Content != "" {
				logger.Printf("[PEER] [CHAT] %v: '%v'\n", peer.Name, p.Content)
			}

		default:
			if !*verbose {
				logger.Printf("[TCP] [PEER] Unexpected packet %v from peer %v: %v\n", reflect.TypeOf(pkt).String()[6:], peer.Name, pkt)
			}
		}
	}

	logger.Printf("[TCP] [PEER] Connection to %v closed\n", peer.Name)
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
	f.JoinCounter++

	var listenPort uint16
	if f.listener != nil {
		listenPort = uint16(f.listener.Addr().(*net.TCPAddr).Port)
	}

	if err := f.send(&w3gs.Join{
		HostCounter:  f.HostCounter,
		EntryKey:     f.EntryKey,
		ListenPort:   listenPort,
		JoinCounter:  f.JoinCounter,
		PlayerName:   f.Name,
		InternalAddr: util.Addr(f.conn.LocalAddr()),
	}); err != nil {
		return err
	}

	if err := f.conn.SetReadDeadline(time.Now().Add(5 * time.Second)); err != nil {
		return err
	}

	pkt, _, err := w3gs.DeserializePacket(f.conn)
	if err != nil {
		return err
	}

	switch p := pkt.(type) {
	case *w3gs.SlotInfoJoin:
		f.ID = p.PlayerID
	case *w3gs.RejectJoin:
		return ErrJoinRejected
	default:
		return ErrInvalidFirstPacket
	}

	return nil
}

func (f *FakePlayer) acceptPeers() {
	defer f.wait.Done()

	logger.Printf("[TCP] [HOST] Listining for peers on %v\n", f.listener.Addr())

	defer f.listener.Close()
	for {
		conn, err := f.listener.Accept()
		if err != nil {
			break
		}
		f.wait.Add(1)
		go f.servePeer(conn.(*net.TCPConn))
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

func (f *FakePlayer) processPackets() {
	defer f.conn.Close()
	defer f.disconnectPeers()
	defer f.wait.Done()

	var rbuf w3gs.DeserializationBuffer

	for {
		// Expecting a ping every 30 seconds
		if err := f.conn.SetReadDeadline(time.Now().Add(33 * time.Second)); err != nil {
			if !errClosed(err) {
				logger.Println("[TCP] [HOST] DEADLINE ERROR", err)
			}
			break
		}

		pkt, _, err := w3gs.DeserializePacketWithBuffer(f.conn, &rbuf)
		if errClosed(err) {
			break
		} else if err != nil {
			logger.Println("[TCP] [HOST] DESERIALIZE ERROR", err)
			break
		}

		if *verbose {
			logger.Println("[TCP] [HOST] Packet", reflect.TypeOf(pkt).String()[6:], pkt)
		}

		switch p := pkt.(type) {
		case *w3gs.Ping:
			if err := f.send(&w3gs.Pong{Ping: w3gs.Ping{Payload: p.Payload}}); err != nil {
				if !errClosed(err) {
					logger.Println("[TCP] [HOST] SEND ERROR", err)
				}
				break
			}

		case *w3gs.MapCheck:
			if err := f.send(&w3gs.MapState{Ready: true, FileSize: p.FileSize}); err != nil {
				if !errClosed(err) {
					logger.Println("[TCP] [HOST] SEND ERROR", err)
				}
				break
			}

		case *w3gs.RejectJoin:
			logger.Println("[HOST] Join denied:", p.Reason)

		case *w3gs.SlotInfo:
			// ignore
		case *w3gs.SlotInfoJoin:
			logger.Println("[HOST] Joined lobby")
			f.ID = p.PlayerID

		case *w3gs.PlayerInfo:
			logger.Printf("[HOST] %v has joined the game\n", p.PlayerName)

			var peer = player{
				Name:        p.PlayerName,
				ID:          p.PlayerID,
				JoinCounter: p.JoinCounter,
			}

			var conn *net.TCPConn
			if c, err := net.DialTCP("tcp4", nil, p.InternalAddr.TCPAddr()); err == nil {
				conn = c
			} else if c, err := net.DialTCP("tcp4", nil, p.ExternalAddr.TCPAddr()); err == nil {
				conn = c
			}

			if conn != nil {
				if _, err := w3gs.SerializePacketWithBuffer(conn, &peer.sbuf, &w3gs.PeerConnect{
					JoinCounter: p.JoinCounter,
					EntryKey:    f.EntryKey,
					PlayerID:    f.ID,
					PeerMask:    f.peerMask,
				}); err != nil {
					conn = nil
				}
			}

			f.peerMutex.Lock()
			f.peers[p.PlayerID] = &peer
			f.peerMutex.Unlock()

			if conn != nil {
				f.wait.Add(1)
				go f.servePeer(conn)
			}

		case *w3gs.PlayerLeft:
			logger.Printf("[HOST] %v has left the game\n", f.playerName(p.PlayerID))

			f.peerMutex.Lock()

			var peer = f.peers[p.PlayerID]
			delete(f.peers, p.PlayerID)

			if peer != nil && peer.conn != nil {
				peer.conn.Close()
			}

			f.peerMutex.Unlock()

		case *w3gs.PlayerKicked:
			logger.Println("[HOST] Kicked from lobby")

		case *w3gs.CountDownStart:
			logger.Println("[HOST] Countdown started")
		case *w3gs.CountDownEnd:
			logger.Println("[HOST] Start loading screen")
			f.send(&w3gs.GameLoaded{})
		case *w3gs.PlayerLoaded:
			// ignore

		case *w3gs.StartLag:
			var laggers []string
			for _, l := range p.Players {
				var peer = f.peers[l.PlayerID]
				if peer != nil {
					laggers = append(laggers, f.playerName(l.PlayerID))
				}
			}
			logger.Printf("[HOST] Laggers %v\n", laggers)
		case *w3gs.StopLag:
			logger.Printf("[HOST] %v stopped lagging\n", f.playerName(p.PlayerID))

		case *w3gs.MessageRelay:
			if p.Content != "" {
				logger.Printf("[HOST] [CHAT] %v: '%v'\n", f.playerName(p.SenderID), p.Content)
			}

		case *w3gs.TimeSlot:
			// Cannot reply to this as we don't know the correct checksum for this round
			// replying with wrong info will result in a desync
			// not replying will result in lagscreen and drop
			//w3gs.SerializePacketWithBuffer(conn, &sbuf, &w3gs.TimeSlotAck{})

			f.gameTicks += uint32(p.TimeIncrementMS)

		default:
			if !*verbose {
				logger.Println("[TCP] [HOST] Unexpected packet", reflect.TypeOf(pkt).String()[6:], pkt)
			}
		}
	}
}

// Leave game
func (f *FakePlayer) Leave(reason w3gs.LeaveReason) error {
	var err = f.send(&w3gs.Leave{Reason: reason})
	f.conn.Close()
	return err
}

// Wait for all goroutines to finish
func (f *FakePlayer) Wait() {
	f.wait.Wait()
}

// Run starts processing packets in the background
func (f *FakePlayer) Run() {
	if f.listener != nil {
		f.wait.Add(1)
		go f.acceptPeers()
	}

	f.wait.Add(1)
	go f.processPackets()
}

// JoinLobby joins a game as a mocked player
func JoinLobby(addr *net.TCPAddr, name string, hostCounter uint32, entryKey uint32, listenPort int) (*FakePlayer, error) {
	var f = FakePlayer{
		player: player{
			Name: name,
		},
		peers:       make(map[uint8]*player),
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

	return &f, nil
}
