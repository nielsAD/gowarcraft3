// Author:  Niels A.D.
// Project: gowarcraft3 (https://github.com/nielsAD/gowarcraft3)
// License: Mozilla Public License, v2.0

package bncs

import (
	"errors"
	"fmt"
)

// Errors
var (
	ErrNoProtocolSig     = errors.New("bncs: Invalid bncs packet (no signature found)")
	ErrInvalidPacketSize = errors.New("bncs: Invalid packet size")
	ErrInvalidChecksum   = errors.New("bncs: Checksum invalid")
	ErrUnexpectedConst   = errors.New("bncs: Unexpected constant value")
)

// ProtocolSig is the BNCS magic number used in the packet header.
const ProtocolSig = 0xFF

// ProtocolGreeting is the BNCS magic number first sent by the client when initiating a connection.
const ProtocolGreeting = 0x01

// BNCS packet type identifiers
const (
	PidNull                   = 0x00 // C -> S | S -> C
	PidStopAdv                = 0x02 // C -> S |
	PidGetAdvListEx           = 0x09 // C -> S | S -> C
	PidEnterChat              = 0x0A // C -> S | S -> C
	PidJoinChannel            = 0x0C // C -> S |
	PidChatCommand            = 0x0E // C -> S |
	PidChatEvent              = 0x0F //        | S -> C
	PidFloodDetected          = 0x13 //        | S -> C
	PidMessageBox             = 0x19 //        | S -> C
	PidStartAdvex3            = 0x1C // C -> S | S -> C
	PidNotifyJoin             = 0x22 // C -> S |
	PidPing                   = 0x25 // C -> S | S -> C
	PidNetGamePort            = 0x45 // C -> S |
	PidAuthInfo               = 0x50 // C -> S | S -> C
	PidAuthCheck              = 0x51 // C -> S | S -> C
	PidAuthAccountCreate      = 0x52 // C -> S | S -> C
	PidAuthAccountLogon       = 0x53 // C -> S | S -> C
	PidAuthAccountLogonProof  = 0x54 // C -> S | S -> C
	PidAuthAccountChange      = 0x55 // C -> S | S -> C
	PidAuthAccountChangeProof = 0x56 // C -> S | S -> C
	PidSetEmail               = 0x59 // C -> S |
	PidClanInfo               = 0x75 //        | S -> C
)

// JoinChannelFlag enum
type JoinChannelFlag uint32

// JoinChannel options
const (
	ChannelJoin         JoinChannelFlag = 0x00
	ChannelJoinFirst    JoinChannelFlag = 0x01
	ChannelJoinOrCreate JoinChannelFlag = 0x02
)

func (f JoinChannelFlag) String() string {
	switch f {
	case ChannelJoin:
		return "Join"
	case ChannelJoinFirst:
		return "JoinFirst"
	case ChannelJoinOrCreate:
		return "JoinOrCreate"
	default:
		return fmt.Sprintf("JoinChannelFlag(0x%02X)", uint32(f))
	}
}

// ChatEventType enum
type ChatEventType uint32

// Chat event types
const (
	ChatShowUser            ChatEventType = 0x01 // User in channel
	ChatJoin                ChatEventType = 0x02 // User joined channel
	ChatLeave               ChatEventType = 0x03 // User left channel
	ChatWhisper             ChatEventType = 0x04 // Recieved whisper
	ChatTalk                ChatEventType = 0x05 // Chat text
	ChatBroadcast           ChatEventType = 0x06 // Server broadcast
	ChatChannelInfo         ChatEventType = 0x07 // Channel information
	ChatUserFlagsUpdate     ChatEventType = 0x09 // Flags update
	ChatWhisperSent         ChatEventType = 0x0A // Sent whisper
	ChatChannelFull         ChatEventType = 0x0D // Channel full
	ChatChannelDoesNotExist ChatEventType = 0x0E // Channel doesn't exist
	ChatChannelRestricted   ChatEventType = 0x0F // Channel is restricted
	ChatInfo                ChatEventType = 0x12 // Information
	ChatError               ChatEventType = 0x13 // Error message
	ChatIgnore              ChatEventType = 0x15 // Notifies that a user has been ignored (DEFUNCT)
	ChatUnignore            ChatEventType = 0x16 // Notifies that a user has been unignored (DEFUNCT)
	ChatEmote               ChatEventType = 0x17 // Emote
)

