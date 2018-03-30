package w3gs

import (
	"math/rand"
	"net"
	"reflect"
	"testing"
)

func TestMarshalPacket(t *testing.T) {
	var types = []Packet{
		&UnknownPacket{
			Blob: []byte{ProtocolSig, 255, 4, 0},
		},
		&Ping{},
		&Ping{
			Payload: 444,
		},
		&Pong{},
		&Pong{
			Ping{Payload: 999},
		},
		&PeerPing{},
		&PeerPing{
			Payload: 123,
		},
		&PeerPong{},
		&PeerPong{
			Ping{Payload: 456},
		},

		&Join{},
		&Join{
			HostCounter: 1,
			EntryKey:    2,
			ListenPort:  3,
			PeerKey:     4,
			PlayerName:  "Grubby",
			InternalAddr: ConnAddr{
				Port: 6,
				IP:   net.IP{7, 8, 9, 10},
			},
		},
		&RejectJoin{},
		&RejectJoin{
			Reason: RejectJoinWrongPass,
		},
		&SlotInfoJoin{},
		&SlotInfoJoin{
			SlotInfo: SlotInfo{
				Slots: []SlotData{SlotData{
					1, 2, 3, true, 5, 6, 7, 8, 9,
				}},
				RandomSeed: 10,
				SlotLayout: LayoutMelee,
				NumPlayers: 12,
			},
			PlayerID: 13,
			ExternalAddr: ConnAddr{
				Port: 14,
				IP:   net.IP{15, 16, 17, 18},
			},
		},
		&SlotInfo{},
		&SlotInfo{
			Slots: []SlotData{SlotData{
				1, 2, 3, true, 5, 6, 7, 8, 9,
			}},
		},
		&PlayerInfo{},
		&PlayerInfo{
			JoinCounter: 1,
			PlayerID:    2,
			PlayerName:  "Moon",
			ExternalAddr: ConnAddr{
				Port: 4,
				IP:   net.IP{5, 6, 7, 8},
			},
			InternalAddr: ConnAddr{
				Port: 9,
				IP:   net.IP{10, 11, 12, 13},
			},
		},

		&Leave{},
		&Leave{
			Reason: LeaveLost,
		},
		&LeaveAck{},
		&PlayerKicked{},
		&PlayerKicked{
			Leave{Reason: LeaveLobby},
		},
		&PlayerLeft{},
		&PlayerLeft{
			PlayerID: 1,
			Reason:   LeaveLost,
		},

		&CountDownStart{},
		&CountDownEnd{},
		&GameLoaded{},
		&PlayerLoaded{},
		&PlayerLoaded{
			PlayerID: 123,
		},

		&StartLag{},
		&StartLag{
			Players: []LagPlayer{
				LagPlayer{1, 2},
				LagPlayer{3, 4},
				LagPlayer{5, 6},
			},
		},
		&StopLag{},
		&StopLag{
			LagPlayer{1, 2},
		},
		&DropLaggers{},

		&GameAction{},
		&GameAction{
			Data: []byte{2, 3, 4, 5, 6, 7, 8, 9},
		},
		&TimeSlot{},
		&TimeSlot{
			Fragment:        false,
			TimeIncrementMS: 50,
			Actions: []PlayerAction{
				PlayerAction{1, make([]byte, 13)},
				PlayerAction{11, make([]byte, 13)},
			},
		},
		&TimeSlot{
			Fragment:        true,
			TimeIncrementMS: 50,
			Actions: []PlayerAction{
				PlayerAction{1, make([]byte, 13)},
				PlayerAction{11, make([]byte, 13)},
			},
		},
		&TimeSlotAck{},
		&TimeSlotAck{
			Checksum: 456,
		},

		&Message{},
		&Message{
			RecipientIDs: []uint8{1, 2, 3},
			SenderID:     4,
			Flags:        ChatMessage,
			Content:      "Tremble before me",
		},
		&Message{
			RecipientIDs: []uint8{1, 2, 3},
			SenderID:     4,
			Flags:        ChatColorChange,
			NewVal:       5,
		},
		&MessageRelay{},
		&MessageRelay{
			Message{
				RecipientIDs: []uint8{1, 2, 3},
				SenderID:     4,
				Flags:        ChatMessage,
				Content:      "I come from the darkness of the pit",
			},
		},
		&MessageRelay{
			Message{
				RecipientIDs: []uint8{1, 2},
				SenderID:     4,
				Flags:        ChatMessageExtra,
				ExtraFlags:   5,
				Content:      "Pitiful",
			},
		},

		&SearchGame{},
		&SearchGame{
			GameVersion: GameVersion{
				TFT:     true,
				Version: 666,
			},
		},
		&GameInfo{},
		&GameInfo{
			GameVersion: GameVersion{
				TFT:     true,
				Version: 1,
			},
			HostCounter:    2,
			EntryKey:       112233,
			GameName:       "game1",
			StatString:     "xxxxx",
			SlotsTotal:     24,
			GameTypeFlags:  GameTypeNewGame,
			SlotsAvailable: 22,
			UptimeSec:      8,
			GamePort:       9,
		},
		&CreateGame{},
		&CreateGame{
			GameVersion: GameVersion{
				TFT:     true,
				Version: 2,
			},
			HostCounter: 3,
		},
		&RefreshGame{},
		&RefreshGame{
			HostCounter:    1,
			PlayersInGame:  2,
			SlotsAvailable: 3,
		},
		&DecreateGame{},
		&DecreateGame{
			HostCounter: 777,
		},
		&ClientInfo{},
		&ClientInfo{
			JoinCounter: 1,
			EntryKey:    2,
			PlayerID:    3,
		},

		&MapCheck{},
		&MapCheck{
			FilePath:          "Maps\\BootyBay.w3x",
			FileSize:          2,
			MapInfo:           3,
			FileCrcEncryption: 4,
		},
		&StartDownload{},
		&StartDownload{
			PlayerID: 111,
		},
		&MapSize{},
		&MapSize{
			Ready:   true,
			MapSize: 2,
		},
		&MapPart{},
		&MapPart{
			RecipientID: 1,
			SenderID:    2,
			ChunkPos:    3,
			Data:        []byte{5, 6, 7, 8, 9},
		},
		&MapPartOK{},
		&MapPartOK{
			RecipientID: 1,
			SenderID:    2,
			ChunkPos:    3,
		},
		&MapPartError{},
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
			case *Message, *MessageRelay:
				// Whitelisted
			default:
				t.Fatalf("Capacity mismatch for %v (%v != %v)", reflect.TypeOf(pkt), len(data), cap(data))
			}
		}

		var pkt2, size, e = UnmarshalPacket(data)
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
		if err != ErrWrongSize {
			t.Fatalf("ErrWrongSize expected for %v", reflect.TypeOf(pkt))
		}

		err = pkt.UnmarshalBinary(tooLong)
		if err != ErrWrongSize && err != ErrInvalidChecksum {
			switch pkt.(type) {
			case *UnknownPacket:
				// Whitelisted
			default:
				t.Fatalf("ErrWrongSize expected for %v", reflect.TypeOf(pkt))
			}

		}
	}
}

