// fakeplayer is a mocked Warcraft 3 game client that can be used to add dummy players to games.
package main

import (
	"flag"
	"log"
	"net"
	"os"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/nielsAD/gowarcraft3/fakeplayer"
	"github.com/nielsAD/gowarcraft3/protocol/w3gs"
)

var (
	verbose = flag.Bool("verbose", false, "Print contents of all packets")

	lan      = flag.Bool("lan", false, "Find a game on LAN")
	gametft  = flag.Bool("tft", true, "Search for TFT or ROC games (only used when searching local)")
	gamevers = flag.Int("v", 29, "Game version (only used when searching local)")
	entrykey = flag.Uint("e", 0, "Entry key (only used when entering local game)")

	hostcounter = flag.Uint("c", 1, "Host counter")
	playername  = flag.String("n", "fakeplayer", "Player name")
	listen      = flag.Int("l", 0, "Listen on port (0 to pick automatically)")
	dialpeers   = flag.Bool("dial", true, "Dial peers")
)

var logger = log.New(os.Stdout, "", log.Ltime)

func main() {
	flag.Parse()

	var err error
	var addr *net.TCPAddr
	var hc = uint32(*hostcounter)
	var ek = uint32(*entrykey)

	if *lan {
		var gv = w3gs.ProductTFT
		if !*gametft {
			gv = w3gs.ProductROC
		}
		addr, hc, ek, err = fakeplayer.FindGameOnLAN(&w3gs.GameVersion{Product: gv, Version: uint32(*gamevers)})
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

	f, err := fakeplayer.JoinLobby(addr, *playername, hc, ek, *listen)
	if err != nil {
		logger.Fatal(err)
	}

	logger.Println("[HOST] Joined lobby")

	f.DialPeers = *dialpeers

	f.OnPeerConnected = func(peer *fakeplayer.Peer) {
		logger.Printf("[PEER] Connected to %v\n", peer.Name)
	}
	f.OnPeerDisconnected = func(peer *fakeplayer.Peer) {
		logger.Printf("[PEER] Connection to %v closed\n", peer.Name)
	}
	f.OnPeerPacket = func(peer *fakeplayer.Peer, pkt w3gs.Packet) bool {
		if *verbose {
			logger.Printf("[PEER] Packet %v from peer %v: %v\n", reflect.TypeOf(pkt).String()[6:], peer.Name, pkt)
		}

		switch p := pkt.(type) {
		case *w3gs.PeerMessage:
			if p.Content != "" {
				logger.Printf("[PEER] [CHAT] %v: '%v'\n", peer.Name, p.Content)
			}
			return true

		default:
			return false
		}
	}
	f.OnPacket = func(pkt w3gs.Packet) bool {
		if *verbose {
			logger.Printf("[HOST] Packet %v from host: %v\n", reflect.TypeOf(pkt).String()[6:], pkt)
		}

		switch p := pkt.(type) {

		case *w3gs.PlayerInfo:
			logger.Printf("[HOST] %v has joined the game\n", p.PlayerName)

		case *w3gs.PlayerLeft:
			logger.Printf("[HOST] %v has left the game\n", f.PeerName(p.PlayerID))

		case *w3gs.PlayerKicked:
			logger.Println("[HOST] Kicked from lobby")

		case *w3gs.CountDownStart:
			logger.Println("[HOST] Countdown started")

		case *w3gs.CountDownEnd:
			logger.Println("[HOST] Countdown ended, loading game")

		case *w3gs.StartLag:
			var laggers []string
			for _, l := range p.Players {
				laggers = append(laggers, f.PeerName(l.PlayerID))
			}
			logger.Printf("[HOST] Laggers %v\n", laggers)

		case *w3gs.StopLag:
			logger.Printf("[HOST] %v stopped lagging\n", f.PeerName(p.PlayerID))

		case *w3gs.MessageRelay:
			if p.Content == "" {
				break
			}

			logger.Printf("[HOST] [CHAT] %v: '%v'\n", f.PeerName(p.SenderID), p.Content)
			if p.SenderID != 1 || p.Content[:1] != "!" {
				break
			}

			var cmd = strings.Split(p.Content, " ")
			switch strings.ToLower(cmd[0]) {
			case "!say":
				f.Say(strings.Join(cmd[1:], " "))
			case "!leave":
				f.Leave(w3gs.LeaveLost)
			case "!race":
				if len(cmd) != 2 {
					f.Say("Use like: !race [str]")
					break
				}

				switch strings.ToLower(cmd[1]) {
				case "h", "hu", "human":
					f.ChangeRace(w3gs.RaceHuman)
				case "o", "orc":
					f.ChangeRace(w3gs.RaceOrc)
				case "u", "ud", "und", "undead":
					f.ChangeRace(w3gs.RaceUndead)
				case "n", "ne", "elf", "nightelf":
					f.ChangeRace(w3gs.RaceNightElf)
				case "r", "rnd", "rdm", "random":
					f.ChangeRace(w3gs.RaceRandom)
				default:
					f.Say("Invalid race")
				}
			case "!team":
				if len(cmd) != 2 {
					f.Say("Use like: !team [int]")
					break
				}
				if t, err := strconv.Atoi(cmd[1]); err == nil && t >= 1 {
					f.ChangeTeam(uint8(t - 1))
				}
			case "!color":
				if len(cmd) != 2 {
					f.Say("Use like: !color [int]")
					break
				}
				if c, err := strconv.Atoi(cmd[1]); err == nil && c >= 1 {
					f.ChangeColor(uint8(c - 1))
				}
			case "!handicap":
				if len(cmd) != 2 {
					f.Say("Use like: !handicap [int]")
					break
				}
				if h, err := strconv.Atoi(cmd[1]); err == nil && h >= 0 {
					f.ChangeHandicap(uint8(h))
				}
			}
		}

		return false
	}

	f.Run()

	time.Sleep(time.Second)
	f.Say("I come from the darkness of the pit.")

	f.Wait()
}
