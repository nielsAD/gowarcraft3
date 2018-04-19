// Author:  Niels A.D.
// Project: gowarcraft3 (https://github.com/nielsAD/gowarcraft3)
// License: Mozilla Public License, v2.0

package protocol_test

import (
	"net"
	"testing"

	"github.com/nielsAD/gowarcraft3/protocol"
)

func TestSockAddrConv(t *testing.T) {
	ipA, e := net.ResolveIPAddr("ip", "127.0.0.1")
	if e != nil {
		t.Fatal(e)
	}
	ipB := protocol.Addr(ipA)
	if ipA.String() != ipB.IPAddr().String() {
		t.Fatal("ResolveIPAddr != IpAddr")
	}

	udpA, e := net.ResolveUDPAddr("udp", "127.0.0.1:6112")
	if e != nil {
		t.Fatal(e)
	}
	udpB := protocol.Addr(udpA)
	if udpA.String() != udpB.UDPAddr().String() {
		t.Fatal("ResolveUDPAddr != UDPAddr")
	}

	tcpA, e := net.ResolveTCPAddr("tcp", "127.0.0.1:6112")
	if e != nil {
		t.Fatal(e)
	}
	tcpB := protocol.Addr(tcpA)
	if tcpA.String() != tcpB.TCPAddr().String() {
		t.Fatal("ResolveTCPAddr != TCPAddr")
	}
}
