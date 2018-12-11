// Author:  Niels A.D.
// Project: gowarcraft3 (https://github.com/nielsAD/gowarcraft3)
// License: Mozilla Public License, v2.0

package capi

// Response payload (*.*Response)
type Response struct{}

// Authenticate payload (Botapiauth.AuthenticateRequest)
//
// When connection is established, the client will need to send an authentication request with the API key
type Authenticate struct {
	APIKey string `json:"api_key"`
}

// Command identifier
func (p Authenticate) Command() string { return "Botapiauth.Authenticate" }

// Connect payload (Botapichat.ConnectRequest)
//
// Connect the bot to the gateway and chat channel
type Connect struct{}

// Command identifier
func (p Connect) Command() string { return "Botapichat.Connect" }

// Disconnect payload (Botapichat.DisconnectRequest)
//
// Disconnects the bot from the gateway and chat channel
type Disconnect struct{}

// Command identifier
func (p Disconnect) Command() string { return "Botapichat.Disconnect" }

// SendMessage payload (Botapichat.SendMessageRequest)
//
// Sends a chat message to the channel
type SendMessage struct {
	Message string `json:"message"`
}

// Command identifier
func (p SendMessage) Command() string { return "Botapichat.SendMessage" }

// SendEmote payload (Botapichat.SendEmoteRequest)
//
// Sends an emote on behalf of a bot
type SendEmote struct {
	Message string `json:"message"`
}

// Command identifier
func (p SendEmote) Command() string { return "Botapichat.SendEmote" }

// SendWhisper payload (Botapichat.SendWhisperRequest)
//
// Sends a chat message to one user in the channel
type SendWhisper struct {
	Message string `json:"message"`
	UserID  string `json:"user_id"`
}

// Command identifier
func (p SendWhisper) Command() string { return "Botapichat.SendWhisper" }

// KickUser payload (Botapichat.KickUserRequest)
//
// Kicks a user from the channel
type KickUser struct {
	UserID string `json:"user_id"`
}

// Command identifier
func (p KickUser) Command() string { return "Botapichat.KickUser" }

// BanUser payload (Botapichat.BanUserRequest)
//
// Bans a user from the channel
type BanUser struct {
	UserID string `json:"user_id"`
}

// Command identifier
func (p BanUser) Command() string { return "Botapichat.BanUser" }

// UnbanUser payload (Botapichat.UnbanUserRequest)
//
// Un-Bans a user from the channel
type UnbanUser struct {
	Username string `json:"toon_name"`
}

// Command identifier
func (p UnbanUser) Command() string { return "Botapichat.UnbanUser" }

// SetModerator payload (Botapichat.SendSetModeratorRequest)
//
// Sets the current chat moderator to a member of the current chat.
// Same as a normal user doing /designate followed by /resign.
type SetModerator struct {
	UserID string `json:"user_id"`
}

// Command identifier
func (p SetModerator) Command() string { return "Botapichat.SendSetModerator" }

// ConnectEvent payload (Botapichat.ConnectEventRequest)
type ConnectEvent struct {
	Channel string `json:"channel"`
}

// Command identifier
func (p ConnectEvent) Command() string { return "Botapichat.ConnectEvent" }

// DisconnectEvent payload (Botapichat.DisconnectEventRequest)
type DisconnectEvent struct{}

// Command identifier
func (p DisconnectEvent) Command() string { return "Botapichat.DisconnectEvent" }

// MessageEvent payload (Botapichat.MessageEventRequest)
type MessageEvent struct {
	UserID  string `json:"user_id"`
	Message string `json:"message"`
	Type    string `json:"type"`
}

// Command identifier
func (p MessageEvent) Command() string { return "Botapichat.MessageEvent" }

// UserUpdateEvent payload (Botapichat.UserUpdateEvent)
type UserUpdateEvent struct {
	UserID     string         `json:"user_id"`
	UserName   string         `json:"toon_name"`
	Flags      []string       `json:"flags"`
	Attributes UserAttributes `json:"attributes"`
}

// UserAttributes for UserUpdateEvent
type UserAttributes struct {
	ProgramID string `json:"ProgramId"`
	Rate      string `json:"Rate"`
	Rank      string `json:"Rank"`
	Wins      string `json:"Wins"`
}

// Command identifier
func (p UserUpdateEvent) Command() string { return "Botapichat.UserUpdateEvent" }

// UserLeaveEvent payload (Botapichat.UserLeaveEventRequest)
type UserLeaveEvent struct {
	UserID string `json:"user_id"`
}

// Command identifier
func (p UserLeaveEvent) Command() string { return "Botapichat.UserLeaveEvent" }
