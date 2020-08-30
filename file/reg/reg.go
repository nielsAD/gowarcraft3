// Author:  Niels A.D.
// Project: gowarcraft3 (https://github.com/nielsAD/gowarcraft3)
// License: Mozilla Public License, v2.0

// Package reg implements cross-platform registry utilities for Warcraft III.
package reg

import (
	"errors"
)

// Errors
var (
	ErrNotImplemented = errors.New("reg: Method not implemented for platform")
)

// ReadInt32 reads integer value from registry
func ReadInt32(key string) (uint32, error) {
	return osReadInt32(key)
}

// ReadInt64 reads integer value from registry
func ReadInt64(key string) (uint64, error) {
	return osReadInt64(key)
}

// ReadString reads string value from registry
func ReadString(key string) (string, error) {
	return osReadString(key)
}

// ReadStrings reads string list from registry
func ReadStrings(key string) ([]string, error) {
	return osReadStrings(key)
}

// WriteInt32 writes integer value to registry
func WriteInt32(key string, val uint32) error {
	return osWriteInt32(key, val)
}

// WriteInt64 writes integer value to registry
func WriteInt64(key string, val uint64) error {
	return osWriteInt64(key, val)
}

// WriteString writes string value to registry
func WriteString(key string, val string) error {
	return osWriteString(key, val)
}

// WriteStrings writes string list to registry
func WriteStrings(key string, val []string) error {
	return osWriteStrings(key, val)
}

// Delete key from registry
func Delete(key string) error {
	return osDelete(key)
}
