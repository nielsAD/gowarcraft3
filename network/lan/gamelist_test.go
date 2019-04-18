// Author:  Niels A.D.
// Project: gowarcraft3 (https://github.com/nielsAD/gowarcraft3)
// License: Mozilla Public License, v2.0

package lan_test

import (
	"sync/atomic"
	"testing"
	"time"

	"github.com/nielsAD/gowarcraft3/network"
	"github.com/nielsAD/gowarcraft3/network/lan"
	"github.com/nielsAD/gowarcraft3/protocol/w3gs"
)

const wait = 50 * time.Millisecond

var gameInfo = w3gs.GameInfo{
	HostCounter: 1,
	EntryKey:    123456789,
	GameName:    "Test Game (gowarcraft3)",
	GameSettings: w3gs.GameSettings{
		GameSettingFlags: w3gs.SettingSpeedFast | w3gs.SettingTerrainDefault | w3gs.SettingObsNone | w3gs.SettingTeamsTogether | w3gs.SettingTeamsFixed,
		MapWidth:         116,
		MapHeight:        84,
		MapXoro:          2599102717,
		MapPath:          "Maps/FrozenThrone/(2)EchoIsles.w3x",
		HostName:         "gowarcraft3",
	},
	SlotsTotal:     2,
	SlotsUsed:      1,
	SlotsAvailable: 2,
	UptimeSec:      0,
	GamePort:       6112,
	GameFlags:      w3gs.GameFlagCustomGame | w3gs.GameFlagSignedMap,
}

func testGameList(t *testing.T, g lan.GameList, gv w3gs.GameVersion) {
	var info = gameInfo
	info.GameVersion = gv

	var updates int32
	g.On(lan.Update{}, func(ev *network.Event) {
		atomic.AddInt32(&updates, 1)
	})
	g.On(&network.AsyncError{}, func(ev *network.Event) {
		t.Fatal(ev.Arg.(*network.AsyncError))
	})

	a, err := lan.NewAdvertiser(&info)
	if err != nil {
		t.Fatal(err)
	}
	a.On(&network.AsyncError{}, func(ev *network.Event) {
		t.Fatal(ev.Arg.(*network.AsyncError))
	})

	// Create one game before game list is listening
	go a.Run()
	time.Sleep(wait)

	go g.Run()
	time.Sleep(wait)

	if atomic.LoadInt32(&updates) != 1 {
		t.Fatal("Update{} not fired after early create")
	}
	if len(g.Games()) != 1 {
		t.Fatal("Game not found after early create")
	}

	if err := a.Close(); err != nil {
		t.Fatal(err)
	}
	time.Sleep(wait)

	if atomic.LoadInt32(&updates) != 2 {
		t.Fatal("Update{} not fired after first decreate")
	}
	if len(g.Games()) != 0 {
		t.Fatal("Game found after first decreate")
	}

	info.GameName += "1"
	info.HostCounter++
	if a, err = lan.NewAdvertiser(&info); err != nil {
		t.Fatal(err)
	}
	a.On(&network.AsyncError{}, func(ev *network.Event) {
		t.Fatal(ev.Arg.(*network.AsyncError))
	})

	go a.Run()
	time.Sleep(wait)

	if atomic.LoadInt32(&updates) != 3 {
		t.Fatal("Update{} not fired after late create")
	}
	if len(g.Games()) != 1 {
		t.Fatal("Game not found after late create")
	}

	a.Refresh(info.SlotsUsed+1, info.SlotsAvailable)
	time.Sleep(wait)

	if atomic.LoadInt32(&updates) != 4 {
		t.Fatal("Update{} not fired after refresh")
	}
	if len(g.Games()) != 1 {
		t.Fatal("Game not found after refresh")
	}

	if err := a.Close(); err != nil {
		t.Fatal(err)
	}
	time.Sleep(wait)

	if atomic.LoadInt32(&updates) != 5 {
		t.Fatal("Update{} not fired after second decreate")
	}
	if len(g.Games()) != 0 {
		t.Fatal("Game found after second decreate")
	}

	info.GameName += "2"
	info.HostCounter++
	info.GameVersion.Version++
	if a, err = lan.NewAdvertiser(&info); err != nil {
		t.Fatal(err)
	}

	go a.Run()
	time.Sleep(wait)

	if atomic.LoadInt32(&updates) != 5 {
		t.Fatal("Update{} fired after create with different game version")
	}
	if len(g.Games()) != 0 {
		t.Fatal("Game found after create with different game version")
	}

	if err := a.Close(); err != nil {
		t.Fatal(err)
	}
	if err := g.Close(); err != nil {
		t.Fatal(err)
	}
}

func TestUDP(t *testing.T) {
	var gv = w3gs.GameVersion{
		Product: w3gs.ProductTFT,
		Version: 26,
	}

	// Either advertiser or gamelist needs to bind port 6112
	g, err := lan.NewUDPGameList(gv, 6112)
	if err != nil {
		t.Fatal(err)
	}

	testGameList(t, g, gv)
}

func TestMDNS(t *testing.T) {
	var gv = w3gs.GameVersion{
		Product: w3gs.ProductTFT,
		Version: 10030,
	}

	g, err := lan.NewMDNSGameList(gv)
	if err != nil {
		t.Fatal(err)
	}

	testGameList(t, g, gv)
}
