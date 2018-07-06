// Author:  Niels A.D.
// Project: gowarcraft3 (https://github.com/nielsAD/gowarcraft3)
// License: Mozilla Public License, v2.0

// bncsclient is a mocked Warcraft 3 chat client that can be used to connect to BNCS servers.
package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"golang.org/x/crypto/ssh/terminal"

	"github.com/nielsAD/gowarcraft3/network"
	"github.com/nielsAD/gowarcraft3/network/bnet"
	"github.com/nielsAD/gowarcraft3/protocol/w3gs"
)

var (
	gamevers = flag.Int("v", 0, "Game version")
	binpath  = flag.String("b", "", "Path to game binaries")
	keyroc   = flag.String("roc", "", "ROC CD-key")
	keytft   = flag.String("tft", "", "TFT CD-key")
	username = flag.String("u", "", "Username")
	password = flag.String("p", "", "Password")
	create   = flag.Bool("c", false, "Create account")
)

var logOut = log.New(os.Stdout, "", log.Ltime)
var logErr = log.New(os.Stderr, "", log.Ltime)
var stdin = bufio.NewReader(os.Stdin)

func main() {
	flag.Parse()

	var c = bnet.NewClient(*binpath)

	c.ServerAddr = strings.Join(flag.Args(), " ")
	if c.ServerAddr == "" {
		c.ServerAddr = "uswest.battle.net:6112"
	}

	if *gamevers != 0 {
		c.AuthInfo.GameVersion.Version = uint32(*gamevers)
	}

	if *keyroc != "" {
		if *keytft != "" {
			c.AuthInfo.GameVersion.Product = w3gs.ProductTFT
			c.CDKeys = []string{*keyroc, *keytft}
		} else {
			c.AuthInfo.GameVersion.Product = w3gs.ProductROC
			c.CDKeys = []string{*keyroc}
		}
	}

	c.UserName = *username
	c.Password = *password

	if *username == "" {
		fmt.Print("Enter username: ")
		c.UserName, _ = stdin.ReadString('\n')
		c.UserName = strings.TrimSpace(c.UserName)
	}

	if *password == "" {
		fmt.Print("Enter password: ")
		if b, err := terminal.ReadPassword(int(os.Stdin.Fd())); err == nil {
			c.Password = string(b)
		} else {
			logErr.Fatal("ReadPassword error: ", err)
		}
		fmt.Println()
	}

	c.On(&network.AsyncError{}, func(ev *network.Event) {
		var err = ev.Arg.(*network.AsyncError)
		logErr.Printf("[ERROR] %s\n", err.Error())
	})
	c.On(&bnet.JoinError{}, func(ev *network.Event) {
		var err = ev.Arg.(*bnet.JoinError)
		logErr.Printf("[ERROR] Could not join %s: %v\n", err.Channel, err.Error)
	})
	c.On(&bnet.Channel{}, func(ev *network.Event) {
		var channel = ev.Arg.(*bnet.Channel)
		logOut.Printf("Joined channel '%s'\n", channel.Name)
	})
	c.On(&bnet.UserJoined{}, func(ev *network.Event) {
		var user = ev.Arg.(*bnet.UserJoined)
		logOut.Printf("%s has joined the channel (ping: %dms)\n", user.Name, user.Ping)
	})
	c.On(&bnet.UserLeft{}, func(ev *network.Event) {
		var user = ev.Arg.(*bnet.UserLeft)
		logOut.Printf("%s has left the channel (after %dm)\n", user.Name, int(time.Now().Sub(user.Joined).Minutes()))
	})
	c.On(&bnet.Whisper{}, func(ev *network.Event) {
		var msg = ev.Arg.(*bnet.Whisper)
		logOut.Printf("[WHISPER] %s: %s\n", msg.UserName, msg.Content)
	})
	c.On(&bnet.Chat{}, func(ev *network.Event) {
		var msg = ev.Arg.(*bnet.Chat)
		logOut.Printf("[%s] %s: %s\n", strings.ToUpper(msg.Type.String()), msg.User.Name, msg.Content)
	})
	c.On(&bnet.Say{}, func(ev *network.Event) {
		var say = ev.Arg.(*bnet.Say)
		if say.Content[0] == '/' {
			logOut.Printf("[CHAT] %s\n", say.Content)
		} else {

			logOut.Printf("[CHAT] %s: %s\n", c.UserName, say.Content)
		}
	})
	c.On(&bnet.SystemMessage{}, func(ev *network.Event) {
		var msg = ev.Arg.(*bnet.SystemMessage)
		logOut.Printf("[%s] %s\n", strings.ToUpper(msg.Type.String()), msg.Content)
	})

	if *create {
		if err := c.CreateAccount(); err != nil {
			logErr.Fatal("CreateAccount error: ", err)
		}
		logOut.Printf("Succesfully registered new account '%s'\n", c.UserName)
		return
	}

	if err := c.Logon(); err != nil {
		logErr.Fatal("Logon error: ", err)
	}

	logOut.Printf("Succesfully logged onto %s@%s\n", c.UserName, c.ServerAddr)

	go func() {
		for {
			line, err := stdin.ReadString('\n')
			if err != nil {
				c.Close()
				break
			}

			if err := c.Say(line); err != nil {
				logErr.Printf("[ERROR] %s\n", err.Error())
			}
		}
	}()

	c.Run()
}
