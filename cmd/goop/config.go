// Author:  Niels A.D.
// Project: gowarcraft3 (https://github.com/nielsAD/gowarcraft3)
// License: Mozilla Public License, v2.0

package main

import (
	"encoding"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/imdario/mergo"

	"github.com/nielsAD/gowarcraft3/network/bnet"
)

// DefaultConfig values used as fallback
var DefaultConfig = Config{
	StdIO: StdIOConfig{
		Read:           true,
		Rank:           RankOwner,
		CommandTrigger: "/",
	},
	BNet: BNetConfigWithDefault{
		Default: BNetConfig{
			BNetRealmConfig: BNetRealmConfig{
				ReconnectDelay: 30 * time.Second,
				CommandTrigger: "!",
			},
			Config: bnet.Config{
				BinPath: bnet.DefaultConfig.BinPath,
			},
		},
	},
	Discord: DiscordConfigWithDefault{
		Default: DefaultDiscordConfig{
			DiscordConfig: DiscordConfig{
				Presence:      "Battle.net",
				RankNoChannel: RankIgnore,
			},
			DiscordChannelConfig: DiscordChannelConfig{
				CommandTrigger: "!",
				RankMentions:   RankWhitelist,
			},
		},
	},
}

// Config struct maps the layout of main configuration file
type Config struct {
	StdIO   StdIOConfig
	BNet    BNetConfigWithDefault
	Discord DiscordConfigWithDefault
	Relay   []Relay
}

// StdIOConfig struct maps the layout of StdIO configuration section
type StdIOConfig struct {
	Read           bool
	Rank           Rank
	CommandTrigger string
	AvatarURL      string
}

// BNetConfigWithDefault struct maps the layout of the BNet configuration section
type BNetConfigWithDefault struct {
	Default BNetConfig
	Realms  map[string]*BNetConfig
}

// BNetConfig stores the configuration of a single BNet server
type BNetConfig struct {
	BNetRealmConfig
	bnet.Config
}

// BNetRealmConfig stores the config additions of goop.BNetRealm next to bnet.Client
type BNetRealmConfig struct {
	ReconnectDelay time.Duration
	HomeChannel    string
	CommandTrigger string
	AvatarURL      string

	RankWhisper    Rank
	RankTalk       Rank
	RankNoWarcraft Rank
	RankOperator   *Rank
	RankLevel      map[int]Rank
	RankClanTag    map[string]Rank
	RankUser       map[string]Rank
}

// DiscordConfigWithDefault struct maps the layout of the Discord configuration section
type DiscordConfigWithDefault struct {
	Default  DefaultDiscordConfig
	Sessions map[string]*DiscordConfig
}

// DefaultDiscordConfig struct maps the layout of the Discord.Default configuration section
type DefaultDiscordConfig struct {
	DiscordConfig
	DiscordChannelConfig
}

// DiscordConfig stores the configuration of a Discord session
type DiscordConfig struct {
	AuthToken     string
	Channels      map[string]*DiscordChannelConfig
	Presence      string
	RankDM        Rank
	RankNoChannel Rank
}

// DiscordChannelConfig stores the configuration of a single Discord channel
type DiscordChannelConfig struct {
	CommandTrigger string
	Webhook        string
	RankMentions   Rank
	RankTalk       Rank
	RankRole       map[string]Rank
	RankUser       map[string]Rank
}

// Relay struct maps the layout of Relay configuration section
type Relay struct {
	In  []string
	Out []string

	Log         bool
	System      bool
	Joins       bool
	Chat        bool
	PrivateChat bool

	JoinRank        Rank
	ChatRank        Rank
	PrivateChatRank Rank
}

func deleteDefaults(def map[string]interface{}, dst map[string]interface{}) {
	for k := range def {
		if reflect.DeepEqual(def[k], dst[k]) {
			delete(dst, k)
			continue
		}
		var v, ok = dst[k].(map[string]interface{})
		if ok {
			deleteDefaults(def[k].(map[string]interface{}), v)
		}
	}
}

func imap(val interface{}) interface{} {
	var v = reflect.ValueOf(val)
	switch v.Kind() {
	case reflect.Ptr:
		if v.IsNil() {
			return nil
		}
		return imap(v.Elem().Interface())
	case reflect.Map:
		var m = make(map[string]interface{})
		for _, key := range v.MapKeys() {
			m[fmt.Sprintf("%v", key.Interface())] = imap(v.MapIndex(key).Interface())
		}
		return m
	case reflect.Slice, reflect.Array:
		var r = make([]interface{}, v.Len())
		for i := 0; i < v.Len(); i++ {
			r[i] = imap(v.Index(i).Interface())
		}
		return r
	case reflect.Struct:
		var m = make(map[string]interface{})
		for i := 0; i < v.NumField(); i++ {
			var f = v.Type().Field(i)
			if f.Name == "" {
				continue
			}

			var x = imap(v.Field(i).Interface())
			if xx, ok := x.(map[string]interface{}); f.Anonymous && ok {
				for k, v := range xx {
					m[k] = v
				}
			} else {
				m[f.Name] = x
			}
		}
		return m
	default:
		return v.Interface()
	}
}

