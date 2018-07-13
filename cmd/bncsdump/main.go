// Author:  Niels A.D.
// Project: gowarcraft3 (https://github.com/nielsAD/gowarcraft3)
// License: Mozilla Public License, v2.0

// bncsdump is a tool that decodes and dumps bncs packets via pcap (on the wire or from a file).
package main

import (
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"reflect"
	"strconv"
	"sync"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/google/gopacket/pcap"
	"github.com/google/gopacket/tcpassembly"
	"github.com/google/gopacket/tcpassembly/tcpreader"

	"github.com/nielsAD/gowarcraft3/protocol/bncs"
)

var (
	fname   = flag.String("f", "", "Filename to read from")
	iface   = flag.String("i", "", "Interface to read packets from")
	promisc = flag.Bool("promisc", true, "Set promiscuous mode")
	snaplen = flag.Int("s", 65536, "Snap length (max number of bytes to read per packet")
	port    = flag.Int("p", 6112, "BNCS port to sniff")

	jsonout = flag.Bool("json", false, "Print machine readable format")
	bloblen = flag.Int("b", 128, "Max number of bytes to print per blob ")
)

var logOut = log.New(os.Stdout, "", log.Ltime)
var logErr = log.New(os.Stderr, "", log.Ltime)

func dumpPackets(layer string, netFlow, transFlow gopacket.Flow, r io.Reader) error {
	var buf bncs.DeserializationBuffer

	var srv = strconv.Itoa(*port)
	var src = netFlow.Src().String() + ":" + transFlow.Src().String()
	var dst = netFlow.Dst().String() + ":" + transFlow.Dst().String()
	var prf = fmt.Sprintf("[%-3v] %21v->%-21v", layer, src, dst)

	// Skip connection initializer
	if transFlow.Dst().String() == srv {
		var firstByte = []byte{0}
		if _, err := r.Read(firstByte); err != nil || firstByte[0] != bncs.ProtocolGreeting {
			return bncs.ErrNoProtocolSig
		}
	}

	for {
		var pkt bncs.Packet
		var size int
		var err error
		if transFlow.Src().String() == srv {
			pkt, size, err = bncs.DeserializeServerPacketWithBuffer(r, &buf)
		} else {
			pkt, size, err = bncs.DeserializeClientPacketWithBuffer(r, &buf)
		}
		if err == io.EOF || err == bncs.ErrNoProtocolSig {
			return err
		} else if err != nil {
			logErr.Printf("%v %-14v %v\n", prf, "ERROR", err)

			if size > len(buf.Buffer) {
				size = len(buf.Buffer)
			}

			logErr.Printf("Payload:\n%v", hex.Dump(buf.Buffer[:size]))

			if err == bncs.ErrBufferTooSmall || err == bncs.ErrInvalidPacketSize || err == bncs.ErrInvalidChecksum || err == bncs.ErrUnexpectedConst {
				continue
			} else {
				return err
			}
		}

		// Truncate blobs
		switch p := pkt.(type) {
		case *bncs.UnknownPacket:
			if len(p.Blob) > *bloblen {
				p.Blob = p.Blob[:*bloblen]
			}
		}

		var str = fmt.Sprintf("%+v", pkt)[1:]
		if *jsonout {
			if json, err := json.Marshal(pkt); err == nil {
				str = string(json)
			}
		}

		logOut.Printf("%v %-14v %v\n", prf, reflect.TypeOf(pkt).String()[6:], str)
	}
}

type streamFactory struct{}
type stream struct {
	netFlow   gopacket.Flow
	transFlow gopacket.Flow
	reader    tcpreader.ReaderStream
}

func (f *streamFactory) New(netFlow, transFlow gopacket.Flow) tcpassembly.Stream {
	var s = stream{
		netFlow:   netFlow,
		transFlow: transFlow,
		reader:    tcpreader.NewReaderStream(),
	}

	go s.run()

	return &s.reader
}

func (s *stream) run() {
	dumpPackets("TCP", s.netFlow, s.transFlow, &s.reader)
	io.Copy(ioutil.Discard, &s.reader)
}

func addHandle(h *pcap.Handle, c chan<- gopacket.Packet, wg *sync.WaitGroup) {
	if err := h.SetBPFFilter(fmt.Sprintf("tcp and port %d", *port)); err != nil {
		logErr.Fatal("BPF filter error:", err)
	}

	var src = gopacket.NewPacketSource(h, h.LinkType())

	wg.Add(1)
	go func() {
		defer h.Close()
		defer wg.Done()

		for {
			p, err := src.NextPacket()
			if err == io.EOF {
				break
			} else if err != nil {
				logErr.Println("Sniffing error:", err)
			} else {
				c <- p
			}
		}
	}()
}

func main() {
	flag.Parse()
	if *jsonout {
		logOut.SetFlags(0)
	}

	var wg sync.WaitGroup
	var packets = make(chan gopacket.Packet)

	if *fname != "" {
		var handle, err = pcap.OpenOffline(*fname)
		if err != nil {
			logErr.Fatal("Could not open pcap file:", err)
		}
		addHandle(handle, packets, &wg)
	} else if *iface != "" {
		var handle, err = pcap.OpenLive(*iface, int32(*snaplen), *promisc, pcap.BlockForever)
		if err != nil {
			if devs, e := pcap.FindAllDevs(); e == nil {
				logErr.Print("Following interfaces are available:")
				for _, d := range devs {
					logErr.Printf("%v\t%v\n", d.Name, d.Description)
					for _, a := range d.Addresses {
						logErr.Printf("\t%v\n", a.IP)
					}
				}

				logErr.Fatalf("Could not create pcap handle: %v", err)
			}
		}
		addHandle(handle, packets, &wg)
	} else {
		var devs, err = pcap.FindAllDevs()
		if err != nil {
			logErr.Fatalf("Could not iterate interfaces: %v", err)
		}

		for _, d := range devs {
			if len(d.Addresses) == 0 {
				continue
			}

			var handle, err = pcap.OpenLive(d.Name, int32(*snaplen), *promisc, pcap.BlockForever)
			if err != nil {
				logErr.Fatalf("Could not create pcap handle: %v", err)
			}
			addHandle(handle, packets, &wg)
			logErr.Printf("Sniffing %v\n", d.Name)
		}
	}

	var asm = tcpassembly.NewAssembler(tcpassembly.NewStreamPool(&streamFactory{}))

	go func() {
		for packet := range packets {
			trans, ok := packet.TransportLayer().(*layers.TCP)
			if !ok {
				continue
			}

			asm.Assemble(packet.NetworkLayer().NetworkFlow(), trans)
		}
	}()

	wg.Wait()
	close(packets)
}
