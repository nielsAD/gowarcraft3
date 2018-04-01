// w3gsdump is a tool that decodes and dumps w3gs packets via pcap (on the wire or from a file).
package main

import (
	"bytes"
	"encoding/hex"
	"flag"
	"io"
	"io/ioutil"
	"log"
	"os"
	"reflect"
	"sync"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/google/gopacket/pcap"
	"github.com/google/gopacket/tcpassembly"
	"github.com/google/gopacket/tcpassembly/tcpreader"

	"github.com/nielsAD/noot/pkg/w3gs"
)

var (
	fname   = flag.String("f", "", "Filename to read from")
	iface   = flag.String("i", "", "Interface to read packets from")
	promisc = flag.Bool("promisc", true, "Set promiscuous mode")
	snaplen = flag.Int("s", 65536, "Snap length (max number of bytes to read per packet")
	bloblen = flag.Int("b", 128, "Max number of bytes to print per blob ")
)

var logger = log.New(os.Stdout, "", log.Ltime)

func dumpPackets(layer string, flow gopacket.Flow, r io.Reader) {
	var buf [2048]byte
	for {
		var pkt, size, err = w3gs.UnmarshalPacketWithBuffer(r, buf[:])
		if err == io.EOF || err == w3gs.ErrNoProtocolSig {
			break
		} else if err != nil {
			logger.Printf("[%-3v] %-32v %-14v %v\n", layer, flow, "ERROR", err)
			logger.Printf("Payload:\n%v", hex.Dump(buf[:size]))
			break
		}

		// Truncate blobs
		switch p := pkt.(type) {
		case *w3gs.UnknownPacket:
			if len(p.Blob) > *bloblen {
				p.Blob = p.Blob[:*bloblen]
			}
		case *w3gs.GameAction:
			if len(p.Data) > *bloblen {
				p.Data = p.Data[:*bloblen]
			}
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

		logger.Printf("[%-3v] %-32v %-14v %+v\n", layer, flow, reflect.TypeOf(pkt).String()[6:], pkt)
	}
}

type streamFactory struct{}
type stream struct {
	flow   gopacket.Flow
	reader tcpreader.ReaderStream
}

func (f *streamFactory) New(networkFlow, transportFlow gopacket.Flow) tcpassembly.Stream {
	var s = stream{
		flow:   networkFlow,
		reader: tcpreader.NewReaderStream(),
	}

	go s.run()

	return &s.reader
}

func (s *stream) run() {
	dumpPackets("TCP", s.flow, &s.reader)
	io.Copy(ioutil.Discard, &s.reader)
}

func addHandle(h *pcap.Handle, c chan<- gopacket.Packet, wg *sync.WaitGroup) {
	if err := h.SetBPFFilter("(tcp and portrange 1000-32768) or (udp and port 6112)"); err != nil {
		log.Fatal("BPF filter error:", err)
	}

	var src = gopacket.NewPacketSource(h, h.LinkType())

	wg.Add(1)
	go func() {
		defer h.Close()
		defer wg.Done()
		for p := range src.Packets() {
			c <- p
		}
	}()
}

func main() {
	flag.Parse()

	var wg sync.WaitGroup
	var packets = make(chan gopacket.Packet)

	if *fname != "" {
		var handle, err = pcap.OpenOffline(*fname)
		if err != nil {
			log.Fatal("Could not open pcap file:", err)
		}
		addHandle(handle, packets, &wg)
	} else if *iface != "" {
		var handle, err = pcap.OpenLive(*iface, int32(*snaplen), *promisc, pcap.BlockForever)
		if err != nil {
			log.Fatalf("Could not create pcap handle: %v", err)

			if devs, e := pcap.FindAllDevs(); e == nil {
				log.Print("Following interfaces are available:")
				for _, d := range devs {
					log.Printf("%v\t%v\n", d.Name, d.Description)
					for _, a := range d.Addresses {
						log.Printf("\t%v\n", a.IP)
					}
				}
			}
		}
		addHandle(handle, packets, &wg)
	} else {
		var devs, err = pcap.FindAllDevs()
		if err != nil {
			log.Fatalf("Could not iterate interfaces: %v", err)
		}

		for _, d := range devs {
			var handle, err = pcap.OpenLive(d.Name, int32(*snaplen), *promisc, pcap.BlockForever)
			if err != nil {
				log.Fatalf("Could not create pcap handle: %v", err)
			}
			addHandle(handle, packets, &wg)
			log.Printf("Sniffing %v\n", d.Name)
		}
	}

	var asm = tcpassembly.NewAssembler(tcpassembly.NewStreamPool(&streamFactory{}))

	go func() {
		for packet := range packets {
			switch tcp := packet.TransportLayer().(type) {
			case *layers.TCP:
				asm.Assemble(packet.NetworkLayer().NetworkFlow(), tcp)
			case *layers.UDP:
				dumpPackets("UDP", packet.NetworkLayer().NetworkFlow(), bytes.NewBuffer(packet.ApplicationLayer().Payload()))
			}
		}
	}()

	wg.Wait()
	close(packets)
}
