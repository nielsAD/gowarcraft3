// Author:  Niels A.D.
// Project: gowarcraft3 (https://github.com/nielsAD/gowarcraft3)
// License: Mozilla Public License, v2.0

//go:build !windows && !darwin
// +build !windows,!darwin

package reg

func osReadInt32(key string) (uint32, error) {
	return 0, ErrNotImplemented
}

func osReadInt64(key string) (uint64, error) {
	return 0, ErrNotImplemented
}

func osReadString(key string) (string, error) {
	return "", ErrNotImplemented
}

func osReadStrings(key string) ([]string, error) {
	return nil, ErrNotImplemented
}

func osWriteInt32(key string, val uint32) error {
	return ErrNotImplemented
}

func osWriteInt64(key string, val uint64) error {
	return ErrNotImplemented
}

func osWriteString(key string, val string) error {
	return ErrNotImplemented
}

func osWriteStrings(key string, val []string) error {
	return ErrNotImplemented
}

func osDelete(key string) error {
	return ErrNotImplemented
}
