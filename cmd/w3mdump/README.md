GoWarcraft3/w3mdump
===========
[![Build Status](https://travis-ci.org/nielsAD/gowarcraft3.svg?branch=master)](https://travis-ci.org/nielsAD/gowarcraft3)
[![Build status](https://ci.appveyor.com/api/projects/status/a5cecrpfo0pe14ux/branch/master?svg=true)](https://ci.appveyor.com/project/nielsAD/gowarcraft3)
[![License: MPL 2.0](https://img.shields.io/badge/License-MPL%202.0-brightgreen.svg)](https://opensource.org/licenses/MPL-2.0)

A tool that decodes and dumps w3m/w3x files.

Usage
-----

`./w3mdump [options] [path]`

|   Flag   |  Type  | Description |
|----------|--------|-------------|
|`-preview`|`path`  |Dump preview image to this file|
|`-json`   |`bool`  |Print machine readable format|

Example
-------

```bash
➜ ./w3mdump "(2)BootyBay.w3m"
{
    Info:{
        FileFormat:18
        SaveCount:29
        EditorVersion:6030
        Name:Booty Bay
        Author:Blizzard Entertainment
        Description:Pirates [...] does.
        SuggestedPlayers:1v1
        CamBounds:[-9728 -5120 9728 3584 -9728 3584 9728 -5120]
        CamBoundsCompl:[18 14 8 16]
        Width:160 Height:72
        Flags:Melee|RevealTerrain|WaterWavesOnCliffShores|WaterWavesOnSlopeShores|MapFlags(0x8400)
        Tileset:LordaeronSummer
        LsBackground:4294967295
        LsPath:
        LsText:
        LsTitle:
        LsSubTitle:
        DataSet:4294967295
        PsPath:
        PsText:
        PsTitle:
        PsSubTitle:
        Fog:0
        FogStart:0
        FogEnd:0
        FogDensity:0
        FogColor:0
        WeatherID:
        SoundEnv:
        LightEnv:Tileset(0x00)
        WaterColor:0
        Players:[
            {ID:0 Type:Human Race:Human Flags: Name: StartPosX:-8384 StartPosY:1792 AllyPrioLow:0b0 AllyPrioHigh:0b10}
            {ID:1 Type:Human Race:Orc Flags: Name:Player 2 StartPosX:8448 StartPosY:2048 AllyPrioLow:0b0 AllyPrioHigh:0b1}
        ]
        Forces:[
            {Flags: PlayerSet:0b11111111111111111111111111111111 Name:Force 1}
        ]
        CustomUpgradeAvailabilities:[]
        CustomTechAvailabilities:[]
    }
    Checksum:{
        Xoro:2933683587
        Sha1:[99 89 64 135 240 104 184 17 231 129 73 151 227 159 211 118 160 245 111 24]
    }
}
```

Download
--------

Official binaries for tools are [available](https://github.com/nielsAD/gowarcraft3/releases/latest). Simply download and run.

_Note: additional dependencies may be required (see [build instructions](/README.md#build))._
