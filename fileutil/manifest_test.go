package fileutil

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestManifestEquivalent(t *testing.T) {
	var f1 = FileInfo{Name: "name1", Size: 1, Checksum: "checksum1"}
	var f2 = FileInfo{Name: "name2", Size: 2, Checksum: "checksum2"}
	var f3 = FileInfo{Name: "name3", Size: 3, Checksum: "checksum3"}
	var f4 = FileInfo{Name: "name4", Size: 4, Checksum: "checksum4"}
	var f5 = FileInfo{Name: "name5", Size: 5, Checksum: "checksum5"}
	var f6 = FileInfo{Name: "name6", Size: 6, Checksum: "checksum6"}
	var a, b = &Manifest{}, &Manifest{}

	if !a.Equiv(b) {
		t.Fatalf("Zero value manifests should be equal")
	}

	a.path = "/path"
	a.Created = time.Now()
	a.Files = []FileInfo{f1, f2, f3, f4}

	b.path = a.path
	b.Created = a.Created
	b.Files = []FileInfo{f1, f2, f3, f4}

	if !a.Equiv(b) {
		t.Fatalf("Exact matches should be equivalent")
	}

	a.Files = []FileInfo{f2, f4, f1, f3}
	if !a.Equiv(b) {
		t.Fatalf("Order of files shouldn't change equivalence")
	}

	b.Files = append(b.Files, f3)
	if a.Equiv(b) {
		t.Fatalf("Dupes should still cause non-equivalence")
	}

	a.Files = []FileInfo{f1, f2, f3, f4, f5}
	b.Files = []FileInfo{f1, f2, f3, f4, f6}

	if a.Equiv(b) {
		t.Fatalf("Different file lists shouldn't be equivalent")
	}

	a.Files = b.Files
	a.path = "/foo"
	b.path = "/bar"
	if !a.Equiv(b) {
		t.Fatalf("Having different paths shouldn't affect equivalence")
	}

	a.Created = time.Now()
	if !a.Equiv(b) {
		t.Fatalf("Different create times shouldn't affect equivalence")
	}
}

func _m(t *testing.T) *Manifest {
	var cwd, err = os.Getwd()
	if err != nil {
		t.Fatalf("Error getting current directory: %s", err)
		return nil
	}
	var testdata = filepath.Join(cwd, "testdata")
	return NewManifest(testdata)
}

func _mkf(name string, size int64, checksum string) FileInfo {
	return FileInfo{Name: name, Size: size, Checksum: checksum}
}

// These are the file manifests for what's in the testdata dir
var expectedFiles = []FileInfo{
	_mkf("a.txt", 30, "df879070"),
	_mkf("b.bin", 5000, "df3b5d6a"),
	_mkf("c.null", 0, "00000000"),
}

func TestManifestBuild(t *testing.T) {
	var m = _m(t)
	var err = m.Build()
	if err != nil {
		t.Fatalf("Unable to build manifest: %s", err)
	}

	var expected = len(expectedFiles)
	var got = len(m.Files)
	if expected != got {
		for _, f := range m.Files {
			t.Logf("File: %#v", f)
		}
		t.Fatalf("Invalid manifest: expected to see %d files, but got %d", expected, got)
	}

	m.sortFiles()

	for i := range expectedFiles {
		if m.Files[i] != expectedFiles[i] {
			t.Fatalf("Invalid manifest: expected m.Files[%d] to be %#v, got %#v", i, expectedFiles[i], m.Files[i])
		}
	}
}

func TestManifestWrite(t *testing.T) {
	var m = _m(t)
	m.Build()
	var err = m.Write()
	if err != nil {
		t.Fatalf("Unable to write manifest: %s", err)
	}
}

func TestManifestRead(t *testing.T) {
	var corpus = _m(t)
	corpus.Build()
	corpus.Created = time.Time{}
	var err = corpus.Write()
	if err != nil {
		t.Fatalf("Unable to write fake manifest out: %s", err)
	}

	var m = _m(t)
	m.Read()

	if !m.Created.IsZero() {
		t.Fatalf("Reading existing manifest didn't result in the expected fake time data")
	}

	var expected = len(expectedFiles)
	var got = len(m.Files)
	if expected != got {
		for _, f := range m.Files {
			t.Logf("File: %#v", f)
		}
		t.Fatalf("Invalid manifest: expected to see %d files, but got %d", expected, got)
	}

	m.sortFiles()

	for i := range expectedFiles {
		if m.Files[i] != expectedFiles[i] {
			t.Fatalf("Invalid manifest: expected m.Files[%d] to be %#v, got %#v", i, expectedFiles[i], m.Files[i])
		}
	}
}

func TestManifestChange(t *testing.T) {
	var corpus = _m(t)
	corpus.Build()
	corpus.Write()
	var cwd, _ = os.Getwd()
	var fname = filepath.Join(cwd, "testdata", "foo.dat")

	var pre = _m(t)
	pre.Build()
	if !corpus.Equiv(pre) {
		t.Fatalf("Pre-create, manifests should be the same")
	}

	var err = os.WriteFile(fname, []byte("foo"), 0644)
	if err != nil {
		t.Fatalf("Unable to write file %q: %s", fname, err)
	}
	defer os.Remove(fname)

	var post = _m(t)
	post.Build()
	if corpus.Equiv(post) {
		t.Fatalf("Post-create, manifests should differ")
	}
}
