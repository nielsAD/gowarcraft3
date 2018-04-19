// Author:  Niels A.D.
// Project: gowarcraft3 (https://github.com/nielsAD/gowarcraft3)
// License: Mozilla Public License, v2.0

package protocol

import "net"

// SockAddr stores a single socket connection tuple (port+ip) similar to Windows SOCKADDR_IN.
type SockAddr struct {
	Port uint16
	IP   net.IP
}

// Addr converts net.Addr to SockAddr
func Addr(a net.Addr) SockAddr {
	switch t := a.(type) {
	case *net.UDPAddr:
		return SockAddr{
			Port: uint16(t.Port),
			IP:   t.IP,
		}
	case *net.TCPAddr:
		return SockAddr{
			Port: uint16(t.Port),
			IP:   t.IP,
		}
	case *net.IPAddr:
		return SockAddr{
			IP: t.IP,
		}
	default:
		return SockAddr{}
	}
}

// Equal compares s against o and returns true if they represent the same address
func (s *SockAddr) Equal(o *SockAddr) bool {
	return o.Port == s.Port && ((o.IP == nil && s.IP == nil) || o.IP.Equal(s.IP))
}

// IPAddr converts SockAddr to net.IPAddr (ignores port)
func (s *SockAddr) IPAddr() *net.IPAddr {
	return &net.IPAddr{
		IP: s.IP,
	}
}

// UDPAddr converts SockAddr to net.UDPAddr
func (s *SockAddr) UDPAddr() *net.UDPAddr {
	return &net.UDPAddr{
		IP:   s.IP,
		Port: int(s.Port),
	}
}

// TCPAddr converts SockAddr to net.TCPAddr
func (s *SockAddr) TCPAddr() *net.TCPAddr {
	return &net.TCPAddr{
		IP:   s.IP,
		Port: int(s.Port),
	}
}
