// w3gsdump is a tool that decodes and dumps w3gs packets via pcap (on the wire or from a file).
package main

import (
	"encoding/hex"
	"flag"
	"fmt"
	"log"
	"os"
	"reflect"

	"github.com/google/gopacket"
	"github.com/google/gopacket/pcap"

	"github.com/nielsAD/noot/pkg/w3gs"
)

var (
	fname   = flag.String("f", "", "Filename to read from")
	iface   = flag.String("i", "any", "Interface to read packets from")
	promisc = flag.Bool("promisc", true, "Set promiscuous mode")
	snaplen = flag.Int("s", 65536, "Snap length (number of bytes max to read per packet")
	port    = flag.Int("p", 6112, "Set port to filter")
)

func main() {
	flag.Parse()

	var handle *pcap.Handle
	var err error
	if *fname != "" {
		if handle, err = pcap.OpenOffline(*fname); err != nil {
			log.Fatal("PCAP OpenOffline error:", err)
		}
	} else {
		handle, err = pcap.OpenLive(*iface, int32(*snaplen), *promisc, pcap.BlockForever)
		if err != nil {
			log.Fatalf("Could not create pcap handle: %v", err)
		}
		defer handle.Close()
	}

	if err = handle.SetBPFFilter(fmt.Sprintf("port %v and (tcp or udp)", *port)); err != nil {
		log.Fatal("BPF filter error:", err)
	}

	var logger = log.New(os.Stdout, "", log.Ltime)

	var src = gopacket.NewPacketSource(handle, handle.LinkType())
	var in = src.Packets()
	for {
		select {
		case packet := <-in:
			var app = packet.ApplicationLayer()
			if app == nil {
				continue
			}

			var payload = app.Payload()
			for len(payload) > 0 {
				var w3p, size, err = w3gs.UnmarshalPacket(payload)
				if err != nil {
					if size > 0 {
						payload = payload[:size]
					}

					logger.Printf("[%-3v] %-32v %-17v %v\n", packet.TransportLayer().LayerType(), packet.NetworkLayer().NetworkFlow(), "ERROR", err)
					logger.Printf("Payload:\n%v", hex.Dump(payload))
					break
				}

				payload = payload[size:]
				logger.Printf("[%-3v] %-32v %-17v %+v\n", packet.TransportLayer().LayerType(), packet.NetworkLayer().NetworkFlow(), reflect.TypeOf(w3p).String()[6:], w3p)
			}
		}
	}
}
