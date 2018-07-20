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

	"github.com/BurntSushi/toml"
	"github.com/fatih/color"

	"github.com/nielsAD/gowarcraft3/network"
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

	var conf = DefaultConfig
	for _, f := range flag.Args() {
		md, err := toml.DecodeFile(f, &conf)
		if err != nil {
			logErr.Fatal("Error reading configuration: ", err)
		}
		uk := md.Undecoded()
		if len(uk) > 0 {
			logErr.Printf("Undecoded configuration keys: %v\n", uk)
		}
	}

	if err := conf.MergeDefaults(); err != nil {
		logErr.Fatal("Merging defaults error: ", err)
	}

	g, err := New(&conf)
	if err != nil {
		logErr.Fatal("Initialization error: ", err)
	}

	if *makeconf {
		var m = conf.Map()
		if err := toml.NewEncoder(os.Stdout).Encode(m); err != nil {
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
