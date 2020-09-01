// Author:  Niels A.D.
// Project: gowarcraft3 (https://github.com/nielsAD/gowarcraft3)
// License: Mozilla Public License, v2.0

package dir

import (
	"os"
	"path/filepath"
)

func osUserDir() string {
	return filepath.Join(os.Getenv("HOME"), "Library/Application Support/Blizzard/Warcraft III")
}

func osInstallDirs() []string {
	return []string{
		"/Applications/Warcraft III",
	}
}
