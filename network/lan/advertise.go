// Author:  Niels A.D.
// Project: gowarcraft3 (https://github.com/nielsAD/gowarcraft3)
// License: Mozilla Public License, v2.0

package lan

import (
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"net"
	"os"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/dedis/protobuf"
	"github.com/miekg/dns"
	"golang.org/x/net/ipv4"

	"github.com/nielsAD/gowarcraft3/network"
	"github.com/nielsAD/gowarcraft3/protocol"
	"github.com/nielsAD/gowarcraft3/protocol/w3gs"
)

// Advertiser broadcasts available game information to the Local Area Network
type Advertiser interface {
	network.Listener

	Create() error
	Refresh(slotsUsed uint32, slotsAvailable uint32) error
	Decreate() error

	Run() error
	Close() error
}

// NewAdvertiser initializes proper Advertiser type for game version
func NewAdvertiser(info *w3gs.GameInfo) (Advertiser, error) {
	if info.GameVersion.Version < 30 {
		// Use random port to not occupy port 6112 by default
		return NewUDPAdvertiser(info, 0)
	}

	return NewMDNSAdvertiser(info)
}

// UDPAdvertiser advertises a hosted game in the Local Area Network using UDP broadcast
type UDPAdvertiser struct {
	network.EventEmitter
	network.W3GSPacketConn

	imut sync.Mutex
	info w3gs.GameInfo

	created time.Time

	// Set once before Run(), read-only after that
	BroadcastInterval time.Duration
}

// NewUDPAdvertiser initializes UDPAdvertiser struct
func NewUDPAdvertiser(info *w3gs.GameInfo, port int) (*UDPAdvertiser, error) {
	conn, err := net.ListenUDP("udp4", &net.UDPAddr{Port: port})
	if err != nil {
		return nil, err
	}

	var a = UDPAdvertiser{
		info:              *info,
		created:           time.Now().Add(time.Duration(info.UptimeSec) * -time.Second),
		BroadcastInterval: 3 * time.Second,
	}

	a.InitDefaultHandlers()
	a.SetConn(conn, w3gs.NewFactoryCache(w3gs.DefaultFactory), w3gs.Encoding{GameVersion: info.GameVersion.Version})
	return &a, nil
}

// Create local game
func (a *UDPAdvertiser) Create() error {
	a.imut.Lock()
	var pkt = w3gs.CreateGame{
		GameVersion: a.info.GameVersion,
		HostCounter: a.info.HostCounter,
	}
	a.imut.Unlock()

	_, err := a.Broadcast(&pkt)
	return err
}

func (a *UDPAdvertiser) refresh() error {
	a.imut.Lock()
	var pkt = w3gs.RefreshGame{
		HostCounter:    a.info.HostCounter,
		SlotsUsed:      a.info.SlotsUsed,
		SlotsAvailable: a.info.SlotsAvailable,
	}
	a.imut.Unlock()

	_, err := a.Broadcast(&pkt)
	return err
}

// Refresh game info
func (a *UDPAdvertiser) Refresh(slotsUsed uint32, slotsAvailable uint32) error {
	a.imut.Lock()
	a.info.SlotsUsed = slotsUsed
	a.info.SlotsAvailable = slotsAvailable
	a.imut.Unlock()

	return a.refresh()
}

// Decreate game
func (a *UDPAdvertiser) Decreate() error {
	a.imut.Lock()
	var pkt = w3gs.DecreateGame{
		HostCounter: a.info.HostCounter,
	}
	a.imut.Unlock()

	_, err := a.Broadcast(&pkt)
	return err
}

// Run broadcasts gameinfo in Local Area Network
func (a *UDPAdvertiser) Run() error {
	if err := a.Create(); err != nil {
		return err
	}

	if a.BroadcastInterval > 0 {
		var ticker = time.NewTicker(a.BroadcastInterval)
		defer ticker.Stop()

		go func() {
			for range ticker.C {
				if err := a.refresh(); err != nil && !network.IsConnClosedError(err) {
					a.Fire(&network.AsyncError{Src: "Run[refresh]", Err: err})
				}
			}
		}()
	}

	return a.W3GSPacketConn.Run(&a.EventEmitter, 0)
}

