// Author:  Niels A.D.
// Project: gowarcraft3 (https://github.com/nielsAD/gowarcraft3)
// License: Mozilla Public License, v2.0

package blp_test

import (
	"bytes"
	"crypto/sha512"
	"encoding/base64"
	"image/png"
	"os"
	"testing"

	"github.com/nielsAD/gowarcraft3/file/blp"
)

func TestJPEG(t *testing.T) {
	var TEST_SHA = "cF03T1FzQzhwZwm3F/yp0fo8uDbHe/3qqqOQyJLKcg5HEHQTtk5M08L6mbDoRvzdbWd8SgWNQ+Fb3qSaovCuYg"

	f, err := os.Open("./test.blp")
	if err != nil {
		t.Fatal(f)
	}
	defer f.Close()

	img, err := blp.Decode(f)
	if err != nil {
		t.Fatal(err)
	}

	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		t.Fatal(err)
	}

	var sha = sha512.Sum512(buf.Bytes())
	if base64.RawStdEncoding.EncodeToString(sha[:]) != TEST_SHA {
		t.Fatal("Sha512 mismatch")
	}
}
