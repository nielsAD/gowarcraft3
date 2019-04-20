// Author:  Niels A.D.
// Project: gowarcraft3 (https://github.com/nielsAD/gowarcraft3)
// License: Mozilla Public License, v2.0

package lan

import (
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"net"
	"strconv"
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

// Update event
type Update struct{}

// GameList keeps track of all the hosted games in the Local Area Network
type GameList interface {
	network.Listener
	Games() map[string]w3gs.GameInfo
	Run() error
	Close() error
}

// NewGameList initializes proper GameList type for game version
func NewGameList(gv w3gs.GameVersion) (GameList, error) {
	if gv.Version < 30 {
		// Use random port to not occupy port 6112 by default
		return NewUDPGameList(gv, 0)
	}

	return NewMDNSGameList(gv)
}

type udpIndex struct {
	source string
	gameID uint32
}

type udpRecord struct {
	w3gs.GameInfo
	created time.Time
	expires time.Time
}

// UDPGameList keeps track of all the hosted games in the Local Area Network using UDP broadcast
// Emits events for every received packet and Update{} when the output of Games() changes
// Public methods/fields are thread-safe unless explicitly stated otherwise
type UDPGameList struct {
	network.EventEmitter
	network.W3GSPacketConn

	gmut  sync.Mutex
	games map[udpIndex]*udpRecord

	// Set once before Run(), read-only after that
	GameVersion       w3gs.GameVersion
	BroadcastInterval time.Duration
}

// NewUDPGameList opens a new UDP socket to listen for LAN GameList updates
func NewUDPGameList(gv w3gs.GameVersion, port int) (*UDPGameList, error) {
	conn, err := net.ListenUDP("udp4", &net.UDPAddr{Port: port})
	if err != nil {
		return nil, err
	}

	var g = UDPGameList{
		GameVersion:       gv,
		BroadcastInterval: 15 * time.Second,
	}

	g.InitDefaultHandlers()
	g.SetConn(conn, w3gs.NewFactoryCache(w3gs.DefaultFactory), g.Encoding())
	return &g, nil
}

//Encoding for w3gs packets
func (g *UDPGameList) Encoding() w3gs.Encoding {
	return w3gs.Encoding{
		GameVersion: g.GameVersion.Version,
	}
}

// Games returns the current list of LAN games. Map key is the remote address.
func (g *UDPGameList) Games() map[string]w3gs.GameInfo {
	var res = make(map[string]w3gs.GameInfo)
	var now = time.Now()

	g.gmut.Lock()
	for k, v := range g.games {
		if v.GameVersion != g.GameVersion || v.GamePort == 0 {
			continue
		}

		var host = k.source
		if h, _, err := net.SplitHostPort(host); err == nil {
			host = fmt.Sprintf("%s:%d", h, v.GamePort)
		}

		v.UptimeSec = (uint32)(now.Sub(v.created).Seconds())
		if r, ok := res[host]; ok && r.UptimeSec < v.UptimeSec {
			continue
		}

		res[host] = v.GameInfo
	}
	g.gmut.Unlock()

	return res
}

// Make sure gmut is locked before calling
func (g *UDPGameList) initMap() {
	if g.games != nil {
		return
	}
	g.games = make(map[udpIndex]*udpRecord)
}

func (g *UDPGameList) expire() {
	var now = time.Now()
	var update = false

	g.gmut.Lock()
	for idx, game := range g.games {
		if now.After(game.expires) {
			update = true
			delete(g.games, idx)
		}
	}
	g.gmut.Unlock()

	if update {
		g.Fire(Update{})
	}
}

// InitDefaultHandlers adds the default callbacks for relevant packets
func (g *UDPGameList) InitDefaultHandlers() {
	g.On(&w3gs.RefreshGame{}, g.onRefreshGame)
	g.On(&w3gs.CreateGame{}, g.onCreateGame)
	g.On(&w3gs.DecreateGame{}, g.onDecreateGame)
	g.On(&w3gs.GameInfo{}, g.onGameInfo)
}

// Run reads packets from Conn and emits an event for each received packet
// Not safe for concurrent invocation
func (g *UDPGameList) Run() error {
	var sg = w3gs.SearchGame{
		GameVersion: g.GameVersion,
	}

	if g.BroadcastInterval != 0 {
		var broadcastTicker = time.NewTicker(g.BroadcastInterval)
		defer broadcastTicker.Stop()

		go func() {
			for range broadcastTicker.C {
				if _, err := g.Broadcast(&sg); err != nil && !network.IsConnClosedError(err) {
					g.Fire(&network.AsyncError{Src: "Run[Broadcast]", Err: err})
				}

				g.expire()
			}
		}()
	}

	if _, err := g.Broadcast(&sg); err != nil {
		return err
	}

	return g.W3GSPacketConn.Run(&g.EventEmitter, 0)
}

func (g *UDPGameList) onRefreshGame(ev *network.Event) {
	var pkt = ev.Arg.(*w3gs.RefreshGame)
	var adr = ev.Opt[0].(net.Addr)
	var idx = udpIndex{
		source: adr.String(),
		gameID: pkt.HostCounter,
	}

	g.gmut.Lock()

	game, ok := g.games[idx]
	if ok {
		game.expires = time.Now().Add(g.BroadcastInterval + 5*time.Second)

		var update = game.SlotsUsed != pkt.SlotsUsed || game.SlotsAvailable != pkt.SlotsAvailable
		if update {
			game.SlotsUsed = pkt.SlotsUsed
			game.SlotsAvailable = pkt.SlotsAvailable

			update = game.GameVersion == g.GameVersion
		}

		g.gmut.Unlock()

		if update {
			g.Fire(Update{})
		}

		return
	}

	g.initMap()
	g.games[idx] = &udpRecord{
		expires: time.Now().Add(5 * time.Minute),
	}
	g.gmut.Unlock()

	var sg = w3gs.SearchGame{
		GameVersion: g.GameVersion,
		HostCounter: pkt.HostCounter,
	}

	if _, err := g.Send(adr, &sg); err != nil && !network.IsConnClosedError(err) {
		g.Fire(&network.AsyncError{Src: "onRefreshGame[Send]", Err: err})
	}
}

func (g *UDPGameList) onCreateGame(ev *network.Event) {
	var pkt = ev.Arg.(*w3gs.CreateGame)
	var adr = ev.Opt[0].(net.Addr)
	var idx = udpIndex{
		source: adr.String(),
		gameID: pkt.HostCounter,
	}

	g.gmut.Lock()
	g.initMap()
	g.games[idx] = &udpRecord{
		created: time.Now(),
		expires: time.Now().Add(5 * time.Minute),
	}
	g.gmut.Unlock()

	if pkt.GameVersion != g.GameVersion {
		return
	}

	var sg = w3gs.SearchGame{
		GameVersion: g.GameVersion,
		HostCounter: pkt.HostCounter,
	}

	if _, err := g.Send(adr, &sg); err != nil && !network.IsConnClosedError(err) {
		g.Fire(&network.AsyncError{Src: "onCreateGame[Send]", Err: err})
	}
}

func (g *UDPGameList) onDecreateGame(ev *network.Event) {
	var pkt = ev.Arg.(*w3gs.DecreateGame)
	var adr = ev.Opt[0].(net.Addr)
	var idx = udpIndex{
		source: adr.String(),
		gameID: pkt.HostCounter,
	}

	g.gmut.Lock()
	game, update := g.games[idx]

	if update {
		update = game.GameVersion == g.GameVersion
		delete(g.games, idx)
	}

	g.gmut.Unlock()

	if update {
		g.Fire(Update{})
	}
}

func (g *UDPGameList) onGameInfo(ev *network.Event) {
	var pkt = ev.Arg.(*w3gs.GameInfo)
	var adr = ev.Opt[0].(net.Addr)
	var idx = udpIndex{
		source: adr.String(),
		gameID: pkt.HostCounter,
	}

	var update = pkt.GameVersion == g.GameVersion

	g.gmut.Lock()

	game, ok := g.games[idx]
	if !ok {
		game = &udpRecord{
			created: time.Now().Add(time.Duration(pkt.UptimeSec) * -time.Second),
		}

		g.initMap()
		g.games[idx] = game
	}

	update = update && game.GameInfo != *pkt

	game.expires = time.Now().Add(g.BroadcastInterval + 5*time.Second)
	game.GameInfo = *pkt

	g.gmut.Unlock()

	if update {
		g.Fire(Update{})
	}
}

type mdnsIndex struct {
	source  string
	service string
}

type mdnsRecord struct {
	w3gs.GameInfo
	created time.Time
	expires time.Time
	ipv4    net.IP
	ipv6    net.IP
}

// MDNSGameList keeps track of all the hosted games in the Local Area Network using MDNS (Bonjour)
// Emits events for every received packet and Update{} when the output of Games() changes
// Public methods/fields are thread-safe unless explicitly stated otherwise
type MDNSGameList struct {
	network.EventEmitter
	DNSPacketConn

	gmut  sync.Mutex
	games map[mdnsIndex]*mdnsRecord

	// Set once before Run(), read-only after that
	GameVersion       w3gs.GameVersion
	BroadcastInterval time.Duration
}

// NewMDNSGameList opens a new UDP socket to listen for MDNS GameList updates
func NewMDNSGameList(gv w3gs.GameVersion) (*MDNSGameList, error) {
	conn, err := net.ListenUDP("udp4", nil)
	if err != nil {
		return nil, err
	}

	var g = MDNSGameList{
		GameVersion:       gv,
		BroadcastInterval: 5 * time.Minute,
	}

	g.InitDefaultHandlers()
	g.SetConn(conn)
	return &g, nil
}

// Games returns the current list of LAN games. Map key is the remote address.
func (g *MDNSGameList) Games() map[string]w3gs.GameInfo {
	var res = make(map[string]w3gs.GameInfo)
	var now = time.Now()

	g.gmut.Lock()
	for _, v := range g.games {
		if v.GamePort == 0 {
			continue
		}

		var host = ""
		if v.ipv4 != nil {
			host = fmt.Sprintf("%s:%d", v.ipv4.String(), v.GamePort)
		} else if v.ipv6 != nil {
			host = fmt.Sprintf("[%s]:%d", v.ipv6.String(), v.GamePort)
		} else {
			continue
		}

		v.UptimeSec = (uint32)(now.Sub(v.created).Seconds())
		if r, ok := res[host]; ok && r.UptimeSec < v.UptimeSec {
			continue
		}

		res[host] = v.GameInfo
	}
	g.gmut.Unlock()

	return res
}

// Make sure gmut is locked before calling
func (g *MDNSGameList) initMap() {
	if g.games != nil {
		return
	}
	g.games = make(map[mdnsIndex]*mdnsRecord)
}

func (g *MDNSGameList) expire() {
	var now = time.Now()
	var update = false

	g.gmut.Lock()
	for idx, game := range g.games {
		if now.After(game.expires) {
			update = true
			delete(g.games, idx)
		}
	}
	g.gmut.Unlock()

	if update {
		g.Fire(Update{})
	}
}

func (g *MDNSGameList) queryAll() error {
	var msg = dns.Msg{
		Question: []dns.Question{
			// Multicast
			dns.Question{
				Name:   mdnsService(&g.GameVersion),
				Qtype:  dns.TypePTR,
				Qclass: dns.ClassINET,
			},
			// Unicast
			dns.Question{
				Name:   mdnsService(&g.GameVersion),
				Qtype:  dns.TypePTR,
				Qclass: dns.ClassINET | TypeUnicastResponse,
			},
		},
	}

	// Default reponse to unicast PTR query does not include Type66 as extra record, so specifically ask for it
	g.gmut.Lock()
	if len(g.games) < 100 {
		for idx := range g.games {
			msg.Question = append(msg.Question, dns.Question{
				Name:   idx.service,
				Qtype:  dns.TypeANY,
				Qclass: dns.ClassINET | TypeUnicastResponse,
			})
		}
	}
	g.gmut.Unlock()

	_, err := g.Broadcast(&msg)
	return err
}

func (g *MDNSGameList) queryGameInfo(svc string) error {
	var msg = dns.Msg{
		Question: []dns.Question{
			dns.Question{
				Name:   svc,
				Qtype:  typeGameInfo,
				Qclass: dns.ClassINET | TypeUnicastResponse,
			},
		},
	}

	_, err := g.Broadcast(&msg)
	return err
}

// InitDefaultHandlers adds the default callbacks for relevant packets
func (g *MDNSGameList) InitDefaultHandlers() {
	g.On(&dns.Msg{}, g.onDNS)
}

// Run reads packets from Conn and emits an event for each received packet
// Not safe for concurrent invocation
func (g *MDNSGameList) Run() error {

	// Query on unicast interface for quick response, listen to multicast interface for quick updates
	if m, err := net.ListenMulticastUDP("udp4", nil, &MulticastGroup); err == nil {
		var conn4 = ipv4.NewPacketConn(m)
		conn4.SetMulticastLoopback(true)
		conn4.SetMulticastTTL(255)

		var mc = NewDNSPacketConn(m)
		defer mc.Close()

		go mc.Run(&g.EventEmitter, 0)
	} else {
		g.Fire(&network.AsyncError{Src: "Run[ListenMulticastUDP]", Err: err})
	}

	if g.BroadcastInterval != 0 {
		var broadcastTicker = time.NewTicker(g.BroadcastInterval)
		defer broadcastTicker.Stop()

		go func() {
			for range broadcastTicker.C {
				if err := g.queryAll(); err != nil && !network.IsConnClosedError(err) {
					g.Fire(&network.AsyncError{Src: "Run[queryAll]", Err: err})
				}

				g.expire()
			}
		}()
	}

	if err := g.queryAll(); err != nil {
		return err
	}

	return g.DNSPacketConn.Run(&g.EventEmitter, 0)
}

func (g *MDNSGameList) processPTR(msg *dns.PTR, addr net.Addr) {
	var idx = mdnsIndex{
		source:  addr.String(),
		service: strings.ToLower(msg.Ptr),
	}

	if msg.Hdr.Ttl == 0 {
		delete(g.games, idx)
	} else {
		var expire = time.Duration(msg.Hdr.Ttl) * time.Second
		if expire < g.BroadcastInterval {
			expire = g.BroadcastInterval
		}

		g.initMap()
		g.games[idx] = &mdnsRecord{
			GameInfo: w3gs.GameInfo{
				GameVersion: g.GameVersion,
			},
			created: time.Now(),
			expires: time.Now().Add(expire),
		}
	}
}

func (g *MDNSGameList) processSRV(msg *dns.Msg, srv *dns.SRV, rec *mdnsRecord) {
	for _, x := range msg.Extra {
		if x.Header().Name != srv.Target {
			continue
		}

		switch ip := x.(type) {
		case *dns.A:
			rec.ipv4 = ip.A
		case *dns.AAAA:
			rec.ipv6 = ip.AAAA
		}
	}
}

func (g *MDNSGameList) processGameInfo(data string, rec *mdnsRecord) error {
	buf, err := hex.DecodeString(data)
	if err != nil {
		return err
	}

	var info gameInfo
	if err := protobuf.Decode(buf, &info); err != nil {
		return err
	}

	rec.GameName = info.GameName
	if i, err := strconv.ParseUint(info.Options[idxGameID], 0, 32); err == nil {
		rec.HostCounter = uint32(i)
	}
	if i, err := strconv.ParseUint(info.Options[idxGameSecret], 0, 32); err == nil {
		rec.EntryKey = uint32(i)
	}
	if i, err := strconv.ParseInt(info.Options[idxGameCreateTime], 0, 64); err == nil {
		rec.created = time.Unix(i, 0)
	}
	if i, err := strconv.ParseUint(info.Options[idxPlayersMax], 0, 32); err == nil {
		rec.SlotsAvailable = uint32(i)
	}
	if i, err := strconv.ParseUint(info.Options[idxPlayersNum], 0, 32); err == nil {
		rec.SlotsUsed = uint32(i)
	}
	if info.Options[idxGameData] != "" {
		buf, err = base64.StdEncoding.DecodeString(info.Options[idxGameData])
		if err != nil {
			return err
		}

		var data gameData
		if err := data.DeserializeContent(&protocol.Buffer{Bytes: buf}, &w3gs.Encoding{GameVersion: g.GameVersion.Version}); err != nil {
			return err
		}

		rec.GameFlags = data.GameFlags
		rec.GameSettings = data.GameSettings
		rec.SlotsTotal = data.SlotsTotal
		rec.GameName = data.GameName
		rec.GamePort = data.GamePort
	}

	return nil
}

func (g *MDNSGameList) onDNS(ev *network.Event) {
	var pkt = ev.Arg.(*dns.Msg)
	if len(pkt.Answer) == 0 {
		return
	}

	var addr = ev.Opt[0].(net.Addr)
	var update = false
	var service = mdnsService(&g.GameVersion)
	var incomplete = map[string]struct{}{}

	g.gmut.Lock()

	// First process all PTR records, don't assume answers are sorted
	for _, r := range pkt.Answer {
		ptr, ok := r.(*dns.PTR)
		if !ok || !strings.EqualFold(ptr.Hdr.Name, service) {
			continue
		}

		if ptr.Hdr.Ttl != 0 {
			incomplete[strings.ToLower(ptr.Ptr)] = struct{}{}
		} else {
			update = true
		}

		g.processPTR(ptr, addr)
	}

	for _, r := range append(pkt.Answer, pkt.Extra...) {
		var svc = strings.ToLower(r.Header().Name)
		var idx = mdnsIndex{
			source:  addr.String(),
			service: svc,
		}

		rec, ok := g.games[idx]
		if !ok {
			continue
		}

		var before = rec.GameInfo
		var expire = time.Now().Add(time.Duration(r.Header().Ttl) * time.Second)
		if rec.expires.Before(expire) {
			rec.expires = expire
		}

		switch v := r.(type) {
		case *dns.SRV:
			g.processSRV(pkt, v, rec)
		case *dns.RFC3597:
			if v.Header().Rrtype == typeGameInfo {
				delete(incomplete, svc)
				if err := g.processGameInfo(v.Rdata, rec); err != nil {
					g.Fire(&network.AsyncError{Src: "onDNS[processGameInfo]", Err: err})
				}
			}
		}

		if rec.GameInfo != before {
			update = true
		}
	}

	g.gmut.Unlock()

	if update {
		g.Fire(Update{})
	}

	// Query extra info for PTR records without matching Type66 record in response
	for svc := range incomplete {
		if err := g.queryGameInfo(svc); err != nil && !network.IsConnClosedError(err) {
			g.Fire(&network.AsyncError{Src: "onDNS[queryGameInfo]", Err: err})
		}
	}
}
