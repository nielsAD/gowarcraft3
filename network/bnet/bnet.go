// Author:  Niels A.D.
// Project: gowarcraft3 (https://github.com/nielsAD/gowarcraft3)
// License: Mozilla Public License, v2.0

// Package bnet implements a mocked BNCS client that can be used to interact with BNCS servers.
package bnet

import (
	"io/ioutil"
	"net"
	"os"
	"path"
	"strings"
	"sync"
	"time"
	"unicode"

	"github.com/nielsAD/gowarcraft3/protocol"

	"github.com/nielsAD/gowarcraft3/network"
	"github.com/nielsAD/gowarcraft3/protocol/bncs"
	"github.com/nielsAD/gowarcraft3/protocol/w3gs"
)

// JoinError event
type JoinError struct {
	Channel string
	Error   bncs.ChatEventType
}

// Channel joined event
type Channel struct {
	Name  string
	Flags bncs.ChatChannelFlags
}

// UserJoined event
type UserJoined struct {
	User
	AlreadyInChannel bool
}

// UserLeft event
type UserLeft struct {
	User
}

// UserUpdate event
type UserUpdate struct {
	User
}

// Say event
type Say struct {
	Content string
}

// Chat event
type Chat struct {
	User
	Content string
	Type    bncs.ChatEventType
}

// Whisper event
type Whisper struct {
	UserName string
	Content  string
	Flags    bncs.ChatUserFlags
	Ping     uint32
}

// SystemMessage event
type SystemMessage struct {
	Content string
	Type    bncs.ChatEventType
}

// User in chat
type User struct {
	Name       string
	StatString string
	Flags      bncs.ChatUserFlags
	Ping       uint32
	Joined     time.Time
	LastSeen   time.Time
}

// Client represents a mocked BNCS client
// Public methods/fields are thread-safe unless explicitly stated otherwise
type Client struct {
	network.EventEmitter
	network.BNCSonn

	chatmut sync.Mutex
	channel string
	users   map[string]*User

	// Read-only
	UniqueName string

	// Set once before Dial(), read-only after that
	ServerAddr        string
	KeepAliveInterval time.Duration
	AuthInfo          bncs.AuthInfoReq
	BinPath           string
	UserName          string
	Password          string
	CDKeyOwner        string
	CDKeys            []string
	GamePort          uint16
}

// NewClient initializes a Client struct with default values
func NewClient(searchPaths ...string) *Client {
	var paths = append(searchPaths, []string{
		"C:/Program Files/Warcraft III/",
		"C:/Program Files (x86)/Warcraft III/",
		path.Join(os.Getenv("HOME"), ".wine/drive_c/Program Files/Warcraft III/"),
		path.Join(os.Getenv("HOME"), ".wine/drive_c/Program Files (x86)/Warcraft III/"),
		"./war3",
	}...)

	var bin = "."
	for i := 0; i < len(paths); i++ {
		if _, err := os.Stat(paths[i]); err == nil {
			bin = paths[i]
			break
		}
	}

	var product = w3gs.ProductROC
	var version = uint32(29)
	if exeVersion, _, err := GetExeInfo(path.Join(bin, "Warcraft III.exe")); err == nil {
		version = (exeVersion >> 16) & 0xFF
	}

	var user = "gowarcraft3"
	if _, err := os.Stat(path.Join(bin, "user.w3k")); err == nil {
		if f, err := ioutil.ReadFile(path.Join(bin, "user.w3k")); err == nil {
			user = strings.TrimSpace(string(f))
		}
	}

	var rock = ""
	if _, err := os.Stat(path.Join(bin, "roc.w3k")); err == nil {
		if f, err := ioutil.ReadFile(path.Join(bin, "roc.w3k")); err == nil {
			rock = strings.TrimSpace(string(f))
		}
	}

	var tftk = ""
	if _, err := os.Stat(path.Join(bin, "tft.w3k")); err == nil {
		if f, err := ioutil.ReadFile(path.Join(bin, "tft.w3k")); err == nil {
			tftk = strings.TrimSpace(string(f))
		}
	}

	var keys = []string{}
	if rock != "" {
		if tftk != "" {
			product = w3gs.ProductTFT
			keys = []string{rock, tftk}
		} else {
			keys = []string{rock}
		}
	}

	var c = &Client{
		KeepAliveInterval: 20 * time.Second,
		AuthInfo: bncs.AuthInfoReq{
			PlatformCode:        protocol.DString("IX86"),
			GameVersion:         w3gs.GameVersion{Product: product, Version: version},
			LanguageCode:        protocol.DString("enUS"),
			TimeZoneBias:        4294967176,
			MpqLocaleID:         1033,
			UserLanguageID:      1033,
			CountryAbbreviation: "USA",
			Country:             "United States",
		},
		BinPath:    bin,
		CDKeyOwner: user,
		CDKeys:     keys,
		GamePort:   6112,
	}

	c.InitDefaultHandlers()

	return c
}

