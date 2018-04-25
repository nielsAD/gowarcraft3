// Author:  Niels A.D.
// Project: gowarcraft3 (https://github.com/nielsAD/gowarcraft3)
// License: Mozilla Public License, v2.0
package bncs_test

import (
	"bytes"
	"net"
	"reflect"
	"testing"

	"github.com/nielsAD/gowarcraft3/protocol"
	"github.com/nielsAD/gowarcraft3/protocol/bncs"
	"github.com/nielsAD/gowarcraft3/protocol/w3gs"
)

func TestClientPackets(t *testing.T) {
	var types = []bncs.Packet{
		&bncs.UnknownPacket{
			ID:   255,
			Blob: []byte{bncs.ProtocolSig, 255, 4, 0},
		},
		&bncs.KeepAlive{},
		&bncs.Ping{},
		&bncs.Ping{
			Payload: 123,
		},
		&bncs.EnterChatReq{},
		&bncs.JoinChannel{},
		&bncs.JoinChannel{
			Flag:    bncs.ChannelJoinFirst,
			Channel: "The Void",
		},
		&bncs.ChatCommand{},
		&bncs.ChatCommand{
			Text: "I come from the darkness of the pit.",
		},
		&bncs.GetAdvListReq{},
		&bncs.GetAdvListReq{
			Filter:        w3gs.GameFlagMapTypeMelee,
			FilterMask:    w3gs.GameFlagMapTypeMask,
			NumberOfGames: 2,
			GameName:      "345",
		},
		&bncs.StartAdvex3Req{},
		&bncs.StartAdvex3Req{
			GameStateFlags: bncs.GameStateFlagPrivate,
			UptimeSec:      2,
			GameFlags:      w3gs.GameFlagSizeLarge,
			Ladder:         true,
			GameName:       "Test",
			GameSettings: bncs.GameSettings{
				SlotsFree:   5,
				HostCounter: 6,
				GameSettings: w3gs.GameSettings{
					GameSettingFlags: w3gs.SettingSpeedNormal,
					MapWidth:         1,
					MapHeight:        2,
					MapXoro:          3,
					MapPath:          "4",
					HostName:         "5",
				},
			},
		},
		&bncs.StopAdv{},
		&bncs.NotifyJoin{},
		&bncs.NotifyJoin{
			GameName: "GameGameNameName",
		},
		&bncs.NetGamePort{},
		&bncs.NetGamePort{
			Port: 6112,
		},
		&bncs.AuthInfoReq{},
		&bncs.AuthInfoReq{
			PlatformCode: protocol.DString("ix86"),
			GameVersion: w3gs.GameVersion{
				Product: w3gs.ProductROC,
				Version: 1,
			},
			LanguageCode:        protocol.DString("enUS"),
			LocalIP:             net.IP{1, 1, 1, 1},
			TimeZoneBias:        2,
			MpqLocaleID:         3,
			UserLanguageID:      4,
			CountryAbbreviation: "NLD",
			Country:             "The Netherlands",
		},
		&bncs.AuthCheckReq{},
		&bncs.AuthCheckReq{
			ClientToken: 555,
			ExeVersion:  666,
			ExeHash:     777,
			CDKeys: []bncs.CDKey{
				bncs.CDKey{
					KeyLength:       1,
					KeyProductValue: 2,
					KeyPublicValue:  3,
				},
				bncs.CDKey{
					KeyLength:       4,
					KeyProductValue: 5,
					KeyPublicValue:  6,
				},
			},
			ExeInformation: "Warcraft III.exe",
			KeyOwnerName:   "Niels",
		},
		&bncs.AuthAccountLogonReq{},
		&bncs.AuthAccountLogonReq{
			Username: "Moon",
		},
		&bncs.AuthAccountLogonProofReq{},
	}

	for _, pkt := range types {
		var err error
		var buf = protocol.Buffer{Bytes: make([]byte, 0, 2048)}

		if err = pkt.Serialize(&buf); err != nil {
			t.Log(reflect.TypeOf(pkt))
			t.Fatal(err)
		}

		var buf2 = protocol.Buffer{Bytes: make([]byte, 0, 2048)}
		if _, err = bncs.SerializePacket(&buf2, pkt); err != nil {
			t.Log(reflect.TypeOf(pkt))
			t.Fatal(err)
		}

		if bytes.Compare(buf.Bytes, buf2.Bytes) != 0 {
			t.Fatalf("SerializePacket != packet.Serialize %v", reflect.TypeOf(pkt))
		}

		var pkt2, _, e = bncs.DeserializeClientPacket(&buf)
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
		if err != bncs.ErrInvalidPacketSize {
			t.Fatalf("ErrInvalidPacketSize expected for %v", reflect.TypeOf(pkt))
		}

		err = pkt.Deserialize(&protocol.Buffer{Bytes: make([]byte, 2048)})
		if err != bncs.ErrInvalidPacketSize && err != bncs.ErrInvalidChecksum {
			switch pkt.(type) {
			case *bncs.UnknownPacket:
				// Whitelisted
			default:
				t.Fatalf("ErrInvalidPacketSize expected for %v", reflect.TypeOf(pkt))
			}

		}
	}
}

