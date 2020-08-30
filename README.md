GoWarcraft3
===========
[![Build Status](https://travis-ci.org/nielsAD/gowarcraft3.svg?branch=master)](https://travis-ci.org/nielsAD/gowarcraft3)
[![Build status](https://ci.appveyor.com/api/projects/status/a5cecrpfo0pe14ux/branch/master?svg=true)](https://ci.appveyor.com/project/nielsAD/gowarcraft3)
[![GoDoc](https://godoc.org/github.com/nielsAD/gowarcraft3?status.svg)](https://godoc.org/github.com/nielsAD/gowarcraft3)
[![License: MPL 2.0](https://img.shields.io/badge/License-MPL%202.0-brightgreen.svg)](https://opensource.org/licenses/MPL-2.0)

This library provides a set of go packages and command line utilities that implement Warcraft III protocols and file formats.

Tools
-----

|              Name            | Description |
|------------------------------|-------------|
|[capiclient](./cmd/capiclient)|A command-line interface for the official classic Battle.net chat API.|
|[bncsclient](./cmd/bncsclient)|A mocked Warcraft III chat client that can be used to connect to BNCS servers.|
|[w3gsclient](./cmd/w3gsclient)|A mocked Warcraft III game client that can be used to add dummy players to games.|
|  [bncsdump](./cmd/bncsdump)  |A tool that decodes and dumps BNCS packets via pcap (on the wire or from a file).|
|  [w3gsdump](./cmd/w3gsdump)  |A tool that decodes and dumps W3GS packets via pcap (on the wire or from a file).|
|   [w3gdump](./cmd/w3gdump)   |A tool that decodes and dumps w3g/nwg files.|
|   [w3mdump](./cmd/w3mdump)   |A tool that decodes and dumps w3m/w3x files.|

### Download

Official binaries for tools are [available](https://github.com/nielsAD/gowarcraft3/releases/latest). Simply download and run.

_Note: additional dependencies may be required (see [build](#build))._

### Build

```bash
# Linux dependencies
apt-get install --no-install-recommends -y build-essential cmake git golang-go libgmp-dev libbz2-dev zlib1g-dev libpcap-dev

# OSX dependencies
brew install cmake git go gmp bzip2 zlib libpcap

# Windows dependencies (use MSYS2 -- https://www.msys2.org/)
pacman --needed --noconfirm -S git mingw-w64-x86_64-toolchain mingw-w64-x86_64-go mingw-w64-x86_64-cmake

# Download vendor submodules
git submodule update --init --recursive

# Run tests
make test

# Build release files in ./bin/
make release
```

Packages
--------

|      Name      | Description |
|----------------|-------------|
|`file`          |Package `file` implements common utilities for handling Warcraft III file formats.|
|`file/blp`      |Package `blp` is a BLIzzard Picture image format decoder.|
|`file/fs`       |Package `fs` implements Warcraft III file system utilities.|
|`file/mpq`      |Package `mpq` provides golang bindings to the StormLib library to read MPQ archives.|
|`file/reg`      |Package `reg` implements cross-platform registry utilities for Warcraft III.|
|`file/w3g`      |Package `w3g` implements a decoder and encoder for w3g files.|
|`file/w3m`      |Package `w3m` implements basic information extraction functions for w3m/w3x files.|
|`network`       |Package `network` implements common utilities for higher-level (emulated) Warcraft III network components.|
|`network/chat`  |Package `chat` implements the official classic Battle.net chat API.|
|`network/bnet*` |Package `bnet` implements a mocked BNCS client that can be used to interact with BNCS servers.|
|`network/dummy` |Package `dummy` implements a mocked Warcraft III game client that can be used to add dummy players to lobbies.|
|`network/lan`   |Package `lan` implements a mocked Warcraft III LAN client that can be used to discover local games.|
|`network/lobby` |Package `lobby` implements a mocked Warcraft III game server that can be used to host lobbies.|
|`network/peer`  |Package `peer` implements a mocked Warcraft III client that can be used to manage peer connections in lobbies.|
|`protocol`      |Package `protocol` implements common utilities for Warcraft III network protocols.|
|`protocol/capi` |Package `capi` implements the datastructures for the official classic Battle.net chat API.|
|`protocol/bncs*`|Package `bncs` implements the old Battle.net chat protocol for Warcraft III.|
|`protocol/w3gs` |Package `w3gs` implements the game protocol for Warcraft III.|

**\*note:** BNCS/BNet protocol works up until patch 1.32.

### Download

```bash
go get github.com/nielsAD/gowarcraft3/${PACKAGE_NAME}
```

### Import

```go
import (
    "github.com/nielsAD/gowarcraft3/${PACKAGE_NAME}"
)
```

_Note: additional dependencies may be required (see [build](#build))._

### Documentation

Documentation is available on [godoc.org](https://godoc.org/github.com/nielsAD/gowarcraft3)
