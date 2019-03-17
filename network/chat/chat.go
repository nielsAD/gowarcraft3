// Author:  Niels A.D.
// Project: gowarcraft3 (https://github.com/nielsAD/gowarcraft3)
// License: Mozilla Public License, v2.0

// Package chat implements the official classic Battle.net chat API.
package chat

import (
	"context"
	"math/rand"
	"os"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gorilla/websocket"
	"github.com/nielsAD/gowarcraft3/network"
	"github.com/nielsAD/gowarcraft3/network/bnet"
	"github.com/nielsAD/gowarcraft3/protocol/capi"
)

// Config for chat.Bot
type Config struct {
	Endpoint   string
	APIKey     string
	RPCTimeout time.Duration
}

// Bot implements a basic chat bot using the official classic Battle.net chat API
// Public methods/fields are thread-safe unless explicitly stated otherwise
type Bot struct {
	network.EventEmitter
	network.CAPIConn

	rid uint32

	chatmut sync.Mutex
	channel string
	users   map[int64]*User

	// Set once before Connect(), read-only after that
	Config
}

// NewBot initializes a Bot struct
func NewBot(conf *Config) (*Bot, error) {
	var b = Bot{
		Config: *conf,
	}

	b.InitDefaultHandlers()
	return &b, nil
}

// Channel currently chatting in
func (b *Bot) Channel() string {
	b.chatmut.Lock()
	var res = b.channel
	b.chatmut.Unlock()
	return res
}

// User in channel by id
func (b *Bot) User(uid int64) (*User, bool) {
	b.chatmut.Lock()
	u, ok := b.users[uid]
	if ok {
		copy := *u
		u = &copy
	}
	b.chatmut.Unlock()

	return u, ok
}

// Users in channel
func (b *Bot) Users() map[int64]User {
	var res = make(map[int64]User)

	b.chatmut.Lock()
	for k, v := range b.users {
		res[k] = *v
	}
	b.chatmut.Unlock()

	return res
}

// Connect opens a new connection to server and joins chat
func (b *Bot) Connect() error {
	conn, _, err := websocket.DefaultDialer.Dial(b.Endpoint, nil)
	if err != nil {
		return err
	}

	capiconn := network.NewCAPIConn(conn)

	if _, err := syncRPC(capiconn, 5*time.Second, capi.CmdAuthenticate, capi.Authenticate{
		APIKey: b.Config.APIKey,
	}); err != nil {
		capiconn.Close()
		return err
	}
	if _, err := syncRPC(capiconn, 5*time.Second, capi.CmdConnect); err != nil {
		capiconn.Close()
		return err
	}

	b.SetConn(capiconn.Conn())
	return nil
}

func syncRPC(conn *network.CAPIConn, timeout time.Duration, command string, arg ...interface{}) (interface{}, error) {
	var p interface{}
	switch len(arg) {
	case 0:
		p = nil
	case 1:
		p = arg[0]
	default:
		p = arg
	}

	var rid = int64(rand.Int31())

	if err := conn.Send(&capi.Packet{
		Command:   command + capi.CmdRequestSuffix,
		RequestID: rid,
		Payload:   p,
	}); err != nil {
		return nil, err
	}

	pkt, err := conn.NextPacket(timeout)
	if err != nil {
		return nil, err
	}
	if pkt.Command != command+capi.CmdResponseSuffix {
		return nil, ErrUnexpectedPacket
	}
	if pkt.Status != nil && *pkt.Status != capi.Success {
		return nil, pkt.Status
	}
	return pkt.Payload, nil
}

func (b *Bot) asyncRPC(ctx context.Context, command string, arg ...interface{}) (interface{}, error) {
	var p interface{}
	switch len(arg) {
	case 0:
		p = nil
	case 1:
		p = arg[0]
	default:
		p = arg
	}

	var rid = int64(atomic.AddUint32(&b.rid, 1))
	var cmd = command + capi.CmdResponseSuffix
	var rsp = make(chan capi.Packet)

	var eid = b.On(&capi.Packet{}, func(ev *network.Event) {
		var pkt = ev.Arg.(*capi.Packet)
		if pkt.RequestID != rid || pkt.Command != cmd {
			return
		}

		rsp <- *pkt
	})

	defer b.Off(eid)

	if err := b.Send(&capi.Packet{
		Command:   command + capi.CmdRequestSuffix,
		RequestID: rid,
		Payload:   p,
	}); err != nil {
		return nil, err
	}

	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case pkt := <-rsp:
		if pkt.Status != nil && *pkt.Status != capi.Success {
			return nil, pkt.Status
		}
		return pkt.Payload, nil
	}
}

