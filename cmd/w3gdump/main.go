// Author:  Niels A.D.
// Project: gowarcraft3 (https://github.com/nielsAD/gowarcraft3)
// License: Mozilla Public License, v2.0

// w3gdump is a tool that decodes and dumps w3g files.
package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"reflect"
	"strings"

	"github.com/nielsAD/gowarcraft3/file/w3g"
)

var (
	sanitize = flag.String("sanitize", "", "Dump cleaned up replay to this file (no chat, sane colors)")
	header   = flag.Bool("header", false, "Decode header only")
	stream   = flag.Bool("stream", false, "Stream game to LAN")
	jsonout  = flag.Bool("json", false, "Print machine readable format")
)

var logOut = log.New(os.Stdout, "", 0)
var logErr = log.New(os.Stderr, "", 0)

func print(v interface{}) {
	var str = fmt.Sprintf("%+v", v)[1:]
	if *jsonout {
		if json, err := json.Marshal(v); err == nil {
			str = string(json)
		}
	}

	logOut.Printf("%-14v %v\n", reflect.TypeOf(v).String()[5:], str)
}

func main() {
	flag.Parse()
	var filename = strings.Join(flag.Args(), " ")

	if *stream {
		if err := cast(filename); err != nil {
			logErr.Fatal("Stream error: ", err)
		}
		return
	}

	f, err := os.Open(filename)
	if err != nil {
		logErr.Fatal("Open error: ", err)
	}
	defer f.Close()

	// Find header, nwg files have their own header prepended
	var b = bufio.NewReaderSize(f, 8192)
	if _, err := w3g.FindHeader(b); err != nil {
		logErr.Fatal("Cannot find header: ", err)
	}

	hdr, data, _, err := w3g.DecodeHeader(b, w3g.NewFactoryCache(w3g.DefaultFactory))
	if err != nil {
		logErr.Fatal("DecodeHeader error: ", err)
	}

	var enc *w3g.Encoder
	if *sanitize != "" {
		o, err := os.Create(*sanitize)
		if err != nil {
			logErr.Fatal("Open error: ", err)
		}
		defer o.Close()

		enc, err = w3g.NewEncoder(o, hdr.Encoding())
		if err != nil {
			logErr.Fatal("NewEncoder error: ", err)
		}
		enc.Header = *hdr
	}

	var skip = false
	var maxp uint8 = 24
	if hdr.GameVersion.Version < 29 {
		maxp = 12
	}

	print(hdr)
	if err := data.ForEach(func(r w3g.Record) error {
		if enc != nil {
			var write = true

			switch v := r.(type) {
			case *w3g.ChatMessage:
				write = false
			case *w3g.SlotInfo:
				var c = uint8(0)
				for i := range v.Slots {
					if v.Slots[i].Team >= maxp {
						continue
					}
					v.Slots[i].Color = c
					c++
				}
			}

			if write {
				if _, err := enc.WriteRecord(r); err != nil {
					return err
				}
			}
		}
		if !skip && *header {
			switch r.(type) {
			case *w3g.CountDownStart, *w3g.CountDownEnd, *w3g.GameStart:
				if enc == nil {
					return errBreakEarly
				}

				skip = true
			}
		}

		if !skip {
			print(r)
		}
		return nil
	}); err != nil && err != errBreakEarly {
		logErr.Fatal("Data error: ", err)
	}

	if enc != nil {
		if err := enc.Close(); err != nil {
			logErr.Fatal("Save error: ", err)
		}
	}
}
