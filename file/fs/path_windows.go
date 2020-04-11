// Author:  Niels A.D.
// Project: gowarcraft3 (https://github.com/nielsAD/gowarcraft3)
// License: Mozilla Public License, v2.0

package fs

import (
	"os"
	"path"

	"golang.org/x/sys/windows/registry"
)

func docPath() string {
	if k, err := registry.OpenKey(registry.CURRENT_USER, `SOFTWARE\Microsoft\Windows\CurrentVersion\Explorer\User Shell Folders`, registry.QUERY_VALUE); err == nil {
		if s, _, err := k.GetStringValue("{F42EE2D3-909F-4907-8871-4C22FC0BF756}"); err == nil {
			return s
		}
		k.Close()
	}
	return path.Join(os.Getenv("USERPROFILE"), "Documents")
}

func osUserDir() string {
	return path.Join(docPath(), "Warcraft III")
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
