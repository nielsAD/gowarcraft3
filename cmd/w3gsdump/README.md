GoWarcraft3/w3gsdump
===========
[![Build Status](https://travis-ci.org/nielsAD/gowarcraft3.svg?branch=master)](https://travis-ci.org/nielsAD/gowarcraft3)
[![Build status](https://ci.appveyor.com/api/projects/status/a5cecrpfo0pe14ux/branch/master?svg=true)](https://ci.appveyor.com/project/nielsAD/gowarcraft3)
[![License: MPL 2.0](https://img.shields.io/badge/License-MPL%202.0-brightgreen.svg)](https://opensource.org/licenses/MPL-2.0)

A tool that decodes and dumps W3GS packets via pcap (on the wire or from a file).

Usage
-----

`./w3gsdump [options]`

|   Flag   |  Type  | Description |
|----------|--------|-------------|
|`-f`      |`string`|Filename to read from|
|`-i`      |`string`|Interface to read packets from|
|`-json`   |`bool`  |Print machine readable format|
|`-promisc`|`bool`  |Set promiscuous mode (default true)|
|`-b`      |`int`   |Max number of bytes to print per blob  (default 128)|
|`-s`      |`int`   |Snap length (max number of bytes to read per packet (default 65536)|


Example
-------

```bash
➜ ./w3gsdump
12:00:00 Sniffing enp11s0
12:00:00 Sniffing lo
12:00:00 [UDP]    192.168.0.101:6112->255.255.255.255:6112  RefreshGame    {HostCounter:1 SlotsUsed:1 SlotsAvailable:12}
12:00:00 [UDP]    192.168.0.101:6112->192.168.0.101:43858   GameInfo       {GameVersion:{Product:W3XP Version:29} HostCounter:1 EntryKey:35527635 GameName:Local Game (PLAYERONE) GameSettings:{GameSettingFlags:SpeedFast|TerrainDefault|ObsNone|TeamsTogether|TeamsFixed MapWidth:180 MapHeight:180 MapXoro:2408033753 MapPath:Maps/(12)IceCrown.w3m HostName:PLAYERONE MapSha1:[119 234 175 148 38 63 150 35 25 193 33 41 43 183 187 80 59 131 226 141]} SlotsTotal:12 GameFlags:Official SlotsUsed:1 SlotsAvailable:12 UptimeSec:6838 GamePort:6112}
12:00:00 [TCP]   192.168.0.101:46088->192.168.0.101:6112    Join           {HostCounter:1 EntryKey:35527635 ListenPort:45423 JoinCounter:1 PlayerName:fakeplayer InternalAddr:{Port:45423 IP:<nil>}}
```

Download
--------

Official binaries for tools are [available](https://github.com/nielsAD/gowarcraft3/releases/latest). Simply download and run.

_Note: additional dependencies may be required (see [build instructions](/README.md#build))._
