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
			Blob: []byte{protocolsig, 255, 4, 0},
		},
		&PingFromHost{},
		&PingFromHost{
			Ping: 444,
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
			PlayerID:     13,
			ExternalPort: 14,
			ExternalIP:   net.IP{15, 16, 17, 18},
		},
		&RejectJoin{},
		&RejectJoin{
			Reason: RejectJoinWrongPass,
		},
		&PlayerInfo{},
		&PlayerInfo{
			JoinCounter:  1,
			PlayerID:     2,
			PlayerName:   "Moon",
			ExternalPort: 4,
			ExternalIP:   net.IP{5, 6, 7, 8},
			InternalPort: 9,
			InternalIP:   net.IP{10, 11, 12, 13},
		},
		&PlayerLeft{},
		&PlayerLeft{
			PlayerID: 1,
			Reason:   LeaveLost,
		},
		&PlayerLoaded{},
		&PlayerLoaded{
			PlayerID: 123,
		},
		&SlotInfo{},
		&SlotInfo{
			Slots: []SlotData{SlotData{
				1, 2, 3, true, 5, 6, 7, 8, 9,
			}},
		},
		&CountDownStart{},
		&CountDownEnd{},
		&IncomingAction{},
		&IncomingAction{
			Fragment:     false,
			SendInterval: 50,
			Actions: []PlayerAction{
				PlayerAction{1, make([]byte, 13)},
				PlayerAction{11, make([]byte, 13)},
			},
		},
		&IncomingAction{
			Fragment:     true,
			SendInterval: 50,
			Actions: []PlayerAction{
				PlayerAction{1, make([]byte, 13)},
				PlayerAction{11, make([]byte, 13)},
			},
		},
		&ChatFromHost{},
		&ChatFromHost{
			RecipientIDs: []uint8{1, 2, 3},
			SenderID:     4,
			Flags:        ChatMessage,
			ExtraFlags:   5,
			Message:      "I come from the darkness of the pit",
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
		&LeaveAck{},
		&ReqJoin{},
		&ReqJoin{
			HostCounter:  1,
			EntryKey:     2,
			ListenPort:   3,
			PeerKey:      4,
			PlayerName:   "Grubby",
			InternalPort: 6,
			InternalIP:   net.IP{7, 8, 9, 10},
		},
		&LeaveReq{},
		&LeaveReq{
			Reason: LeaveLost,
		},
		&GameLoadedSelf{},
		&OutgoingAction{},
		&OutgoingAction{
			[]byte{2, 3, 4, 5, 6, 7, 8, 9},
		},
		&OutgoingKeepAlive{},
		&OutgoingKeepAlive{
			Checksum: 456,
		},
		&ChatToHost{},
		&ChatToHost{
			Messages: []ChatToHostMessage{
				ChatToHostMessage{
					RecipientID: 1,
					SenderID:    2,
					Flags:       ChatMessage,
					Message:     "What a foolish boy!",
				},
				ChatToHostMessage{
					RecipientID: 3,
					SenderID:    4,
					Flags:       ChatColorChange,
					NewVal:      5,
				},
			},
		},
		&DropReq{},
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
			HostCounter:       2,
			EntryKey:          112233,
			GameName:          "game1",
			StatString:        "xxxxx",
			SlotsTotal:        24,
			GameTypeFlags:     GameTypeNewGame,
			SlotsAvailable:    22,
			TimeSinceCreation: 8,
			GamePort:          9,
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
		&PingFromOthers{},
		&PingFromOthers{
			Ping: 123,
		},
		&PongToOthers{},
		&PongToOthers{
			Pong: 456,
		},
		&ClientInfo{},
		&ClientInfo{
			JoinCounter: 1,
			PlayerID:    2,
		},
		&MapCheck{},
		&MapCheck{
			FilePath:          "/beans.w3x",
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
			RecipientID:         1,
			SenderID:            2,
			ChunkPositionInFile: 3,
			Data:                []byte{5, 6, 7, 8, 9},
		},
		&MapPartOK{},
		&MapPartOK{
			RecipientID:         1,
			SenderID:            2,
			ChunkPositionInFile: 3,
		},
		&MapPartError{},
		&PongToHost{},
		&PongToHost{
			Pong: 999,
		},
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
			t.Logf("Capacity mismatch for %v (%v != %v)", reflect.TypeOf(pkt), len(data), cap(data))
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
		if err != errMalformedData {
			t.Fatalf("errMalformedData expected for %v", reflect.TypeOf(pkt))
		}

		err = pkt.UnmarshalBinary(tooLong)
		if _, ok := pkt.(*UnknownPacket); err != errMalformedData && err != errInvalidChecksum && !ok {
			t.Fatalf("errMalformedData expected for %v", reflect.TypeOf(pkt))
		}
	}
}

func TestUnmarshalPacket(t *testing.T) {
	if _, _, e := UnmarshalPacket([]byte{protocolsig, 255}); e != errMalformedData {
		t.Fatal("errMalformedData expected if no size")
	}
	if _, _, e := UnmarshalPacket([]byte{protocolsig, 255, 255, 0}); e != errMalformedData {
		t.Fatal("errMalformedData expected if invalid size")
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
			PlayerID:     uint8(rand.Intn(2552)),
			ExternalPort: uint16(rand.Intn(65534)),
			ExternalIP:   net.IPv4bcast,
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
		PlayerID:     uint8(rand.Intn(2552)),
		ExternalPort: uint16(rand.Intn(65534)),
		ExternalIP:   net.IPv4bcast,
	}
	var data, _ = pkt.MarshalBinary()

	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		var res SlotInfoJoin
		res.UnmarshalBinary(data)
	}
}
