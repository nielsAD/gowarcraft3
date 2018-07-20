// Author:  Niels A.D.
// Project: gowarcraft3 (https://github.com/nielsAD/gowarcraft3)
// License: Mozilla Public License, v2.0

package main

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/bwmarrin/discordgo"

	"github.com/nielsAD/gowarcraft3/network"
)

// DiscordSession manages a Discord connection
type DiscordSession struct {
	network.EventEmitter
	*discordgo.Session
	*DiscordConfig

	Channels map[string]*DiscordChannel
}

// DiscordChannel manages a Discord channel
type DiscordChannel struct {
	network.EventEmitter
	*DiscordChannelConfig

	wg      *sync.WaitGroup
	id      string
	session *discordgo.Session
}

// NewDiscordSession initializes a new DiscordRealm struct
func NewDiscordSession(conf *DiscordConfig) (*DiscordSession, error) {
	s, err := discordgo.New("Bot " + conf.AuthToken)
	if err != nil {
		return nil, err
	}

	s.SyncEvents = true
	s.State.TrackEmojis = false
	s.State.TrackVoice = false
	s.State.MaxMessageCount = 0

	var d = DiscordSession{
		Session:       s,
		DiscordConfig: conf,
		Channels:      make(map[string]*DiscordChannel),
	}

	var wg sync.WaitGroup
	wg.Add(1)
	d.Once(Connected{}, func(ev *network.Event) {
		go func() {
			time.Sleep(time.Second)
			wg.Done()
		}()
	})

	for id, c := range d.DiscordConfig.Channels {
		d.Channels[id] = &DiscordChannel{
			DiscordChannelConfig: c,

			wg:      &wg,
			id:      id,
			session: s,
		}
	}

	d.InitDefaultHandlers()

	return &d, nil
}

// Run reads packets and emits an event for each received packet
func (d *DiscordSession) Run(ctx context.Context) error {
	var err error
	for i := 1; i < 60 && ctx.Err() == nil; i++ {
		err = d.Session.Open()
		if err == nil {
			break
		}

		d.Fire(&network.AsyncError{Src: "Run[Open]", Err: err})

		select {
		case <-time.After(2 * time.Minute):
		case <-ctx.Done():
		}
	}

	if err != nil {
		return err
	}

	<-ctx.Done()
	d.Close()

	return ctx.Err()
}

// InitDefaultHandlers adds the default callbacks for relevant packets
func (d *DiscordSession) InitDefaultHandlers() {
	d.AddHandler(d.onConnect)
	d.AddHandler(d.onDisconnect)
	d.AddHandler(d.onPresenceUpdate)
	d.AddHandler(d.onMessageCreate)
}

func (d *DiscordSession) onConnect(s *discordgo.Session, msg *discordgo.Connect) {
	if d.Presence != "" {
		go func() {
			time.Sleep(time.Second)
			if err := s.UpdateStatus(0, d.Presence); err != nil {
				d.Fire(&network.AsyncError{Src: "onConnect[UpdateStatus]", Err: err})
			}
		}()
	}
	d.Fire(Connected{})
}

func (d *DiscordSession) onDisconnect(s *discordgo.Session, msg *discordgo.Disconnect) {
	d.Fire(Disconnected{})
}

func (d *DiscordSession) onPresenceUpdate(s *discordgo.Session, msg *discordgo.PresenceUpdate) {
	old, _ := d.Session.State.Presence(msg.GuildID, msg.User.ID)
	if old == nil || msg.Presence.Status != old.Status {
		fmt.Println(msg)
	}
}

