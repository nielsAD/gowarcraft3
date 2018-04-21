// Author:  Niels A.D.
// Project: gowarcraft3 (https://github.com/nielsAD/gowarcraft3)
// License: Mozilla Public License, v2.0

// w3gsdump is a tool that decodes and dumps w3m/w3x files.
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/nielsAD/gowarcraft3/file/w3m"
)

var (
	jsonout = flag.Bool("json", false, "Print machine readable format")
)

var logOut = log.New(os.Stdout, "", 0)
var logErr = log.New(os.Stderr, "", 0)

func main() {
	flag.Parse()
	var filename = strings.Join(flag.Args(), " ")

	m, err := w3m.Load(filename)
	if err != nil {
		logErr.Fatal(filename, " ", err)
	}

	var str = fmt.Sprintf("%+v", *m)
	if *jsonout {
		if json, err := json.MarshalIndent(*m, "", "  "); err == nil {
			str = string(json)
		}
	}

	logOut.Println(str)
}
