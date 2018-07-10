GoWarcraft3/bncsdump
===========
[![Build Status](https://travis-ci.org/nielsAD/gowarcraft3.svg?branch=master)](https://travis-ci.org/nielsAD/gowarcraft3)
[![Build status](https://ci.appveyor.com/api/projects/status/a5cecrpfo0pe14ux/branch/master?svg=true)](https://ci.appveyor.com/project/nielsAD/gowarcraft3)
[![License: MPL 2.0](https://img.shields.io/badge/License-MPL%202.0-brightgreen.svg)](https://opensource.org/licenses/MPL-2.0)

A tool that decodes and dumps BNCS packets via pcap (on the wire or from a file).

Usage
-----

`./bncsdump [options]`

|   Flag   |  Type  | Description |
|----------|--------|-------------|
|`-f`      |`string`|Filename to read from|
|`-i`      |`string`|Interface to read packets from|
|`-json`   |`bool`  |Print machine readable format|
|`-promisc`|`bool`  |Set promiscuous mode (default true)|
|`-p`      |`int`   |BNCS port to sniff (default 6112)|
|`-b`      |`int`   |Max number of bytes to print per blob  (default 128)|
|`-s`      |`int`   |Snap length (max number of bytes to read per packet (default 65536)|

Example
-------

```bash
➜ ./bncsdump
12:00:00 Sniffing enp11s0
12:00:00 Sniffing lo
12:00:00 [TCP]   192.168.0.101:39562->37.244.29.100:6112    AuthInfoReq    {PlatformCode:IX86 GameVersion:{Product:W3XP Version:29} LanguageCode:enUS LocalIP:192.168.0.101 TimeZoneBias:4294967176 MpqLocaleID:1033 UserLanguageID:1033 CountryAbbreviation:USA Country:United States}
12:00:00 [TCP]    37.244.29.100:6112->192.168.0.101:39562   Ping           {Payload:672995397}
12:00:00 [TCP]   192.168.0.101:39562->37.244.29.100:6112    Ping           {Payload:672995397}
12:00:00 [TCP]    37.244.29.100:6112->192.168.0.101:39562   AuthInfoResp   {ServerToken:1693987612 Unknown1:2635 MpqFileTime:131088735080000000 MpqFileName:ver-IX86-4.mpq ValueString:A=156513096 B=2831108732 C=3097134736 4 A=A-S B=B^C C=C^A A=A^B ServerSignature:[...]}
```

Download
--------

Official binaries for tools are [available](https://github.com/nielsAD/gowarcraft3/releases/latest). Simply download and run.

_Note: additional dependencies may be required (see [build instructions](/README.md#build))._
