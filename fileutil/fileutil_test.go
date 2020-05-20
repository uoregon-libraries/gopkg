package fileutil

import (
	"os"
	"strings"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
)

// TestFind verifies that Find ... doesn't crash.  This needs a mock for the
// Readdir wrapper function so we can get actual high-level testing without
// relying on a completely unknown filesystem....
func TestFind(t *testing.T) {
	var _, err = Find(os.TempDir(), 1)
	if err != nil {
		t.Fatalf("Got an error trying to read the filesystem!  %s", err)
	}
}

func TestReaddir(t *testing.T) {
	var infos, err = ReaddirSorted(os.TempDir())
	if err != nil {
		t.Fatalf("Got an error trying to read the filesystem!  %s", err)
	}

	t.Logf("Found %d Files:", len(infos))
	for _, info := range infos {
		t.Log("  - " + info.Name())
	}
}

func TestNumberify(t *testing.T) {
	var tests = map[string]struct {
		input string
		want  int
	}{
		"simple":                {"10", 10},
		"prefixed with zeroes":  {"0010", 10},
		"starts with non-digit": {"x1052", 0},
		"ends with non-digits":  {"1052.pdf", 1052},
		"too big number":        {"10523214435342534321423143215435432521343214231423142314.pdf", 0},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			var got = numberify(tc.input)
			if got != tc.want {
				t.Fatalf("expected %d, got %d", tc.want, got)
			}
		})
	}
}

type fi string

func (i fi) Name() string       { return string(i) }
func (i fi) Size() int64        { return 0 }
func (i fi) Mode() os.FileMode  { return 0 }
func (i fi) ModTime() time.Time { return time.Time{} }
func (i fi) IsDir() bool        { return false }
func (i fi) Sys() interface{}   { return nil }

func TestNumericFilenameSort(t *testing.T) {
	var tests = map[string]struct {
		list []os.FileInfo
		want string
	}{
		"simple": {
			list: []os.FileInfo{fi("0001.pdf"), fi("0003.tif"), fi("0002"), fi("0010.tif")},
			want: "0001.pdf, 0002, 0003.tif, 0010.tif",
		},
		"different digit lengths": {
			list: []os.FileInfo{fi("1.pdf"), fi("9"), fi("009.tif"), fi("000002"), fi("10.tif")},
			want: "1.pdf, 000002, 009.tif, 9, 10.tif",
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			SortFileInfosNumerically(tc.list)
			var fnames = make([]string, len(tc.list))
			for i, info := range tc.list {
				fnames[i] = info.Name()
			}
			var diff = cmp.Diff(tc.want, strings.Join(fnames, ", "))
			if diff != "" {
				t.Fatalf(diff)
			}
		})
	}
}
