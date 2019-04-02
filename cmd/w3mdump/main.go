// Author:  Niels A.D.
// Project: gowarcraft3 (https://github.com/nielsAD/gowarcraft3)
// License: Mozilla Public License, v2.0

// w3gsdump is a tool that decodes and dumps w3m/w3x files.
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"image/png"
	"io"
	"log"
	"os"
	"strings"

	"github.com/nielsAD/gowarcraft3/file/w3m"
)

var (
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
		logErr.Fatal(filename, " ", err)
	}

	info, err := m.Info()
	if err != nil {
		logErr.Fatal(filename, " ", err)
	}

	var defaultFiles = make(map[string]io.Reader)

	if f, err := os.Open("common.j"); err == nil {
		defaultFiles["scripts\\common.j"] = f
		defer f.Close()
	}

	if f, err := os.Open("blizzard.j"); err == nil {
		defaultFiles["scripts\\blizzard.j"] = f
		defer f.Close()
	}

	hash, err := m.Checksum(defaultFiles)
	if err != nil {
		logErr.Fatal(filename, " ", err)
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
			logErr.Fatal(err)
		}

		out, err := os.Create(*preview)
		if err != nil {
			logErr.Fatal(err)
		}
		defer out.Close()

		if err := png.Encode(out, img); err != nil {
			logErr.Fatal(err)
		}
	}
}
