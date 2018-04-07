// fakeplayer is a mocked Warcraft 3 game client that can be used to add dummy players to games.
package main

import (
	"flag"
	"io"
	"log"
	"net"
	"os"
	"reflect"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/nielsAD/noot/pkg/util"
	"github.com/nielsAD/noot/pkg/w3gs"
)

var (
	verbose = flag.Bool("verbose", false, "Print contents of all packets")

	local    = flag.Bool("local", false, "Find a local game")
	gametft  = flag.Bool("tft", true, "Search for TFT or ROC games (only used when searching local)")
	gamevers = flag.Int("v", 28, "Game version (only used when searching local)")
	entrykey = flag.Uint("e", 0, "Entry key (only used when entering local game)")

	hostcounter = flag.Uint("c", 1, "Host counter")
	playername  = flag.String("n", "fakeplayer", "Player name")
)

var logger = log.New(os.Stdout, "", log.Ltime)

func sendUDP(conn *net.UDPConn, addr *net.UDPAddr, pkt w3gs.Packet) (int, error) {
	var buf util.PacketBuffer
	if err := pkt.Serialize(&buf); err != nil {
		return 0, nil
	}
	return conn.WriteToUDP(buf.Bytes, addr)
}

func findLocalGame(gameVersion *w3gs.GameVersion) (*net.TCPAddr, error) {
	conn, err := net.ListenUDP("udp4", &net.UDPAddr{Port: 6112})
	if err != nil {
		return nil, err
	}

	defer conn.Close()
	logger.Printf("[UDP] Listening on %v for local game %+v\n", conn.LocalAddr(), *gameVersion)

	var search = w3gs.SearchGame{GameVersion: *gameVersion, Counter: 2}
	if _, err := sendUDP(conn, &net.UDPAddr{IP: net.IPv4bcast, Port: 6112}, &search); err != nil {
		return nil, err
	}

	buf := make([]byte, 2048)
	for {
		size, addr, err := conn.ReadFromUDP(buf)
		if err != nil {
			logger.Fatal(err)
		}

		pkt, _, err := w3gs.DeserializePacket(&util.PacketBuffer{Bytes: buf[:size]})
		if err != nil {
			return nil, err
		}

		switch p := pkt.(type) {
		case *w3gs.CreateGame:
			if p.GameVersion == *gameVersion {
				if _, err := sendUDP(conn, addr, &search); err != nil {
					return nil, err
				}
			}
		case *w3gs.GameInfo:
			logger.Printf("[UDP] Found local game '%v'\n", p.GameName)
			*hostcounter = uint(p.HostCounter)
			*entrykey = uint(p.EntryKey)
			return &net.TCPAddr{IP: addr.IP, Port: int(p.GamePort)}, nil
		}
	}
}

