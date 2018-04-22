// Author:  Niels A.D.
// Project: gowarcraft3 (https://github.com/nielsAD/gowarcraft3)
// License: Mozilla Public License, v2.0
package w3gs_test

import (
	"bytes"
	"net"
	"reflect"
	"testing"

	"github.com/nielsAD/gowarcraft3/protocol"
	"github.com/nielsAD/gowarcraft3/protocol/w3gs"
)

var sd = []w3gs.SlotData{
	w3gs.SlotData{
		PlayerID:       1,
		DownloadStatus: 2,
		SlotStatus:     3,
		Computer:       true,
		Team:           5,
		Color:          6,
		Race:           7,
		ComputerType:   8,
		Handicap:       9,
	},
	w3gs.SlotData{
		PlayerID:       9,
		DownloadStatus: 8,
		SlotStatus:     7,
		Computer:       false,
		Team:           5,
		Color:          4,
		Race:           3,
		ComputerType:   2,
		Handicap:       1,
	},
}

func TestPackets(t *testing.T) {
	var types = []w3gs.Packet{
		&w3gs.UnknownPacket{
			ID:   255,
			Blob: []byte{w3gs.ProtocolSig, 255, 4, 0},
		},
		&w3gs.Ping{},
		&w3gs.Ping{
			Payload: 444,
		},
		&w3gs.Pong{},
		&w3gs.Pong{
			Ping: w3gs.Ping{Payload: 999},
		},
		&w3gs.PeerPing{},
		&w3gs.PeerPing{
			Payload:   123,
			PeerSet:   protocol.BS32(true, false, true),
			GameTicks: 789,
		},
		&w3gs.PeerPong{},
		&w3gs.PeerPong{
			Ping: w3gs.Ping{Payload: 1011},
		},

		&w3gs.Join{},
		&w3gs.Join{
			HostCounter: 1,
			EntryKey:    2,
			ListenPort:  3,
			JoinCounter: 4,
			PlayerName:  "Grubby",
			InternalAddr: protocol.SockAddr{
				Port: 6,
				IP:   net.IP{7, 8, 9, 10},
			},
		},
		&w3gs.RejectJoin{},
		&w3gs.RejectJoin{
			Reason: w3gs.RejectJoinWrongKey,
		},
		&w3gs.SlotInfoJoin{},
		&w3gs.SlotInfoJoin{
			SlotInfo: w3gs.SlotInfo{
				Slots:      sd,
				RandomSeed: 10,
				SlotLayout: w3gs.LayoutMelee,
				NumPlayers: 12,
			},
			PlayerID: 13,
			ExternalAddr: protocol.SockAddr{
				Port: 14,
				IP:   net.IP{15, 16, 17, 18},
			},
		},
		&w3gs.SlotInfo{},
		&w3gs.SlotInfo{
			Slots: sd,
		},
		&w3gs.PlayerInfo{},
		&w3gs.PlayerInfo{
			JoinCounter: 1,
			PlayerID:    2,
			PlayerName:  "Moon",
			ExternalAddr: protocol.SockAddr{
				Port: 4,
				IP:   net.IP{5, 6, 7, 8},
			},
			InternalAddr: protocol.SockAddr{
				Port: 9,
				IP:   net.IP{10, 11, 12, 13},
			},
		},

		&w3gs.Leave{},
		&w3gs.Leave{
			Reason: w3gs.LeaveLost,
		},
		&w3gs.LeaveAck{},
		&w3gs.PlayerKicked{},
		&w3gs.PlayerKicked{
			Leave: w3gs.Leave{Reason: w3gs.LeaveLobby},
		},
		&w3gs.PlayerLeft{},
		&w3gs.PlayerLeft{
			PlayerID: 1,
			Reason:   w3gs.LeaveLost,
		},

		&w3gs.CountDownStart{},
		&w3gs.CountDownEnd{},
		&w3gs.GameLoaded{},
		&w3gs.PlayerLoaded{},
		&w3gs.PlayerLoaded{
			PlayerID: 12,
		},
		&w3gs.GameOver{},
		&w3gs.GameOver{
			PlayerID: 34,
		},

		&w3gs.StartLag{},
		&w3gs.StartLag{
			Players: []w3gs.LagPlayer{
				w3gs.LagPlayer{PlayerID: 1, LagDurationMS: 2},
				w3gs.LagPlayer{PlayerID: 3, LagDurationMS: 4},
				w3gs.LagPlayer{PlayerID: 5, LagDurationMS: 6},
			},
		},
		&w3gs.StopLag{},
		&w3gs.StopLag{
			LagPlayer: w3gs.LagPlayer{PlayerID: 1, LagDurationMS: 2},
		},
		&w3gs.DropLaggers{},

		&w3gs.GameAction{},
		&w3gs.GameAction{
			Data: []byte{2, 3, 4, 5, 6, 7, 8, 9},
		},
		&w3gs.TimeSlot{},
		&w3gs.TimeSlot{
			Fragment:        false,
			TimeIncrementMS: 50,
			Actions: []w3gs.PlayerAction{
				w3gs.PlayerAction{PlayerID: 1, Data: make([]byte, 23)},
				w3gs.PlayerAction{PlayerID: 12, Data: make([]byte, 3)},
			},
		},
		&w3gs.TimeSlot{
			Fragment:        true,
			TimeIncrementMS: 50,
			Actions: []w3gs.PlayerAction{
				w3gs.PlayerAction{PlayerID: 1, Data: make([]byte, 23)},
				w3gs.PlayerAction{PlayerID: 12, Data: make([]byte, 3)},
			},
		},
		&w3gs.TimeSlotAck{},
		&w3gs.TimeSlotAck{
			Checksum: 456,
		},
		&w3gs.Desync{},
		&w3gs.Desync{
			Checksum:       789,
			PlayersInState: []uint8{1, 2, 3},
		},

		&w3gs.Message{},
		&w3gs.Message{
			RecipientIDs: []uint8{1, 2, 3},
			SenderID:     4,
			Type:         w3gs.MsgChat,
			Content:      "Tremble before me",
		},
		&w3gs.Message{
			RecipientIDs: []uint8{1, 2, 3},
			SenderID:     4,
			Type:         w3gs.MsgColorChange,
			NewVal:       5,
		},
		&w3gs.MessageRelay{},
		&w3gs.MessageRelay{
			Message: w3gs.Message{
				RecipientIDs: []uint8{1, 2, 3},
				SenderID:     4,
				Type:         w3gs.MsgChat,
				Content:      "I come from the darkness of the pit",
			},
		},
		&w3gs.MessageRelay{
			Message: w3gs.Message{
				RecipientIDs: []uint8{1, 2},
				SenderID:     4,
				Type:         w3gs.MsgChatExtra,
				ExtraFlags:   5,
				Content:      "Pitiful",
			},
		},
		&w3gs.PeerMessage{},
		&w3gs.PeerMessage{
			Message: w3gs.Message{
				RecipientIDs: []uint8{1, 2, 3},
				SenderID:     4,
				Type:         w3gs.MsgChat,
				Content:      "You fail to amuse me",
			},
		},

		&w3gs.SearchGame{},
		&w3gs.SearchGame{
			GameVersion: w3gs.GameVersion{
				Product: w3gs.ProductDemo,
				Version: 666,
			},
			Counter: 1,
		},
		&w3gs.GameInfo{},
		&w3gs.GameInfo{
			GameVersion: w3gs.GameVersion{
				Product: w3gs.ProductROC,
				Version: 1,
			},
			HostCounter:    2,
			EntryKey:       112233,
			GameName:       "game1",
			StatString:     "xxxxx",
			SlotsTotal:     24,
			GameType:       w3gs.GameTypeNewGame,
			SlotsUsed:      1,
			SlotsAvailable: 24,
			UptimeSec:      8,
			GamePort:       9,
		},
		&w3gs.CreateGame{},
		&w3gs.CreateGame{
			GameVersion: w3gs.GameVersion{
				Product: w3gs.ProductTFT,
				Version: 2,
			},
			HostCounter: 3,
		},
		&w3gs.RefreshGame{},
		&w3gs.RefreshGame{
			HostCounter:    1,
			SlotsUsed:      2,
			SlotsAvailable: 3,
		},
		&w3gs.DecreateGame{},
		&w3gs.DecreateGame{
			HostCounter: 777,
		},

		&w3gs.PeerConnect{},
		&w3gs.PeerConnect{
			JoinCounter: 1,
			EntryKey:    2,
			PlayerID:    3,
			PeerSet:     protocol.BS32(false, true, false),
		},
		&w3gs.PeerSet{},
		&w3gs.PeerSet{
			PeerSet: protocol.BS16(true, false, true),
		},

		&w3gs.MapCheck{},
		&w3gs.MapCheck{
			FilePath: "Maps\\BootyBay.w3x",
			FileSize: 2,
			FileCRC:  3,
			MapXoro:  4,
		},
		&w3gs.StartDownload{},
		&w3gs.StartDownload{
			PlayerID: 111,
		},
		&w3gs.MapState{},
		&w3gs.MapState{
			Ready:    true,
			FileSize: 2,
		},
		&w3gs.MapPart{},
		&w3gs.MapPart{
			RecipientID: 1,
			SenderID:    2,
			ChunkPos:    3,
			Data:        []byte{5, 6, 7, 8, 9},
		},
		&w3gs.MapPartOK{},
		&w3gs.MapPartOK{
			RecipientID: 1,
			SenderID:    2,
			ChunkPos:    3,
		},
		&w3gs.MapPartError{},
	}

	for _, pkt := range types {
		var err error
		var buf = protocol.Buffer{Bytes: make([]byte, 0, 2048)}

		if err = pkt.Serialize(&buf); err != nil {
			t.Log(reflect.TypeOf(pkt))
			t.Fatal(err)
		}

		var buf2 = protocol.Buffer{Bytes: make([]byte, 0, 2048)}
		if _, err = w3gs.SerializePacket(&buf2, pkt); err != nil {
			t.Log(reflect.TypeOf(pkt))
			t.Fatal(err)
		}

		if bytes.Compare(buf.Bytes, buf2.Bytes) != 0 {
			t.Fatalf("SerializePacket != packet.Serialize %v", reflect.TypeOf(pkt))
		}

		var pkt2, _, e = w3gs.DeserializePacket(&buf)
		if e != nil {
			t.Log(reflect.TypeOf(pkt))
			t.Fatal(e)
		}
		if buf.Size() > 0 {
			t.Fatalf("DeserializePacket size mismatch for %v", reflect.TypeOf(pkt))
		}
		if reflect.TypeOf(pkt2) != reflect.TypeOf(pkt) {
			t.Fatalf("DeserializePacket type mismatch %v != %v", reflect.TypeOf(pkt2), reflect.TypeOf(pkt))
		}
		if !reflect.DeepEqual(pkt, pkt2) {
			t.Logf("I: %+v", pkt)
			t.Logf("O: %+v", pkt2)
			t.Errorf("DeserializePacket value mismatch for %v", reflect.TypeOf(pkt))
		}

		err = pkt.Deserialize(&protocol.Buffer{Bytes: make([]byte, 0)})
		if err != w3gs.ErrInvalidPacketSize {
			t.Fatalf("ErrInvalidPacketSize expected for %v", reflect.TypeOf(pkt))
		}

		err = pkt.Deserialize(&protocol.Buffer{Bytes: make([]byte, 2048)})
		if err != w3gs.ErrInvalidPacketSize && err != w3gs.ErrInvalidChecksum {
			switch pkt.(type) {
			case *w3gs.UnknownPacket:
				// Whitelisted
			default:
				t.Fatalf("ErrInvalidPacketSize expected for %v", reflect.TypeOf(pkt))
			}

		}
	}
}