func (e ChatEventType) String() string {
	switch e {
	case ChatShowUser:
		return "ShowUser"
	case ChatJoin:
		return "Join"
	case ChatLeave:
		return "Leave"
	case ChatWhisper:
		return "Whisper"
	case ChatTalk:
		return "Chat"
	case ChatBroadcast:
		return "Broadcast"
	case ChatChannelInfo:
		return "ChannelInfo"
	case ChatUserFlagsUpdate:
		return "UserFlagsUpdate"
	case ChatWhisperSent:
		return "WhisperSent"
	case ChatChannelFull:
		return "ChannelFull"
	case ChatChannelDoesNotExist:
		return "ChannelDoesNotExist"
	case ChatChannelRestricted:
		return "ChannelRestricted"
	case ChatInfo:
		return "Info"
	case ChatError:
		return "Error"
	case ChatIgnore:
		return "Ignore"
	case ChatUnignore:
		return "Unignore"
	case ChatEmote:
		return "Emote"
	default:
		return fmt.Sprintf("ChatEventType(0x%02X)", uint32(e))
	}
}

// ChatUserFlags enum
type ChatUserFlags uint32

// ChatUser Flags
const (
	ChatUserFlagBlizzard  ChatUserFlags = 0x01 // Blizzard Representative
	ChatUserFlagOperator  ChatUserFlags = 0x02 // Channel Operator
	ChatUserFlagSpeaker   ChatUserFlags = 0x04 // Channel Speaker
	ChatUserFlagAdmin     ChatUserFlags = 0x08 // Battle.net Administrator
	ChatUserFlagSquelched ChatUserFlags = 0x20 // Squelched
)

func (f ChatUserFlags) String() string {
	var res string
	if f&ChatUserFlagBlizzard != 0 {
		res += "|Blizzard"
		f &= ^ChatUserFlagBlizzard
	}
	if f&ChatUserFlagOperator != 0 {
		res += "|Operator"
		f &= ^ChatUserFlagOperator
	}
	if f&ChatUserFlagSpeaker != 0 {
		res += "|Speaker"
		f &= ^ChatUserFlagSpeaker
	}
	if f&ChatUserFlagAdmin != 0 {
		res += "|Admin"
		f &= ^ChatUserFlagAdmin
	}
	if f&ChatUserFlagSquelched != 0 {
		res += "|Squelched"
		f &= ^ChatUserFlagSquelched
	}
	if f != 0 {
		res += fmt.Sprintf("|ChatUserFlags(0x%02X)", uint32(f))
	}
	if res != "" {
		res = res[1:]
	}
	return res
}

// ChatChannelFlags enum
type ChatChannelFlags uint32

// ChatChannel Flags
const (
	ChatChannelFlagPublic      ChatChannelFlags = 0x00001 // Public Channel
	ChatChannelFlagModerated   ChatChannelFlags = 0x00002 // Moderated
	ChatChannelFlagRestricted  ChatChannelFlags = 0x00004 // Restricted
	ChatChannelFlagSilent      ChatChannelFlags = 0x00008 // Silent
	ChatChannelFlagSystem      ChatChannelFlags = 0x00010 // System
	ChatChannelFlagProduct     ChatChannelFlags = 0x00020 // Product-Specific
	ChatChannelFlagGlobal      ChatChannelFlags = 0x01000 // Globally Accessible
	ChatChannelFlagRedirected  ChatChannelFlags = 0x04000 // Redirected
	ChatChannelFlagChat        ChatChannelFlags = 0x08000 // Chat
	ChatChannelFlagTechSupport ChatChannelFlags = 0x10000 // Tech Support
)

func (f ChatChannelFlags) String() string {
	var res string
	if f&ChatChannelFlagPublic != 0 {
		res += "|Public"
		f &= ^ChatChannelFlagPublic
	}
	if f&ChatChannelFlagModerated != 0 {
		res += "|Moderated"
		f &= ^ChatChannelFlagModerated
	}
	if f&ChatChannelFlagRestricted != 0 {
		res += "|Restricted"
		f &= ^ChatChannelFlagPublic
	}
	if f&ChatChannelFlagSilent != 0 {
		res += "|Silent"
		f &= ^ChatChannelFlagSilent
	}
	if f&ChatChannelFlagSystem != 0 {
		res += "|System"
		f &= ^ChatChannelFlagSystem
	}
	if f&ChatChannelFlagProduct != 0 {
		res += "|Product"
		f &= ^ChatChannelFlagProduct
	}
	if f&ChatChannelFlagGlobal != 0 {
		res += "|Global"
		f &= ^ChatChannelFlagGlobal
	}
	if f&ChatChannelFlagRedirected != 0 {
		res += "|Redirected"
		f &= ^ChatChannelFlagRedirected
	}
	if f&ChatChannelFlagChat != 0 {
		res += "|Chat"
		f &= ^ChatChannelFlagChat
	}
	if f&ChatChannelFlagTechSupport != 0 {
		res += "|TechSupport"
		f &= ^ChatChannelFlagTechSupport
	}
	if f != 0 {
		res += fmt.Sprintf("|ChatChannelFlags(0x%02X)", uint32(f))
	}
	if res != "" {
		res = res[1:]
	}
	return res
}