func servePeer(conn *net.TCPConn, pid *uint8, peerkey uint32) {
	conn.SetNoDelay(true)
	defer conn.Close()

	if peerkey == 0 {
		return
	}

	if *pid == 0 {
		logger.Println("No pid assigned yet")
		return
	}

	var mutex = &sync.Mutex{}
	var sbuf w3gs.SerializationBuffer
	var rbuf w3gs.DeserializationBuffer

	var send = func(p w3gs.Packet) {
		mutex.Lock()
		defer mutex.Unlock()

		conn.SetWriteDeadline(time.Now().Add(5 * time.Millisecond))
		if _, err := w3gs.SerializePacketWithBuffer(conn, &sbuf, p); err != nil {
			logger.Fatal(err)
		}
	}

	if err := conn.SetReadDeadline(time.Now().Add(5 * time.Second)); err != nil {
		logger.Fatal(err)
	}

	pkt, _, err := w3gs.DeserializePacketWithBuffer(conn, &rbuf)
	if err == io.EOF {
		return
	} else if err != nil {
		logger.Fatal(err)
	}

	switch p := pkt.(type) {
	case *w3gs.PeerConnect:
		if p.EntryKey != uint32(*entrykey) {
			logger.Println("Peer tried to connect with wrong entry key")
			return
		}
		if peerkey == 0 {
			peerkey = p.PeerMask
			send(&w3gs.PeerConnect{
				JoinCounter: 4,
				EntryKey:    uint32(*entrykey),
				PlayerID:    *pid,
				PeerMask:    1 << (*pid - 1),
			})
		}

		logger.Printf("Peer %v connected from %v with joincounter %v and peerkey %v\n", p.PlayerID, conn.RemoteAddr(), p.JoinCounter, peerkey)
	default:
		logger.Fatal("Invalid first packet")
	}

	pingTicker := time.NewTicker(10 * time.Second)
	defer pingTicker.Stop()

	go func() {
		for range pingTicker.C {
			send(&w3gs.PeerPing{
				Payload:  uint32(time.Now().UnixNano() / 1000000),
				PeerMask: peerkey,
			})
		}
	}()

	for {
		if err := conn.SetReadDeadline(time.Now().Add(time.Minute)); err != nil {
			logger.Fatal(err)
		}

		pkt, _, err := w3gs.DeserializePacketWithBuffer(conn, &rbuf)
		if err == io.EOF {
			break
		} else if err != nil {
			logger.Fatal(err)
		}

		if *verbose {
			logger.Println("[TCP] [PEER] Packet ", reflect.TypeOf(pkt).String()[6:], pkt)
		}

		switch p := pkt.(type) {
		case *w3gs.PeerPing:
			send(&w3gs.PeerPong{
				Ping: w3gs.Ping{Payload: p.Payload},
			})
		case *w3gs.PeerPong:
			logger.Printf("Ping: %vms\n", uint32(time.Now().UnixNano()/1000000)-p.Payload)
		default:
			if !*verbose {
				logger.Println("[TCP] [PEER] Unexpected packet ", reflect.TypeOf(pkt).String()[6:], pkt)
			}
		}
	}

	logger.Println("[TCP] [PEER] connection closed")
}

