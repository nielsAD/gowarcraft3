// Author:  Niels A.D.
// Project: gowarcraft3 (https://github.com/nielsAD/gowarcraft3)
// License: Mozilla Public License, v2.0

package lan_test

import (
	"context"
	"fmt"
	"time"

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

	fmt.Printf("Found game at %s (id: %d, secret %d)\n", addr, id, secret)
}
