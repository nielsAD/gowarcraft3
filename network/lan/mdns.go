// Author:  Niels A.D.
// Project: gowarcraft3 (https://github.com/nielsAD/gowarcraft3)
// License: Mozilla Public License, v2.0

package lan

import (
	"fmt"
	"io"
	"net"
	"strings"
	"time"

	"github.com/miekg/dns"

	"github.com/nielsAD/gowarcraft3/network"
	"github.com/nielsAD/gowarcraft3/protocol/w3gs"
)

// MulticastGroup endpoint
var MulticastGroup = net.UDPAddr{IP: net.IPv4(224, 0, 0, 251), Port: 5353}

// TypeCacheFlush bit
var TypeCacheFlush uint16 = 1 << 15

// TypeUnicastResponse bit
var TypeUnicastResponse uint16 = 1 << 15

func mdnsService(gv *w3gs.GameVersion) string {
	return strings.ToLower(fmt.Sprintf("_%s%x._sub._blizzard._udp.local.", gv.Product.String(), gv.Version))
}

// DNSPacketConn manages a UDP connection that transfers DNS packets.
// Public methods/fields are thread-safe unless explicitly stated otherwise
type DNSPacketConn struct {
	cmut network.RWMutex
	conn net.PacketConn
	wto  time.Duration

	msg dns.Msg
	buf [2048]byte
}

// NewDNSPacketConn returns conn wrapped in DNSPacketConn
func NewDNSPacketConn(conn net.PacketConn) *DNSPacketConn {
	return &DNSPacketConn{
		conn: conn,
	}
}

// Conn returns the underlying net.PacketConn
func (c *DNSPacketConn) Conn() net.PacketConn {
	c.cmut.RLock()
	var conn = c.conn
	c.cmut.RUnlock()
	return conn
}

// SetConn closes the old connection and starts using the new net.PacketConn
func (c *DNSPacketConn) SetConn(conn net.PacketConn) {
	c.Close()
	c.cmut.Lock()
	c.conn = conn
	c.cmut.Unlock()
}

// SetWriteTimeout for Send() calls
func (c *DNSPacketConn) SetWriteTimeout(wto time.Duration) {
	c.cmut.Lock()
	c.wto = wto
	c.cmut.Unlock()
}

// Close the connection
func (c *DNSPacketConn) Close() error {
	c.cmut.RLock()

	var err error
	if c.conn != nil {
		err = c.conn.Close()
	}

	c.cmut.RUnlock()

	return err
}

// Send pkt to addr over net.PacketConn
func (c *DNSPacketConn) Send(addr net.Addr, pkt *dns.Msg) (int, error) {
	c.cmut.RLock()

	if c.conn == nil {
		c.cmut.RUnlock()
		return 0, io.EOF
	}

	var n = 0

	raw, err := pkt.Pack()
	if err == nil && c.wto >= 0 {
		err = c.conn.SetWriteDeadline(network.Deadline(c.wto))
	}
	if err == nil {
		n, err = c.conn.WriteTo(raw, addr)
	}
	c.cmut.RUnlock()

	return n, err
}

// Broadcast a packet over LAN
func (c *DNSPacketConn) Broadcast(pkt *dns.Msg) (int, error) {
	return c.Send(&MulticastGroup, pkt)
}

// NextPacket waits for the next packet (with given timeout) and returns its deserialized representation
// Not safe for concurrent invocation
func (c *DNSPacketConn) NextPacket(timeout time.Duration) (*dns.Msg, net.Addr, error) {
	c.cmut.RLock()

	if c.conn == nil {
		c.cmut.RUnlock()
		return nil, nil, io.EOF
	}

	if timeout >= 0 {
		if err := c.conn.SetReadDeadline(network.Deadline(timeout)); err != nil {
			c.cmut.RUnlock()
			return nil, nil, err
		}
	}

	size, addr, err := c.conn.ReadFrom(c.buf[:])
	if err != nil {
		c.cmut.RUnlock()
		return nil, nil, err
	}

	err = c.msg.Unpack(c.buf[:size])
	c.cmut.RUnlock()

	if err != nil {
		return nil, nil, err
	}

	return &c.msg, addr, err
}

// Run reads packets (with given max time between packets) from Conn and emits an event for each received packet
// Not safe for concurrent invocation
func (c *DNSPacketConn) Run(f network.Emitter, timeout time.Duration) error {
	c.cmut.RLock()
	f.Fire(network.RunStart{})
	for {
		pkt, addr, err := c.NextPacket(timeout)

		if err != nil {
			switch err.(type) {
			// Connection is still valid after these errors, only deserialization failed
			case *dns.Error:
				f.Fire(&network.AsyncError{Src: "Run[NextPacket]", Err: err})
				continue
			default:
				f.Fire(network.RunStop{})
				c.cmut.RUnlock()
				return err
			}
		}

		f.Fire(pkt, addr)
	}
}
