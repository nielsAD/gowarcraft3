GoWarcraft3/w3gdump
===========
[![Build Status](https://travis-ci.org/nielsAD/gowarcraft3.svg?branch=master)](https://travis-ci.org/nielsAD/gowarcraft3)
[![Build status](https://ci.appveyor.com/api/projects/status/a5cecrpfo0pe14ux/branch/master?svg=true)](https://ci.appveyor.com/project/nielsAD/gowarcraft3)
[![License: MPL 2.0](https://img.shields.io/badge/License-MPL%202.0-brightgreen.svg)](https://opensource.org/licenses/MPL-2.0)

A tool that decodes and dumps w3g files.

Usage
-----

`./w3gdump [options] [path]`

|   Flag   |  Type  | Description |
|----------|--------|-------------|
|`-header` |`bool`  |Decode header only|
|`-json`   |`bool`  |Print machine readable format|

Example
-------

```bash
➜ ./w3gdump "lastreplay.w3g"
Header         {GameVersion:{Product:W3XP Version:10030} BuildNumber:6061 DurationMS:640650 SinglePlayer:true}
GameInfo       {HostPlayer:{ID:1 Name:niels Race:RacePref(0x00) JoinCounter:0} GameName:Local Game GameSettings:{GameSettingFlags:SpeedFast|TerrainDefault|ObsNone|TeamsTogether|TeamsFixed MapWidth:116 MapHeight:84 MapXoro:2599102717 MapPath:Maps/FrozenThrone//(2)EchoIsles.w3x HostName:niels MapSha1:[]} GameFlags:Custom|SignedMap NumSlots:24 LanguageID:0}
SlotInfo       {SlotInfo:{Slots:[{PlayerID:1 DownloadStatus:100 SlotStatus:Occupied Computer:false Team:0 Color:0 Race:Random(Selectable) ComputerType:Normal Handicap:100} {PlayerID:0 DownloadStatus:100 SlotStatus:Occupied Computer:true Team:1 Color:1 Race:Random(Selectable) ComputerType:Normal Handicap:100}] RandomSeed:40053178 SlotLayout:Melee NumPlayers:2}}
CountDownStart {GameStart:{}}
CountDownEnd   {GameStart:{}}
GameStart      {}
TimeSlot       {TimeSlot:{Fragment:false TimeIncrementMS:100 Actions:[]}}
TimeSlotAck    {Checksum:[3 171 6 32]}

...

TimeSlot       {TimeSlot:{Fragment:false TimeIncrementMS:101 Actions:[]}}
TimeSlotAck    {Checksum:[69 44 194 113]}
PlayerLeft     {Local:true PlayerID:1 Reason:Lost Counter:4}
```

Download
--------

Official binaries for tools are [available](https://github.com/nielsAD/gowarcraft3/releases/latest). Simply download and run.

_Note: additional dependencies may be required (see [build instructions](/README.md#build))._