func (d *DiscordSession) onMessageCreate(s *discordgo.Session, msg *discordgo.MessageCreate) {
	if msg.Author.Bot {
		return
	}

	var chat = Chat{
		User: User{
			ID:        msg.Author.ID,
			Name:      msg.Author.Username,
			Rank:      d.RankNoChannel,
			AvatarURL: msg.Author.AvatarURL(""),
		},
		Channel: Channel{
			ID:   msg.ChannelID,
			Name: msg.ChannelID,
		},
		Content: msg.Content,
	}

	if c, err := msg.ContentWithMoreMentionsReplaced(s); err == nil {
		chat.Content = c
	}

	var channel = d.Channels[msg.ChannelID]
	if channel != nil {
		chat.User.Rank = channel.RankTalk
	}

	if ch, err := s.State.Channel(msg.ChannelID); err == nil {
		if channel == nil && ch.Type == discordgo.ChannelTypeDM {
			chat.User.Rank = d.RankDM
			d.Fire(&PrivateChat{
				User:    chat.User,
				Content: chat.Content,
			})
			return
		}

		if g, err := s.State.Guild(ch.GuildID); err == nil {
			chat.Channel.Name = fmt.Sprintf("%s.%s", g.Name, ch.Name)
		} else {
			chat.Channel.Name = ch.Name
		}

		if channel != nil && channel.RankRole != nil {
			if m, err := s.State.Member(ch.GuildID, msg.Author.ID); err == nil {
				for _, rid := range m.Roles {
					if r, err := s.State.Role(ch.GuildID, rid); err == nil {
						var rank, ok = channel.RankRole[strings.ToLower(r.Name)]
						if ok {
							chat.User.Rank = rank
						}
					}
				}
			}
		}
	}

	if channel != nil {
		var rank, ok = channel.RankUser[msg.Author.ID]
		if ok {
			chat.User.Rank = rank
		}

		channel.Fire(&chat)
	} else {
		d.Fire(&chat)
	}
}

// Relay placeholder to implement Realm interface
// Events should instead be relayed directly to a DiscordChannel
func (d *DiscordSession) Relay(ev *network.Event, sender string) {
}

// Run placeholder to implement Realm interface
func (c *DiscordChannel) Run(ctx context.Context) error {
	var done = make(chan struct{})

	go func() {
		c.wg.Wait()
		var name = c.id
		if ch, err := c.session.State.Channel(c.id); err == nil {
			if g, err := c.session.State.Guild(ch.GuildID); err == nil {
				name = fmt.Sprintf("%s.%s", g.Name, ch.Name)
			} else {
				name = ch.Name
			}
		}
		c.Fire(&Channel{ID: c.id, Name: name})
		done <- struct{}{}
	}()

	select {
	case <-done:
	case <-ctx.Done():
	}

	return nil
}

// Say sends a chat message
func (c *DiscordChannel) Say(s string) error {
	_, err := c.session.ChannelMessageSend(c.id, s)
	return err
}

// WebhookOrSay sends a chat message preferably via webhook
func (c *DiscordChannel) WebhookOrSay(p *discordgo.WebhookParams) error {
	if c.Webhook != "" {
		_, err := c.session.RequestWithBucketID("POST", c.Webhook, p, discordgo.EndpointWebhookToken("", ""))
		return err
	}

	var s = p.Content
	if p.Username != "" {
		s = fmt.Sprintf("**<%s>** %s", p.Username, p.Content)
	}
	return c.Say(s)
}

func (c *DiscordChannel) filter(s string, r Rank) string {
	if r < c.RankMentions {
		s = strings.Replace(s, "@", "@"+string('\u200B'), -1)
	}
	return s
}

// Relay dumps the event content in channel
func (c *DiscordChannel) Relay(ev *network.Event, sender string) {
	var err error

	sender = strings.SplitN(sender, RealmDelimiter, 2)[0]

	switch msg := ev.Arg.(type) {
	case Connected:
		err = c.Say(fmt.Sprintf("*Established connection to %s*", sender))
	case Disconnected:
		err = c.Say(fmt.Sprintf("*Connection to %s closed*", sender))
	case *Channel:
		err = c.Say(fmt.Sprintf("*Joined %s on %s*", msg.Name, sender))
	case *SystemMessage:
		err = c.Say(fmt.Sprintf("📢 **%s** %s", sender, msg.Content))
	case *Join:
		err = c.Say(fmt.Sprintf("➡️ **%s@%s** has joined the channel", msg.User.Name, sender))
	case *Leave:
		err = c.Say(fmt.Sprintf("⬅️ **%s@%s** has left the channel", msg.User.Name, sender))
	case *Chat:
		err = c.WebhookOrSay(&discordgo.WebhookParams{
			Content:   c.filter(msg.Content, msg.User.Rank),
			Username:  fmt.Sprintf("%s@%s", msg.User.Name, sender),
			AvatarURL: msg.User.AvatarURL,
		})
	case *PrivateChat:
		err = c.WebhookOrSay(&discordgo.WebhookParams{
			Content:   c.filter(msg.Content, msg.User.Rank),
			Username:  fmt.Sprintf("%s@%s (Direct Message)", msg.User.Name, sender),
			AvatarURL: msg.User.AvatarURL,
		})
	default:
		err = ErrUnknownEvent
	}

	if err != nil && !network.IsConnClosedError(err) {
		c.Fire(&network.AsyncError{Src: "Relay", Err: err})
	}
}
