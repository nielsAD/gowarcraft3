package main

import (
	"net"

	"github.com/nielsAD/noot/pkg/util"
	"github.com/nielsAD/noot/pkg/w3gs"
)

func sendUDP(conn *net.UDPConn, addr *net.UDPAddr, pkt w3gs.Packet) (int, error) {
	var buf util.PacketBuffer
	if err := pkt.Serialize(&buf); err != nil {
		return 0, nil
	}
	return conn.WriteToUDP(buf.Bytes, addr)
}

func findGameOnLAN(gameVersion *w3gs.GameVersion) (*net.TCPAddr, uint32, uint32, error) {
	conn, err := net.ListenUDP("udp4", &net.UDPAddr{Port: 6112})
	if err != nil {
		return nil, 0, 0, err
	}

	defer conn.Close()
	logger.Printf("[UDP] Listening on %v for local game %+v\n", conn.LocalAddr(), *gameVersion)

	if _, err := sendUDP(conn, &net.UDPAddr{IP: net.IPv4bcast, Port: 6112}, &w3gs.SearchGame{GameVersion: *gameVersion, Counter: 0}); err != nil {
		return nil, 0, 0, err
	}

	var counter = make(map[string]uint32)

	buf := make([]byte, 2048)
	for {
		size, addr, err := conn.ReadFromUDP(buf)
		if err != nil {
			return nil, 0, 0, err
		}

		pkt, _, err := w3gs.DeserializePacket(&util.PacketBuffer{Bytes: buf[:size]})
		if err != nil {
			return nil, 0, 0, err
		}

		switch p := pkt.(type) {
		case *w3gs.RefreshGame:
			if counter[addr.IP.String()] == 0 {
				counter[addr.IP.String()]++
				if _, err := sendUDP(conn, addr, &w3gs.SearchGame{GameVersion: *gameVersion, Counter: counter[addr.IP.String()]}); err != nil {
					return nil, 0, 0, err
				}
			}
		case *w3gs.CreateGame:
			if p.GameVersion == *gameVersion {
				counter[addr.IP.String()]++
				if _, err := sendUDP(conn, addr, &w3gs.SearchGame{GameVersion: *gameVersion, Counter: counter[addr.IP.String()]}); err != nil {
					return nil, 0, 0, err
				}
			}
		case *w3gs.GameInfo:
			logger.Printf("[UDP] Found LAN game '%v'\n", p.GameName)
			return &net.TCPAddr{IP: addr.IP, Port: int(p.GamePort)}, p.HostCounter, p.EntryKey, nil
		}
	}
}