func fakeplayer(addr *net.TCPAddr) {
	serv, err := net.ListenTCP("tcp4", nil)
	if err != nil {
		logger.Fatal(err)
	}

	defer serv.Close()
	logger.Printf("[TCP] Listening on %v for peers\n", serv.Addr())

	var pid uint8
	var players = make(map[uint8]string)

	go func() {
		for {
			conn, err := serv.Accept()
			if err != nil {
				logger.Fatal("[TCP] [PEER] Error accepting: ", err)
			}
			go servePeer(conn.(*net.TCPConn), &pid, 0)
		}
	}()

	conn, err := net.DialTCP("tcp4", nil, addr)
	if err != nil {
		logger.Fatal(err)
	}

	conn.SetNoDelay(true)
	defer conn.Close()
	logger.Printf("[TCP] Joining game at %v with {HostCounter=%v EntryKey=%v}\n", addr.String(), *hostcounter, *entrykey)

	var sbuf w3gs.SerializationBuffer
	var rbuf w3gs.DeserializationBuffer

	w3gs.SerializePacketWithBuffer(conn, &sbuf, &w3gs.Join{
		HostCounter:  uint32(*hostcounter),
		EntryKey:     uint32(*entrykey),
		ListenPort:   uint16(serv.Addr().(*net.TCPAddr).Port),
		JoinCounter:  1,
		PlayerName:   *playername,
		InternalAddr: w3gs.Addr(conn.LocalAddr()),
	})

	for {
		if err := conn.SetReadDeadline(time.Now().Add(time.Minute)); err != nil {
			logger.Fatal(err)
		}

		pkt, _, err := w3gs.DeserializePacketWithBuffer(conn, &rbuf)
		if err == io.EOF {
			break
		} else if err != nil {
			logger.Fatal(err)
		}

		if *verbose {
			logger.Println("[TCP] [HOST] Packet ", reflect.TypeOf(pkt).String()[6:], pkt)
		}

		conn.SetWriteDeadline(time.Now().Add(5 * time.Millisecond))

		switch p := pkt.(type) {
		case *w3gs.Ping:
			w3gs.SerializePacketWithBuffer(conn, &sbuf, &w3gs.Pong{Ping: w3gs.Ping{Payload: p.Payload}})
		case *w3gs.MapCheck:
			w3gs.SerializePacketWithBuffer(conn, &sbuf, &w3gs.MapState{Ready: true, FileSize: p.FileSize})

		case *w3gs.RejectJoin:
			var reason string
			switch p.Reason {
			case w3gs.RejectJoinInvalid:
				reason = "RejectJoinInvalid"
			case w3gs.RejectJoinFull:
				reason = "RejectJoinFull"
			case w3gs.RejectJoinStarted:
				reason = "RejectJoinStarted"
			case w3gs.RejectJoinWrongKey:
				reason = "RejectJoinWrongKey"
			default:
				reason = strconv.Itoa(int(p.Reason))
			}
			logger.Println("Join denied: ", reason)

		case *w3gs.SlotInfo:
			// ignore
		case *w3gs.SlotInfoJoin:
			logger.Println("Joined lobby")
			pid = p.PlayerID
			players[p.PlayerID] = *playername
		case *w3gs.PlayerInfo:
			players[p.PlayerID] = p.PlayerName

			// if conn, err := net.DialTCP("tcp4", nil, w3gs.TCPAddr(&p.InternalAddr)); err == nil {
			// 	go servePeer(conn, &pid, 4)
			// 	if _, err := w3gs.SerializePacketWithBuffer(conn, &sbuf, &w3gs.PeerConnect{
			// 		JoinCounter: p.JoinCounter,
			// 		EntryKey:    uint32(*entrykey),
			// 		PlayerID:    pid,
			// 		PeerKey:     1 << pid,
			// 	}); err != nil {
			// 		logger.Fatal(err)
			// 	}
			// 	logger.Printf("Connected to peer %v with 4\n", p.PlayerName)
			// } else if conn, err := net.DialTCP("tcp4", nil, w3gs.TCPAddr(&p.ExternalAddr)); err == nil {
			// 	go servePeer(conn, &pid, 4)
			// 	if _, err := w3gs.SerializePacketWithBuffer(conn, &sbuf, &w3gs.PeerConnect{
			// 		JoinCounter: p.JoinCounter,
			// 		EntryKey:    uint32(*entrykey),
			// 		PlayerID:    pid,
			// 		PeerKey:     1 << pid,
			// 	}); err != nil {
			// 		logger.Fatal(err)
			// 	}
			// 	logger.Printf("Connected to peer %v with 4\n", p.PlayerName)
			// } else if p.InternalAddr.IP != nil || p.ExternalAddr.IP != nil {
			// 	logger.Printf("Could not connect to peer %v (%v)\n", p.PlayerName, p.ExternalAddr)
			// }

		case *w3gs.PlayerLeft:
			delete(players, p.PlayerID)
		case *w3gs.PlayerKicked:
			logger.Println("Kicked from lobby")

		case *w3gs.CountDownStart:
			logger.Println("Countdown started")
		case *w3gs.CountDownEnd:
			logger.Println("Start loading screen")
			w3gs.SerializePacketWithBuffer(conn, &sbuf, &w3gs.GameLoaded{})
		case *w3gs.PlayerLoaded:
			// ignore

		case *w3gs.StartLag:
			var laggers []string
			for _, l := range p.Players {
				laggers = append(laggers, players[l.PlayerID])
			}
			logger.Printf("Laggers %v\n", laggers)
		case *w3gs.StopLag:
			logger.Printf("%v stopped lagging\n", players[p.PlayerID])

		case *w3gs.MessageRelay:
			if p.Content != "" {
				logger.Printf("[CHAT] %v: '%v'\n", players[p.SenderID], p.Content)
			}

		case *w3gs.TimeSlot:
			// Cannot reply to this as we don't know the correct checksum for this round
			// replying with wrong info will result in a desync
			// not replying will result in lagscreen and drop
			//w3gs.SerializePacketWithBuffer(conn, &sbuf, &w3gs.TimeSlotAck{})

		default:
			if !*verbose {
				logger.Println("[TCP] [HOST] Unexpected packet ", reflect.TypeOf(pkt).String()[6:], pkt)
			}
		}
	}

	logger.Println("[TCP] [HOST] connection closed")
}

func main() {
	flag.Parse()

	var err error
	var addr *net.TCPAddr

	if *local {
		addr, err = findLocalGame(&w3gs.GameVersion{TFT: *gametft, Version: uint32(*gamevers)})
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

	fakeplayer(addr)
}
