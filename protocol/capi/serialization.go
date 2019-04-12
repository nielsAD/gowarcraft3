// Author:  Niels A.D.
// Project: gowarcraft3 (https://github.com/nielsAD/gowarcraft3)
// License: Mozilla Public License, v2.0

package capi

import (
	"encoding/json"
	"io"
	"strings"
)

// PayloadFactory returns a struct of the appropiate type for a Payload ID
type PayloadFactory interface {
	NewPayload(cmd string) interface{}
}

// FactoryFunc creates new payload container
type FactoryFunc func() interface{}

// MapFactory implements PayloadFactory using a map
type MapFactory map[string]FactoryFunc

// NewPayload implements PayloadFactory interface
func (f MapFactory) NewPayload(cmd string) interface{} {
	fun, ok := f[cmd]
	if !ok {
		if strings.HasSuffix(cmd, CmdResponseSuffix) {
			return &Response{}
		}
		return &map[string]interface{}{}
	}
	return fun()
}

type rawPacket struct {
	Command   string          `json:"command"`
	RequestID int64           `json:"request_id"`
	Status    *Status         `json:"status,omitempty"`
	Payload   json.RawMessage `json:"payload"`
}

func (r *rawPacket) toPacket(f PayloadFactory) (*Packet, error) {
	var p = &Packet{
		Command:   r.Command,
		RequestID: r.RequestID,
		Status:    r.Status,
	}

	if f == nil {
		f = DefaultFactory
	}

	p.Payload = f.NewPayload(r.Command)
	if p.Payload != nil {
		if err := json.Unmarshal(r.Payload, p.Payload); err != nil {
			return nil, err
		}
	}

	return p, nil
}

// DeserializeWithFactory reads exactly one packet from b and returns it in the proper (deserialized) packet type.
func DeserializeWithFactory(b []byte, f PayloadFactory) (*Packet, error) {
	var raw rawPacket
	if err := json.Unmarshal(b, &raw); err != nil {
		return nil, err
	}

	return raw.toPacket(f)
}

// ReadWithFactory exactly one packet from r and returns it in the proper (deserialized) packet type.
func ReadWithFactory(r io.Reader, f PayloadFactory) (*Packet, error) {
	var raw rawPacket
	if err := json.NewDecoder(r).Decode(&raw); err != nil {
		return nil, err
	}

	return raw.toPacket(f)
}

// Serialize packet and returns its byte representation.
func Serialize(p *Packet) ([]byte, error) {
	return json.Marshal(p)
}

// Deserialize reads exactly one packet from b and returns it in the proper (deserialized) packet type.
func Deserialize(b []byte) (*Packet, error) {
	return DeserializeWithFactory(b, nil)
}

// Read exactly one packet from r and returns it in the proper (deserialized) packet type.
func Read(r io.Reader) (*Packet, error) {
	return ReadWithFactory(r, nil)
}

// Write serializes p and writes it to w.
func Write(w io.Writer, p *Packet) error {
	return json.NewEncoder(w).Encode(p)
}
