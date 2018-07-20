// Author:  Niels A.D.
// Project: gowarcraft3 (https://github.com/nielsAD/gowarcraft3)
// License: Mozilla Public License, v2.0

package main

import (
	"bufio"
	"context"
	"log"
	"os"
	"strings"
	"unicode"

	"github.com/fatih/color"

	"github.com/nielsAD/gowarcraft3/network"
)

var logOut = log.New(color.Output, "", 0)
var logErr = log.New(color.Error, "", 0)
var stdin = bufio.NewReader(os.Stdin)

// StdIO relays between stdin/stdout
type StdIO struct {
	network.EventEmitter
	*StdIOConfig
}

func (o *StdIO) read() error {
	for {
		line, err := stdin.ReadString('\n')
		if err != nil {
			return err
		}
		line = strings.TrimRightFunc(line, unicode.IsSpace)
		if line == "" {
			continue
		}

		o.Fire(&Chat{
			User: User{
				ID:        "STDIO",
				Name:      "stdin",
				Rank:      o.Rank,
				AvatarURL: o.AvatarURL,
			},
			Channel: Channel{
				ID:   "STDIO",
				Name: "stdin",
			},
			Content: line,
		})
	}
}

// Run reads packets and emits an event for each received packet
func (o *StdIO) Run(ctx context.Context) error {
	if !o.Read {
		return nil
	}

	var res = make(chan error)
	go func() {
		res <- o.read()
	}()

	select {
	case err := <-res:
		return err
	case <-ctx.Done():
		return ctx.Err()
	}
}

// Relay dumps the event content to stdout
func (o *StdIO) Relay(ev *network.Event, sender string) {
	switch msg := ev.Arg.(type) {
	case Connected:
		logOut.Println(color.MagentaString("Established connection to %s", sender))
	case Disconnected:
		logOut.Println(color.MagentaString("Connection to %s closed", sender))
	case *Channel:
		logOut.Println(color.MagentaString("Joined %s on %s", msg.Name, sender))
	case *SystemMessage:
		logOut.Println(color.CyanString("[SYSTEM][%s] %s", sender, msg.Content))
	case *Join:
		logOut.Println(color.YellowString("[CHAT][%s#%s] %s has joined the channel", sender, msg.Channel.Name, msg.User.Name))
	case *Leave:
		logOut.Println(color.YellowString("[CHAT][%s#%s] %s has left the channel", sender, msg.Channel.Name, msg.User.Name))
	case *Chat:
		logOut.Printf("[CHAT][%s#%s] <%s> %s\n", sender, msg.Channel.Name, msg.User.Name, msg.Content)
	case *PrivateChat:
		logOut.Println(color.GreenString("[PRIVATE][%s] <%s> %s", sender, msg.User.Name, msg.Content))
	default:
		o.Fire(&network.AsyncError{Src: "Relay", Err: ErrUnknownEvent})
	}
}