// AdvListResult enum
type AdvListResult uint32

// Game status
const (
	AdvListSuccess           AdvListResult = 0x00 // OK
	AdvListNotFound          AdvListResult = 0x01 // Game doesn't exist
	AdvListIncorrectPassword AdvListResult = 0x02 // Incorrect password
	AdvListFull              AdvListResult = 0x03 // Game full
	AdvListStarted           AdvListResult = 0x04 // Game already started
	AdvListCDKeyNotAllowed   AdvListResult = 0x05 // Spawned CD-Key not allowed
	AdvListRequestRate       AdvListResult = 0x06 // Too many server requests
)

func (s AdvListResult) String() string {
	switch s {
	case AdvListSuccess:
		return "Success"
	case AdvListNotFound:
		return "No game found"
	case AdvListIncorrectPassword:
		return "Incorrect password"
	case AdvListFull:
		return "Game full"
	case AdvListStarted:
		return "Game already started"
	case AdvListCDKeyNotAllowed:
		return "Spawned CD-Key not allowed"
	case AdvListRequestRate:
		return "Too many server requests"
	default:
		return fmt.Sprintf("AdvListResult(0x%02X)", uint32(s))
	}
}

// GameStateFlags enum
type GameStateFlags uint32

// GameState Flags
const (
	GameStateFlagPrivate    GameStateFlags = 0x01 // Game is private
	GameStateFlagFull       GameStateFlags = 0x02 // Game is full
	GameStateFlagHasPlayers GameStateFlags = 0x04 // Game contains players (other than creator)
	GameStateFlagInProgress GameStateFlags = 0x08 // Game is in progress
	GameStateFlagOpen       GameStateFlags = 0x10 // Game is open for players (?)
	GameStateFlagReplay     GameStateFlags = 0x80 // Game is a replay
)

func (f GameStateFlags) String() string {
	var res string
	if f&GameStateFlagOpen != 0 {
		res += "|Open"
		f &= ^GameStateFlagOpen
	}
	if f&GameStateFlagPrivate != 0 {
		res += "|Private"
		f &= ^GameStateFlagPrivate
	}
	if f&GameStateFlagFull != 0 {
		res += "|Full"
		f &= ^GameStateFlagFull
	}
	if f&GameStateFlagHasPlayers != 0 {
		res += "|HasPlayers"
		f &= ^GameStateFlagHasPlayers
	}
	if f&GameStateFlagInProgress != 0 {
		res += "|InProgress"
		f &= ^GameStateFlagInProgress
	}
	if f&GameStateFlagReplay != 0 {
		res += "|Replay"
		f &= ^GameStateFlagReplay
	}
	if f&GameStateFlagInProgress != 0 {
		res += "|InProgress"
		f &= ^GameStateFlagInProgress
	}
	if f&GameStateFlagReplay != 0 {
		res += "|Replay"
		f &= ^GameStateFlagReplay
	}
	if f != 0 {
		res += fmt.Sprintf("|GameStateFlags(0x%02X)", uint32(f))
	}
	if res != "" {
		res = res[1:]
	}
	return res
}

// AuthResult enum
type AuthResult uint32

// AuthCheck result
const (
	AuthSuccess            AuthResult = 0x000 // Passed challenge
	AuthUpgradeRequired    AuthResult = 0x100 // Old game version (Additional info field supplies patch MPQ filename)
	AuthInvalidVersion     AuthResult = 0x101 // Invalid version
	AuthInvalidVersionMask AuthResult = 0x0FF // Invalid version (error code correlates to exact version)
	AuthDowngradeRequired  AuthResult = 0x102 // Game version must be downgraded (Additional info field supplies patch MPQ filename)
	AuthCDKeyInvalid       AuthResult = 0x200 // Invalid CD key (If you receive this status, official Battle.net servers will IP ban you for 1 to 14 days. Before June 15, 2011, this used to exclusively be 14 days)
	AuthCDKeyInUse         AuthResult = 0x201 // CD key in use (Additional info field supplies name of user)
	AuthCDKeyBanned        AuthResult = 0x202 // Banned key
	AuthWrongProduct       AuthResult = 0x203 // Wrong product
)

func (r AuthResult) String() string {
	var invalidVersion = r & AuthInvalidVersionMask
	switch r {
	case AuthSuccess:
		return "Success"
	case AuthUpgradeRequired:
		return "Upgrade required"
	case AuthInvalidVersion, invalidVersion:
		return "Invalid version"
	case AuthDowngradeRequired:
		return "Downgrade required"
	case AuthCDKeyInvalid:
		return "CD key invalid"
	case AuthCDKeyInUse:
		return "CD key in use"
	case AuthCDKeyBanned:
		return "CD key banned"
	case AuthWrongProduct:
		return "Wrong product"
	default:
		return fmt.Sprintf("AuthResult(0x%03X)", uint32(r))
	}
}

