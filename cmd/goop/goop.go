// Author:  Niels A.D.
// Project: gowarcraft3 (https://github.com/nielsAD/gowarcraft3)
// License: Mozilla Public License, v2.0

package main

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"github.com/nielsAD/gowarcraft3/network"
)

// Errors
var (
	ErrUnkownRealm      = errors.New("goop: Unknown realm")
	ErrUnknownEvent     = errors.New("goop: Unknown event")
	ErrUnknownConfigKey = errors.New("goop: Unknown config key")
	ErrInvalidType      = errors.New("goop: Type mismatch")
	ErrChanBufferFull   = errors.New("goop: Channel buffer full")
)

// Goop main
type Goop struct {
	network.EventEmitter

	// Read-only
	Realms map[string]Realm
	Config Config
}

// New initializes a Goop struct
func New(conf *Config) (*Goop, error) {
	var g = Goop{
		Config: *conf,
		Realms: map[string]Realm{
			"STDIO": &StdIO{StdIOConfig: &conf.StdIO},
		},
	}

	var realms = []string{"STDIO"}

	for k, r := range g.Config.BNet.Realms {
		realm, err := NewBNetRealm(r)
		if err != nil {
			return nil, err
		}

		g.Realms[k] = realm
		realms = append(realms, k)
	}

	for k, r := range g.Config.Discord.Sessions {
		discord, err := NewDiscordSession(r)
		if err != nil {
			return nil, err
		}

		g.Realms[k] = discord
		realms = append(realms, k)

		for cid, c := range discord.Channels {
			var idx = k + RealmDelimiter + cid
			g.Realms[idx] = c
			realms = append(realms, idx)
		}
	}

	for i := 0; i < len(g.Config.Relay); i++ {
		var r = g.Config.Relay[i]
		if r.In == nil {
			r.In = realms
		}
		if r.Out == nil {
			r.Out = realms
		}
		for _, in := range r.In {
			var r1 = g.Realms[in]
			if r1 == nil {
				return nil, ErrUnkownRealm
			}

			for _, out := range r.Out {
				var r2 = g.Realms[out]
				if r2 == nil {
					return nil, ErrUnkownRealm
				}
				if r1 == r2 {
					continue
				}

				var sender = in
				var handler = func(ev *network.Event) { r2.Relay(ev, sender) }

				if r.Log {
					r1.On(Connected{}, handler)
					r1.On(Disconnected{}, handler)
					r1.On(&Channel{}, handler)
				}
				if r.System {
					r1.On(&SystemMessage{}, handler)
				}

				if r.Joins {
					r1.On(&Join{}, func(ev *network.Event) {
						var user = ev.Arg.(*Join)
						if user.Rank < r.JoinRank {
							return
						}
						r2.Relay(ev, sender)
					})
					r1.On(&Leave{}, func(ev *network.Event) {
						var user = ev.Arg.(*Leave)
						if user.Rank < r.JoinRank {
							return
						}
						r2.Relay(ev, sender)
					})
				}

				if r.Chat {
					r1.On(&Chat{}, func(ev *network.Event) {
						var msg = ev.Arg.(*Chat)
						if msg.User.Rank < r.ChatRank {
							return
						}
						r2.Relay(ev, sender)
					})
				}

				if r.PrivateChat {
					r1.On(&PrivateChat{}, func(ev *network.Event) {
						var msg = ev.Arg.(*PrivateChat)
						if msg.User.Rank < r.PrivateChatRank {
							return
						}
						r2.Relay(ev, sender)
					})
				}
			}
		}
	}

	return &g, nil
}

// Run connects to each realm and returns when all connections have ended
func (g *Goop) Run(ctx context.Context) {
	var wg sync.WaitGroup
	for i := range g.Realms {
		wg.Add(1)

		var k = i
		var r = g.Realms[k]
		go func() {
			if err := r.Run(ctx); err != nil && err != context.Canceled {
				g.Fire(&network.AsyncError{Src: fmt.Sprintf("Run[realm:%s]", k), Err: err})
			}
			wg.Done()
		}()
	}

	wg.Wait()
}