// Channel currently chatting in
func (b *Client) Channel() string {
	b.chatmut.Lock()
	var res = b.channel
	b.chatmut.Unlock()
	return res
}

// Users in channel
func (b *Client) Users() map[string]User {
	var res = make(map[string]User)

	b.chatmut.Lock()
	for k, v := range b.users {
		res[k] = *v
	}
	b.chatmut.Unlock()

	return res
}

// Dial opens a new connection to server, verifies game version, and authenticates with CD keys
//
// Dial sequence:
//   1. C > S [0x50] SID_AUTH_INFO
//   2. S > C [0x25] SID_PING
//   3. C > S [0x25] SID_PING (optional)
//   4. S > C [0x50] SID_AUTH_INFO
//   5. C > S [0x51] SID_AUTH_CHECK
//   6. S > C [0x51] SID_AUTH_CHECK
//   7. Client gets icons file, TOS file, and server list file:
//     1. C > S [0x2D] SID_GETICONDATA (optional)
//     2. S > C [0x2D] SID_GETICONDATA (optional response)
//     3. C > S [0x33] SID_GETFILETIME (returned icons file name) (optional)
//     4. C > S [0x33] SID_GETFILETIME ("tos_USA.txt") (optional)
//     5. C > S [0x33] SID_GETFILETIME ("bnserver.ini") (optional)
//     6. S > C [0x33] SID_GETFILETIME (one for each request)
//     7. Connection to BNFTPv2 to do file downloads
//
func (b *Client) Dial() (*network.BNCSonn, error) {
	if !strings.ContainsRune(b.ServerAddr, ':') {
		b.ServerAddr += ":6112"
	}

	addr, err := net.ResolveTCPAddr("tcp4", b.ServerAddr)
	if err != nil {
		return nil, err
	}

	conn, err := net.DialTCP("tcp4", nil, addr)
	if err != nil {
		return nil, err
	}

	conn.SetNoDelay(true)
	conn.SetLinger(3)
	conn.Write([]byte{bncs.ProtocolGreeting})

	bncsconn := network.NewBNCSonn(conn)

	authInfo, err := b.sendAuthInfo(bncsconn)
	if err != nil {
		bncsconn.Close()
		return nil, err
	}

	clientToken := uint32(time.Now().Unix())
	authCheck, err := b.sendAuthCheck(bncsconn, clientToken, authInfo)
	if err != nil {
		bncsconn.Close()
		return nil, err
	}

	if authCheck.Result != bncs.AuthSuccess {
		bncsconn.Close()
		return nil, AuthResultToError(authCheck.Result)
	}

	return bncsconn, nil
}

