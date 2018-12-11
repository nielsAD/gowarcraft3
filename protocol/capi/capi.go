// Author:  Niels A.D.
// Project: gowarcraft3 (https://github.com/nielsAD/gowarcraft3)
// License: Mozilla Public License, v2.0

// Package capi implements the official classic Battle.net chat API.
//
// The Chat API uses JSON with UTF8 encoding as its protocol with secure websockets as the transport.
package capi

// Endpoint for websocket connection
//
// It is recommended that the certificate is checked to ensure that the common name matches *.classic.blizzard.com
const Endpoint = "connect-bot.classic.blizzard.com/v1/rpc/chat"

// Packet structure
type Packet struct {
	Command   string      `json:"command"`
	RequestID int         `json:"request_id"`
	Status    *Status     `json:"status,omitempty"`
	Payload   interface{} `json:"payload"`
}
