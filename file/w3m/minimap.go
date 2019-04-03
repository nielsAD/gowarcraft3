// Author:  Niels A.D.
// Project: gowarcraft3 (https://github.com/nielsAD/gowarcraft3)
// License: Mozilla Public License, v2.0

package w3m

import (
	"image"
	"image/color"
	"image/draw"
	"io"

	"github.com/ftrvxmtrx/tga"
	"github.com/nielsAD/gowarcraft3/file/blp"
	"github.com/nielsAD/gowarcraft3/protocol"
)

// Preview returns a preview image
func (m *Map) Preview() (image.Image, error) {
	f, err := m.Archive.Open("war3mapPreview.tga")
	if err != nil {
		return nil, err
	}
	defer f.Close()

	return tga.Decode(f)
}

// Minimap returns an image with the minimap
func (m *Map) Minimap() (image.Image, error) {
	f, err := m.Archive.Open("war3mapMap.blp")
	if err != nil {
		return nil, err
	}
	defer f.Close()

	return blp.Decode(f)
}

// MinimapIcons returns an image with (just the) minimap icons
func (m *Map) MinimapIcons() (image.Image, error) {
	mmp, err := m.Archive.Open("war3map.mmp")
	if err != nil {
		return nil, err
	}
	defer mmp.Close()

	var b protocol.Buffer
	if _, err := io.Copy(&b, mmp); err != nil {
		return nil, err
	}

	// Format version
	if b.ReadUInt32() != 0 {
		return nil, ErrBadFormat
	}

	var num = int(b.ReadUInt32())
	if b.Size() != num*16 {
		return nil, ErrBadFormat
	}

	var img = image.NewRGBA(image.Rect(0, 0, 256, 256))

	for i := 0; i < num; i++ {
		var icon = MinimapIcon(b.ReadUInt32())
		var posx = int(b.ReadUInt32())
		var posy = int(b.ReadUInt32())
		var rgba = color.RGBA{
			B: b.ReadUInt8(),
			G: b.ReadUInt8(),
			R: b.ReadUInt8(),
			A: b.ReadUInt8(),
		}

		var i = icon.Icon()
		if i == nil {
			continue
		}

		var irect = i.Bounds()
		var orect = image.Rect(posx, posy, posx+irect.Dx(), posy+irect.Dy())
		orect = orect.Sub(image.Point{irect.Dx() / 2, irect.Dy() / 2})

		if (rgba == color.RGBA{255, 255, 255, 255}) {
			draw.Draw(img, orect, i, irect.Min, draw.Over)
		} else {
			draw.DrawMask(img, orect, &image.Uniform{rgba}, irect.Min, i, irect.Min, draw.Over)
		}
	}

	return img, nil
}

// MenuMinimap returns the minimap with icons
func (m *Map) MenuMinimap() (image.Image, error) {
	img, err := m.Minimap()
	if err != nil {
		return nil, err
	}

	icons, err := m.MinimapIcons()
	if err != nil {
		return nil, err
	}

	var res = image.NewRGBA(img.Bounds())
	draw.Draw(res, res.Rect, img, img.Bounds().Min, draw.Src)
	draw.Draw(res, res.Rect, icons, icons.Bounds().Min, draw.Over)

	return res, nil
}
