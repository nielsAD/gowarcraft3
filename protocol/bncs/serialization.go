// Author:  Niels A.D.
// Project: gowarcraft3 (https://github.com/nielsAD/gowarcraft3)
// License: Mozilla Public License, v2.0

package bncs

import (
	"io"
	"io/ioutil"

	"github.com/nielsAD/gowarcraft3/protocol"
)

// SerializationBuffer is used by SerializePacketWithBuffer to bring amortized allocs to 0 for repeated calls
type SerializationBuffer = protocol.Buffer

// SerializePacketWithBuffer serializes p and writes it to w.
func SerializePacketWithBuffer(w io.Writer, b *SerializationBuffer, p Packet) (int, error) {
	b.Truncate()
	if err := p.Serialize(b); err != nil {
		return 0, err
	}
	return w.Write(b.Bytes)
}

// SerializePacket serializes p and writes it to w.
func SerializePacket(w io.Writer, p Packet) (int, error) {
	return SerializePacketWithBuffer(w, &SerializationBuffer{}, p)
}

// DeserializationBuffer is used by DeserializePacketWithBuffer to bring amortized allocs to 0 for repeated calls
type DeserializationBuffer struct {
	Buffer                    [4096]byte
	keepAlive                 KeepAlive
	ping                      Ping
	enterChatReq              EnterChatReq
	enterChatResp             EnterChatResp
	joinChannel               JoinChannel
	chatCommand               ChatCommand
	chatEvent                 ChatEvent
	floodDetected             FloodDetected
	messageBox                MessageBox
	getAdvListResp            GetAdvListResp
	getAdvListReq             GetAdvListReq
	startAdvex3Resp           StartAdvex3Resp
	startAdvex3Req            StartAdvex3Req
	stopAdv                   StopAdv
	notifyJoin                NotifyJoin
	netGamePort               NetGamePort
	authInfoResp              AuthInfoResp
	authInfoReq               AuthInfoReq
	authCheckResp             AuthCheckResp
	authCheckReq              AuthCheckReq
	authAccountCreateResp     AuthAccountCreateResp
	authAccountCreateReq      AuthAccountCreateReq
	authAccountLogonResp      AuthAccountLogonResp
	authAccountLogonReq       AuthAccountLogonReq
	authAccountLogonProofResp AuthAccountLogonProofResp
	authAccountLogonProofReq  AuthAccountLogonProofReq
	setEmail                  SetEmail
	unknownPacket             UnknownPacket
}

// ReadPacketWithBuffer reads exactly one packet from r and returns its raw bytes.
func ReadPacketWithBuffer(r io.Reader, b *DeserializationBuffer) ([]byte, int, error) {
	if n, err := io.ReadFull(r, b.Buffer[:4]); err != nil {
		if err == io.ErrUnexpectedEOF {
			err = ErrNoProtocolSig
		}

		return nil, n, err
	}

	if b.Buffer[0] != ProtocolSig {
		return nil, 4, ErrNoProtocolSig
	}

	var size = int(uint16(b.Buffer[3])<<8 | uint16(b.Buffer[2]))
	if size < 4 {
		return nil, 4, ErrNoProtocolSig
	} else if size > len(b.Buffer) {
		if n, err := io.CopyN(ioutil.Discard, r, int64(size-4)); err != nil {
			return nil, int(n + 4), err
		}
		return nil, size, ErrBufferTooSmall
	}

	if n, err := io.ReadFull(r, b.Buffer[4:size]); err != nil {
		if err == io.EOF {
			err = io.ErrUnexpectedEOF
		}
		return nil, n + 4, err
	}

	return b.Buffer[:size], size, nil
}

