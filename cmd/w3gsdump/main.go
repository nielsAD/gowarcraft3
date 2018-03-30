// w3gsdump is a tool that decodes and dumps w3gs packets via pcap (on the wire or from a file).
package main

import (
	"encoding/hex"
	"flag"
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
	snaplen = flag.Int("s", 65536, "Snap length (max number of bytes to read per packet")
	bloblen = flag.Int("b", 128, "Max number of bytes to print per blob ")
)

func main() {
	flag.Parse()

	var handle *pcap.Handle
	var err error
	if *fname != "" {
		if handle, err = pcap.OpenOffline(*fname); err != nil {
			log.Fatal("Could not open pcap file:", err)
		}
	} else {
		handle, err = pcap.OpenLive(*iface, int32(*snaplen), *promisc, pcap.BlockForever)
		if err != nil {
			log.Fatalf("Could not create pcap handle: %v", err)
		}
		defer handle.Close()
	}

	if err = handle.SetBPFFilter("tcp or udp"); err != nil {
		log.Fatal("BPF filter error:", err)
	}

	var logger = log.New(os.Stdout, "", log.Ltime)

	var src = gopacket.NewPacketSource(handle, handle.LinkType())
	for packet := range src.Packets() {
		var app = packet.ApplicationLayer()
		if app == nil {
			continue
		}

		var payload = app.Payload()
		for len(payload) > 0 && payload[0] == w3gs.ProtocolSig {
			var w3p, size, err = w3gs.UnmarshalPacket(payload)
			if err != nil {
				logger.Printf("[%-3v] %-32v %-14v %v\n", packet.TransportLayer().LayerType(), packet.NetworkLayer().NetworkFlow(), "ERROR", err)
				logger.Printf("Payload:\n%v", hex.Dump(payload))
				break
			}

			// Truncate blobs
			switch p := w3p.(type) {
			case *w3gs.UnknownPacket:
				p.Blob = p.Blob[:*bloblen]
			case *w3gs.GameAction:
				p.Data = p.Data[:*bloblen]
			case *w3gs.TimeSlot:
				var blobsize = 0
				for i := 0; i < len(p.Actions); i++ {
					if blobsize+len(p.Actions[i].Data) > *bloblen {
						p.Actions[i].Data = p.Actions[i].Data[:*bloblen-blobsize]
					}
					blobsize += len(p.Actions[i].Data)
					if blobsize >= *bloblen {
						p.Actions = p.Actions[:i+1]
						break
					}
				}
			case *w3gs.MapPart:
				p.Data = p.Data[:*bloblen]
			}

			payload = payload[size:]
			logger.Printf("[%-3v] %-32v %-14v %+v\n", packet.TransportLayer().LayerType(), packet.NetworkLayer().NetworkFlow(), reflect.TypeOf(w3p).String()[6:], w3p)
		}
	}
}
