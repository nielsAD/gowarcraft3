// Author:  Niels A.D.
// Project: gowarcraft3 (https://github.com/nielsAD/gowarcraft3)
// License: Mozilla Public License, v2.0

// bncsclient is a mocked Warcraft III chat client that can be used to connect to BNCS servers.
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
	binpath     = flag.String("b", "", "Path to game binaries")
	exeinfo     = flag.String("ei", "", "Exe version string")
	exevers     = flag.Uint("ev", 0, "Exe version number")
	exehash     = flag.Uint("eh", 0, "Exe hash")
	keyroc      = flag.String("roc", "", "ROC CD-key")
	keytft      = flag.String("tft", "", "TFT CD-key")
	username    = flag.String("u", "", "Username")
	password    = flag.String("p", "", "Password")
	newpassword = flag.String("np", "", "New password")
	verify      = flag.Bool("verify", false, "Verify server signature")
	sha1        = flag.Bool("sha1", false, "SHA1 password authentication (used in old PvPGN servers)")
	create      = flag.Bool("create", false, "Create account")
	changepass  = flag.Bool("changepass", false, "Change password")
)

var logOut = log.New(color.Output, "", log.Ltime)
var logErr = log.New(color.Error, "", log.Ltime)
var stdin = bufio.NewReader(os.Stdin)

func main() {
	flag.Parse()

	c, err := bnet.NewClient(&bnet.Config{
		BinPath:         *binpath,
		ExeInfo:         *exeinfo,
		ExeVersion:      uint32(*exevers),
		ExeHash:         uint32(*exehash),
		VerifySignature: *verify,
		SHA1Auth:        *sha1,
	})
	if err != nil {
		logErr.Fatal("NewClient error: ", err)
	}

	c.ServerAddr = strings.Join(flag.Args(), ":")
	if c.ServerAddr == "" {
		c.ServerAddr = "uswest.battle.net:6112"
	}

	if *keyroc != "" {
		if *keytft != "" {
			c.Platform.GameVersion.Product = w3gs.ProductTFT
			c.CDKeys = []string{*keyroc, *keytft}
		} else {
			c.Platform.GameVersion.Product = w3gs.ProductROC
			c.CDKeys = []string{*keyroc}
		}
	}

	c.Username = *username
	c.Password = *password

	if *username == "" {
		fmt.Print("Enter username: ")
		c.Username, _ = stdin.ReadString('\n')
		c.Username = strings.TrimSpace(c.Username)
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
		logOut.Println(color.GreenString("[WHISPER] %s: %s", msg.Username, msg.Content))
	})
	c.On(&bnet.Chat{}, func(ev *network.Event) {
		var msg = ev.Arg.(*bnet.Chat)
		logOut.Printf("[%s] %s: %s\n", strings.ToUpper(msg.Type.String()), msg.User.Name, msg.Content)
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
		logOut.Println(color.MagentaString("Succesfully registered new account '%s'", c.Username))
		return
	}

	if *changepass {
		var pass = *newpassword
		if pass == "" {
			fmt.Print("Enter new password: ")
			if b, err := terminal.ReadPassword(int(os.Stdin.Fd())); err == nil {
				pass = string(b)
			} else {
				logErr.Fatal("ReadPassword error: ", err)
			}
			fmt.Println()
		}

		if err := c.ChangePassword(pass); err != nil {
			logErr.Fatal("ChangePassword error: ", err)
		}
		logOut.Println(color.MagentaString("Succesfully changed password"))
		return
	}

	if err := c.Logon(); err != nil {
		logErr.Fatal("Logon error: ", err)
	}

	logOut.Println(color.MagentaString("Succesfully logged onto %s@%s", c.Username, c.ServerAddr))

	go func() {
		for {
			line, err := stdin.ReadString('\n')
			if err != nil {
				c.Close()
				break
			}

			if err := c.Say(strings.TrimRight(line, "\r\n")); err != nil {
				logErr.Println(color.RedString("[ERROR] %s", err.Error()))
			}
		}
	}()

	c.Run()
}