func TestServerPackets(t *testing.T) {
	var types = []bncs.Packet{
		&bncs.UnknownPacket{
			ID:   255,
			Blob: []byte{bncs.ProtocolSig, 255, 4, 0},
		},
		&bncs.KeepAlive{},
		&bncs.Ping{},
		&bncs.Ping{
			Payload: 123,
		},
		&bncs.EnterChatResp{},
		&bncs.EnterChatResp{
			UniqueName:  "He",
			StatString:  "lo wo",
			AccountName: "rld",
		},
		&bncs.ChatEvent{},
		&bncs.ChatEvent{
			Type:      bncs.ChatTalk,
			UserFlags: 2,
			Ping:      3,
			UserName:  "Grubby",
			Text:      "Oh hi, Mark!",
		},
		&bncs.FloodDetected{},
		&bncs.MessageBox{},
		&bncs.MessageBox{
			Style:   1,
			Text:    "They came from behind",
			Caption: "Gyrocopter",
		},
		&bncs.GetAdvListResp{},
		&bncs.GetAdvListResp{
			Result: bncs.AdvListFull,
		},
		&bncs.GetAdvListResp{
			Games: []bncs.GetAdvListGame{
				bncs.GetAdvListGame{
					GameFlags:  w3gs.GameFlagCreatorUser,
					LanguageID: 1,
					Addr: protocol.SockAddr{
						Port: 6,
						IP:   net.IP{7, 8, 9, 10},
					},
					GameStateFlags: bncs.GameStateFlagInProgress,
					UptimeSec:      4,
					GameName:       "ShortName",
					GameSettings: bncs.GameSettings{
						SlotsFree:   5,
						HostCounter: 6,
						GameSettings: w3gs.GameSettings{
							GameSettingFlags: w3gs.SettingSpeedNormal,
							MapWidth:         1,
							MapHeight:        2,
							MapXoro:          3,
							MapPath:          "4",
							HostName:         "5",
						},
					},
				},
			},
		},
		&bncs.StartAdvex3Resp{},
		&bncs.StartAdvex3Resp{
			Failed: true,
		},
		&bncs.AuthInfoResp{},
		&bncs.AuthInfoResp{
			ServerToken: 2,
			MpqFileTime: 3,
			MpqFileName: "456",
			ValueString: "789",
		},
		&bncs.AuthCheckResp{},
		&bncs.AuthCheckResp{
			Result:                111,
			AdditionalInformation: "222",
		},
		&bncs.AuthAccountLogonResp{},
		&bncs.AuthAccountLogonResp{
			Result: bncs.LogonUpgradeRequired,
		},
		&bncs.AuthAccountLogonProofResp{},
		&bncs.AuthAccountLogonProofResp{
			Result: bncs.LogonProofPasswordIncorrect,
		},
		&bncs.AuthAccountLogonProofResp{
			Result:                bncs.LogonProofCustomError,
			AdditionalInformation: "Foo, bar.",
		},
	}

	for _, pkt := range types {
		var err error
		var buf = protocol.Buffer{Bytes: make([]byte, 0, 2048)}

		if err = pkt.Serialize(&buf); err != nil {
			t.Log(reflect.TypeOf(pkt))
			t.Fatal(err)
		}

		var buf2 = protocol.Buffer{Bytes: make([]byte, 0, 2048)}
		if _, err = bncs.SerializePacket(&buf2, pkt); err != nil {
			t.Log(reflect.TypeOf(pkt))
			t.Fatal(err)
		}

		if bytes.Compare(buf.Bytes, buf2.Bytes) != 0 {
			t.Fatalf("SerializePacket != packet.Serialize %v", reflect.TypeOf(pkt))
		}

		var pkt2, _, e = bncs.DeserializeServerPacket(&buf)
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
		if err != bncs.ErrInvalidPacketSize {
			t.Fatalf("ErrInvalidPacketSize expected for %v", reflect.TypeOf(pkt))
		}

		err = pkt.Deserialize(&protocol.Buffer{Bytes: make([]byte, 2048)})
		if err != bncs.ErrInvalidPacketSize && err != bncs.ErrInvalidChecksum {
			switch pkt.(type) {
			case *bncs.UnknownPacket:
				// Whitelisted
			default:
				t.Fatalf("ErrInvalidPacketSize expected for %v", reflect.TypeOf(pkt))
			}

		}
	}
}
