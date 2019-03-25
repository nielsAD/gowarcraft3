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
|`-b`         |`path`  |Path to game binaries (will guess if omitted)|
|`-ei`        |`string`|Override exe version string|
|`-ev`        |`uint`  |Override exe version number|
|`-eh`        |`uint`  |Override exe hash|
|`-roc`       |`string`|ROC CD-key|
|`-tft`       |`string`|TFT CD-key|
|`-verify`    |`bool`  |Verify server signature|
|`-sha1`      |`bool`  |SHA1 password authentication (used in old PvPGN servers)|
|`-create`    |`bool`  |Create account|
|`-changepass`|`bool`  |Change password|

Example
-------

Loading version and CD-key info from the default Warcraft III installation directory:

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

Preset version and CD-key info, SHA1 (PvPGN) password authentication:

```bash
➜ ./bncsclient -sha1 -u niels -roc FFFFFFFFFFFFFFFFFFFFFFFFFF -tft FFFFFFFFFFFFFFFFFFFFFFFFFF -ev 0x011b01ad -eh 0xaaaba048 rubattle.net
12:00:00 Succesfully logged onto niels@rubattle.net:6112
12:00:00 Joined channel 'Warcraft III RUS-1'
12:00:00 niels has joined the channel (ping: 41ms)
12:00:00 [INFO] Obey the law, rules are the law!
12:00:00 [INFO] Hello niels, welcome to Rubattle.net!
```

Download
--------

Official binaries for tools are [available](https://github.com/nielsAD/gowarcraft3/releases/latest). Simply download and run.

_Note: additional dependencies may be required (see [build instructions](/README.md#build))._
