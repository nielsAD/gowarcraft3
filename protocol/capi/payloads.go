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

// Connect payload (Botapichat.ConnectRequest)
//
// Connect the bot to the gateway and chat channel
type Connect struct{}

// Disconnect payload (Botapichat.DisconnectRequest)
//
// Disconnects the bot from the gateway and chat channel
type Disconnect struct{}

// SendMessage payload (Botapichat.SendMessageRequest)
//
// Sends a chat message to the channel
type SendMessage struct {
	Message string `json:"message"`
}

// SendEmote payload (Botapichat.SendEmoteRequest)
//
// Sends an emote on behalf of a bot
type SendEmote struct {
	Message string `json:"message"`
}

// SendWhisper payload (Botapichat.SendWhisperRequest)
//
// Sends a chat message to one user in the channel
type SendWhisper struct {
	Message string `json:"message"`
	UserID  int64  `json:"user_id"`
}

// KickUser payload (Botapichat.KickUserRequest)
//
// Kicks a user from the channel
type KickUser struct {
	UserID int64 `json:"user_id"`
}

// BanUser payload (Botapichat.BanUserRequest)
//
// Bans a user from the channel
type BanUser struct {
	UserID int64 `json:"user_id"`
}

// UnbanUser payload (Botapichat.UnbanUserRequest)
//
// Un-Bans a user from the channel
type UnbanUser struct {
	Username string `json:"toon_name"`
}

// SetModerator payload (Botapichat.SendSetModeratorRequest)
//
// Sets the current chat moderator to a member of the current chat.
// Same as a normal user doing /designate followed by /resign.
type SetModerator struct {
	UserID int64 `json:"user_id"`
}

// ConnectEvent payload (Botapichat.ConnectEventRequest)
type ConnectEvent struct {
	Channel string `json:"channel"`
}

// DisconnectEvent payload (Botapichat.DisconnectEventRequest)
type DisconnectEvent struct{}

// MessageEvent payload (Botapichat.MessageEventRequest)
type MessageEvent struct {
	UserID  int64            `json:"user_id"`
	Message string           `json:"message"`
	Type    MessageEventType `json:"type"`
}

// UserUpdateEvent payload (Botapichat.UserUpdateEvent)
type UserUpdateEvent struct {
	UserID     int64           `json:"user_id"`
	Username   string          `json:"toon_name,omitempty"`
	Flags      []string        `json:"flag,omitempty"`
	Attributes []UserAttribute `json:"attribute,omitempty"`
}

// UserAttribute for UserUpdateEvent
type UserAttribute struct {
	Key   string
	Value string
}

// UserLeaveEvent payload (Botapichat.UserLeaveEventRequest)
type UserLeaveEvent struct {
	UserID int64 `json:"user_id"`
}
