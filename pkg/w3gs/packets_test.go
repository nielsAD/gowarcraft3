package w3gs_test

import (
	"math/rand"
	"net"
	"reflect"
	"testing"

	"github.com/nielsAD/noot/pkg/w3gs"
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

func TestMarshalPacket(t *testing.T) {
	var types = []w3gs.Packet{
		&w3gs.UnknownPacket{
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
			Payload: 123,
		},
		&w3gs.PeerPong{},
		&w3gs.PeerPong{
			Ping: w3gs.Ping{Payload: 456},
		},

		&w3gs.Join{},
		&w3gs.Join{
			HostCounter: 1,
			EntryKey:    2,
			ListenPort:  3,
			PeerKey:     4,
			PlayerName:  "Grubby",
			InternalAddr: w3gs.ConnAddr{
				Port: 6,
				IP:   net.IP{7, 8, 9, 10},
			},
		},
		&w3gs.RejectJoin{},
		&w3gs.RejectJoin{
			Reason: w3gs.RejectJoinWrongPass,
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
			ExternalAddr: w3gs.ConnAddr{
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
			ExternalAddr: w3gs.ConnAddr{
				Port: 4,
				IP:   net.IP{5, 6, 7, 8},
			},
			InternalAddr: w3gs.ConnAddr{
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
			PlayerID: 123,
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

		&w3gs.Message{},
		&w3gs.Message{
			RecipientIDs: []uint8{1, 2, 3},
			SenderID:     4,
			Flags:        w3gs.ChatMessage,
			Content:      "Tremble before me",
		},
		&w3gs.Message{
			RecipientIDs: []uint8{1, 2, 3},
			SenderID:     4,
			Flags:        w3gs.ChatColorChange,
			NewVal:       5,
		},
		&w3gs.MessageRelay{},
		&w3gs.MessageRelay{
			Message: w3gs.Message{
				RecipientIDs: []uint8{1, 2, 3},
				SenderID:     4,
				Flags:        w3gs.ChatMessage,
				Content:      "I come from the darkness of the pit",
			},
		},
		&w3gs.MessageRelay{
			Message: w3gs.Message{
				RecipientIDs: []uint8{1, 2},
				SenderID:     4,
				Flags:        w3gs.ChatMessageExtra,
				ExtraFlags:   5,
				Content:      "Pitiful",
			},
		},

		&w3gs.SearchGame{},
		&w3gs.SearchGame{
			GameVersion: w3gs.GameVersion{
				TFT:     true,
				Version: 666,
			},
		},
		&w3gs.GameInfo{},
		&w3gs.GameInfo{
			GameVersion: w3gs.GameVersion{
				TFT:     true,
				Version: 1,
			},
			HostCounter:    2,
			EntryKey:       112233,
			GameName:       "game1",
			StatString:     "xxxxx",
			SlotsTotal:     24,
			GameTypeFlags:  w3gs.GameTypeNewGame,
			SlotsAvailable: 22,
			UptimeSec:      8,
			GamePort:       9,
		},
		&w3gs.CreateGame{},
		&w3gs.CreateGame{
			GameVersion: w3gs.GameVersion{
				TFT:     true,
				Version: 2,
			},
			HostCounter: 3,
		},
		&w3gs.RefreshGame{},
		&w3gs.RefreshGame{
			HostCounter:    1,
			PlayersInGame:  2,
			SlotsAvailable: 3,
		},
		&w3gs.DecreateGame{},
		&w3gs.DecreateGame{
			HostCounter: 777,
		},
		&w3gs.ClientInfo{},
		&w3gs.ClientInfo{
			JoinCounter: 1,
			EntryKey:    2,
			PlayerID:    3,
		},

		&w3gs.MapCheck{},
		&w3gs.MapCheck{
			FilePath:          "Maps\\BootyBay.w3x",
			FileSize:          2,
			MapInfo:           3,
			FileCrcEncryption: 4,
		},
		&w3gs.StartDownload{},
		&w3gs.StartDownload{
			PlayerID: 111,
		},
		&w3gs.MapSize{},
		&w3gs.MapSize{
			Ready:   true,
			MapSize: 2,
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

	var tooShort = make([]byte, 0)
	var tooLong = make([]byte, 2048)

	for _, pkt := range types {
		var data, err = pkt.MarshalBinary()
		if err != nil {
			t.Log(reflect.TypeOf(pkt))
			t.Fatal(err)
		}

		if len(data) != cap(data) {
			switch pkt.(type) {
			case *w3gs.Message, *w3gs.MessageRelay:
				// Whitelisted
			default:
				t.Fatalf("Capacity mismatch for %v (%v != %v)", reflect.TypeOf(pkt), len(data), cap(data))
			}
		}

		var pkt2, size, e = w3gs.UnmarshalPacket(data)
		if e != nil {
			t.Log(reflect.TypeOf(pkt))
			t.Fatal(e)
		}
		if size != len(data) {
			t.Fatalf("UnmarshalPacket size mismatch for %v", reflect.TypeOf(pkt))
		}
		if reflect.TypeOf(pkt2) != reflect.TypeOf(pkt) {
			t.Fatalf("UnmarshalPacket type mismatch %v != %v", reflect.TypeOf(pkt2), reflect.TypeOf(pkt))
		}
		if !reflect.DeepEqual(pkt, pkt2) {
			t.Logf("I: %+v", pkt)
			t.Logf("O: %+v", pkt2)
			t.Errorf("UnmarshalPacket value mismatch for %v", reflect.TypeOf(pkt))
		}

		err = pkt.UnmarshalBinary(tooShort)
		if err != w3gs.ErrWrongSize {
			t.Fatalf("ErrWrongSize expected for %v", reflect.TypeOf(pkt))
		}

		err = pkt.UnmarshalBinary(tooLong)
		if err != w3gs.ErrWrongSize && err != w3gs.ErrInvalidChecksum {
			switch pkt.(type) {
			case *w3gs.UnknownPacket:
				// Whitelisted
			default:
				t.Fatalf("ErrWrongSize expected for %v", reflect.TypeOf(pkt))
			}

		}
	}
}

func TestUnmarshalPacket(t *testing.T) {
	if _, _, e := w3gs.UnmarshalPacket([]byte{w3gs.ProtocolSig, 255}); e != w3gs.ErrNoProtocolSig {
		t.Fatal("ErrNoProtocolSig expected if no size")
	}
	if _, _, e := w3gs.UnmarshalPacket([]byte{w3gs.ProtocolSig, 255, 255, 0}); e != w3gs.ErrMalformedData {
		t.Fatal("errMalformedData expected if invalid size")
	}

	var packet = make([]byte, 2048)
	packet[0] = w3gs.ProtocolSig
	packet[1] = w3gs.PidSlotInfoJoin
	packet[3] = 8
	if _, _, e := w3gs.UnmarshalPacket(packet); e != w3gs.ErrWrongSize {
		t.Fatal("ErrWrongSize expected if invalid data")
	}
}

func BenchmarkMarshalBinary(b *testing.B) {
	for n := 0; n < b.N; n++ {
		var pkt = w3gs.SlotInfoJoin{
			SlotInfo: w3gs.SlotInfo{
				Slots:      sd,
				RandomSeed: rand.Uint32(),
				SlotLayout: w3gs.LayoutMelee,
				NumPlayers: uint8(rand.Intn(24)),
			},
			PlayerID: uint8(rand.Intn(2552)),
			ExternalAddr: w3gs.ConnAddr{
				Port: uint16(rand.Intn(65534)),
				IP:   net.IPv4bcast,
			},
		}
		pkt.MarshalBinary()
	}
}

func BenchmarkUnmarshalBinary(b *testing.B) {
	var pkt = w3gs.SlotInfoJoin{
		SlotInfo: w3gs.SlotInfo{
			Slots:      sd,
			RandomSeed: rand.Uint32(),
			SlotLayout: w3gs.LayoutMelee,
			NumPlayers: uint8(rand.Intn(24)),
		},
		PlayerID: uint8(rand.Intn(2552)),
		ExternalAddr: w3gs.ConnAddr{
			Port: uint16(rand.Intn(65534)),
			IP:   net.IPv4bcast,
		},
	}
	var data, _ = pkt.MarshalBinary()

	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		var res w3gs.SlotInfoJoin
		res.UnmarshalBinary(data)
	}
}
