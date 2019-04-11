// Author:  Niels A.D.
// Project: gowarcraft3 (https://github.com/nielsAD/gowarcraft3)
// License: Mozilla Public License, v2.0

package bncs

import (
	"io"

	"github.com/nielsAD/gowarcraft3/protocol"
)

// Encoder keeps amortized allocs at 0 for repeated Packet.Serialize calls.
// Byte slices are valid until the next Serialize() call.
type Encoder struct {
	Encoding
	buf protocol.Buffer
}

// NewEncoder initialization
func NewEncoder(e Encoding) *Encoder {
	return &Encoder{
		Encoding: e,
	}
}

// Serialize packet and returns its byte representation.
// Result is valid until the next Serialize() call.
func (enc *Encoder) Serialize(p Packet) ([]byte, error) {
	enc.buf.Truncate()
	if err := p.Serialize(&enc.buf, &enc.Encoding); err != nil {
		return nil, err
	}
	return enc.buf.Bytes, nil
}

// Write serializes p and writes it to w.
func (enc *Encoder) Write(w io.Writer, p Packet) (int, error) {
	b, err := enc.Serialize(p)
	if err != nil {
		return 0, err
	}

	return w.Write(b)
}

// Serialize serializes p and returns its byte representation.
func Serialize(p Packet, e Encoding) ([]byte, error) {
	return NewEncoder(e).Serialize(p)
}

// Write serializes p and writes it to w.
func Write(w io.Writer, p Packet, e Encoding) (int, error) {
	return NewEncoder(e).Write(w, p)
}

// Decoder keeps amortized allocs at 0 for repeated Packet.Deserialize calls.
// Packets are valid until the next Deserialize() call.
type Decoder struct {
	Encoding
	bufRaw protocol.Buffer
	bufDes protocol.Buffer

	keepAlive                      KeepAlive
	ping                           Ping
	enterChatReq                   EnterChatReq
	enterChatResp                  EnterChatResp
	joinChannel                    JoinChannel
	chatCommand                    ChatCommand
	chatEvent                      ChatEvent
	floodDetected                  FloodDetected
	messageBox                     MessageBox
	getAdvListResp                 GetAdvListResp
	getAdvListReq                  GetAdvListReq
	startAdvex3Resp                StartAdvex3Resp
	startAdvex3Req                 StartAdvex3Req
	stopAdv                        StopAdv
	notifyJoin                     NotifyJoin
	netGamePort                    NetGamePort
	authInfoResp                   AuthInfoResp
	authInfoReq                    AuthInfoReq
	authCheckResp                  AuthCheckResp
	authCheckReq                   AuthCheckReq
	authAccountCreateResp          AuthAccountCreateResp
	authAccountCreateReq           AuthAccountCreateReq
	authAccountLogonResp           AuthAccountLogonResp
	authAccountLogonReq            AuthAccountLogonReq
	authAccountLogonProofResp      AuthAccountLogonProofResp
	authAccountLogonProofReq       AuthAccountLogonProofReq
	authAccountChangePassResp      AuthAccountChangePassResp
	authAccountChangePassReq       AuthAccountChangePassReq
	authAccountChangePassProofResp AuthAccountChangePassProofResp
	authAccountChangePassProofReq  AuthAccountChangePassProofReq
	setEmail                       SetEmail
	clanInfo                       ClanInfo
	unknownPacket                  UnknownPacket
}

// NewDecoder initialization
func NewDecoder(e Encoding) *Decoder {
	return &Decoder{
		Encoding: e,
	}
}