// DeserializeClientPacketWithBuffer reads exactly one client packet from r and returns it in the proper (deserialized) packet type.
func DeserializeClientPacketWithBuffer(r io.Reader, b *DeserializationBuffer) (Packet, int, error) {
	var bytes, n, err = ReadPacketWithBuffer(r, b)
	if err != nil {
		return nil, n, err
	}

	var pbuf = protocol.Buffer{Bytes: bytes}

	var pkt Packet

	// Explicitly call deserialize on type instead of interface for compiler optimizations
	switch b.Buffer[1] {
	case PidNull:
		err = b.keepAlive.Deserialize(&pbuf)
		pkt = &b.keepAlive
	case PidStopAdv:
		err = b.stopAdv.Deserialize(&pbuf)
		pkt = &b.stopAdv
	case PidGetAdvListEx:
		err = b.getAdvListReq.Deserialize(&pbuf)
		pkt = &b.getAdvListReq
	case PidEnterChat:
		err = b.enterChatReq.Deserialize(&pbuf)
		pkt = &b.enterChatReq
	case PidJoinChannel:
		err = b.joinChannel.Deserialize(&pbuf)
		pkt = &b.joinChannel
	case PidChatCommand:
		err = b.chatCommand.Deserialize(&pbuf)
		pkt = &b.chatCommand
	case PidStartAdvex3:
		err = b.startAdvex3Req.Deserialize(&pbuf)
		pkt = &b.startAdvex3Req
	case PidNotifyJoin:
		err = b.notifyJoin.Deserialize(&pbuf)
		pkt = &b.notifyJoin
	case PidPing:
		err = b.ping.Deserialize(&pbuf)
		pkt = &b.ping
	case PidNetGamePort:
		err = b.netGamePort.Deserialize(&pbuf)
		pkt = &b.netGamePort
	case PidAuthInfo:
		err = b.authInfoReq.Deserialize(&pbuf)
		pkt = &b.authInfoReq
	case PidAuthCheck:
		err = b.authCheckReq.Deserialize(&pbuf)
		pkt = &b.authCheckReq
	case PidAuthAccountCreate:
		err = b.authAccountCreateReq.Deserialize(&pbuf)
		pkt = &b.authAccountCreateReq
	case PidAuthAccountLogon:
		err = b.authAccountLogonReq.Deserialize(&pbuf)
		pkt = &b.authAccountLogonReq
	case PidAuthAccountLogonProof:
		err = b.authAccountLogonProofReq.Deserialize(&pbuf)
		pkt = &b.authAccountLogonProofReq
	case PidSetEmail:
		err = b.setEmail.Deserialize(&pbuf)
		pkt = &b.setEmail
	default:
		err = b.unknownPacket.Deserialize(&pbuf)
		pkt = &b.unknownPacket
	}

	if err != nil {
		return nil, n, err
	}

	return pkt, n, nil
}

// DeserializeServerPacketWithBuffer reads exactly one server packet from r and returns it in the proper (deserialized) packet type.
func DeserializeServerPacketWithBuffer(r io.Reader, b *DeserializationBuffer) (Packet, int, error) {
	var bytes, n, err = ReadPacketWithBuffer(r, b)
	if err != nil {
		return nil, n, err
	}

	var pbuf = protocol.Buffer{Bytes: bytes}

	var pkt Packet

	// Explicitly call deserialize on type instead of interface for compiler optimizations
	switch b.Buffer[1] {
	case PidNull:
		err = b.keepAlive.Deserialize(&pbuf)
		pkt = &b.keepAlive
	case PidGetAdvListEx:
		err = b.getAdvListResp.Deserialize(&pbuf)
		pkt = &b.getAdvListResp
	case PidEnterChat:
		err = b.enterChatResp.Deserialize(&pbuf)
		pkt = &b.enterChatResp
	case PidChatEvent:
		err = b.chatEvent.Deserialize(&pbuf)
		pkt = &b.chatEvent
	case PidFloodDetected:
		err = b.floodDetected.Deserialize(&pbuf)
		pkt = &b.floodDetected
	case PidMessageBox:
		err = b.messageBox.Deserialize(&pbuf)
		pkt = &b.messageBox
	case PidStartAdvex3:
		err = b.startAdvex3Resp.Deserialize(&pbuf)
		pkt = &b.startAdvex3Resp
	case PidPing:
		err = b.ping.Deserialize(&pbuf)
		pkt = &b.ping
	case PidAuthInfo:
		err = b.authInfoResp.Deserialize(&pbuf)
		pkt = &b.authInfoResp
	case PidAuthCheck:
		err = b.authCheckResp.Deserialize(&pbuf)
		pkt = &b.authCheckResp
	case PidAuthAccountCreate:
		err = b.authAccountCreateResp.Deserialize(&pbuf)
		pkt = &b.authAccountCreateResp
	case PidAuthAccountLogon:
		err = b.authAccountLogonResp.Deserialize(&pbuf)
		pkt = &b.authAccountLogonResp
	case PidAuthAccountLogonProof:
		err = b.authAccountLogonProofResp.Deserialize(&pbuf)
		pkt = &b.authAccountLogonProofResp
	default:
		err = b.unknownPacket.Deserialize(&pbuf)
		pkt = &b.unknownPacket
	}

	if err != nil {
		return nil, n, err
	}

	return pkt, n, nil
}

// ReadPacket reads exactly one packet from r and returns it in the proper (deserialized) packet type.
func ReadPacket(r io.Reader, b *DeserializationBuffer) ([]byte, int, error) {
	return ReadPacketWithBuffer(r, &DeserializationBuffer{})
}

// DeserializeClientPacket reads exactly one packet from r and returns it in the proper (deserialized) packet type.
func DeserializeClientPacket(r io.Reader) (Packet, int, error) {
	return DeserializeClientPacketWithBuffer(r, &DeserializationBuffer{})
}

// DeserializeServerPacket reads exactly one packet from r and returns it in the proper (deserialized) packet type.
func DeserializeServerPacket(r io.Reader) (Packet, int, error) {
	return DeserializeServerPacketWithBuffer(r, &DeserializationBuffer{})
}
