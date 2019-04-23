// Author:  Niels A.D.
// Project: gowarcraft3 (https://github.com/nielsAD/gowarcraft3)
// License: Mozilla Public License, v2.0

// Package dummy implements a mocked Warcraft III game client that can be used to add dummy players to lobbies.
package dummy

import (
	"net"
	"strings"
	"time"
	"unicode"

	"github.com/nielsAD/gowarcraft3/network"
	"github.com/nielsAD/gowarcraft3/network/peer"
	"github.com/nielsAD/gowarcraft3/protocol"
	"github.com/nielsAD/gowarcraft3/protocol/w3gs"
)

// Chat event
type Chat struct {
	Sender  *w3gs.PlayerInfo
	Content string
}

// Say event
type Say struct {
	Content string
}

// Player represents a mocked player that can join a game lobby
type Player struct {
	peer.Host
	network.W3GSConn

	// Set once before Join(), read-only after that
	HostAddr    string
	HostCounter uint32
	DialPeers   bool
}

// Join a game lobby as a mocked player
func Join(addr string, name string, hostCounter uint32, entryKey uint32, listenPort int, encoding w3gs.Encoding) (*Player, error) {
	var p = Player{
		Host: peer.Host{
			PlayerInfo: w3gs.PlayerInfo{
				PlayerName: name,
			},
			Encoding:     encoding,
			EntryKey:     entryKey,
			PingInterval: 10 * time.Second,
		},
		HostAddr:    addr,
		HostCounter: hostCounter,
		DialPeers:   true,
	}

	p.InitDefaultHandlers()

	if listenPort >= 0 {
		p.Host.PlayerInfo.InternalAddr.Port = uint16(listenPort)
		if err := p.ListenAndServe(); err != nil {
			return nil, err
		}
	}

	var err = p.Join()
	return &p, err
}

// Join opens a new connection to host
// Not safe for concurrent invocation
func (p *Player) Join() error {
	addr, err := net.ResolveTCPAddr("tcp4", p.HostAddr)
	if err != nil {
		return err
	}

	conn, err := net.DialTCP("tcp4", nil, addr)
	if err != nil {
		return err
	}

	conn.SetKeepAlive(false)
	conn.SetNoDelay(true)
	conn.SetLinger(3)

	w3gsconn := network.NewW3GSConn(conn, nil, p.Encoding)

	p.PlayerInfo.JoinCounter++
	if _, err := w3gsconn.Send(&w3gs.Join{
		HostCounter:  p.HostCounter,
		EntryKey:     p.EntryKey,
		ListenPort:   p.PlayerInfo.InternalAddr.Port,
		JoinCounter:  p.PlayerInfo.JoinCounter,
		PlayerName:   p.PlayerInfo.PlayerName,
		InternalAddr: p.PlayerInfo.InternalAddr,
	}); err != nil {
		w3gsconn.Close()
		return err
	}

	pkt, err := w3gsconn.NextPacket(5 * time.Second)
	if err != nil {
		w3gsconn.Close()
		return err
	}

	switch r := pkt.(type) {
	case *w3gs.SlotInfoJoin:
		p.PlayerInfo.PlayerID = r.PlayerID
	case *w3gs.RejectJoin:
		w3gsconn.Close()
		return RejectReasonToError(r.Reason)
	default:
		w3gsconn.Close()
		return ErrInvalidFirstPacket
	}

	p.SetConn(conn, w3gs.NewFactoryCache(w3gs.DefaultFactory), p.Encoding)
	return nil
}

// SendOrClose sends pkt to player, closes connection on failure
func (p *Player) SendOrClose(pkt w3gs.Packet) (int, error) {
	n, err := p.W3GSConn.Send(pkt)
	if err == nil || network.IsCloseError(err) {
		return n, nil
	}

	p.Close()
	return n, err
}

// Close closes all connections to host and peers
func (p *Player) Close() error {
	p.Host.Close()
	return p.W3GSConn.Close()
}

// Run reads packets and emits an event for each received packet
// Not safe for concurrent invocation
func (p *Player) Run() error {
	var err = p.W3GSConn.Run(&p.EventEmitter, 35*time.Second)
	p.Leave(w3gs.LeaveLobby)

	return err
}

// Say sends a chat message
func (p *Player) Say(s string) error {
	s = strings.Map(func(r rune) rune {
		if !unicode.IsPrint(r) {
			return -1
		}
		return r
	}, s)
	if len(s) == 0 {
		return nil
	}

	var nonPeers = p.Host.Say(s)
	if _, err := p.Send(&w3gs.Message{
		RecipientIDs: nonPeers,
		SenderID:     p.PlayerInfo.PlayerID,
		Type:         w3gs.MsgChat,
		Content:      s,
	}); err != nil {
		return err
	}

	p.Fire(&Say{Content: s})
	return nil
}

func (p *Player) changeVal(t w3gs.MessageType, v uint8) error {
	var _, err = p.SendOrClose(&w3gs.Message{
		RecipientIDs: []uint8{1},
		SenderID:     p.PlayerInfo.PlayerID,
		Type:         t,
		NewVal:       v,
	})
	return err
}