// DeserializeClient reads exactly one client packet from b and returns it in the proper (deserialized) packet type.
// Result is valid until the next Deserialize() call.
func (dec *Decoder) DeserializeClient(b []byte) (Packet, int, error) {
	dec.bufDes.Reset(b)

	var size = dec.bufDes.Size()
	if size < 4 || b[0] != ProtocolSig {
		return nil, 0, ErrNoProtocolSig
	}

	var pkt Packet
	var err error

	// Explicitly call deserialize on type instead of interface for compiler optimizations
	switch b[1] {
	case PidNull:
		err = dec.keepAlive.Deserialize(&dec.bufDes, &dec.Encoding)
		pkt = &dec.keepAlive
	case PidStopAdv:
		err = dec.stopAdv.Deserialize(&dec.bufDes, &dec.Encoding)
		pkt = &dec.stopAdv
	case PidGetAdvListEx:
		err = dec.getAdvListReq.Deserialize(&dec.bufDes, &dec.Encoding)
		pkt = &dec.getAdvListReq
	case PidEnterChat:
		err = dec.enterChatReq.Deserialize(&dec.bufDes, &dec.Encoding)
		pkt = &dec.enterChatReq
	case PidJoinChannel:
		err = dec.joinChannel.Deserialize(&dec.bufDes, &dec.Encoding)
		pkt = &dec.joinChannel
	case PidChatCommand:
		err = dec.chatCommand.Deserialize(&dec.bufDes, &dec.Encoding)
		pkt = &dec.chatCommand
	case PidStartAdvex3:
		err = dec.startAdvex3Req.Deserialize(&dec.bufDes, &dec.Encoding)
		pkt = &dec.startAdvex3Req
	case PidNotifyJoin:
		err = dec.notifyJoin.Deserialize(&dec.bufDes, &dec.Encoding)
		pkt = &dec.notifyJoin
	case PidPing:
		err = dec.ping.Deserialize(&dec.bufDes, &dec.Encoding)
		pkt = &dec.ping
	case PidNetGamePort:
		err = dec.netGamePort.Deserialize(&dec.bufDes, &dec.Encoding)
		pkt = &dec.netGamePort
	case PidAuthInfo:
		err = dec.authInfoReq.Deserialize(&dec.bufDes, &dec.Encoding)
		pkt = &dec.authInfoReq
	case PidAuthCheck:
		err = dec.authCheckReq.Deserialize(&dec.bufDes, &dec.Encoding)
		pkt = &dec.authCheckReq
	case PidAuthAccountCreate:
		err = dec.authAccountCreateReq.Deserialize(&dec.bufDes, &dec.Encoding)
		pkt = &dec.authAccountCreateReq
	case PidAuthAccountLogon:
		err = dec.authAccountLogonReq.Deserialize(&dec.bufDes, &dec.Encoding)
		pkt = &dec.authAccountLogonReq
	case PidAuthAccountLogonProof:
		err = dec.authAccountLogonProofReq.Deserialize(&dec.bufDes, &dec.Encoding)
		pkt = &dec.authAccountLogonProofReq
	case PidAuthAccountChange:
		err = dec.authAccountChangePassReq.Deserialize(&dec.bufDes, &dec.Encoding)
		pkt = &dec.authAccountChangePassReq
	case PidAuthAccountChangeProof:
		err = dec.authAccountChangePassProofReq.Deserialize(&dec.bufDes, &dec.Encoding)
		pkt = &dec.authAccountChangePassProofReq
	case PidSetEmail:
		err = dec.setEmail.Deserialize(&dec.bufDes, &dec.Encoding)
		pkt = &dec.setEmail
	default:
		err = dec.unknownPacket.Deserialize(&dec.bufDes, &dec.Encoding)
		pkt = &dec.unknownPacket
	}

	var n = size - dec.bufDes.Size()
	if err != nil {
		return nil, n, err
	}

	return pkt, n, nil
}

// DeserializeServer reads exactly one server packet from b and returns it in the proper (deserialized) packet type.
// Result is valid until the next Deserialize() call.
func (dec *Decoder) DeserializeServer(b []byte) (Packet, int, error) {
	dec.bufDes.Reset(b)

	var size = dec.bufDes.Size()
	if size < 4 || b[0] != ProtocolSig {
		return nil, 0, ErrNoProtocolSig
	}

	var pkt Packet
	var err error

	// Explicitly call deserialize on type instead of interface for compiler optimizations
	switch b[1] {
	case PidNull:
		err = dec.keepAlive.Deserialize(&dec.bufDes, &dec.Encoding)
		pkt = &dec.keepAlive
	case PidGetAdvListEx:
		err = dec.getAdvListResp.Deserialize(&dec.bufDes, &dec.Encoding)
		pkt = &dec.getAdvListResp
	case PidEnterChat:
		err = dec.enterChatResp.Deserialize(&dec.bufDes, &dec.Encoding)
		pkt = &dec.enterChatResp
	case PidChatEvent:
		err = dec.chatEvent.Deserialize(&dec.bufDes, &dec.Encoding)
		pkt = &dec.chatEvent
	case PidFloodDetected:
		err = dec.floodDetected.Deserialize(&dec.bufDes, &dec.Encoding)
		pkt = &dec.floodDetected
	case PidMessageBox:
		err = dec.messageBox.Deserialize(&dec.bufDes, &dec.Encoding)
		pkt = &dec.messageBox
	case PidStartAdvex3:
		err = dec.startAdvex3Resp.Deserialize(&dec.bufDes, &dec.Encoding)
		pkt = &dec.startAdvex3Resp
	case PidPing:
		err = dec.ping.Deserialize(&dec.bufDes, &dec.Encoding)
		pkt = &dec.ping
	case PidAuthInfo:
		err = dec.authInfoResp.Deserialize(&dec.bufDes, &dec.Encoding)
		pkt = &dec.authInfoResp
	case PidAuthCheck:
		err = dec.authCheckResp.Deserialize(&dec.bufDes, &dec.Encoding)
		pkt = &dec.authCheckResp
	case PidAuthAccountCreate:
		err = dec.authAccountCreateResp.Deserialize(&dec.bufDes, &dec.Encoding)
		pkt = &dec.authAccountCreateResp
	case PidAuthAccountLogon:
		err = dec.authAccountLogonResp.Deserialize(&dec.bufDes, &dec.Encoding)
		pkt = &dec.authAccountLogonResp
	case PidAuthAccountLogonProof:
		err = dec.authAccountLogonProofResp.Deserialize(&dec.bufDes, &dec.Encoding)
		pkt = &dec.authAccountLogonProofResp
	case PidAuthAccountChange:
		err = dec.authAccountChangePassResp.Deserialize(&dec.bufDes, &dec.Encoding)
		pkt = &dec.authAccountChangePassResp
	case PidAuthAccountChangeProof:
		err = dec.authAccountChangePassProofResp.Deserialize(&dec.bufDes, &dec.Encoding)
		pkt = &dec.authAccountChangePassProofResp
	case PidClanInfo:
		err = dec.clanInfo.Deserialize(&dec.bufDes, &dec.Encoding)
		pkt = &dec.clanInfo
	default:
		err = dec.unknownPacket.Deserialize(&dec.bufDes, &dec.Encoding)
		pkt = &dec.unknownPacket
	}

	var n = size - dec.bufDes.Size()
	if err != nil {
		return nil, n, err
	}

	return pkt, n, nil
}

