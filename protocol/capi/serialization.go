// Author:  Niels A.D.
// Project: gowarcraft3 (https://github.com/nielsAD/gowarcraft3)
// License: Mozilla Public License, v2.0

package capi

import (
	"encoding/json"
	"io"
	"strings"
)

// Serialize packet and returns its byte representation.
func Serialize(p *Packet) ([]byte, error) {
	return json.Marshal(p)
}

// Write serializes p and writes it to w.
func Write(w io.Writer, p *Packet) error {
	return json.NewEncoder(w).Encode(p)
}

type rawPacket struct {
	Command   string          `json:"command"`
	RequestID int64           `json:"request_id"`
	Status    *Status         `json:"status,omitempty"`
	Payload   json.RawMessage `json:"payload"`
}

func (r *rawPacket) toPacket() (*Packet, error) {
	var p = &Packet{
		Command:   r.Command,
		RequestID: r.RequestID,
		Status:    r.Status,
	}

	if strings.HasSuffix(r.Command, CmdResponseSuffix) {
		p.Payload = &Response{}
	} else {
		switch r.Command {
		case CmdAuthenticate + CmdRequestSuffix:
			p.Payload = &Authenticate{}
		case CmdConnect + CmdRequestSuffix:
			p.Payload = &Connect{}
		case CmdDisconnect + CmdRequestSuffix:
			p.Payload = &Disconnect{}
		case CmdSendMessage + CmdRequestSuffix:
			p.Payload = &SendMessage{}
		case CmdSendEmote + CmdRequestSuffix:
			p.Payload = &SendEmote{}
		case CmdSendWhisper + CmdRequestSuffix:
			p.Payload = &SendWhisper{}
		case CmdKickUser + CmdRequestSuffix:
			p.Payload = &KickUser{}
		case CmdBanUser + CmdRequestSuffix:
			p.Payload = &BanUser{}
		case CmdUnbanUser + CmdRequestSuffix:
			p.Payload = &UnbanUser{}
		case CmdSetModerator + CmdRequestSuffix:
			p.Payload = &SetModerator{}
		case CmdConnectEvent + CmdRequestSuffix:
			p.Payload = &ConnectEvent{}
		case CmdDisconnectEvent + CmdRequestSuffix:
			p.Payload = &DisconnectEvent{}
		case CmdMessageEvent + CmdRequestSuffix:
			p.Payload = &MessageEvent{}
		case CmdUserUpdateEvent + CmdRequestSuffix:
			p.Payload = &UserUpdateEvent{}
		case CmdUserLeaveEvent + CmdRequestSuffix:
			p.Payload = &UserLeaveEvent{}
		default:
			p.Payload = &map[string]interface{}{}
		}
	}

	if err := json.Unmarshal(r.Payload, p.Payload); err != nil {
		return nil, err
	}

	return p, nil
}

// Deserialize reads exactly one packet from b and returns it in the proper (deserialized) packet type.
func Deserialize(b []byte) (*Packet, error) {
	var raw rawPacket
	if err := json.Unmarshal(b, &raw); err != nil {
		return nil, err
	}

	return raw.toPacket()
}

// Read exactly one packet from r and returns it in the proper (deserialized) packet type.
func Read(r io.Reader) (*Packet, error) {
	var raw rawPacket
	if err := json.NewDecoder(r).Decode(&raw); err != nil {
		return nil, err
	}

	return raw.toPacket()
}
