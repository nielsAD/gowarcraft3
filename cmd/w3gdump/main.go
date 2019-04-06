// Author:  Niels A.D.
// Project: gowarcraft3 (https://github.com/nielsAD/gowarcraft3)
// License: Mozilla Public License, v2.0

// w3gdump is a tool that decodes and dumps w3g files.
package main

import (
	"bufio"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"reflect"
	"strings"

	"github.com/nielsAD/gowarcraft3/file/w3g"
)

var (
	header  = flag.Bool("header", false, "Decode header only")
	jsonout = flag.Bool("json", false, "Print machine readable format")
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

var errBreakEarly = errors.New("early break")

func main() {
	flag.Parse()
	var filename = strings.Join(flag.Args(), " ")

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

	hdr, data, _, err := w3g.DecodeHeader(b)
	if err != nil {
		logErr.Fatal("DecodeHeader error: ", err)
	}
	defer data.Close()

	print(hdr)
	if err := data.ForEach(func(r w3g.Record) error {
		if *header {
			switch r.(type) {
			case *w3g.CountDownStart, *w3g.CountDownEnd, *w3g.GameStart:
				return errBreakEarly
			}
		}
		print(r)
		return nil
	}); err != nil && err != errBreakEarly {
		logErr.Fatal("Data error: ", err)
	}
}