func TestUnmarshalPacket(t *testing.T) {
	if _, _, e := UnmarshalPacket([]byte{ProtocolSig, 255}); e != ErrNoProtocolSig {
		t.Fatal("ErrNoProtocolSig expected if no size")
	}
	if _, _, e := UnmarshalPacket([]byte{ProtocolSig, 255, 255, 0}); e != ErrMalformedData {
		t.Fatal("errMalformedData expected if invalid size")
	}

	var packet = make([]byte, 2048)
	packet[0] = ProtocolSig
	packet[1] = PidSlotInfoJoin
	packet[3] = 8
	if _, _, e := UnmarshalPacket(packet); e != ErrWrongSize {
		t.Fatal("ErrWrongSize expected if invalid data")
	}
}

func BenchmarkMarshalBinary(b *testing.B) {
	for n := 0; n < b.N; n++ {
		var pkt = SlotInfoJoin{
			SlotInfo: SlotInfo{
				Slots: []SlotData{
					SlotData{0, 0, 0, false, 0, 0, 0, 0, 0},
					SlotData{1, 0, 0, true, 1, 1, 0, 0, 0},
					SlotData{2, 0, 0, false, 2, 2, 0, 0, 0},
					SlotData{3, 0, 0, true, 3, 3, 0, 0, 0},
					SlotData{4, 0, 0, false, 4, 4, 0, 0, 0},
					SlotData{5, 0, 0, true, 5, 5, 0, 0, 0},
					SlotData{6, 0, 0, false, 6, 6, 0, 0, 0},
					SlotData{7, 0, 0, true, 7, 7, 0, 0, 0},
				},
				RandomSeed: rand.Uint32(),
				SlotLayout: LayoutMelee,
				NumPlayers: uint8(rand.Intn(24)),
			},
			PlayerID: uint8(rand.Intn(2552)),
			ExternalAddr: ConnAddr{
				Port: uint16(rand.Intn(65534)),
				IP:   net.IPv4bcast,
			},
		}
		pkt.MarshalBinary()
	}
}

func BenchmarkUnmarshalBinary(b *testing.B) {
	var pkt = SlotInfoJoin{
		SlotInfo: SlotInfo{
			Slots: []SlotData{
				SlotData{0, 0, 0, false, 0, 0, 0, 0, 0},
				SlotData{1, 0, 0, true, 1, 1, 0, 0, 0},
				SlotData{2, 0, 0, false, 2, 2, 0, 0, 0},
				SlotData{3, 0, 0, true, 3, 3, 0, 0, 0},
				SlotData{4, 0, 0, false, 4, 4, 0, 0, 0},
				SlotData{5, 0, 0, true, 5, 5, 0, 0, 0},
				SlotData{6, 0, 0, false, 6, 6, 0, 0, 0},
				SlotData{7, 0, 0, true, 7, 7, 0, 0, 0},
			},
			RandomSeed: rand.Uint32(),
			SlotLayout: LayoutMelee,
			NumPlayers: uint8(rand.Intn(24)),
		},
		PlayerID: uint8(rand.Intn(2552)),
		ExternalAddr: ConnAddr{
			Port: uint16(rand.Intn(65534)),
			IP:   net.IPv4bcast,
		},
	}
	var data, _ = pkt.MarshalBinary()

	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		var res SlotInfoJoin
		res.UnmarshalBinary(data)
	}
}