func flatten(prf string, val reflect.Value, dst map[string]reflect.Value) {
	switch val.Kind() {
	case reflect.Ptr:
		if val.IsNil() {
			dst[strings.ToLower(prf)] = val
		} else {
			flatten(prf, val.Elem(), dst)
		}
	case reflect.Map:
		for _, key := range val.MapKeys() {
			var pre string
			if prf == "" {
				pre = fmt.Sprintf("%v", key.Interface())
			} else {
				pre = fmt.Sprintf("%s.%v", prf, key.Interface())
			}
			flatten(pre, val.MapIndex(key), dst)
		}
	case reflect.Slice, reflect.Array:
		for i := 0; i < val.Len(); i++ {
			var pre string
			if prf == "" {
				pre = fmt.Sprintf("%d", i)
			} else {
				pre = fmt.Sprintf("%s.%d", prf, i)
			}
			flatten(pre, val.Index(i), dst)
		}
	case reflect.Struct:
		for i := 0; i < val.NumField(); i++ {
			var f = val.Type().Field(i)
			if f.Name == "" {
				continue
			}

			var pre = f.Name
			if f.Anonymous {
				pre = prf
			} else if prf != "" {
				pre = fmt.Sprintf("%s.%v", prf, f.Name)
			}
			flatten(pre, val.Field(i), dst)
		}
	default:
		dst[strings.ToLower(prf)] = val
	}
}

// MergeDefaults applies default configuration for unset fields
func (c *Config) MergeDefaults() error {
	for _, r := range c.BNet.Realms {
		if err := mergo.Merge(r, c.BNet.Default); err != nil {
			return err
		}
	}

	for _, s := range c.Discord.Sessions {
		if err := mergo.Merge(s, c.Discord.Default.DiscordConfig); err != nil {
			return err
		}
		for _, n := range s.Channels {
			if err := mergo.Merge(n, c.Discord.Default.DiscordChannelConfig); err != nil {
				return err
			}
		}
	}

	return nil
}

// type alias for easy type casts
type mi = map[string]interface{}

// Map converts Config to a map[string]interface{} representation
func (c *Config) Map() map[string]interface{} {
	var m = imap(c).(mi)

	var bn = m["BNet"].(mi)["Default"].(mi)
	for _, r := range m["BNet"].(mi)["Realms"].(mi) {
		deleteDefaults(bn, r.(mi))
	}

	var dc = m["Discord"].(mi)["Default"].(mi)
	for _, a := range m["Discord"].(mi)["Sessions"].(mi) {
		for _, c := range a.(mi)["Channels"].(mi) {
			deleteDefaults(dc, c.(mi))
		}
		deleteDefaults(dc, a.(mi))
	}

	return m
}

// Flat list all the (nested) config keys
func (c *Config) Flat() map[string]reflect.Value {
	var f = make(map[string]reflect.Value)
	flatten("", reflect.ValueOf(&c), f)
	return f
}

// Get config value via flat index string
func (c *Config) Get(key string) (interface{}, error) {
	var f, ok = c.Flat()[strings.ToLower(key)]
	if !ok {
		return nil, ErrUnknownConfigKey
	}
	return f.Interface(), nil
}

// Set config value via flat index string
func (c *Config) Set(key string, val interface{}) error {
	var dst, ok = c.Flat()[strings.ToLower(key)]
	if !ok || !dst.CanSet() {
		return ErrUnknownConfigKey
	}

	var src = reflect.ValueOf(val)
	if !src.Type().AssignableTo(dst.Type()) {
		if !src.Type().ConvertibleTo(dst.Type()) {
			return ErrInvalidType
		}
		src = src.Convert(dst.Type())
	}

	dst.Set(src)
	return nil
}

// GetString config value via flat index string
func (c *Config) GetString(key string) (string, error) {
	val, err := c.Get(key)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%v", val), nil
}

// SetString config value via flat index string
func (c *Config) SetString(key string, val string) error {
	var dst, ok = c.Flat()[strings.ToLower(key)]
	if !ok || !dst.CanSet() {
		return ErrUnknownConfigKey
	}

	if i, ok := dst.Interface().(encoding.TextUnmarshaler); ok {
		return i.UnmarshalText([]byte(val))
	}

	switch dst.Kind() {
	case reflect.String:
		dst.SetString(val)
		return nil
	case reflect.Bool:
		b, err := strconv.ParseBool(val)
		if err != nil {
			return err
		}
		dst.SetBool(b)
		return nil
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		n, err := strconv.ParseInt(val, 10, 64)
		if err != nil {
			return err
		}
		dst.SetInt(n)
		return nil
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		n, err := strconv.ParseUint(val, 10, 64)
		if err != nil {
			return err
		}
		dst.SetUint(n)
		return nil
	default:
		return ErrInvalidType
	}
}
