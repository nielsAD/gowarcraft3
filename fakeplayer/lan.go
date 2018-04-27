// Author:  Niels A.D.
// Project: gowarcraft3 (https://github.com/nielsAD/gowarcraft3)
// License: Mozilla Public License, v2.0

package fakeplayer

import (
	"net"
	"time"

	"github.com/nielsAD/gowarcraft3/protocol"
	"github.com/nielsAD/gowarcraft3/protocol/w3gs"
)

func sendUDP(conn *net.UDPConn, addr *net.UDPAddr, pkt w3gs.Packet) (int, error) {
	var buf protocol.Buffer
	if err := pkt.Serialize(&buf); err != nil {
		return 0, nil
	}
	return conn.WriteToUDP(buf.Bytes, addr)
}

// FindGameOnLAN returns the first game found on LAN
// Returns (HostAddress, HostCounter, EntryKey, Error)
func FindGameOnLAN(gameVersion *w3gs.GameVersion) (*net.TCPAddr, uint32, uint32, error) {
	conn, err := net.ListenUDP("udp4", &net.UDPAddr{Port: 6112})
	if err != nil {
		conn, err = net.ListenUDP("udp4", &net.UDPAddr{})
	}
	if err != nil {
		return nil, 0, 0, err
	}
	defer conn.Close()

	var lastSeen = make(map[string]uint32)

	buf := make([]byte, 2048)
	for {
		if _, err := sendUDP(conn, &net.UDPAddr{IP: net.IPv4bcast, Port: 6112}, &w3gs.SearchGame{GameVersion: *gameVersion, HostCounter: 0}); err != nil {
			return nil, 0, 0, err
		}

		conn.SetReadDeadline(time.Now().Add(5 * time.Second))

		size, addr, err := conn.ReadFromUDP(buf)
		if err != nil {
			return nil, 0, 0, err
		}

		pkt, _, err := w3gs.DeserializePacket(&protocol.Buffer{Bytes: buf[:size]})
		if err != nil {
			return nil, 0, 0, err
		}

		switch p := pkt.(type) {
		case *w3gs.RefreshGame:
			if lastSeen[addr.IP.String()] != p.HostCounter {
				lastSeen[addr.IP.String()] = p.HostCounter
				if _, err := sendUDP(conn, addr, &w3gs.SearchGame{GameVersion: *gameVersion, HostCounter: p.HostCounter}); err != nil {
					return nil, 0, 0, err
				}
			}
		case *w3gs.CreateGame:
			if p.GameVersion == *gameVersion {
				lastSeen[addr.IP.String()] = p.HostCounter
				if _, err := sendUDP(conn, addr, &w3gs.SearchGame{GameVersion: *gameVersion, HostCounter: p.HostCounter}); err != nil {
					return nil, 0, 0, err
				}
			}
		case *w3gs.GameInfo:
			return &net.TCPAddr{IP: addr.IP, Port: int(p.GamePort)}, p.HostCounter, p.EntryKey, nil
		}
	}
}