// Logon opens a new connection to server, logs on, and joins chat
//
// Logon sequence:
//   1. Client starts with Dial sequence ([0x50] SID_AUTH_INFO and [0x51] SID_AUTH_CHECK)
//   2. Client waits for user to enter account information (standard logon shown, uses NLS):
//     1. C > S [0x53] SID_AUTH_ACCOUNTLOGON
//     2. S > C [0x53] SID_AUTH_ACCOUNTLOGON
//     3. C > S [0x54] SID_AUTH_ACCOUNTLOGONPROOF
//     4. S > C [0x54] SID_AUTH_ACCOUNTLOGONPROOF
//   3. C > S [0x45] SID_NETGAMEPORT (optional)
//   4. C > S [0x0A] SID_ENTERCHAT
//   5. S > C [0x0A] SID_ENTERCHAT
//   6. C > S [0x44] SID_WARCRAFTGENERAL (WID_TOURNAMENT) (optional)
//   7. S > C [0x44] SID_WARCRAFTGENERAL (WID_TOURNAMENT) (optional response)
//   8. C > S [0x46] SID_NEWS_INFO (optional)
//   9. S > C [0x46] SID_NEWS_INFO (optional response)
//  10. Client waits until user wants to Enter Chat.
//  11. C > S [0x0C] SID_JOINCHANNEL (First Join, "W3")
//  12. S > C [0x0F] SID_CHATEVENT
//  13. A sequence of chat events for entering chat follow.
//
func (b *Client) Logon() error {
	nls, err := NewNLS(b.UserName, b.Password)
	if err != nil {
		return err
	}

	defer nls.Free()

	bncsconn, err := b.Dial()
	if err != nil {
		return err
	}

	logon, err := b.sendLogon(bncsconn, nls)
	if err != nil {
		bncsconn.Close()
		return err
	}

	if logon.Result != bncs.LogonSuccess {
		bncsconn.Close()
		return LogonResultToError(logon.Result)
	}

	proof, err := b.sendLogonProof(bncsconn, nls, logon)
	if err != nil {
		bncsconn.Close()
		return err
	}

	switch proof.Result {
	case bncs.LogonProofSuccess:
		//nothing
	case bncs.LogonProofRequireEmail:
		if _, err := bncsconn.Send(&bncs.SetEmail{EmailAddress: ""}); err != nil {
			bncsconn.Close()
			return err
		}
	default:
		bncsconn.Close()
		return LogonProofResultToError(proof.Result)
	}

	chat, err := b.sendEnterChat(bncsconn)
	if err != nil {
		bncsconn.Close()
		return err
	}

	if _, err := bncsconn.Send(&bncs.JoinChannel{Flag: bncs.ChannelJoinFirst, Channel: "W3"}); err != nil {
		bncsconn.Close()
		return err
	}

	b.UniqueName = chat.UniqueName
	b.SetConn(bncsconn.Conn())
	return nil
}

// CreateAccount registers a new account
//
// CreateAccount sequence:
//  1. Client starts with Dial sequence
//  2. Client waits for user to enter new account information:
//    1. C > S [0x52] SID_AUTH_ACCOUNTCREATE
//    2. S > C [0x52] SID_AUTH_ACCOUNTCREATE
//  3. Client can continue with logon ([0x53] SID_AUTH_ACCOUNTLOGON)
//
func (b *Client) CreateAccount() error {
	nls, err := NewNLS(b.UserName, b.Password)
	if err != nil {
		return err
	}

	defer nls.Free()

	bncsconn, err := b.Dial()
	if err != nil {
		return err
	}

	defer bncsconn.Close()

	create, err := b.sendCreateAccount(bncsconn, nls)
	if err != nil {
		return err
	}

	if create.Result != bncs.AccountCreateSuccess {
		return AccountCreateResultToError(create.Result)
	}

	return nil
}

func (b *Client) sendAuthInfo(conn *network.BNCSonn) (*bncs.AuthInfoResp, error) {
	if _, err := conn.Send(&b.AuthInfo); err != nil {
		return nil, err
	}

	pkt, err := conn.NextServerPacket(5 * time.Second)
	if err != nil {
		return nil, err
	}
	switch p := pkt.(type) {
	case *bncs.Ping:
		if _, err := conn.Send(p); err != nil {
			return nil, err
		}
	default:
		return nil, ErrUnexpectedPacket
	}

	pkt, err = conn.NextServerPacket(5 * time.Second)
	if err != nil {
		return nil, err
	}
	switch p := pkt.(type) {
	case *bncs.AuthInfoResp:
		return p, nil
	default:
		return nil, ErrUnexpectedPacket
	}
}

