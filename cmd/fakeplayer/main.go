// Author:  Niels A.D.
// Project: gowarcraft3 (https://github.com/nielsAD/gowarcraft3)
// License: Mozilla Public License, v2.0

// fakeplayer is a mocked Warcraft 3 game client that can be used to add dummy players to games.
package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/nielsAD/gowarcraft3/mock/lan"
	"github.com/nielsAD/gowarcraft3/mock/player"
	"github.com/nielsAD/gowarcraft3/protocol/w3gs"
)

var (
	verbose = flag.Bool("verbose", false, "Print contents of all packets")

	findlan  = flag.Bool("lan", false, "Find a game on LAN")
	gametft  = flag.Bool("tft", true, "Search for TFT or ROC games (only used when searching local)")
	gamevers = flag.Int("v", 29, "Game version (only used when searching local)")
	entrykey = flag.Uint("e", 0, "Entry key (only used when entering local game)")

	hostcounter = flag.Uint("c", 1, "Host counter")
	playername  = flag.String("n", "fakeplayer", "Player name")
	listen      = flag.Int("l", 0, "Listen on port (0 to pick automatically)")
	dialpeers   = flag.Bool("dial", true, "Dial peers")
)

var logOut = log.New(os.Stdout, "", log.Ltime)
var logErr = log.New(os.Stderr, "", log.Ltime)

func main() {
	flag.Parse()

	logOut.SetPrefix(fmt.Sprintf("[%v] ", *playername))
	logErr.SetPrefix(fmt.Sprintf("[%v] ", *playername))

	var err error
	var addr *net.TCPAddr
	var hc = uint32(*hostcounter)
	var ek = uint32(*entrykey)

	if *findlan {
		// Search local game for 3 seconds
		var ctx, cancel = context.WithTimeout(context.Background(), 20*time.Second)

		var p = w3gs.ProductTFT
		if !*gametft {
			p = w3gs.ProductROC
		}

		var address string
		address, hc, ek, err = lan.FindGame(ctx, w3gs.GameVersion{Product: p, Version: uint32(*gamevers)})
		cancel()

		if err == nil {
			addr, err = net.ResolveTCPAddr("tcp4", address)
		}
	} else {
		var address = strings.Join(flag.Args(), " ")
		if address == "" {
			address = "127.0.0.1:6112"
		}
		addr, err = net.ResolveTCPAddr("tcp4", address)
	}
	if err != nil {
		logErr.Fatal(err)
	}

	f, err := player.JoinLobby(addr, *playername, hc, ek, *listen)
	if err != nil {
		logErr.Fatal("JoinLobby error: ", err)
	}

	logOut.Println("[HOST] Joined lobby")

	f.DialPeers = *dialpeers

	f.OnError = func(err error) {
		logErr.Printf("[HOST] [ERROR] Packet: %v\n", err)
	}
	f.OnPeerError = func(err error) {
		logErr.Printf("[PEER] [ERROR] Accept: %v\n", err)
	}
	f.OnPeerAccept = func(conn *net.TCPConn) bool {
		logOut.Printf("[PEER] Accepting connection from %v\n", conn.RemoteAddr())
		return true
	}
	f.OnPeerConnect = func(peer *player.Peer) bool {
		logOut.Printf("[PEER] Connected to %v\n", peer.Name)

		peer.OnPacket = func(pkt w3gs.Packet) bool {
			if *verbose {
				logOut.Printf("[PEER] Packet %v from peer %v: %v\n", reflect.TypeOf(pkt).String()[6:], peer.Name, pkt)
			}

			switch p := pkt.(type) {
			case *w3gs.PeerMessage:
				if p.Content != "" {
					logOut.Printf("[PEER] [CHAT] %v: '%v'\n", peer.Name, p.Content)
				}
				return true

			default:
				return false
			}
		}
		peer.OnError = func(err error) {
			logErr.Printf("[PEER] [ERROR] Packet: %v\n", err)
		}

		return true
	}
	f.OnPeerDisconnect = func(peer *player.Peer) {
		logOut.Printf("[PEER] Connection to %v closed\n", peer.Name)
	}
	f.OnPacket = func(pkt w3gs.Packet) bool {
		if *verbose {
			logOut.Printf("[HOST] Packet %v from host: %v\n", reflect.TypeOf(pkt).String()[6:], pkt)
		}

		switch p := pkt.(type) {

		case *w3gs.PlayerInfo:
			logOut.Printf("[HOST] %v has joined the game\n", p.PlayerName)

		case *w3gs.PlayerLeft:
			logOut.Printf("[HOST] %v has left the game\n", f.PeerName(p.PlayerID))

		case *w3gs.PlayerKicked:
			logOut.Println("[HOST] Kicked from lobby")

		case *w3gs.CountDownStart:
			logOut.Println("[HOST] Countdown started")

		case *w3gs.CountDownEnd:
			logOut.Println("[HOST] Countdown ended, loading game")

		case *w3gs.StartLag:
			var laggers []string
			for _, l := range p.Players {
				laggers = append(laggers, f.PeerName(l.PlayerID))
			}
			logOut.Printf("[HOST] Laggers %v\n", laggers)

		case *w3gs.StopLag:
			logOut.Printf("[HOST] %v stopped lagging\n", f.PeerName(p.PlayerID))

		case *w3gs.MessageRelay:
			if p.Content == "" {
				break
			}

			logOut.Printf("[HOST] [CHAT] %v: '%v'\n", f.PeerName(p.SenderID), p.Content)
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
				case "h", "hu", "hum", "human":
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
