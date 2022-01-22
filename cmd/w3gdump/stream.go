// Author:  Niels A.D.
// Project: gowarcraft3 (https://github.com/nielsAD/gowarcraft3)
// License: Mozilla Public License, v2.0

package main

import (
	"errors"
	"fmt"
	"hash/crc32"
	"io"
	"net"
	"strconv"
	"strings"
	"sync/atomic"
	"time"

	"github.com/nielsAD/gowarcraft3/file/fs"
	"github.com/nielsAD/gowarcraft3/file/fs/dir"
	"github.com/nielsAD/gowarcraft3/file/w3g"
	"github.com/nielsAD/gowarcraft3/network"
	"github.com/nielsAD/gowarcraft3/network/lan"
	"github.com/nielsAD/gowarcraft3/protocol/w3gs"
)

var (
	errBreakEarly       = errors.New("w3gdump: Early break")
	errUnexpectedPacket = errors.New("w3gdump: Unexpected packet")
	errMapUnavailable   = errors.New("w3gdump: Map unavailable")
)

func mapCRC(name string) (uint32, uint32) {
	var stor = fs.Open(dir.InstallDir(), dir.UserDir())
	defer stor.Close()

	if f, err := stor.Open(name); err == nil {
		defer f.Close()

		var crc = crc32.NewIEEE()
		if size, err := io.Copy(crc, f); err == nil {
			return uint32(size), crc.Sum32()
		}
	}

	return 1, 1
}

func speedString(s int64) string {
	if s < 0 {
		return fmt.Sprintf("1/%dx", -s+1)
	}
	return fmt.Sprintf("%dx", s+1)
}

