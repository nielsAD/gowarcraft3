// Author:  Niels A.D.
// Project: gowarcraft3 (https://github.com/nielsAD/gowarcraft3)
// License: Mozilla Public License, v2.0

// w3mdump is a tool that decodes and dumps w3m/w3x files.
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"image/png"
	"log"
	"os"
	"strings"

	"github.com/nielsAD/gowarcraft3/file/fs"
	"github.com/nielsAD/gowarcraft3/file/w3m"
)

var (
	binpath = flag.String("b", fs.FindInstallationDir(), "Path to game binaries")
	preview = flag.String("preview", "", "Dump preview image to this file")
	jsonout = flag.Bool("json", false, "Print machine readable format")
)

var logOut = log.New(os.Stdout, "", 0)
var logErr = log.New(os.Stderr, "", 0)

func main() {
	flag.Parse()
	var filename = strings.Join(flag.Args(), " ")

	m, err := w3m.Open(filename)
	if err != nil {
		logErr.Fatal("Open error: ", err)
	}

	info, err := m.Info()
	if err != nil {
		logErr.Fatal("Info error: ", err)
	}

	stor := fs.Open(*binpath)
	defer stor.Close()

	hash, err := m.Checksum(stor)
	if err != nil {
		logErr.Fatal("Checksum error: ", err)
	}

	var print = struct {
		Info     w3m.Info
		Checksum w3m.Hash
	}{
		*info,
		*hash,
	}

	var str = fmt.Sprintf("%+v", print)
	if *jsonout {
		if json, err := json.MarshalIndent(print, "", "  "); err == nil {
			str = string(json)
		}
	}

	logOut.Println(str)

	if *preview != "" {
		img, err := m.Preview()
		if err == os.ErrNotExist {
			img, err = m.MenuMinimap()
		}
		if err != nil {
			logErr.Fatal("Preview error: ", err)
		}

		out, err := os.Create(*preview)
		if err != nil {
			logErr.Fatal("os.Create error: ", err)
		}
		defer out.Close()

		if err := png.Encode(out, img); err != nil {
			logErr.Fatal("png.Encode error: ", err)
		}
	}
}
