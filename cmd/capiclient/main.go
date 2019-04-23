// Author:  Niels A.D.
// Project: gowarcraft3 (https://github.com/nielsAD/gowarcraft3)
// License: Mozilla Public License, v2.0

// capiclient is a command-line interface for the official classic Battle.net chat API.
package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/fatih/color"
	"github.com/nielsAD/gowarcraft3/network"
	"github.com/nielsAD/gowarcraft3/network/chat"
	"github.com/nielsAD/gowarcraft3/protocol/capi"
	"golang.org/x/crypto/ssh/terminal"
)

var (
	endpoint = flag.String("e", capi.Endpoint, "Endpoint")
	apikey   = flag.String("k", "", "API Key")
)

var logOut = log.New(color.Output, "", log.Ltime)
var logErr = log.New(color.Error, "", log.Ltime)
var stdin = bufio.NewReader(os.Stdin)

func main() {
	flag.Parse()

	if *apikey == "" {
		fmt.Print("Enter API key: ")
		if b, err := terminal.ReadPassword(int(os.Stdin.Fd())); err == nil {
			*apikey = string(b)
		} else {
			logErr.Fatal("ReadPassword error: ", err)
		}
		fmt.Println()
	}

	b, err := chat.NewBot(&chat.Config{
		Endpoint: *endpoint,
		APIKey:   *apikey,
	})
	if err != nil {
		logErr.Fatal("NewBot error: ", err)
	}

	b.On(&network.AsyncError{}, func(ev *network.Event) {
		var err = ev.Arg.(*network.AsyncError)
		logErr.Println(color.RedString("[ERROR] %s", err.Error()))
	})
	b.On(&capi.ConnectEvent{}, func(ev *network.Event) {
		var event = ev.Arg.(*capi.ConnectEvent)
		logOut.Println(color.MagentaString("Joined channel '%s'", event.Channel))
	})
	b.On(&capi.UserUpdateEvent{}, func(ev *network.Event) {
		var event = ev.Arg.(*capi.UserUpdateEvent)
		logOut.Println(color.YellowString("%s has been updated (%+v)", event.Username, event))
	})
	b.On(&capi.UserLeaveEvent{}, func(ev *network.Event) {
		var event = ev.Arg.(*capi.UserLeaveEvent)
		if u, ok := b.User(event.UserID); ok {
			logOut.Println(color.YellowString("%s has left the channel (after %dm)", u.Username, int(time.Now().Sub(u.Joined).Minutes())))
		} else {
			logOut.Println(color.YellowString("%s has left the channel", event.UserID))
		}
	})
	b.On(&capi.MessageEvent{}, func(ev *network.Event) {
		var event = ev.Arg.(*capi.MessageEvent)
		if u, ok := b.User(event.UserID); ok {
			logOut.Printf("[%s] %s: %s\n", strings.ToUpper(event.Type.String()), u.Username, event.Message)
		} else {
			logOut.Printf("[%s] %s\n", strings.ToUpper(event.Type.String()), event.Message)
		}
	})

	if err := b.Connect(); err != nil {
		logErr.Fatal("Connect error: ", err)
	}

	logOut.Println(color.MagentaString("Succesfully connected to %s", *endpoint))

	go func() {
		for {
			line, err := stdin.ReadString('\n')
			if err != nil {
				b.Close()
				break
			}

			if err := b.SendMessage(strings.TrimRight(line, "\r\n")); err != nil {
				logErr.Println(color.RedString("[ERROR] %s", err.Error()))
			}
		}
	}()

	if err := b.Run(); err != nil && !network.IsCloseError(err) {
		logErr.Println(color.RedString("[ERROR] %s", err.Error()))
	}
}
