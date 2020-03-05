// Author:  Niels A.D.
// Project: gowarcraft3 (https://github.com/nielsAD/gowarcraft3)
// License: Mozilla Public License, v2.0

// Package fs implements Warcraft 3 file system utilities.
package fs

import (
	"bytes"
	"io"
	"io/ioutil"
	"os"
	"path"
	"strings"

	"github.com/jybp/casc"
	"github.com/jybp/casc/common"

	"github.com/nielsAD/gowarcraft3/file/mpq"
)

// Storage provider for Warcraft III file system
type Storage struct {
	dir       []string
	mpq       []*mpq.Archive
	casc      *casc.Explorer
	cascFiles map[string]string
}

var mpqFiles = []string{
	"War3Patch.mpq",
	"War3xlocal.mpq",
	"War3x.mpq",
	"war3.mpq",
}

// Open Warcraft III storage from installPath and userPath
func Open(installPath string, userPaths ...string) *Storage {
	installPath = path.Clean(installPath)

	var stor = Storage{
		dir: append([]string{installPath}, userPaths...),
	}

	for _, mpqFileName := range mpqFiles {
		if archive, err := mpq.OpenArchive(path.Join(installPath, mpqFileName)); err == nil {
			stor.mpq = append(stor.mpq, archive)
		}
	}

	if explorer, err := casc.Local(installPath); err == nil {
		stor.casc = explorer

		if files, err := explorer.Files(); err == nil {
			stor.cascFiles = map[string]string{}
			for _, f := range files {
				stor.cascFiles[strings.ToLower(f)] = f
			}
		}
	}

	return &stor
}

// Close storage
func (stor *Storage) Close() error {
	var err error
	for _, archive := range stor.mpq {
		if e := archive.Close(); e != nil {
			err = e
		}
	}
	return err
}

var cascPrefixes = []string{
	"",
	"War3.w3mod:",
	"War3.mpq:",
	"enUS-War3Local.mpq:",
	"enUS-",
}

// Open subFileName from storage
func (stor *Storage) Open(subFileName string) (io.ReadCloser, error) {
	subFileName = strings.Replace(subFileName, "\\", "/", -1)

	for _, dir := range stor.dir {
		if file, err := os.Open(path.Join(dir, subFileName)); !os.IsNotExist(err) {
			return file, err
		}
	}

	for _, archive := range stor.mpq {
		file, err := archive.Open(subFileName)
		if file != nil {
			return file, err
		} else if !os.IsNotExist(err) {
			return nil, err
		}
	}

	for _, cascPrefix := range cascPrefixes {
		var file, ok = stor.cascFiles[strings.ToLower(common.CleanPath(cascPrefix+subFileName))]
		if !ok {
			continue
		}

		var b, err = stor.casc.Extract(file)
		if err != nil {
			return nil, err
		}

		return ioutil.NopCloser(bytes.NewReader(b)), err
	}

	return nil, os.ErrNotExist
}
