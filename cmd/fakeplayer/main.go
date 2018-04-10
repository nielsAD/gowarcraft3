// fakeplayer is a mocked Warcraft 3 game client that can be used to add dummy players to games.
package main

import (
	"flag"
	"log"
	"net"
	"os"
	"strings"

	"github.com/nielsAD/noot/pkg/w3gs"
)

var (
	verbose = flag.Bool("verbose", false, "Print contents of all packets")

	lan      = flag.Bool("lan", false, "Find a game on LAN")
	gametft  = flag.Bool("tft", true, "Search for TFT or ROC games (only used when searching local)")
	gamevers = flag.Int("v", 28, "Game version (only used when searching local)")
	entrykey = flag.Uint("e", 0, "Entry key (only used when entering local game)")

	hostcounter = flag.Uint("c", 1, "Host counter")
	playername  = flag.String("n", "fakeplayer", "Player name")
	listen      = flag.Int("l", 0, "Listen on port (0 to pick automatically)")
)

var logger = log.New(os.Stdout, "", log.Ltime)

func main() {
	flag.Parse()

	var err error
	var addr *net.TCPAddr
	var hc = uint32(*hostcounter)
	var ek = uint32(*entrykey)

	if *lan {
		addr, hc, ek, err = findGameOnLAN(&w3gs.GameVersion{TFT: *gametft, Version: uint32(*gamevers)})
	} else {
		var address = strings.Join(flag.Args(), " ")
		if address == "" {
			address = "127.0.0.1:6112"
		}
		addr, err = net.ResolveTCPAddr("tcp4", address)
	}
	if err != nil {
		logger.Fatal(err)
	}

	f, err := JoinLobby(addr, *playername, hc, ek, *listen)
	if err != nil {
		logger.Fatal(err)
	}

	f.Run()
	f.Wait()
}
