// Author:  Niels A.D.
// Project: gowarcraft3 (https://github.com/nielsAD/gowarcraft3)
// License: Mozilla Public License, v2.0

package main

import (
	"context"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/nielsAD/gowarcraft3/network"
	"github.com/nielsAD/gowarcraft3/network/bnet"
	"github.com/nielsAD/gowarcraft3/protocol/bncs"
	"github.com/nielsAD/gowarcraft3/protocol/w3gs"
)

type bnetConfig struct {
	ReconnectDelay time.Duration
	FirstChannel   string
	CommandTrigger string
	AvatarURL      string

	RankWhisper    Rank
	RankTalk       Rank
	RankOperator   Rank
	RankNoWarcraft Rank
	RankClanTag    map[string]Rank
	RankLevel      map[int]Rank
}

// BNetConfig stores the configuration of a single BNet server
type BNetConfig struct {
	bnet.Config
	bnetConfig
}

// BNetRealm manages a BNet connection
type BNetRealm struct {
	*bnet.Client

	smut sync.Mutex
	say  chan string

	// Set once before Run(), read-only after that
	bnetConfig
}

// NewBNetRealm initializes a new BNetRealm struct
func NewBNetRealm(conf *BNetConfig) (*BNetRealm, error) {
	c, err := bnet.NewClient(&conf.Config)
	if err != nil {
		return nil, err
	}

	var b = BNetRealm{
		Client:     c,
		bnetConfig: conf.bnetConfig,
	}

	b.InitDefaultHandlers()

	return &b, nil
}

// Say sends a chat message
func (b *BNetRealm) Say(s string) error {
	b.smut.Lock()
	if b.say == nil {
		b.say = make(chan string, 16)

		go func() {
			for s := range b.say {
				b.Client.Say(s)
			}
		}()
	}
	b.smut.Unlock()

	select {
	case b.say <- s:
		return nil
	default:
		return ErrChanBufferFull
	}
}

// Run reads packets and emits an event for each received packet
func (b *BNetRealm) Run(ctx context.Context) error {
	go func() {
		<-ctx.Done()
		b.Close()
	}()

	var backoff = b.ReconnectDelay
	for ctx.Err() == nil {
		if backoff > 4*time.Hour {
			backoff = 4 * time.Hour
		}

		var err = b.Client.Logon()
		if err != nil {
			var reconnect bool
			switch err {
			case bnet.ErrCDKeyInUse, bnet.ErrUnexpectedPacket:
				reconnect = true
			default:
				reconnect = network.IsConnClosedError(err) || os.IsTimeout(err)
			}

			if reconnect && ctx.Err() == nil {
				b.Fire(&network.AsyncError{Src: "Run[Logon]", Err: err})
				time.Sleep(backoff)
				backoff = time.Duration(float64(backoff) * 1.5)
				continue
			}

			return err
		}

		b.Fire(Connected{})

		var channel = b.Channel()
		if channel == "" {
			channel = b.FirstChannel
		}
		if channel != "" {
			b.Say("/join " + channel)
		}

		backoff = b.ReconnectDelay
		if err := b.Client.Run(); err != nil && ctx.Err() == nil {
			b.Fire(&network.AsyncError{Src: "Run[Client]", Err: err})
		}

		b.Fire(Disconnected{})
	}

	return ctx.Err()
}

func (b *BNetRealm) channel() Channel {
	var name = b.Channel()
	return Channel{
		ID:   name,
		Name: name,
	}
}

func (b *BNetRealm) user(u *bnet.User) User {
	var res = User{
		Name: u.Name,
		Rank: b.RankTalk,
	}
	if b.RankOperator > res.Rank && u.Operator() {
		res.Rank = b.RankOperator
	}

	var prod, icon, lvl, tag = u.Stat()
	if prod != 0 {
		switch prod {
		case w3gs.ProductDemo, w3gs.ProductROC, w3gs.ProductTFT:
			// Expected
		default:
			res.Rank = b.RankNoWarcraft
			return res
		}

		res.AvatarURL = strings.Replace(b.AvatarURL, "${ICON}", icon.String(), -1)

		if lvl > 0 && b.RankLevel != nil {
			for l, r := range b.RankLevel {
				if lvl >= l && r > res.Rank {
					res.Rank = r
				}
			}
		}
		if tag != 0 && b.RankClanTag != nil {
			var rank = b.RankClanTag[tag.String()]
			if rank > res.Rank {
				res.Rank = rank
			}
		}
	}

	return res
}

