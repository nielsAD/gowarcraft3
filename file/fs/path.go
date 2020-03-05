// Author:  Niels A.D.
// Project: gowarcraft3 (https://github.com/nielsAD/gowarcraft3)
// License: Mozilla Public License, v2.0

package fs

import (
	"os"
)

// LocalDir overrides any global directory
const LocalDir = "./war3"

// UserDir returns the Warcraft III user directory
func UserDir() string {
	if dir, ok := os.LookupEnv("WC3_USER_DIR"); ok {
		return dir
	}
	if _, err := os.Stat(LocalDir); err == nil {
		return LocalDir
	}
	return osUserDir()
}

// FindInstallationDir checks common Warcraft III installation directories
// and returns the first one that exists
func FindInstallationDir() string {
	if dir, ok := os.LookupEnv("WC3_INSTALL_DIR"); ok {
		return dir
	}
	if _, err := os.Stat(LocalDir); err == nil {
		return LocalDir
	}

	var paths = osInstallDirs()
	for _, p := range paths {
		if _, err := os.Stat(p); err == nil {
			return p
		}
	}

	return ""
}
