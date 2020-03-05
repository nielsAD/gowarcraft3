// Author:  Niels A.D.
// Project: gowarcraft3 (https://github.com/nielsAD/gowarcraft3)
// License: Mozilla Public License, v2.0

package fs

import (
	"os"
	"path"

	"golang.org/x/sys/windows/registry"
)

func osUserDir() string {
	return path.Join(os.Getenv("USERPROFILE"), "Documents/Warcraft III")
}

func osInstallDirs() []string {
	var res = []string{
		"C:/Program Files/Warcraft III",
		"C:/Program Files (x86)/Warcraft III",
	}

	if k, err := registry.OpenKey(registry.LOCAL_MACHINE, `SOFTWARE\Microsoft\Windows\CurrentVersion\Uninstall\Warcraft III`, registry.QUERY_VALUE); err == nil {
		if s, _, err := k.GetStringValue("InstallLocation"); err == nil {
			res = append([]string{s}, res...)
		}
		k.Close()
	}
	if k, err := registry.OpenKey(registry.LOCAL_MACHINE, `SOFTWARE\WOW6432Node\Microsoft\Windows\CurrentVersion\Uninstall\Warcraft III`, registry.QUERY_VALUE); err == nil {
		if s, _, err := k.GetStringValue("InstallLocation"); err == nil {
			res = append([]string{s}, res...)
		}
		k.Close()
	}

	return res
}