// Close the connection
func (a *UDPAdvertiser) Close() error {
	if err := a.Decreate(); err != nil && !network.IsConnClosedError(err) {
		a.Fire(&network.AsyncError{Src: "Close[Decreate]", Err: err})
	}
	return a.W3GSPacketConn.Close()
}

// InitDefaultHandlers adds the default callbacks for relevant packets
func (a *UDPAdvertiser) InitDefaultHandlers() {
	a.On(&w3gs.SearchGame{}, a.onSearchGame)
}

func (a *UDPAdvertiser) onSearchGame(ev *network.Event) {
	var pkt = ev.Arg.(*w3gs.SearchGame)

	a.imut.Lock()
	if pkt.Product != a.info.Product {
		a.imut.Unlock()
		return
	}

	a.info.UptimeSec = (uint32)(time.Now().Sub(a.created).Seconds())
	if _, err := a.Send(ev.Opt[0].(net.Addr), &a.info); err != nil && !network.IsConnClosedError(err) {
		a.Fire(&network.AsyncError{Src: "onSearchGame[Send]", Err: err})
	}
	a.imut.Unlock()
}

// MDNSAdvertiser advertises a hosted game in the Local Area Network using MDNS (Bonjour)
type MDNSAdvertiser struct {
	network.EventEmitter
	DNSPacketConn

	imut sync.Mutex
	info w3gs.GameInfo

	msgc    int32
	created time.Time

	// Set once before Run(), read-only after that
	BroadcastInterval time.Duration
}

// NewMDNSAdvertiser initializes MDNSAdvertiser struct
func NewMDNSAdvertiser(info *w3gs.GameInfo) (*MDNSAdvertiser, error) {
	conn, err := net.ListenMulticastUDP("udp4", nil, &MulticastGroup)
	if err != nil {
		return nil, err
	}

	var conn4 = ipv4.NewPacketConn(conn)
	conn4.SetMulticastLoopback(true)
	conn4.SetMulticastTTL(255)

	var a = MDNSAdvertiser{
		info:              *info,
		created:           time.Now().Add(time.Duration(info.UptimeSec) * -time.Second),
		BroadcastInterval: 3 * time.Minute,
	}

	a.InitDefaultHandlers()
	a.SetConn(conn)
	return &a, nil
}

var illegalChars = regexp.MustCompile("\\W")

func (a *MDNSAdvertiser) mdnsService() string {
	a.imut.Lock()
	var res = mdnsService(&a.info.GameVersion)
	a.imut.Unlock()
	return res
}

func (a *MDNSAdvertiser) mdnsName() string {
	a.imut.Lock()
	var name = a.info.GameName
	a.imut.Unlock()

	name = illegalChars.ReplaceAllStringFunc(name, func(s string) string {
		return "\\" + s
	})

	return fmt.Sprintf("_%s._blizzard._udp.local.", name)
}

func hostname() string {
	var hostname, _ = os.Hostname()
	if strings.IndexByte(hostname, '.') == -1 {
		hostname += ".local"
	}
	return hostname + "."
}

func newMsg(id uint16) *dns.Msg {
	return &dns.Msg{
		MsgHdr: dns.MsgHdr{
			Id:            id,
			Response:      true,
			Authoritative: true,
		},
		Compress: true,
	}
}

func (a *MDNSAdvertiser) addPtr(msg *dns.Msg) {
	msg.Answer = append(msg.Answer, &dns.PTR{
		Hdr: dns.RR_Header{
			Name:   a.mdnsService(),
			Rrtype: dns.TypePTR,
			Class:  dns.ClassINET | TypeCacheFlush,
			Ttl:    4500,
		},
		Ptr: a.mdnsName(),
	})
}