// AccountCreateResult enum
type AccountCreateResult uint32

// AuthAccountCreate result
const (
	AccountCreateSuccess        AccountCreateResult = 0x00 // Successfully created account name.
	AccountCreateNameExists     AccountCreateResult = 0x04 // Name already exists.
	AccountCreateNameTooShort   AccountCreateResult = 0x07 // Name is too short/blank.
	AccountCreateIllegalChar    AccountCreateResult = 0x08 // Name contains an illegal character.
	AccountCreateBlacklist      AccountCreateResult = 0x09 // Name contains an illegal word.
	AccountCreateTooFewAlphaNum AccountCreateResult = 0x0A // Name contains too few alphanumeric characters.
	AccountCreateAdjacentPunct  AccountCreateResult = 0x0B // Name contains adjacent punctuation characters.
	AccountCreateTooManyPunct   AccountCreateResult = 0x0C // Name contains too many punctuation characters.
)

func (r AccountCreateResult) String() string {
	switch r {
	case AccountCreateSuccess:
		return "Successfully created account name"
	case AccountCreateNameExists:
		return "Name already exist."
	case AccountCreateNameTooShort:
		return "Name is too short/blank"
	case AccountCreateIllegalChar:
		return "Name contains an illegal character"
	case AccountCreateBlacklist:
		return "Name contains an illegal word"
	case AccountCreateTooFewAlphaNum:
		return "Name contains too few alphanumeric characters"
	case AccountCreateAdjacentPunct:
		return "Name contains adjacent punctuation characters"
	case AccountCreateTooManyPunct:
		return "Name contains too many punctuation characters"
	default:
		return fmt.Sprintf("AccountCreateResult(0x%02X)", uint32(r))
	}
}

// LogonResult enum
type LogonResult uint32

// AuthAccountLogon result
const (
	LogonSuccess         LogonResult = 0x00 // Logon accepted, requires proof.
	LogonInvalidAccount  LogonResult = 0x01 // Account doesn't exist.
	LogonUpgradeRequired LogonResult = 0x05 // Account requires upgrade.
)

func (r LogonResult) String() string {
	switch r {
	case LogonSuccess:
		return "Success"
	case LogonInvalidAccount:
		return "Account does not exist"
	case LogonUpgradeRequired:
		return "Account requires upgrade"
	default:
		return fmt.Sprintf("LogonResult(0x%02X)", uint32(r))
	}
}

// LogonProofResult enum
type LogonProofResult uint32

// AuthAccountLogonProof result
const (
	LogonProofSuccess           LogonProofResult = 0x00 // Logon successful.
	LogonProofPasswordIncorrect LogonProofResult = 0x02 // Incorrect password.
	LogonProofAccountClosed     LogonProofResult = 0x06 // Account closed.
	LogonProofRequireEmail      LogonProofResult = 0x0E // An email address should be registered for this account.
	LogonProofCustomError       LogonProofResult = 0x0F // Custom error. A string at the end of this message contains the error.
)

func (r LogonProofResult) String() string {
	switch r {
	case LogonProofSuccess:
		return "Success"
	case LogonProofPasswordIncorrect:
		return "Password incorrect"
	case LogonProofAccountClosed:
		return "Account closed"
	case LogonProofRequireEmail:
		return "An email address should be registered for this account"
	case LogonProofCustomError:
		return "Custom error"
	default:
		return fmt.Sprintf("LogonProofResult(0x%02X)", uint32(r))
	}
}

// ClanRank enum
type ClanRank uint8

// Clan rank
const (
	ClanRankNew      ClanRank = 0x00 // Initiate (Peon icon), in clan less than one week
	ClanRankInitiate ClanRank = 0x01 // Initiate (Peon icon)
	ClanRankMember   ClanRank = 0x02 // Member (Grunt icon)
	ClanRankOfficer  ClanRank = 0x03 // Officer (Shaman icon)
	ClanRankLeader   ClanRank = 0x04 // Leader (Chieftain icon)
)

func (r ClanRank) String() string {
	switch r {
	case ClanRankNew:
		return "Initiate"
	case ClanRankInitiate:
		return "Initiate"
	case ClanRankMember:
		return "Member"
	case ClanRankOfficer:
		return "Officer"
	case ClanRankLeader:
		return "Leader"
	default:
		return fmt.Sprintf("ClanRank(0x%02X)", uint8(r))
	}
}