func cast(name string) error {
	replay, err := w3g.Open(name)
	if err != nil {
		return err
	}

	l, err := net.ListenTCP("tcp4", nil)
	if err != nil {
		return err
	}
	defer l.Close()
	adv, err := lan.NewAdvertiser(&w3gs.GameInfo{
		GameVersion:    replay.GameVersion,
		HostCounter:    1,
		EntryKey:       0xDEADBEEF,
		GameName:       replay.GameName,
		GameSettings:   replay.GameSettings,
		GameFlags:      replay.GameFlags,
		SlotsTotal:     (uint32)(len(replay.Slots)),
		SlotsUsed:      0,
		SlotsAvailable: 1,
		GamePort:       uint16(l.Addr().(*net.TCPAddr).Port),
	})
	if err != nil {
		return err
	}
	defer adv.Close()

	go adv.Run()
	logOut.Printf("Streaming game '%s' on %s (game version: %v), please join the lobby\n", replay.GameName, l.Addr(), replay.GameVersion)

	l.SetDeadline(time.Now().Add(3 * time.Minute))
	tcp, err := l.AcceptTCP()
	if err != nil {
		return err
	}
	defer tcp.Close()

	tcp.SetNoDelay(true)

	conn := network.NewW3GSConn(tcp, w3gs.NewFactoryCache(w3gs.DefaultFactory), w3gs.Encoding{GameVersion: replay.GameVersion.Version})
	pkt, err := conn.NextPacket(10 * time.Second)
	if err != nil {
		return err
	}

	switch v := pkt.(type) {
	case *w3gs.Join:
		if v.HostCounter == 1 && v.EntryKey == 0xDEADBEEF {
			logOut.Printf("%s joined the lobby, starting game..\n", v.PlayerName)
			break
		}
		conn.Send(&w3gs.RejectJoin{Reason: w3gs.RejectJoinWrongKey})
		return errUnexpectedPacket
	default:
		conn.Send(&w3gs.RejectJoin{Reason: w3gs.RejectJoinInvalid})
		return errUnexpectedPacket
	}

	// Close advertiser early
	adv.Close()

	var hostID = replay.HostPlayer.ID
	for _, s := range replay.Slots {
		if s.SlotStatus == w3gs.SlotOccupied && !s.Computer {
			// Hope player in lowest slot is an observer
			hostID = s.PlayerID
		}
	}

	if _, err := conn.Send(&w3gs.SlotInfoJoin{
		SlotInfo: replay.SlotInfo.SlotInfo,
		PlayerID: hostID,
	}); err != nil {
		return err
	}

	for _, p := range replay.PlayerInfo {
		if p.ID == hostID {
			continue
		}
		if _, err := conn.Send(&w3gs.PlayerInfo{
			JoinCounter: p.JoinCounter,
			PlayerID:    p.ID,
			PlayerName:  p.Name,
		}); err != nil {
			return err
		}
	}

	var size, crc = mapCRC(strings.Replace(replay.GameSettings.MapPath, "\\", "/", -1))
	if _, err := conn.Send(&w3gs.MapCheck{
		FilePath: replay.GameSettings.MapPath,
		FileSize: size,
		FileCRC:  crc,
		MapXoro:  replay.GameSettings.MapXoro,
		MapSha1:  replay.GameSettings.MapSha1,
	}); err != nil {
		return err
	}

	pkt, err = conn.NextPacket(10 * time.Second)
	for {
		if err != nil {
			return err
		}
		switch m := pkt.(type) {
		case *w3gs.PlayerExtra:
			pkt, err = conn.NextPacket(network.NoTimeout)
			continue
		case *w3gs.MapState:
			if !m.Ready {
				return errMapUnavailable
			}
			// Break out of loop
		default:
			return errUnexpectedPacket
		}
		break
	}

	for _, p := range replay.PlayerExtra {
		if _, err := conn.Send(&p.PlayerExtra); err != nil {
			return err
		}
	}

	time.Sleep(1 * time.Second)
	conn.Send(&w3gs.CountDownStart{})
	conn.Send(&w3gs.CountDownEnd{})

	for _, p := range replay.PlayerInfo {
		if p.ID == hostID {
			continue
		}
		if _, err := conn.Send(&w3gs.PlayerLoaded{
			PlayerID: p.ID,
		}); err != nil {
			return err
		}
	}

	pkt, err = conn.NextPacket(time.Minute * 5)
	for {
		if err != nil {
			return err
		}
		switch pkt.(type) {
		case *w3gs.PlayerExtra:
			pkt, err = conn.NextPacket(network.NoTimeout)
			continue
		case *w3gs.GameLoaded:
			// Break out of loop
		default:
			return errUnexpectedPacket
		}
		break
	}

	var msec int64
	var speed int64

	var say = func(s string) {
		if _, err := conn.Send(&w3gs.MessageRelay{Message: w3gs.Message{
			SenderID: hostID,
			Type:     w3gs.MsgChatExtra,
			Scope:    w3gs.ScopeAll,
			Content:  s,
		}}); err != nil {
			logErr.Println("Say error: ", err)
			conn.Close()
		}
	}

	var events = network.EventEmitter{}
	events.On(&w3gs.Leave{}, func(_ *network.Event) {
		conn.Send(&w3gs.LeaveAck{})
		conn.Close()
	})
	events.On(&w3gs.Message{}, func(ev *network.Event) {
		var msg = ev.Arg.(*w3gs.Message)
		if !strings.HasPrefix(msg.Content, ".") {
			return
		}

		var cmd = strings.Fields(msg.Content)
		switch strings.ToLower(cmd[0]) {
		case ".time":
			var t = (time.Duration)(atomic.LoadInt64(&msec)) * time.Millisecond
			say("Time: " + t.String())
		case ".speed":
			var s = atomic.LoadInt64(&speed)

			if len(cmd) > 1 {
				if strings.HasPrefix(cmd[1], "1/") {
					if i, err := strconv.ParseInt(cmd[1][2:], 0, 64); err == nil {
						s = -(i - 1)
					}
				} else {
					if i, err := strconv.ParseInt(cmd[1], 0, 64); err == nil {
						s = i - 1
					}
				}
				atomic.StoreInt64(&speed, s)
			}

			say("Replay speed: " + speedString(s))
		}
	})

	go func() {
		err := conn.Run(&events, 3*time.Second)
		if err != nil && !network.IsCloseError(err) {
			logErr.Println("Connection error: ", err)
			conn.Close()
		}
	}()

	if _, err := conn.Send(&w3gs.PlayerLoaded{
		PlayerID: hostID,
	}); err != nil {
		return err
	}

	for _, rec := range replay.Records {
		var pkt w3gs.Packet
		switch v := rec.(type) {
		case *w3g.PlayerLeft:
			if v.PlayerID == hostID {
				continue
			}
			pkt = &w3gs.PlayerLeft{
				PlayerID: v.PlayerID,
				Reason:   v.Reason,
			}
		case *w3g.TimeSlot:
			var s = atomic.LoadInt64(&speed)
			if s >= 0 {
				time.Sleep(time.Duration(v.TimeIncrementMS) * time.Millisecond / (time.Duration)(s+1))
			} else {
				time.Sleep(time.Duration(v.TimeIncrementMS) * time.Millisecond * (time.Duration)(-s+1))
			}
			atomic.AddInt64(&msec, int64(v.TimeIncrementMS))

			pkt = &v.TimeSlot
		case *w3g.Desync:
			pkt = &v.Desync
		case *w3g.ChatMessage:
			pkt = &w3gs.MessageRelay{Message: v.Message}
		default:
			continue
		}

		if _, err := conn.Send(pkt); err != nil {
			return err
		}
	}

	return nil
}
