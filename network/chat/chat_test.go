// Author:  Niels A.D.
// Project: gowarcraft3 (https://github.com/nielsAD/gowarcraft3)
// License: Mozilla Public License, v2.0

package chat_test

import (
	"fmt"

	"github.com/nielsAD/gowarcraft3/network"
	"github.com/nielsAD/gowarcraft3/network/chat"
	"github.com/nielsAD/gowarcraft3/protocol/capi"
)

func Example() {
	bot, err := chat.NewBot(&chat.Config{
		Endpoint: capi.Endpoint + ".example",
		APIKey:   "12345678901234567890",
	})
	if err != nil {
		fmt.Println(err)
		return
	}
	defer bot.Close()

	// Connect to server
	if err = bot.Connect(); err != nil {
		fmt.Println(err)
		return
	}

	// Print incoming chat messages
	bot.On(&capi.MessageEvent{}, func(ev *network.Event) {
		var msg = ev.Arg.(*capi.MessageEvent)
		fmt.Printf("[%d] %s\n", msg.UserID, msg.Message)
	})

	// Run() blocks until the connection is closed
	bot.Run()
}