func BenchmarkSerialize(b *testing.B) {
	var pkt = w3gs.SlotInfo{
		Slots: sd,
	}

	var buf = protocol.Buffer{Bytes: make([]byte, 0, 2048)}
	pkt.Serialize(&buf)

	b.SetBytes(int64(buf.Size()))
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		buf.Truncate()
		pkt.Serialize(&buf)
	}
}

func BenchmarkDeserialize(b *testing.B) {
	var pkt = w3gs.SlotInfo{
		Slots: sd,
	}

	var res w3gs.SlotInfo
	var buf = protocol.Buffer{Bytes: make([]byte, 0, 2048)}
	pkt.Serialize(&buf)

	b.SetBytes(int64(buf.Size()))
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		res.Deserialize(&protocol.Buffer{Bytes: buf.Bytes})
	}
}

func BenchmarkCreateAndSerialize(b *testing.B) {
	var size = 0
	var buf = protocol.Buffer{Bytes: make([]byte, 0, 2048)}

	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		var pkt = w3gs.SlotInfo{
			Slots: sd,
		}

		buf.Truncate()
		pkt.Serialize(&buf)
		size = buf.Size()
	}
	b.SetBytes(int64(size))
}

func BenchmarkCreateAndDeserialize(b *testing.B) {
	var pkt = w3gs.SlotInfo{
		Slots: sd,
	}
	var buf = protocol.Buffer{Bytes: make([]byte, 0, 2048)}
	pkt.Serialize(&buf)

	b.SetBytes(int64(buf.Size()))
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		var res w3gs.SlotInfo
		res.Deserialize(&protocol.Buffer{Bytes: buf.Bytes})
	}
}
