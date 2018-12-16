GoWarcraft3/bncsclient
===========
[![Build Status](https://travis-ci.org/nielsAD/gowarcraft3.svg?branch=master)](https://travis-ci.org/nielsAD/gowarcraft3)
[![Build status](https://ci.appveyor.com/api/projects/status/a5cecrpfo0pe14ux/branch/master?svg=true)](https://ci.appveyor.com/project/nielsAD/gowarcraft3)
[![License: MPL 2.0](https://img.shields.io/badge/License-MPL%202.0-brightgreen.svg)](https://opensource.org/licenses/MPL-2.0)

A mocked Warcraft III chat client that can be used to connect to BNCS servers.

Usage
-----

`./bncsclient [options] [server address]`

|     Flag    |  Type  | Description |
|-------------|--------|-------------|
|`-u`         |`string`|Username (will query if omitted)|
|`-p`         |`string`|Password (will query if omitted)|
|`-np`        |`string`|New password (used by `-changepass`, will query if omitted)|
|`-b`         |`path`  |Path to game binaries|
|`-v`         |`int`   |Game version|
|`-roc`       |`string`|ROC CD-key|
|`-tft`       |`string`|TFT CD-key|
|`-create`    |`bool`  |Create account|
|`-changepass`|`bool`  |Change password|

Example
-------

```bash
➜ ./bncsclient -u niels europe.battle.net
Enter password:
12:00:00 Succesfully logged onto niels@europe.battle.net:6112
12:00:00 Joined channel 'W3 En-21'
12:00:00 niels has joined the channel (ping: 31ms)
12:00:00 [INFO] There are currently 2391 users playing 393 games of Warcraft III The Frozen Throne, and 15384 users playing 11699 games on Battle.net.
12:00:00 [INFO] Last logon: Fri Jul 6  7:52 PM
# Type "hello" in terminal
12:00:05 [CHAT] niels: hello
12:00:05 [INFO] No one hears you.
```

Download
--------

Official binaries for tools are [available](https://github.com/nielsAD/gowarcraft3/releases/latest). Simply download and run.

_Note: additional dependencies may be required (see [build instructions](/README.md#build))._