func (b *Client) sendAuthCheck(conn *network.BNCSonn, clientToken uint32, authinfo *bncs.AuthInfoResp) (*bncs.AuthCheckResp, error) {
	exePath := path.Join(b.BinPath, "Warcraft III.exe")
	if _, err := os.Stat(exePath); err != nil {
		return nil, err
	}

	exeVersion, exeInfo, err := GetExeInfo(exePath)
	if err != nil {
		return nil, err
	}

	var files = []string{exePath}
	if b.AuthInfo.GameVersion.Version < 29 {
		stormPath := path.Join(b.BinPath, "Storm.dll")
		if _, err := os.Stat(stormPath); err != nil {
			return nil, err
		}
		gamePath := path.Join(b.BinPath, "game.dll")
		if _, err := os.Stat(gamePath); err != nil {
			return nil, err
		}
		files = append(files, stormPath, gamePath)
	}

	exeHash, err := CheckRevision(authinfo.ValueString, files, ExtractMPQNumber(authinfo.MpqFileName))
	if err != nil {
		return nil, err
	}

	var cdkeys = make([]bncs.CDKey, len(b.CDKeys))
	for i := 0; i < len(b.CDKeys); i++ {
		info, err := CreateBNCSKeyInfo(b.CDKeys[i], clientToken, authinfo.ServerToken)
		if err != nil {
			return nil, err
		}

		cdkeys[i] = *info
	}

	var req = &bncs.AuthCheckReq{
		ClientToken:    clientToken,
		ExeVersion:     exeVersion,
		ExeHash:        exeHash,
		CDKeys:         cdkeys,
		ExeInformation: exeInfo,
		KeyOwnerName:   b.CDKeyOwner,
	}

	if _, err := conn.Send(req); err != nil {
		return nil, err
	}

	pkt, err := conn.NextServerPacket(5 * time.Second)
	if err != nil {
		return nil, err
	}
	switch p := pkt.(type) {
	case *bncs.AuthCheckResp:
		return p, nil
	default:
		return nil, ErrUnexpectedPacket
	}
}

func (b *Client) sendLogon(conn *network.BNCSonn, nls *NLS) (*bncs.AuthAccountLogonResp, error) {
	var req = &bncs.AuthAccountLogonReq{
		ClientKey: nls.ClientKey(),
		UserName:  b.UserName,
	}

	if _, err := conn.Send(req); err != nil {
		return nil, err
	}

	pkt, err := conn.NextServerPacket(15 * time.Second)
	if err != nil {
		return nil, err
	}
	switch p := pkt.(type) {
	case *bncs.AuthAccountLogonResp:
		return p, nil
	default:
		return nil, ErrUnexpectedPacket
	}
}

func (b *Client) sendLogonProof(conn *network.BNCSonn, nls *NLS, logon *bncs.AuthAccountLogonResp) (*bncs.AuthAccountLogonProofResp, error) {
	var req = &bncs.AuthAccountLogonProofReq{
		ClientPasswordProof: nls.SessionKey(&logon.ServerKey, &logon.Salt),
	}

	if _, err := conn.Send(req); err != nil {
		return nil, err
	}

	pkt, err := conn.NextServerPacket(5 * time.Second)
	if err != nil {
		return nil, err
	}
	switch p := pkt.(type) {
	case *bncs.AuthAccountLogonProofResp:
		return p, nil
	default:
		return nil, ErrUnexpectedPacket
	}
}

func (b *Client) sendCreateAccount(conn *network.BNCSonn, nls *NLS) (*bncs.AuthAccountCreateResp, error) {

	salt, verifier, err := nls.AccountCreate()
	if err != nil {
		return nil, err
	}

	var req = &bncs.AuthAccountCreateReq{UserName: b.UserName}
	copy(req.Salt[:], salt)
	copy(req.Verifier[:], verifier)

	if _, err := conn.Send(req); err != nil {
		return nil, err
	}

	pkt, err := conn.NextServerPacket(5 * time.Second)
	if err != nil {
		return nil, err
	}
	switch p := pkt.(type) {
	case *bncs.AuthAccountCreateResp:
		return p, nil
	default:
		return nil, ErrUnexpectedPacket
	}
}

func (b *Client) sendEnterChat(conn *network.BNCSonn) (*bncs.EnterChatResp, error) {
	if _, err := conn.Send(&bncs.NetGamePort{Port: b.GamePort}); err != nil {
		return nil, err
	}

	if _, err := conn.Send(&bncs.EnterChatReq{}); err != nil {
		return nil, err
	}

	pkt, err := conn.NextServerPacket(5 * time.Second)

rcv:
	if err != nil {
		return nil, err
	}
	switch p := pkt.(type) {
	case *bncs.ClanInfo:
		b.Fire(pkt)
		pkt, err = conn.NextServerPacket(0)
		goto rcv
	case *bncs.EnterChatResp:
		return p, nil
	default:
		return nil, ErrUnexpectedPacket
	}
}

// Run reads packets and emits an event for each received packet
// Not safe for concurrent invocation
func (b *Client) Run() error {
	if b.KeepAliveInterval != 0 {
		var keepaliveTicker = time.NewTicker(b.KeepAliveInterval)
		defer keepaliveTicker.Stop()

		var pkt bncs.KeepAlive
		go func() {
			for range keepaliveTicker.C {
				if _, err := b.Send(&pkt); err != nil && !network.IsConnClosedError(err) {
					b.Fire(&network.AsyncError{Src: "Run[KeepAlive]", Err: err})
				}
			}
		}()
	}

	return b.BNCSonn.RunClient(&b.EventEmitter, 30*time.Second)
}

