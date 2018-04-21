// Author:  Niels A.D.
// Project: gowarcraft3 (https://github.com/nielsAD/gowarcraft3)
// License: Mozilla Public License, v2.0

// Package w3m implements basic information extraction functions for w3m/w3x files.
package w3m

import (
	"github.com/nielsAD/gowarcraft3/file/mpq"
)

// Map refers to an w3m/w3x map (MPQ archive)
type Map struct {
	Archive *mpq.Archive
	ts      map[int]string
}

// Open a w3m/w3x map file
func Open(fileName string) (*Map, error) {
	var archive, err = mpq.OpenArchive(fileName)
	if err != nil {
		return nil, err
	}
	return &Map{Archive: archive}, nil
}

// Close a w3m/w3x map file
func (m *Map) Close() error {
	return m.Archive.Close()
}

// Signed checks if map is signed with a strong signature
func (m *Map) Signed() bool {
	return m.Archive.StrongSigned()
}
