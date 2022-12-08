package fileutil

import (
	"io/fs"
	"os"
	"path/filepath"
	"testing"
)

func TestSyncDirectory(t *testing.T) {
	var cwd, _ = os.Getwd()
	var src = filepath.Join(cwd, "testdata")
	var dst, err = os.MkdirTemp("", "fileutil-*")
	if err != nil {
		t.Fatalf("Unable to create temp dir: %s", err)
	}
	defer os.RemoveAll(dst)
	t.Logf("Created temp dir %q", dst)

	err = SyncDirectory(src, dst)
	if err != nil {
		t.Fatalf("Unable to sync %q to %q: %s", src, dst, err)
	}

	var srcI, dstI []fs.FileInfo
	srcI, err = ReaddirSorted(src)
	if err != nil {
		t.Fatalf("Got an error trying to read source dir %q: %s", src, err)
	}

	dstI, err = ReaddirSorted(dst)
	if err != nil {
		t.Fatalf("Got an error trying to read dest dir %q: %s", dst, err)
	}

	// VERY basic sanity check here
	if len(srcI) != len(dstI) {
		t.Fatalf("Source and dest have different number of files")
	}
	for i := range srcI {
		var a, b = srcI[i], dstI[i]
		if a.Name() != b.Name() || a.Size() != b.Size() {
			t.Fatalf("Source and dest files not equivalent: %#v != %#v", a, b)
		}
	}
}

func TestSyncDirectoryExcluding(t *testing.T) {
	var cwd, _ = os.Getwd()
	var src = filepath.Join(cwd, "testdata")
	var dst, err = os.MkdirTemp("", "fileutil-*")
	if err != nil {
		t.Fatalf("Unable to create temp dir: %s", err)
	}
	defer os.RemoveAll(dst)
	t.Logf("Created temp dir %q", dst)

	err = SyncDirectoryExcluding(src, dst, []string{"*.xml", "*.bin"})
	if err != nil {
		t.Fatalf("Unable to sync %q to %q: %s", src, dst, err)
	}

	var srcI, dstI []fs.FileInfo
	srcI, err = ReaddirSorted(src)
	if err != nil {
		t.Fatalf("Got an error trying to read source dir %q: %s", src, err)
	}

	dstI, err = ReaddirSorted(dst)
	if err != nil {
		t.Fatalf("Got an error trying to read dest dir %q: %s", dst, err)
	}

	if len(srcI)-1 != len(dstI) {
		t.Fatalf("Dest should have one fewer file than source")
	}

	// Remove the file we skipped from srcI so we can do a simple compare again
	srcI = []fs.FileInfo{srcI[0], srcI[2]}
	for i := range srcI {
		var a, b = srcI[i], dstI[i]
		if a.Name() != b.Name() || a.Size() != b.Size() {
			t.Fatalf("Source and dest files not equivalent: %#v != %#v", a, b)
		}
	}
}
