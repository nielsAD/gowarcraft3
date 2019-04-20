// Author:  Niels A.D.
// Project: gowarcraft3 (https://github.com/nielsAD/gowarcraft3)
// License: Mozilla Public License, v2.0

package bnet_test

import (
	"fmt"

	"github.com/nielsAD/gowarcraft3/network"
	"github.com/nielsAD/gowarcraft3/network/bnet"
)

func Example() {
	client, err := bnet.NewClient(&bnet.Config{
		ServerAddr: "europe.battle.net.example",
		Username:   "gowarcraft3",
		Password:   "gowarcraft3",
	})
	if err != nil {
		fmt.Println(err)
		return
	}
	defer client.Close()

	// Log on
	if err = client.Logon(); err != nil {
		fmt.Println(err)
		return
	}

	// Print incoming chat messages
	client.On(&bnet.Chat{}, func(ev *network.Event) {
		var msg = ev.Arg.(*bnet.Chat)
		fmt.Printf("[%s] %s\n", msg.User.Name, msg.Content)
	})

	// Run() blocks until the connection is closed
	client.Run()
}