// ChangeRace to r
func (p *Player) ChangeRace(r w3gs.RacePref) error {
	return p.changeVal(w3gs.MsgRaceChange, uint8(r&w3gs.RaceMask))
}

// ChangeTeam to t
func (p *Player) ChangeTeam(t uint8) error {
	return p.changeVal(w3gs.MsgTeamChange, t)
}

// ChangeColor to c
func (p *Player) ChangeColor(c uint8) error {
	return p.changeVal(w3gs.MsgColorChange, c)
}

// ChangeHandicap to h
func (p *Player) ChangeHandicap(h uint8) error {
	return p.changeVal(w3gs.MsgHandicapChange, h)
}

// Leave game
func (p *Player) Leave(reason w3gs.LeaveReason) error {
	var _, err = p.Send(&w3gs.Leave{
		Reason: reason,
	})

	p.Close()
	return err
}

// InitDefaultHandlers adds the default callbacks for relevant packets
func (p *Player) InitDefaultHandlers() {
	p.On(&peer.Connected{}, p.onPeerConnected)
	p.On(&peer.Chat{}, p.onPeerChat)
	p.On(&w3gs.Ping{}, p.onPing)
	p.On(&w3gs.MapCheck{}, p.onMapCheck)
	p.On(&w3gs.MessageRelay{}, p.onMessageRelay)
	p.On(&w3gs.PlayerInfo{}, p.onPlayerInfo)
	p.On(&w3gs.PlayerLeft{}, p.onPlayerLeft)
	p.On(&w3gs.CountDownEnd{}, p.onCountDownEnd)
	p.On(&w3gs.TimeSlot{}, p.onTimeSlot)
}

func (p *Player) onPeerConnected(ev *network.Event) {
	if _, err := p.SendOrClose(&w3gs.PeerSet{PeerSet: protocol.BitSet16(p.PeerSet())}); err != nil {
		p.Fire(&network.AsyncError{Src: "onPeerConnected[Send]", Err: err})
	}
}

func (p *Player) onPeerChat(ev *network.Event) {
	var chat = ev.Arg.(*peer.Chat)
	p.Fire(&Chat{
		Sender:  &chat.Peer.PlayerInfo,
		Content: chat.Content,
	})
}

func (p *Player) onPing(ev *network.Event) {
	var pkt = ev.Arg.(*w3gs.Ping)

	if _, err := p.SendOrClose(&w3gs.Pong{Ping: w3gs.Ping{Payload: pkt.Payload}}); err != nil {
		p.Fire(&network.AsyncError{Src: "onPing[Send]", Err: err})
	}
}

func (p *Player) onMapCheck(ev *network.Event) {
	var pkt = ev.Arg.(*w3gs.MapCheck)

	if _, err := p.SendOrClose(&w3gs.MapState{Ready: true, FileSize: pkt.FileSize}); err != nil {
		p.Fire(&network.AsyncError{Src: "onMapCheck[Send]", Err: err})
	}
}

func (p *Player) onMessageRelay(ev *network.Event) {
	var pkt = ev.Arg.(*w3gs.MessageRelay)
	if pkt.Content == "" {
		return
	}

	var chat = Chat{Content: pkt.Content}
	var peer = p.Peer(pkt.SenderID)
	if peer != nil {
		chat.Sender = &peer.PlayerInfo
	}

	p.Fire(&chat)
}

func (p *Player) onPlayerInfo(ev *network.Event) {
	var pkt = ev.Arg.(*w3gs.PlayerInfo)

	peer, err := p.Register(pkt)
	if err != nil {
		p.Fire(&network.AsyncError{Src: "onPlayerInfo[Register]", Err: err})
		return
	}

	if !p.DialPeers || (peer.PlayerInfo.InternalAddr.IP == nil && peer.PlayerInfo.ExternalAddr.IP == nil) {
		return
	}

	go func() {
		if _, err := p.Host.Dial(peer.PlayerInfo.PlayerID); err != nil && !network.IsRefusedError(err) {
			p.Fire(&network.AsyncError{Src: "onPlayerInfo[Dial]", Err: err})
		}
	}()
}

func (p *Player) onPlayerLeft(ev *network.Event) {
	var pkt = ev.Arg.(*w3gs.PlayerLeft)

	p.Deregister(pkt.PlayerID)
}

func (p *Player) onCountDownEnd(ev *network.Event) {
	if _, err := p.SendOrClose(&w3gs.GameLoaded{}); err != nil {
		p.Fire(&network.AsyncError{Src: "onCountDownEnd[Send]", Err: err})
	}
}

func (p *Player) onTimeSlot(ev *network.Event) {
	// Cannot reply to this as we don't know the correct checksum for this round
	// replying with wrong info will result in a desync
	// not replying will result in lagscreen and drop

	var pkt = ev.Arg.(*w3gs.TimeSlot)
	p.IncGameTicks(uint32(pkt.TimeIncrementMS))
}
