// Author:  Niels A.D.
// Project: gowarcraft3 (https://github.com/nielsAD/gowarcraft3)
// License: Mozilla Public License, v2.0

// +build !windows,!darwin

package fs

import (
	"os"
	"path"
)

func osUserDir() string {
	return path.Join(os.Getenv("HOME"), "Documents/Warcraft III")
}

func osInstallDirs() []string {
	return []string{
		path.Join(os.Getenv("HOME"), ".wine/drive_c/Program Files/Warcraft III"),
		path.Join(os.Getenv("HOME"), ".wine/drive_c/Program Files (x86)/Warcraft III"),
	}
}
