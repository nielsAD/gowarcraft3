GoWarcraft3/w3gsclient
===========
[![Build Status](https://travis-ci.org/nielsAD/gowarcraft3.svg?branch=master)](https://travis-ci.org/nielsAD/gowarcraft3)
[![Build status](https://ci.appveyor.com/api/projects/status/a5cecrpfo0pe14ux/branch/master?svg=true)](https://ci.appveyor.com/project/nielsAD/gowarcraft3)
[![License: MPL 2.0](https://img.shields.io/badge/License-MPL%202.0-brightgreen.svg)](https://opensource.org/licenses/MPL-2.0)

A mocked Warcraft III game client that can be used to add dummy players to games.

Usage
-----

`./w3gsclient [options] [host address]`

| Flag  |  Type  | Description |
|-------|--------|-------------|
|`-lan` |`bool`  |Find a game on LAN|
|`-tft` |`bool`  |Search for TFT instead of ROC games (only used when searching local) (default `true`)|
|`-v`   |`int`   |Game version (only used when searching local) (default `29`)|
|`-e`   |`uint`  |Entry key (only used when entering local game)|
|`-c`   |`uint`  |Host counter (default `1`)|
|`-dial`|`bool`  |Dial peers (default `true`)|
|`-l`   |`int`   |Listen on port (0 to pick automatically)|
|`-n`   |`string`|Player name (default `fakeplayer`)|

Example
-------

```bash
➜ ./w3gsclient -lan -n niels
[niels] 12:00:00 Joined lobby with (ID: 2)
[niels] 12:00:00 PLAYERONE has joined the game (ID: 1)
# Type "hello" in terminal
[niels] 12:00:05 [CHAT] niels (ID: 2): hello
[niels] 12:00:15 [CHAT] PLAYERONE (ID: 1): hi
```

Download
--------

Official binaries for tools are [available](https://github.com/nielsAD/gowarcraft3/releases/latest). Simply download and run.

_Note: additional dependencies may be required (see [build instructions](/README.md#build))._