func (a *MDNSAdvertiser) addTxt(msg *dns.Msg) {
	msg.Answer = append(msg.Answer, &dns.TXT{
		Hdr: dns.RR_Header{
			Name:   a.mdnsName(),
			Rrtype: dns.TypeTXT,
			Class:  dns.ClassINET | TypeCacheFlush,
			Ttl:    4500,
		},
		Txt: []string{""},
	})
}

func (a *MDNSAdvertiser) addSrv(msg *dns.Msg) {
	a.imut.Lock()
	var port = a.info.GamePort
	a.imut.Unlock()

	msg.Answer = append(msg.Answer, &dns.SRV{
		Hdr: dns.RR_Header{
			Name:   a.mdnsName(),
			Rrtype: dns.TypeSRV,
			Class:  dns.ClassINET | TypeCacheFlush,
			Ttl:    120,
		},
		Priority: 0,
		Weight:   0,
		Port:     port,
		Target:   hostname(),
	})

	a.addIP(msg)
}

func (a *MDNSAdvertiser) addIP(msg *dns.Msg) {
	addrs, _ := net.InterfaceAddrs()
	for _, a := range addrs {
		ipnet, ok := a.(*net.IPNet)
		if !ok || ipnet.IP.IsLoopback() || ipnet.IP.IsMulticast() {
			continue
		}

		if ip := ipnet.IP.To4(); ip != nil {
			msg.Extra = append(msg.Extra, &dns.A{
				Hdr: dns.RR_Header{
					Name:   hostname(),
					Rrtype: dns.TypeA,
					Class:  dns.ClassINET | TypeCacheFlush,
					Ttl:    120,
				},
				A: ip,
			})
		} else {
			msg.Extra = append(msg.Extra, &dns.AAAA{
				Hdr: dns.RR_Header{
					Name:   hostname(),
					Rrtype: dns.TypeAAAA,
					Class:  dns.ClassINET | TypeCacheFlush,
					Ttl:    120,
				},
				AAAA: ipnet.IP,
			})
		}
	}
}

func (a *MDNSAdvertiser) addGameInfo(msg *dns.Msg) {
	a.imut.Lock()

	var data = gameData{
		GameFlags:    a.info.GameFlags,
		GameSettings: a.info.GameSettings,
		SlotsTotal:   a.info.SlotsTotal,
		GameName:     a.info.GameName,
		GamePort:     a.info.GamePort,
	}

	var buf = protocol.Buffer{}
	if err := data.SerializeContent(&buf, &w3gs.Encoding{GameVersion: a.info.GameVersion.Version}); err != nil {
		buf.Truncate()
		a.Fire(&network.AsyncError{Src: "gameInfo[serializeContent]", Err: err})
	}

	a.msgc++
	var info = gameInfo{
		GameName:  a.info.GameName,
		MessageID: a.msgc,
		Options: map[string]string{
			idxGameID:         fmt.Sprintf("%d", a.info.HostCounter),
			idxGameSecret:     fmt.Sprintf("%d", a.info.EntryKey),
			idxGameCreateTime: fmt.Sprintf("%d", a.created.Unix()),
			idxPlayersMax:     fmt.Sprintf("%d", a.info.SlotsAvailable),
			idxPlayersNum:     fmt.Sprintf("%d", a.info.SlotsUsed),
			idxGameData:       base64.StdEncoding.EncodeToString(buf.Bytes),
		},
	}

	a.imut.Unlock()

	raw, err := protobuf.Encode(&info)
	if err != nil {
		a.Fire(&network.AsyncError{Src: "gameInfo[protobuf.Encode]", Err: err})
	}

	msg.Answer = append(msg.Answer, &dns.RFC3597{
		Hdr: dns.RR_Header{
			Name:   a.mdnsName(),
			Rrtype: typeGameInfo,
			Class:  dns.ClassINET | TypeCacheFlush,
			Ttl:    4500,
		},
		Rdata: hex.EncodeToString(raw),
	})
}