// RPC executes Remote Procedure Call cmd asynchronously, retries on timeout/rate-limit
// Needs to called in a goroutine while Run() is running asynchronously to process incoming packets
func (b *Bot) RPC(command string, arg ...interface{}) (interface{}, error) {
	var t = b.RPCTimeout
	if t == 0 {
		t = 5 * time.Second
	}

	var d = time.Second
	for {
		var ctx, cancel = context.WithTimeout(context.Background(), t)
		var res, err = b.asyncRPC(ctx, command, arg...)
		cancel()

		if err == nil {
			return res, nil
		}
		if !os.IsTimeout(err) || d >= 10*time.Second {
			return nil, err
		}
		time.Sleep(d)
		d *= 2
	}
}

// Run reads packets and emits an event for each received packet
// Not safe for concurrent invocation
func (b *Bot) Run() error {
	return b.CAPIConn.Run(&b.EventEmitter, 12*time.Hour)
}

// SendMessage sends a chat message to the channel
func (b *Bot) SendMessage(s string) error {
	s = bnet.FilterChat(s)
	if len(s) == 0 {
		return nil
	}

	if _, err := b.RPC(capi.CmdSendMessage, &capi.SendMessage{Message: s}); err != nil {
		return err
	}

	return nil
}

// SendEmote sends an emote on behalf of a bot
func (b *Bot) SendEmote(s string) error {
	_, err := b.RPC(capi.CmdSendEmote, &capi.SendEmote{Message: s})
	return err
}

// SendWhisper sends a chat message to one user in the channel
func (b *Bot) SendWhisper(uid int64, s string) error {
	_, err := b.RPC(capi.CmdSendWhisper, &capi.SendWhisper{UserID: uid, Message: s})
	return err
}

// KickUser kicks a user from the channel
func (b *Bot) KickUser(uid int64) error {
	_, err := b.RPC(capi.CmdKickUser, &capi.KickUser{UserID: uid})
	return err
}

// BanUser bans a user from the channel
func (b *Bot) BanUser(uid int64) error {
	_, err := b.RPC(capi.CmdBanUser, &capi.BanUser{UserID: uid})
	return err
}

// UnbanUser un-bans a user from the channel
func (b *Bot) UnbanUser(username string) error {
	_, err := b.RPC(capi.CmdUnbanUser, &capi.UnbanUser{Username: username})
	return err
}

// SetModerator sets the current chat moderator to a member of the current chat
func (b *Bot) SetModerator(uid int64) error {
	_, err := b.RPC(capi.CmdSetModerator, &capi.SetModerator{UserID: uid})
	return err
}

// InitDefaultHandlers adds the default callbacks for relevant packets
func (b *Bot) InitDefaultHandlers() {
	b.On(&capi.Packet{}, b.onPacket)
	b.On(&capi.DisconnectEvent{}, b.onDisconnectEvent)
	b.On(&capi.ConnectEvent{}, b.onConnectEvent)
	b.On(&capi.MessageEvent{}, b.onMessageEvent)
	b.On(&capi.UserUpdateEvent{}, b.onUserUpdateEvent)
	b.On(&capi.UserLeaveEvent{}, b.onUserLeaveEvent)
}

func (b *Bot) onPacket(ev *network.Event) {
	var pkt = ev.Arg.(*capi.Packet)
	if pkt.Status != nil && *pkt.Status == capi.ErrNotConnected {
		b.Fire(&network.AsyncError{Src: "onPacket", Err: pkt.Status})
		b.Close()
	}
}

func (b *Bot) onDisconnectEvent(ev *network.Event) {
	b.Close()
}

func (b *Bot) onConnectEvent(ev *network.Event) {
	var pkt = ev.Arg.(*capi.ConnectEvent)

	b.chatmut.Lock()
	b.channel = pkt.Channel
	b.users = nil
	b.chatmut.Unlock()
}

func (b *Bot) onMessageEvent(ev *network.Event) {
	var pkt = ev.Arg.(*capi.MessageEvent)

	b.chatmut.Lock()
	var u = b.users[pkt.UserID]
	if u != nil {
		u.LastSeen = time.Now()
	}
	b.chatmut.Unlock()
}

func (b *Bot) onUserUpdateEvent(ev *network.Event) {
	var pkt = ev.Arg.(*capi.UserUpdateEvent)

	b.chatmut.Lock()
	var p = b.users[pkt.UserID]
	if p == nil {
		if b.users == nil {
			b.users = make(map[int64]*User)
		}

		var t = time.Now()
		p = &User{
			Joined:   t,
			LastSeen: t,
		}

		b.users[pkt.UserID] = p
	}
	p.Update(pkt)
	b.chatmut.Unlock()
}

func (b *Bot) onUserLeaveEvent(ev *network.Event) {
	var pkt = ev.Arg.(*capi.UserLeaveEvent)

	b.chatmut.Lock()
	delete(b.users, pkt.UserID)
	b.chatmut.Unlock()
}
