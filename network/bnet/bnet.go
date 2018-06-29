// Author:  Niels A.D.
// Project: gowarcraft3 (https://github.com/nielsAD/gowarcraft3)
// License: Mozilla Public License, v2.0

// Package bnet implements a mocked BNCS client that can be used to interact with BNCS servers.
package bnet

import (
	"net"
	"os"
	"path"
	"strings"
	"time"

	"github.com/nielsAD/gowarcraft3/network"
	"github.com/nielsAD/gowarcraft3/protocol/bncs"
)

// Client represents a mocked BNCS client
// Public methods/fields are thread-safe unless explicitly stated otherwise
type Client struct {
	network.EventEmitter
	network.BNCSonn

	// Set once before Dial(), read-only after that
	ServerAddr        net.TCPAddr
	KeepAliveInterval time.Duration
	AuthInfo          bncs.AuthInfoReq
	BinPath           string
	Username          string
	Password          string
	CDKeyOwner        string
	CDKeys            []string
	GamePort          uint16
}

// Dial opens a new connection to server
// Not safe for concurrent invocation
//
// Logon sequence:
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
//   8. Client waits for user to enter account information (standard logon shown, uses NLS):
//     1. C > S [0x53] SID_AUTH_ACCOUNTLOGON
//     2. S > C [0x53] SID_AUTH_ACCOUNTLOGON
//     3. C > S [0x54] SID_AUTH_ACCOUNTLOGONPROOF
//     4. S > C [0x54] SID_AUTH_ACCOUNTLOGONPROOF
//   9. C > S [0x45] SID_NETGAMEPORT (optional)
//  10. C > S [0x0A] SID_ENTERCHAT
//  11. S > C [0x0A] SID_ENTERCHAT
//  12. C > S [0x44] SID_WARCRAFTGENERAL (WID_TOURNAMENT) (optional)
//  13. S > C [0x44] SID_WARCRAFTGENERAL (WID_TOURNAMENT) (optional response)
//  14. C > S [0x46] SID_NEWS_INFO (optional)
//  15. S > C [0x46] SID_NEWS_INFO (optional response)
//  16. Client waits until user wants to Enter Chat.
//  17. C > S [0x0C] SID_JOINCHANNEL (First Join, "W3")
//  18. S > C [0x0F] SID_CHATEVENT
//  19. A sequence of chat events for entering chat follow.
//
func (b *Client) Dial() error {
	nls, err := NewNLS(b.Username, b.Password)
	if err != nil {
		return err
	}

	defer nls.Free()

	conn, err := net.DialTCP("tcp4", nil, &b.ServerAddr)
	if err != nil {
		return err
	}

	conn.SetNoDelay(true)
	conn.SetLinger(3)
	conn.Write([]byte{bncs.ProtocolGreeting})

	bncsconn := network.NewBNCSonn(conn)

	authInfo, err := b.sendAuthInfo(bncsconn)
	if err != nil {
		bncsconn.Close()
		return err
	}

	clientToken := uint32(time.Now().Unix())
	authCheck, err := b.sendAuthCheck(bncsconn, clientToken, authInfo)
	if err != nil {
		bncsconn.Close()
		return err
	}

	if authCheck.Result != bncs.AuthSuccess {
		bncsconn.Close()
		return AuthResultToError(authCheck.Result)
	}

	logon, err := b.sendLogon(bncsconn, nls)
	if err != nil {
		bncsconn.Close()
		return err
	}

	if logon.Result != bncs.LogonSuccess {
		bncsconn.Close()
		return ErrInvalidAccount
	}

	proof, err := b.sendLogonProof(bncsconn, nls, logon)
	if err != nil {
		bncsconn.Close()
		return err
	}

	if proof.Result != bncs.LogonProofSuccess {
		bncsconn.Close()
		return LogonResultToError(proof.Result)
	}

	if _, err := b.sendEnterChat(bncsconn); err != nil {
		bncsconn.Close()
		return err
	}

	if _, err := bncsconn.Send(&bncs.JoinChannel{Flag: bncs.ChannelJoinFirst, Channel: "W3"}); err != nil {
		bncsconn.Close()
		return err
	}

	b.SetConn(conn)
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
		Username:  b.Username,
	}

	if _, err := conn.Send(req); err != nil {
		return nil, err
	}

	pkt, err := conn.NextServerPacket(5 * time.Second)
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

func (b *Client) sendEnterChat(conn *network.BNCSonn) (*bncs.EnterChatResp, error) {
	if _, err := conn.Send(&bncs.NetGamePort{Port: b.GamePort}); err != nil {
		return nil, err
	}

	if _, err := conn.Send(&bncs.EnterChatReq{}); err != nil {
		return nil, err
	}

	pkt, err := conn.NextServerPacket(5 * time.Second)
	if err != nil {
		return nil, err
	}
	switch p := pkt.(type) {
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
		if r >= 32 && r != 127 {
			return r
		}
		return -1
	}, s)

	if len(s) == 0 {
		return nil
	}
	if len(s) > 254 {
		s = s[:254]
	}

	_, err := b.SendRL(&bncs.ChatCommand{Text: s})
	return err
}

// InitDefaultHandlers adds the default callbacks for relevant packets
func (b *Client) InitDefaultHandlers() {
	b.On(&bncs.ChatEvent{}, b.onChatEvent)
}

func (b *Client) onChatEvent(ev *network.Event) {
	// var pkt = ev.Arg.(*bncs.ChatEvent)
}