// InitDefaultHandlers adds the default callbacks for relevant packets
func (b *BNetRealm) InitDefaultHandlers() {
	b.On(&bnet.UserJoined{}, b.onUserJoined)
	b.On(&bnet.UserLeft{}, b.onUserLeft)
	b.On(&bnet.Chat{}, b.onChat)
	b.On(&bnet.Whisper{}, b.onWhisper)
	b.On(&bnet.Channel{}, b.onChannel)
	b.On(&bnet.SystemMessage{}, b.onSystemMessage)
	b.On(&bncs.FloodDetected{}, b.onFloodDetected)
}

func (b *BNetRealm) onUserJoined(ev *network.Event) {
	var user = ev.Arg.(*bnet.UserJoined)
	if user.Name == b.UniqueName {
		return
	}

	b.Fire(&Join{
		User:    b.user(&user.User),
		Channel: b.channel(),
	})
}

func (b *BNetRealm) onUserLeft(ev *network.Event) {
	var user = ev.Arg.(*bnet.UserLeft)
	if user.Name == b.UniqueName {
		return
	}

	b.Fire(&Leave{
		User:    b.user(&user.User),
		Channel: b.channel(),
	})
}

func (b *BNetRealm) onChat(ev *network.Event) {
	var msg = ev.Arg.(*bnet.Chat)

	var chat = Chat{
		User:    b.user(&msg.User),
		Channel: b.channel(),
		Content: msg.Content,
	}

	switch msg.Type {
	case bncs.ChatEmote:
		chat.Content = fmt.Sprintf("%s %s", msg.User.Name, msg.Content)
	default:
		chat.Content = msg.Content
	}

	b.Fire(&chat)
}

func (b *BNetRealm) onWhisper(ev *network.Event) {
	var msg = ev.Arg.(*bnet.Whisper)

	if msg.Username[:1] == "#" {
		b.Fire(&SystemMessage{Content: fmt.Sprintf("[%s] %s", msg.Username, msg.Content)})
		return
	}

	b.Fire(&PrivateChat{
		User: User{
			Name: msg.Username,
			Rank: b.RankWhisper,
		},
		Content: msg.Content,
	})
}

func (b *BNetRealm) onChannel(ev *network.Event) {
	var c = ev.Arg.(*bnet.Channel)
	b.Fire(&Channel{ID: c.Name, Name: c.Name})
}

func (b *BNetRealm) onSystemMessage(ev *network.Event) {
	var msg = ev.Arg.(*bnet.SystemMessage)

	if msg.Type == bncs.ChatInfo && msg.Content == "No one hears you." {
		return
	}

	b.Fire(&SystemMessage{Content: fmt.Sprintf("[%s] %s", strings.ToUpper(msg.Type.String()), msg.Content)})
}

func (b *BNetRealm) onFloodDetected(ev *network.Event) {
	b.Fire(&SystemMessage{Content: "FLOOD DETECTED"})
}

// Relay dumps the event content in current channel
func (b *BNetRealm) Relay(ev *network.Event, sender string) {
	var err error

	sender = strings.SplitN(sender, RealmDelimiter, 2)[0]

	switch msg := ev.Arg.(type) {
	case Connected:
		err = b.Say(fmt.Sprintf("Established connection to %s", sender))
	case Disconnected:
		err = b.Say(fmt.Sprintf("Connection to %s closed", sender))
	case *Channel:
		err = b.Say(fmt.Sprintf("Joined %s on %s", msg.Name, sender))
	case *SystemMessage:
		err = b.Say(fmt.Sprintf("[%s] %s", sender, msg.Content))
	case *Join:
		err = b.Say(fmt.Sprintf("%s@%s has joined the channel", msg.User.Name, sender))
	case *Leave:
		err = b.Say(fmt.Sprintf("%s@%s has left the channel", msg.User.Name, sender))
	case *Chat:
		err = b.Say(fmt.Sprintf("<%s@%s> %s", msg.User.Name, sender, msg.Content))
	case *PrivateChat:
		err = b.Say(fmt.Sprintf("[DM] <%s@%s> %s", msg.User.Name, sender, msg.Content))
	default:
		err = ErrUnknownEvent
	}

	if err != nil && !network.IsConnClosedError(err) {
		b.Fire(&network.AsyncError{Src: "Relay", Err: err})
	}
}
