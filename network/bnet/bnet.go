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

	"github.com/imdario/mergo"
	"github.com/kyokomi/emoji"
	"github.com/nielsAD/gowarcraft3/network"
	"github.com/nielsAD/gowarcraft3/protocol"
	"github.com/nielsAD/gowarcraft3/protocol/bncs"
	"github.com/nielsAD/gowarcraft3/protocol/w3gs"
)

// Config for bnet.Client
type Config struct {
	Platform          bncs.AuthInfoReq
	ServerAddr        string
	KeepAliveInterval time.Duration
	BinPath           string
	Username          string
	Password          string
	CDKeyOwner        string
	CDKeys            []string
	GamePort          uint16
}

// Client represents a mocked BNCS client
// Public methods/fields are thread-safe unless explicitly stated otherwise
type Client struct {
	network.EventEmitter
	network.BNCSConn

	chatmut sync.Mutex
	channel string
	users   map[string]*User

	// Read-only
	UniqueName string

	// Set once before Dial(), read-only after that
	Config
}

// DefaultConfig for bnet.Client
var DefaultConfig = Config{
	Platform: bncs.AuthInfoReq{
		PlatformCode:        protocol.DString("IX86"),
		GameVersion:         w3gs.GameVersion{Product: w3gs.ProductROC, Version: w3gs.CurrentGameVersion},
		LanguageCode:        protocol.DString("enUS"),
		TimeZoneBias:        4294967176,
		MpqLocaleID:         1033,
		UserLanguageID:      1033,
		CountryAbbreviation: "USA",
		Country:             "United States",
	},
	KeepAliveInterval: 30 * time.Second,
	CDKeyOwner:        "gowarcraft3",
	GamePort:          6112,

	BinPath: func() string {
		var paths = []string{
			"./war3",
			"C:/Program Files/Warcraft III/",
			"C:/Program Files (x86)/Warcraft III/",
			path.Join(os.Getenv("HOME"), ".wine/drive_c/Program Files/Warcraft III/"),
			path.Join(os.Getenv("HOME"), ".wine/drive_c/Program Files (x86)/Warcraft III/"),
		}

		for i := 0; i < len(paths); i++ {
			if _, err := os.Stat(paths[i]); err == nil {
				return paths[i]
			}
		}

		return "."
	}(),
}

// NewClient initializes a Client struct
func NewClient(conf *Config) (*Client, error) {
	var c = Client{
		Config: *conf,
	}

	c.InitDefaultHandlers()

	if err := mergo.Merge(&c.Config, DefaultConfig); err != nil {
		return nil, err
	}

	if conf.Platform.GameVersion.Version == 0 {
		if exeVersion, _, err := GetExeInfo(path.Join(c.BinPath, "Warcraft III.exe")); err == nil {
			c.Platform.GameVersion.Version = (exeVersion >> 16) & 0xFF
		}
		if exeVersion, _, err := GetExeInfo(path.Join(c.BinPath, "war3.exe")); err == nil {
			c.Platform.GameVersion.Version = (exeVersion >> 16) & 0xFF
		}
	}

	if conf.Username == "" {
		if _, err := os.Stat(path.Join(c.BinPath, "user.w3k")); err == nil {
			if f, err := ioutil.ReadFile(path.Join(c.BinPath, "user.w3k")); err == nil {
				c.Username = strings.TrimSpace(string(f))
			}
		}
	}

	if len(conf.CDKeys) == 0 {
		var rock = ""
		if _, err := os.Stat(path.Join(c.BinPath, "roc.w3k")); err == nil {
			if f, err := ioutil.ReadFile(path.Join(c.BinPath, "roc.w3k")); err == nil {
				rock = strings.TrimSpace(string(f))
			}
		}

		var tftk = ""
		if _, err := os.Stat(path.Join(c.BinPath, "tft.w3k")); err == nil {
			if f, err := ioutil.ReadFile(path.Join(c.BinPath, "tft.w3k")); err == nil {
				tftk = strings.TrimSpace(string(f))
			}
		}

		if rock != "" {
			if tftk != "" {
				c.CDKeys = []string{rock, tftk}
			} else {
				c.CDKeys = []string{rock}
			}
		}
	}

	if conf.Platform.GameVersion.Product == 0 {
		switch len(c.CDKeys) {
		case 1:
			c.Platform.GameVersion.Product = w3gs.ProductROC
		case 2:
			c.Platform.GameVersion.Product = w3gs.ProductTFT
		}
	}

	return &c, nil
}

