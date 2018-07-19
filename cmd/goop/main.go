// Author:  Niels A.D.
// Project: gowarcraft3 (https://github.com/nielsAD/gowarcraft3)
// License: Mozilla Public License, v2.0

// Goop (GO OPerator) is a BNet Channel Operator.
package main

import (
	"context"
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/fatih/color"

	"github.com/nielsAD/gowarcraft3/network"
	"github.com/nielsAD/gowarcraft3/network/bnet"
)

var (
	makeconf = flag.Bool("makeconf", false, "Generate a configuration file")
	logtime  = flag.Bool("logtime", true, "Prepend log output with time")
)

func main() {
	flag.Parse()

	if *logtime {
		logOut.SetFlags(log.Ltime)
		logErr.SetFlags(log.Ltime)
	}

	var conf = &Config{
		StdIO: StdIOConfig{
			Rank:           RankOwner,
			CommandTrigger: "/",
		},
		BNet: BNetConfigWithDefault{
			Default: BNetConfig{
				bnetConfig: bnetConfig{
					ReconnectDelay: 30 * time.Second,
					CommandTrigger: "!",
				},
				Config: bnet.Config{
					CDKeyOwner: "goop",
				},
			},
		},
		Discord: DiscordConfigWithDefault{
			Default: DefaultDiscordConfig{
				DiscordConfig: DiscordConfig{
					Presence:      "Battle.net",
					RankNoChannel: RankIgnore,
				},
				DiscordChannelConfig: DiscordChannelConfig{
					CommandTrigger: "!",
					RankMentions:   RankWhitelist,
				},
			},
		},
	}
	for _, f := range flag.Args() {
		md, err := toml.DecodeFile(f, conf)
		if err != nil {
			logErr.Fatal("Error reading configuration: ", err)
		}
		uk := md.Undecoded()
		if len(uk) > 0 {
			logErr.Printf("Undecoded configuration keys: %v\n", uk)
		}
	}

	g, err := New(conf)
	if err != nil {
		logErr.Fatal("Initialization error: ", err)
	}

	if *makeconf {
		if err := toml.NewEncoder(os.Stdout).Encode(conf); err != nil {
			logErr.Fatal("Configuration encoding error: ", err)
		}
		return
	}

	g.On(&network.AsyncError{}, func(ev *network.Event) {
		var err = ev.Arg.(*network.AsyncError)
		logErr.Println(color.RedString("[ERROR] %s", err.Error()))
	})

	for i, r := range g.Realms {
		var k = i
		r.On(&network.AsyncError{}, func(ev *network.Event) {
			var err = ev.Arg.(*network.AsyncError)
			logErr.Println(color.RedString("[ERROR][%s] %s", k, err.Error()))
		})
	}

	var ctx, cancel = context.WithCancel(context.Background())
	go func() {
		sig := make(chan os.Signal, 1)
		signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
		<-sig
		cancel()
	}()

	g.Run(ctx)
}
