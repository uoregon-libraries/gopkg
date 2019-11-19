package pdf

import (
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestNewImageFromRaw(t *testing.T) {
	var line = "   1     0 image    2574  3649  icc     1   8  jpeg   no       508  0   151   151 1048K  11%"
	var i, err = newImageFromRaw(line)
	if err != nil {
		t.Errorf("Expected success, got %s", err)
	}

	var expectedImage = &Image{
		Page:     "1",
		Num:      "0",
		Type:     "image",
		Width:    "2574",
		Height:   "3649",
		Color:    "icc",
		Comp:     "1",
		BPC:      "8",
		Encoding: "jpeg",
		Interp:   "no",
		Object:   "508",
		ID:       "0",
		XPPI:     "151",
		YPPI:     "151",
		Size:     "1048K",
		Ratio:    "11%",
	}
	var diff = cmp.Diff(expectedImage, i)
	if diff != "" {
		t.Fatalf(diff)
	}
}

func TestNewImageFromRawInlineObject(t *testing.T) {
	var line = "   1    43 image       8     4  cmyk    4   8  image  no   [inline]     307   355  128B 100%"
	var i, err = newImageFromRaw(line)
	if err != nil {
		t.Errorf("Expected success, got %s", err)
	}

	var expectedImage = &Image{
		Page:     "1",
		Num:      "43",
		Type:     "image",
		Width:    "8",
		Height:   "4",
		Color:    "cmyk",
		Comp:     "4",
		BPC:      "8",
		Encoding: "image",
		Interp:   "no",
		Object:   "[inline]",
		ID:       "",
		XPPI:     "307",
		YPPI:     "355",
		Size:     "128B",
		Ratio:    "100%",
	}
	var diff = cmp.Diff(expectedImage, i)
	if diff != "" {
		t.Fatalf(diff)
	}
}