// Channel currently chatting in
func (b *Client) Channel() string {
	b.chatmut.Lock()
	var res = b.channel
	b.chatmut.Unlock()
	return res
}

// User in channel by name
func (b *Client) User(name string) (*User, bool) {
	b.chatmut.Lock()
	u, ok := b.users[strings.ToLower(name)]
	if ok {
		copy := *u
		u = &copy
	}
	b.chatmut.Unlock()

	return u, ok
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
func (b *Client) Dial() (*network.BNCSConn, error) {
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

	bncsconn := network.NewBNCSConn(conn)

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
	nls, err := NewNLS(b.Username, b.Password)
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
	nls, err := NewNLS(b.Username, b.Password)
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

// ChangePassword of an existing account
//
// ChangePassword sequence:
//  1. Client starts with Dial sequence
//  2. Client waits for user to enter account information and new password:
//    1. C > S [0x55] SID_AUTH_ACCOUNTCHANGE
//    2. S > C [0x55] SID_AUTH_ACCOUNTCHANGE
//    3. C > S [0x56] SID_AUTH_ACCOUNTCHANGEPROOF
//    4. S > C [0x56] SID_AUTH_ACCOUNTCHANGEPROOF
//  3. Client can continue with logon ([0x53] SID_AUTH_ACCOUNTLOGON)
//
func (b *Client) ChangePassword(newPassword string) error {
	oldNLS, err := NewNLS(b.Username, b.Password)
	if err != nil {
		return err
	}

	defer oldNLS.Free()

	newNLS, err := NewNLS(b.Username, newPassword)
	if err != nil {
		return err
	}

	defer newNLS.Free()

	bncsconn, err := b.Dial()
	if err != nil {
		return err
	}

	defer bncsconn.Close()

	resp, err := b.sendChangePass(bncsconn, oldNLS)
	if err != nil {
		return err
	}

	if resp.Result != bncs.LogonSuccess {
		return LogonResultToError(resp.Result)
	}

	proof, err := b.sendChangePassProof(bncsconn, oldNLS, newNLS, resp)
	if err != nil {
		return err
	}

	if proof.Result != bncs.LogonProofSuccess {
		return LogonProofResultToError(proof.Result)
	}

	b.Password = newPassword
	return nil
}

func (b *Client) sendAuthInfo(conn *network.BNCSConn) (*bncs.AuthInfoResp, error) {
	if _, err := conn.Send(&b.Platform); err != nil {
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

func (b *Client) sendAuthCheck(conn *network.BNCSConn, clientToken uint32, authinfo *bncs.AuthInfoResp) (*bncs.AuthCheckResp, error) {
	exePath := path.Join(b.BinPath, "Warcraft III.exe")
	if _, err := os.Stat(exePath); err != nil {
		return nil, err
	}

	exeVersion, exeInfo, err := GetExeInfo(exePath)
	if err != nil {
		return nil, err
	}

	var files = []string{exePath}
	if b.Platform.GameVersion.Version < 29 {
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

func (b *Client) sendLogon(conn *network.BNCSConn, nls *NLS) (*bncs.AuthAccountLogonResp, error) {
	var req = &bncs.AuthAccountLogonReq{
		ClientKey: nls.ClientKey(),
		Username:  b.Username,
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

func (b *Client) sendLogonProof(conn *network.BNCSConn, nls *NLS, logon *bncs.AuthAccountLogonResp) (*bncs.AuthAccountLogonProofResp, error) {
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

func (b *Client) sendCreateAccount(conn *network.BNCSConn, nls *NLS) (*bncs.AuthAccountCreateResp, error) {
	salt, verifier, err := nls.AccountCreate()
	if err != nil {
		return nil, err
	}

	var req = &bncs.AuthAccountCreateReq{Username: b.Username}
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

func (b *Client) sendChangePass(conn *network.BNCSConn, nls *NLS) (*bncs.AuthAccountChangePassResp, error) {
	var req = &bncs.AuthAccountChangePassReq{
		AuthAccountLogonReq: bncs.AuthAccountLogonReq{
			ClientKey: nls.ClientKey(),
			Username:  b.Username,
		},
	}

	if _, err := conn.Send(req); err != nil {
		return nil, err
	}

	pkt, err := conn.NextServerPacket(15 * time.Second)
	if err != nil {
		return nil, err
	}
	switch p := pkt.(type) {
	case *bncs.AuthAccountChangePassResp:
		return p, nil
	default:
		return nil, ErrUnexpectedPacket
	}
}

func (b *Client) sendChangePassProof(conn *network.BNCSConn, oldNLS *NLS, newNLS *NLS, resp *bncs.AuthAccountChangePassResp) (*bncs.AuthAccountChangePassProofResp, error) {
	salt, verifier, err := newNLS.AccountCreate()
	if err != nil {
		return nil, err
	}

	var req = &bncs.AuthAccountChangePassProofReq{
		ClientPasswordProof: oldNLS.SessionKey(&resp.ServerKey, &resp.Salt),
	}
	copy(req.NewSalt[:], salt)
	copy(req.NewVerifier[:], verifier)

	if _, err := conn.Send(req); err != nil {
		return nil, err
	}

	pkt, err := conn.NextServerPacket(5 * time.Second)
	if err != nil {
		return nil, err
	}
	switch p := pkt.(type) {
	case *bncs.AuthAccountChangePassProofResp:
		return p, nil
	default:
		return nil, ErrUnexpectedPacket
	}
}

func (b *Client) sendEnterChat(conn *network.BNCSConn) (*bncs.EnterChatResp, error) {
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

	return b.BNCSConn.RunClient(&b.EventEmitter, 30*time.Second)
}

var emojiToText = func() *strings.Replacer {
	var r []string
	for txt, uc := range emoji.CodeMap() {
		r = append(r, uc, txt)
	}

	return strings.NewReplacer(r...)
}()

// FilterChat makes the chat message suitable for bnet.
// It filters out control characters, replaces emoji with text, and truncates length.
func FilterChat(s string) string {
	s = emojiToText.Replace(s)

	s = strings.Map(func(r rune) rune {
		if !unicode.IsPrint(r) {
			return -1
		}
		return r
	}, s)

	if len(s) > 254 {
		s = s[:254]
	}
	return s
}

// Say sends a chat message
// May block while rate-limiting packets
func (b *Client) Say(s string) error {
	s = FilterChat(s)
	if len(s) == 0 {
		return nil
	}

	if _, err := b.SendRL(&bncs.ChatCommand{Text: s}); err != nil {
		return err
	}

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
			Name:       pkt.Username,
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
		var p = b.users[strings.ToLower(pkt.Username)]
		if p != nil {
			u.Joined = p.Joined
			u.LastSeen = p.LastSeen
		}
		b.users[strings.ToLower(pkt.Username)] = &u
		b.chatmut.Unlock()

		if p == nil {
			b.Fire(&UserJoined{User: u, AlreadyInChannel: pkt.Type == bncs.ChatShowUser})
		} else {
			b.Fire(&UserUpdate{User: u})
		}
	case bncs.ChatUserFlagsUpdate:
		var e UserUpdate

		b.chatmut.Lock()
		var u = b.users[strings.ToLower(pkt.Username)]
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
		var u = b.users[strings.ToLower(pkt.Username)]
		delete(b.users, strings.ToLower(pkt.Username))
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
		var u = b.users[strings.ToLower(pkt.Username)]
		if u != nil {
			u.LastSeen = time.Now()
			e.User = *u
		}
		b.chatmut.Unlock()

		if u != nil {
			b.Fire(&e)
		}
	case bncs.ChatWhisper:
		b.Fire(&Whisper{Username: pkt.Username, Content: pkt.Text, Flags: pkt.UserFlags, Ping: pkt.Ping})
	case bncs.ChatChannelFull, bncs.ChatChannelDoesNotExist, bncs.ChatChannelRestricted:
		b.Fire(&JoinError{Channel: pkt.Text, Error: pkt.Type})
	case bncs.ChatBroadcast, bncs.ChatInfo, bncs.ChatError:
		b.Fire(&SystemMessage{Content: pkt.Text, Type: pkt.Type})
	}
}
