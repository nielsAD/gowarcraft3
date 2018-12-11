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
func SerializePacket(w io.Writer, p Packet) error {
	return json.NewEncoder(w).Encode(p)
}

// DeserializationBuffer is used by DeserializePacketWithBuffer to bring amortized allocs to 0 for repeated calls
type DeserializationBuffer struct {
	packet          Packet
	response        Response
	authenticate    Authenticate
	connect         Connect
	disconnect      Disconnect
	sendMessage     SendMessage
	sendEmote       SendEmote
	sendWhisper     SendWhisper
	kickUser        KickUser
	banUser         BanUser
	unbanUser       UnbanUser
	setModerator    SetModerator
	connectEvent    ConnectEvent
	disconnectEvent DisconnectEvent
	messageEvent    MessageEvent
	userUpdateEvent UserUpdateEvent
	userLeaveEvent  UserLeaveEvent
}

// DeserializePacketWithBuffer reads exactly one packet from r and returns it in the proper (deserialized) packet type.
func DeserializePacketWithBuffer(r io.Reader, b *DeserializationBuffer) (*Packet, error) {
	type rawPacket struct {
		Command   string          `json:"command"`
		RequestID int             `json:"request_id"`
		Status    *Status         `json:"status,omitempty"`
		Payload   json.RawMessage `json:"payload"`
	}

	var raw rawPacket
	if err := json.NewDecoder(r).Decode(&raw); err != nil {
		return nil, err
	}

	var p = &b.packet
	*p = Packet{
		Command:   raw.Command,
		RequestID: raw.RequestID,
		Status:    raw.Status,
	}

	if strings.HasSuffix(raw.Command, "Response") {
		p.Payload = &b.response
	} else {
		switch strings.TrimSuffix(raw.Command, "Request") {
		case b.authenticate.Command():
			p.Payload = &b.authenticate
		case b.connect.Command():
			p.Payload = &b.connect
		case b.disconnect.Command():
			p.Payload = &b.disconnect
		case b.sendMessage.Command():
			p.Payload = &b.sendMessage
		case b.sendEmote.Command():
			p.Payload = &b.sendEmote
		case b.sendWhisper.Command():
			p.Payload = &b.sendWhisper
		case b.kickUser.Command():
			p.Payload = &b.kickUser
		case b.banUser.Command():
			p.Payload = &b.banUser
		case b.unbanUser.Command():
			p.Payload = &b.unbanUser
		case b.setModerator.Command():
			p.Payload = &b.setModerator
		case b.connectEvent.Command():
			p.Payload = &b.connectEvent
		case b.disconnectEvent.Command():
			p.Payload = &b.disconnectEvent
		case b.messageEvent.Command():
			p.Payload = &b.messageEvent
		case b.userUpdateEvent.Command():
			p.Payload = &b.userUpdateEvent
		case b.userLeaveEvent.Command():
			p.Payload = &b.userLeaveEvent
		default:
			p.Payload = map[string]interface{}{}
		}
	}

	if err := json.Unmarshal(raw.Payload, p.Payload); err != nil {
		return nil, err
	}

	return p, nil
}

// DeserializePacket reads exactly one packet from r and returns it in the proper (deserialized) packet type.
func DeserializePacket(r io.Reader) (*Packet, error) {
	return DeserializePacketWithBuffer(r, &DeserializationBuffer{})
}
