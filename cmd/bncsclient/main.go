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

	"github.com/fatih/color"
	"golang.org/x/crypto/ssh/terminal"

	"github.com/nielsAD/gowarcraft3/network"
	"github.com/nielsAD/gowarcraft3/network/bnet"
	"github.com/nielsAD/gowarcraft3/protocol/bncs"
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

var logOut = log.New(color.Output, "", log.Ltime)
var logErr = log.New(color.Error, "", log.Ltime)
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
		logErr.Println(color.RedString("[ERROR] %s", err.Error()))
	})
	c.On(&bnet.JoinError{}, func(ev *network.Event) {
		var err = ev.Arg.(*bnet.JoinError)
		logErr.Println(color.RedString("[ERROR] Could not join %s: %v", err.Channel, err.Error))
	})
	c.On(&bnet.Channel{}, func(ev *network.Event) {
		var channel = ev.Arg.(*bnet.Channel)
		logOut.Println(color.MagentaString("Joined channel '%s'", channel.Name))
	})
	c.On(&bnet.UserJoined{}, func(ev *network.Event) {
		var user = ev.Arg.(*bnet.UserJoined)
		logOut.Println(color.YellowString("%s has joined the channel (ping: %dms)", user.Name, user.Ping))
	})
	c.On(&bnet.UserLeft{}, func(ev *network.Event) {
		var user = ev.Arg.(*bnet.UserLeft)
		logOut.Println(color.YellowString("%s has left the channel (after %dm)", user.Name, int(time.Now().Sub(user.Joined).Minutes())))
	})
	c.On(&bnet.Whisper{}, func(ev *network.Event) {
		var msg = ev.Arg.(*bnet.Whisper)
		logOut.Println(color.GreenString("[WHISPER] %s: %s", msg.UserName, msg.Content))
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
		logOut.Println(color.CyanString("[%s] %s", strings.ToUpper(msg.Type.String()), msg.Content))
	})
	c.On(&bncs.FloodDetected{}, func(ev *network.Event) {
		logErr.Println(color.RedString("[ERROR] Flood detected!"))
	})

	if *create {
		if err := c.CreateAccount(); err != nil {
			logErr.Fatal("CreateAccount error: ", err)
		}
		logOut.Println(color.MagentaString("Succesfully registered new account '%s'", c.UserName))
		return
	}

	if err := c.Logon(); err != nil {
		logErr.Fatal("Logon error: ", err)
	}

	logOut.Println(color.MagentaString("Succesfully logged onto %s@%s", c.UserName, c.ServerAddr))

	go func() {
		for {
			line, err := stdin.ReadString('\n')
			if err != nil {
				c.Close()
				break
			}

			if err := c.Say(line); err != nil {
				logErr.Println(color.RedString("[ERROR] %s", err.Error()))
			}
		}
	}()

	c.Run()
}
