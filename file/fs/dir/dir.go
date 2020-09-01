// Author:  Niels A.D.
// Project: gowarcraft3 (https://github.com/nielsAD/gowarcraft3)
// License: Mozilla Public License, v2.0

// Package dir locates the Warcraft III installation and user directories.
package dir

import (
	"os"
)

// Local directory that overrides any global directory
const Local = "./war3"

// Environment variables
const (
	EnvUser    = "WC3_USER_DIR"
	EnvInstall = "WC3_INSTALL_DIR"
)

// UserDir returns the Warcraft III user directory
func UserDir() string {
	if dir, ok := os.LookupEnv(EnvUser); ok {
		return dir
	}
	if _, err := os.Stat(Local); err == nil {
		return Local
	}
	return osUserDir()
}

// InstallDir checks common Warcraft III installation directories
// and returns the first one that exists
func InstallDir() string {
	if dir, ok := os.LookupEnv(EnvInstall); ok {
		return dir
	}
	if _, err := os.Stat(Local); err == nil {
		return Local
	}

	var paths = osInstallDirs()
	for _, p := range paths {
		if _, err := os.Stat(p); err == nil {
			return p
		}
	}

	return ""
}