// Create local game
func (a *MDNSAdvertiser) Create() error {
	var msg = newMsg(0)
	a.addTxt(msg)
	a.addPtr(msg)
	a.addSrv(msg)
	a.addGameInfo(msg)

	_, err := a.Broadcast(msg)
	return err
}

func (a *MDNSAdvertiser) refresh() error {
	var msg = newMsg(0)
	a.addGameInfo(msg)

	_, err := a.Broadcast(msg)
	return err
}

// Refresh game info
func (a *MDNSAdvertiser) Refresh(slotsUsed uint32, slotsAvailable uint32) error {
	a.imut.Lock()
	a.info.SlotsUsed = slotsUsed
	a.info.SlotsAvailable = slotsAvailable
	a.imut.Unlock()

	return a.refresh()
}

// Decreate game
func (a *MDNSAdvertiser) Decreate() error {
	var msg = newMsg(0)
	a.addPtr(msg)

	msg.Answer[0].(*dns.PTR).Hdr.Ttl = 0

	_, err := a.Broadcast(msg)
	return err
}

// Run broadcasts gameinfo in Local Area Network
func (a *MDNSAdvertiser) Run() error {
	if err := a.Create(); err != nil {
		return err
	}
	defer a.Decreate()

	if a.BroadcastInterval > 0 {
		var ticker = time.NewTicker(a.BroadcastInterval)
		defer ticker.Stop()

		go func() {
			for range ticker.C {
				if err := a.refresh(); err != nil && !network.IsConnClosedError(err) {
					a.Fire(&network.AsyncError{Src: "Run[refresh]", Err: err})
				}
			}
		}()
	}

	return a.DNSPacketConn.Run(&a.EventEmitter, 0)
}

// Close the connection
func (a *MDNSAdvertiser) Close() error {
	if err := a.Decreate(); err != nil && !network.IsConnClosedError(err) {
		a.Fire(&network.AsyncError{Src: "Close[Decreate]", Err: err})
	}
	return a.DNSPacketConn.Close()
}

// InitDefaultHandlers adds the default callbacks for relevant packets
func (a *MDNSAdvertiser) InitDefaultHandlers() {
	a.On(&dns.Msg{}, a.onDNS)
}

func (a *MDNSAdvertiser) onDNS(ev *network.Event) {
	var msg = ev.Arg.(*dns.Msg)
	if len(msg.Question) == 0 {
		return
	}

	var addr = ev.Opt[0].(net.Addr)
	var service = a.mdnsService()
	var name = a.mdnsName()

	var addTxt = false
	var addPtr = false
	var addSrv = false
	var addInfo = false

	for _, q := range msg.Question {
		if q.Qclass&TypeUnicastResponse == 0 {
			addr = &MulticastGroup
		}

		if strings.EqualFold(q.Name, service) && (q.Qtype == dns.TypePTR || q.Qtype == dns.TypeANY) {
			addTxt = true
			addPtr = true
			addSrv = true
			addInfo = true
		} else if strings.EqualFold(q.Name, name) {
			switch q.Qtype {
			case dns.TypeANY:
				addSrv = true
				addTxt = true
				addInfo = true
			case dns.TypeTXT:
				addTxt = true
			case dns.TypeSRV:
				addSrv = true
			case typeGameInfo:
				addInfo = true
			}
		}
	}

	if addPtr || addTxt || addSrv || addInfo {
		var ans = newMsg(msg.Id)
		if addTxt {
			a.addTxt(ans)
		}
		if addPtr {
			a.addPtr(ans)
		}
		if addSrv {
			a.addSrv(ans)
		}
		if addInfo {
			a.addGameInfo(ans)
		}
		if _, err := a.Send(addr, ans); err != nil && !network.IsConnClosedError(err) {
			a.Fire(&network.AsyncError{Src: "onDNS[Send]", Err: err})
		}
	}
}