// Say sends a chat message
// May block while rate-limiting packets
func (b *Client) Say(s string) error {

	s = strings.Map(func(r rune) rune {
		if !unicode.IsPrint(r) {
			return -1
		}
		return r
	}, s)

	if len(s) == 0 {
		return nil
	}
	if len(s) > 254 {
		s = s[:254]
	}

	if _, err := b.SendRL(&bncs.ChatCommand{Text: s}); err != nil {
		return err
	}

	b.Fire(&Say{Content: s})
	return nil
}

// InitDefaultHandlers adds the default callbacks for relevant packets
func (b *Client) InitDefaultHandlers() {
	b.On(&bncs.Ping{}, b.onPing)
	b.On(&bncs.ChatEvent{}, b.onChatEvent)
}

func (b *Client) onPing(ev *network.Event) {
	var pkt = ev.Arg.(*bncs.Ping)

	if _, err := b.Send(pkt); err != nil && !network.IsConnClosedError(err) {
		b.Fire(&network.AsyncError{Src: "onPing[Send]", Err: err})
	}
}

func (b *Client) onChatEvent(ev *network.Event) {
	var pkt = ev.Arg.(*bncs.ChatEvent)

	switch pkt.Type {
	case bncs.ChatChannelInfo:
		b.chatmut.Lock()
		b.channel = pkt.Text
		b.users = nil
		b.chatmut.Unlock()

		b.Fire(&Channel{Name: pkt.Text, Flags: pkt.ChannelFlags})
	case bncs.ChatShowUser, bncs.ChatJoin:
		var t = time.Now()
		var u = User{
			Name:       pkt.UserName,
			StatString: pkt.Text,
			Flags:      pkt.UserFlags,
			Ping:       pkt.Ping,
			Joined:     t,
			LastSeen:   t,
		}

		b.chatmut.Lock()
		if b.users == nil {
			b.users = make(map[string]*User)
		}
		var p = b.users[strings.ToLower(pkt.UserName)]
		if p != nil {
			u.Joined = p.Joined
			u.LastSeen = p.LastSeen
		}
		b.users[strings.ToLower(pkt.UserName)] = &u
		b.chatmut.Unlock()

		if p == nil {
			b.Fire(&UserJoined{User: u, AlreadyInChannel: pkt.Type == bncs.ChatShowUser})
		} else {
			b.Fire(&UserUpdate{User: u})
		}
	case bncs.ChatUserFlagsUpdate:
		var e UserUpdate

		b.chatmut.Lock()
		var u = b.users[strings.ToLower(pkt.UserName)]
		if u != nil {
			u.Flags = pkt.UserFlags
			e.User = *u
		}
		b.chatmut.Unlock()

		if u != nil {
			b.Fire(&e)
		}
	case bncs.ChatLeave:
		b.chatmut.Lock()
		var u = b.users[strings.ToLower(pkt.UserName)]
		delete(b.users, strings.ToLower(pkt.UserName))
		b.chatmut.Unlock()

		if u != nil {
			b.Fire(&UserLeft{User: *u})
		}
	case bncs.ChatTalk, bncs.ChatEmote:
		var e = Chat{
			Content: pkt.Text,
			Type:    pkt.Type,
		}

		b.chatmut.Lock()
		var u = b.users[strings.ToLower(pkt.UserName)]
		if u != nil {
			u.LastSeen = time.Now()
			e.User = *u
		}
		b.chatmut.Unlock()

		if u != nil {
			b.Fire(&e)
		}
	case bncs.ChatWhisper:
		b.Fire(&Whisper{UserName: pkt.UserName, Content: pkt.Text, Flags: pkt.UserFlags, Ping: pkt.Ping})
	case bncs.ChatChannelFull, bncs.ChatChannelDoesNotExist, bncs.ChatChannelRestricted:
		b.Fire(&JoinError{Channel: pkt.Text, Error: pkt.Type})
	case bncs.ChatBroadcast, bncs.ChatInfo, bncs.ChatError:
		b.Fire(&SystemMessage{Content: pkt.Text, Type: pkt.Type})
	}
}
