// Author:  Niels A.D.
// Project: gowarcraft3 (https://github.com/nielsAD/gowarcraft3)
// License: Mozilla Public License, v2.0

package reg

import (
	"golang.org/x/sys/windows/registry"
)

const regPath = `SOFTWARE\Blizzard Entertainment\Warcraft III`

func osReadInt32(key string) (uint32, error) {
	k, err := registry.OpenKey(registry.CURRENT_USER, regPath, registry.READ)
	if err != nil {
		return 0, err
	}

	val, typ, err := k.GetIntegerValue(key)
	k.Close()

	if typ != registry.DWORD {
		return 0, registry.ErrUnexpectedType
	}

	return (uint32)(val), err
}

func osReadInt64(key string) (uint64, error) {
	k, err := registry.OpenKey(registry.CURRENT_USER, regPath, registry.READ)
	if err != nil {
		return 0, err
	}

	val, typ, err := k.GetIntegerValue(key)
	k.Close()

	if typ != registry.QWORD {
		return 0, registry.ErrUnexpectedType
	}

	return (uint64)(val), err
}

func osReadString(key string) (string, error) {
	k, err := registry.OpenKey(registry.CURRENT_USER, regPath, registry.READ)
	if err != nil {
		return "", err
	}

	val, _, err := k.GetStringValue(key)
	k.Close()

	return val, err
}

func osReadStrings(key string) ([]string, error) {
	k, err := registry.OpenKey(registry.CURRENT_USER, regPath, registry.READ)
	if err != nil {
		return nil, err
	}

	val, _, err := k.GetStringsValue(key)
	k.Close()

	return val, err
}

func osWriteInt32(key string, val uint32) error {
	k, _, err := registry.CreateKey(registry.CURRENT_USER, regPath, registry.WRITE)
	if err == nil {
		err = k.SetDWordValue(key, val)
		k.Close()
	}
	return err
}

func osWriteInt64(key string, val uint64) error {
	k, _, err := registry.CreateKey(registry.CURRENT_USER, regPath, registry.WRITE)
	if err == nil {
		err = k.SetQWordValue(key, val)
		k.Close()
	}
	return err
}

func osWriteString(key string, val string) error {
	k, _, err := registry.CreateKey(registry.CURRENT_USER, regPath, registry.WRITE)
	if err == nil {
		err = k.SetStringValue(key, val)
		k.Close()
	}
	return err
}

func osWriteStrings(key string, val []string) error {
	k, _, err := registry.CreateKey(registry.CURRENT_USER, regPath, registry.WRITE)
	if err == nil {
		err = k.SetStringsValue(key, val)
		k.Close()
	}
	return err
}

func osDelete(key string) error {
	return registry.DeleteKey(registry.CURRENT_USER, regPath+`\`+key)
}
