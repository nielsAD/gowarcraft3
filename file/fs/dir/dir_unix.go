// Author:  Niels A.D.
// Project: gowarcraft3 (https://github.com/nielsAD/gowarcraft3)
// License: Mozilla Public License, v2.0

// +build !windows,!darwin

package dir

import (
	"os"
	"path/filepath"
)

func osUserDir() string {
	return filepath.Join(os.Getenv("HOME"), "Documents/Warcraft III")
}

func osInstallDirs() []string {
	return []string{
		filepath.Join(os.Getenv("HOME"), ".wine/drive_c/Program Files/Warcraft III"),
		filepath.Join(os.Getenv("HOME"), ".wine/drive_c/Program Files (x86)/Warcraft III"),
		filepath.Join(os.Getenv("HOME"), "Games/warcraft-iii-reforged/drive_c/Program Files (x86)/Warcraft III"),
	}
}
