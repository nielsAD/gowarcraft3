// Author:  Niels A.D.
// Project: gowarcraft3 (https://github.com/nielsAD/gowarcraft3)
// License: Mozilla Public License, v2.0

package dummy_test

import (
	"context"
	"fmt"
	"time"

	"github.com/nielsAD/gowarcraft3/network"
	"github.com/nielsAD/gowarcraft3/network/dummy"
	"github.com/nielsAD/gowarcraft3/network/lan"
	"github.com/nielsAD/gowarcraft3/protocol/w3gs"
)

func Example() {
	var game = w3gs.GameVersion{
		Product: w3gs.ProductTFT,
		Version: w3gs.CurrentGameVersion,
	}

	// Find LAN game
	ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond)
	addr, id, secret, err := lan.FindGame(ctx, game)

	cancel()
	if err != nil {
		fmt.Println(err)
		return
	}

	// Join game
	player, err := dummy.Join(addr, "DummyPlayer", id, secret, -1, w3gs.Encoding{GameVersion: game.Version})
	if err != nil {
		fmt.Println(err)
		return
	}
	defer player.Close()

	// Print incoming chat messages
	player.On(&dummy.Chat{}, func(ev *network.Event) {
		var msg = ev.Arg.(*dummy.Chat)
		fmt.Printf("[%s] %s\n", msg.Sender.PlayerName, msg.Content)
	})

	// Run() blocks until the connection is closed
	player.Run()
}
