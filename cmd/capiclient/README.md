GoWarcraft3/capiclient
===========
[![Build Status](https://travis-ci.org/nielsAD/gowarcraft3.svg?branch=master)](https://travis-ci.org/nielsAD/gowarcraft3)
[![Build status](https://ci.appveyor.com/api/projects/status/a5cecrpfo0pe14ux/branch/master?svg=true)](https://ci.appveyor.com/project/nielsAD/gowarcraft3)
[![License: MPL 2.0](https://img.shields.io/badge/License-MPL%202.0-brightgreen.svg)](https://opensource.org/licenses/MPL-2.0)

Command-line interface for the official classic Battle.net chat API.

Usage
-----

`./capi [options] [server address]`

| Flag |  Type  | Description |
|------|--------|-------------|
|`-e`  |`string`|Endpoint|
|`-k`  |`string`|API Key (will query if omitted)|

Example
-------

```bash
➜ ./capiclient
Enter API key:
12:00:00 Succesfully connected to wss://connect-bot.classic.blizzard.com/v1/rpc/chat
12:00:00 Joined channel 'Clan 1uk1'
12:00:00 niels has joined the channel
# Type "hello" in terminal
12:00:05 [CHAT] niels: hello
```

Download
--------

Official binaries for tools are [available](https://github.com/nielsAD/gowarcraft3/releases/latest). Simply download and run.

_Note: additional dependencies may be required (see [build instructions](/README.md#build))._