// ReadRaw reads exactly one packet from r and returns its raw bytes.
// Result is valid until the next ReadRaw() call.
func (dec *Decoder) ReadRaw(r io.Reader) ([]byte, int, error) {
	dec.bufRaw.Truncate()

	if n, err := dec.bufRaw.ReadSizeFrom(r, 4); err != nil {
		if err == io.ErrUnexpectedEOF {
			err = ErrNoProtocolSig
		}

		return nil, int(n), err
	}

	if dec.bufRaw.Bytes[0] != ProtocolSig {
		return nil, 4, ErrNoProtocolSig
	}

	var size = int(uint16(dec.bufRaw.Bytes[3])<<8 | uint16(dec.bufRaw.Bytes[2]))
	if size < 4 {
		return nil, 4, ErrNoProtocolSig
	}

	if n, err := dec.bufRaw.ReadSizeFrom(r, size-4); err != nil {
		if err == io.EOF {
			err = io.ErrUnexpectedEOF
		}
		return nil, int(n) + 4, err
	}

	return dec.bufRaw.Bytes, size, nil
}

// ReadClient exactly one client packet from r and returns it in the proper (deserialized) packet type.
func (dec *Decoder) ReadClient(r io.Reader) (Packet, int, error) {
	b, n, err := dec.ReadRaw(r)
	if err != nil {
		return nil, n, err
	}

	p, m, err := dec.DeserializeClient(b)
	if err != nil {
		return nil, n, err
	}
	if m != n {
		return nil, n, ErrInvalidPacketSize
	}

	return p, n, nil
}

// ReadServer exactly one server packet from r and returns it in the proper (deserialized) packet type.
func (dec *Decoder) ReadServer(r io.Reader) (Packet, int, error) {
	b, n, err := dec.ReadRaw(r)
	if err != nil {
		return nil, n, err
	}

	p, m, err := dec.DeserializeServer(b)
	if err != nil {
		return nil, n, err
	}
	if m != n {
		return nil, n, ErrInvalidPacketSize
	}

	return p, n, nil
}

// DeserializeClient reads exactly one client packet from b and returns it in the proper (deserialized) packet type.
func DeserializeClient(b []byte, e Encoding) (Packet, int, error) {
	return NewDecoder(e).DeserializeClient(b)
}

// DeserializeServer reads exactly one server packet from b and returns it in the proper (deserialized) packet type.
func DeserializeServer(b []byte, e Encoding) (Packet, int, error) {
	return NewDecoder(e).DeserializeServer(b)
}

// ReadClient exactly one client packet from r and returns it in the proper (deserialized) packet type.
func ReadClient(r io.Reader, e Encoding) (Packet, int, error) {
	return NewDecoder(e).ReadClient(r)
}

// ReadServer exactly one server packet from r and returns it in the proper (deserialized) packet type.
func ReadServer(r io.Reader, e Encoding) (Packet, int, error) {
	return NewDecoder(e).ReadServer(r)
}
