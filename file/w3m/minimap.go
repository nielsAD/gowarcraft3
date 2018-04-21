// Author:  Niels A.D.
// Project: gowarcraft3 (https://github.com/nielsAD/gowarcraft3)
// License: Mozilla Public License, v2.0

package w3m

import (
	"image"

	"github.com/ftrvxmtrx/tga"
	"github.com/nielsAD/gowarcraft3/file/blp"
)

// Minimap returns an image with the minimap
func (m *Map) Minimap() (image.Image, error) {
	f, err := m.Archive.Open("war3mapMap.blp")
	if err != nil {
		return nil, err
	}
	defer f.Close()

	return blp.Decode(f)
}

// Preview returns a preview image
func (m *Map) Preview() (image.Image, error) {
	f, err := m.Archive.Open("war3mapPreview.tga")
	if err != nil {
		return nil, err
	}
	defer f.Close()

	return tga.Decode(f)
}
