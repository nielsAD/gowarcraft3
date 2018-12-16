// Author:  Niels A.D.
// Project: gowarcraft3 (https://github.com/nielsAD/gowarcraft3)
// License: Mozilla Public License, v2.0

package capi

import (
	"encoding/json"
	"io"
	"strings"
)

// SerializePacket serializes p and writes it to w.
func SerializePacket(w io.Writer, p *Packet) error {
	return json.NewEncoder(w).Encode(p)
}

// DeserializePacket reads exactly one packet from r and returns it in the proper (deserialized) packet type.
func DeserializePacket(r io.Reader) (*Packet, error) {
	type rawPacket struct {
		Command   string          `json:"command"`
		RequestID int64           `json:"request_id"`
		Status    *Status         `json:"status,omitempty"`
		Payload   json.RawMessage `json:"payload"`
	}

	var raw rawPacket
	if err := json.NewDecoder(r).Decode(&raw); err != nil {
		return nil, err
	}

	var p = &Packet{
		Command:   raw.Command,
		RequestID: raw.RequestID,
		Status:    raw.Status,
	}

	if strings.HasSuffix(raw.Command, CmdResponseSuffix) {
		p.Payload = &Response{}
	} else {
		switch raw.Command {
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

	if err := json.Unmarshal(raw.Payload, p.Payload); err != nil {
		return nil, err
	}

	return p, nil
}
