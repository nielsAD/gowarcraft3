// Author:  Niels A.D.
// Project: gowarcraft3 (https://github.com/nielsAD/gowarcraft3)
// License: Mozilla Public License, v2.0

package reg

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

type defAction string

const (
	defDomain = "com.blizzard.Warcraft III"

	actRead   defAction = "read"
	actWrite  defAction = "write"
	actDelete defAction = "delete"
)

func defaults(action defAction, key string, arg ...string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	cmd := exec.CommandContext(ctx, "defaults", append([]string{(string)(action), key, defDomain}, arg...)...)

	var out bytes.Buffer
	cmd.Stdout = &out

	err := cmd.Run()
	cancel()

	return strings.TrimSpace(out.String()), err
}

func osReadInt32(key string) (uint32, error) {
	if out, err := defaults(actRead, key); err != nil {
		return 0, err
	} else if i, err := strconv.ParseUint(out, 10, 32); err != nil {
		return 0, err
	} else {
		return (uint32)(i), nil
	}
}

func osReadInt64(key string) (uint64, error) {
	if out, err := defaults(actRead, key); err != nil {
		return 0, err
	} else if i, err := strconv.ParseUint(out, 10, 64); err != nil {
		return 0, err
	} else {
		return (uint64)(i), nil
	}
}

func osReadString(key string) (string, error) {
	return defaults(actRead, key)
}

func osReadStrings(key string) ([]string, error) {
	out, err := defaults(actRead, key)
	return strings.Split(out, "\n"), err
}

func osWriteInt32(key string, val uint32) error {
	_, err := defaults(actWrite, key, "-int", fmt.Sprintf("%d", val))
	return err
}

func osWriteInt64(key string, val uint64) error {
	_, err := defaults(actWrite, key, "-int", fmt.Sprintf("%d", val))
	return err
}

func osWriteString(key string, val string) error {
	_, err := defaults(actWrite, key, "-string", val)
	return err
}

func osWriteStrings(key string, val []string) error {
	_, err := defaults(actWrite, key, append([]string{"-array"}, val...)...)
	return err
}

func osDelete(key string) error {
	_, err := defaults(actDelete, key)
	return err
}
