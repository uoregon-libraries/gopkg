package manifest

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/uoregon-libraries/gopkg/hasher"
)

func TestEquivalent(t *testing.T) {
	var f1 = FileInfo{Name: "name1", Size: 1, Mode: 0644, ModTime: time.Date(1999, 1, 1, 0, 0, 0, 0, time.UTC)}
	var f2 = FileInfo{Name: "name2", Size: 2, Mode: 0644, ModTime: time.Date(1999, 1, 2, 0, 0, 0, 0, time.UTC)}
	var f3 = FileInfo{Name: "name3", Size: 3, Mode: 0644, ModTime: time.Date(1999, 1, 3, 0, 0, 0, 0, time.UTC)}
	var f4 = FileInfo{Name: "name4", Size: 4, Mode: 0644, ModTime: time.Date(1999, 1, 4, 0, 0, 0, 0, time.UTC)}
	var f5 = FileInfo{Name: "name5", Size: 5, Mode: 0644, ModTime: time.Date(1999, 1, 5, 0, 0, 0, 0, time.UTC)}
	var f6 = FileInfo{Name: "name6", Size: 6, Mode: 0644, ModTime: time.Date(1999, 1, 6, 0, 0, 0, 0, time.UTC)}
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
		t.Fatalf("Different manifest create times shouldn't affect equivalence")
	}

	a.Files = []FileInfo{f1, f3}
	b.Files = []FileInfo{f1, f3}
	a.Files[0].Mode = 0755
	if a.Equiv(b) {
		t.Fatalf("Different file modes should mean differing manifests")
	}

	a.Files = []FileInfo{f1, f3}
	b.Files[1].ModTime = time.Now()
	if a.Equiv(b) {
		t.Fatalf("Different file modtime should mean differing manifests")
	}
}

func _m(t *testing.T) *Manifest {
	var cwd, err = os.Getwd()
	if err != nil {
		t.Fatalf("Error getting current directory: %s", err)
		return nil
	}
	var testdata = filepath.Join(cwd, "testdata")
	return New(testdata)
}

func _mkf(name string, size int64, mode fs.FileMode) FileInfo {
	var cwd, _ = os.Getwd()
	var fullpath = filepath.Join(cwd, "testdata", name)
	var info, err = os.Stat(fullpath)
	if err != nil {
		panic(fmt.Sprintf("Unable to read %q in _mkf: %s", fullpath, err))
	}
	return FileInfo{Name: name, Size: size, Mode: mode, ModTime: info.ModTime()}
}

// These are the file manifests for what's in the testdata dir
var expectedFiles = []FileInfo{
	_mkf("a.txt", 30, 0644),
	_mkf("b.bin", 5000, 0644),
	_mkf("c.null", 0, 0644),
}

func TestBuild(t *testing.T) {
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

func TestWrite(t *testing.T) {
	var m = _m(t)
	m.Build()
	var err = m.Write()
	if err != nil {
		t.Fatalf("Unable to write manifest: %s", err)
	}
}

func TestRead(t *testing.T) {
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

func TestFileInfoEqual(t *testing.T) {
	var a = FileInfo{Name: "name", Size: 12345, Mode: 0644, ModTime: time.Now()}
	var b = a

	if !a.Equal(b) {
		t.Fatalf("%#v should be equal to %#v", a, b)
	}

	b.ModTime = time.Unix(0, a.ModTime.UnixNano())
	if a == b {
		t.Fatalf("After hacking time to not have monotonic clock, a still equals b")
	}
	if !a.Equal(b) {
		t.Fatalf("After hacking time to not have monotonic clock, a.Equal(b) should be true")
	}
}

func TestChange(t *testing.T) {
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

func TestManifestWithHash(t *testing.T) {
	var m = _m(t)
	m.Hasher = hasher.SHA256()

	var err = m.Build()
	if err != nil {
		t.Fatalf("Unable to build manifest with hash: %s", err)
	}

	for _, f := range m.Files {
		t.Logf("%q: %s", f.Name, f.Sum)
		if f.Sum == "" {
			t.Errorf("Expected file %q to have a non-empty sum", f.Name)
		}
	}

	// Check hashes with output from sha256sum
	var sums = map[string]string{
		"a.txt":  "6d9edf0206a454f22d168dc5ca1e2ce422d0bed2fd0fa5f7092ea720900eb4ae",
		"b.bin":  "0ddc84e6ee1159d24fa9223df52fa5c1dd15593217c2fcf53e48fa0b334f84b8",
		"c.null": "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855",
	}
	for _, f := range m.Files {
		var expected, exist = sums[f.Name]
		if !exist {
			t.Errorf("Unexpected file found in testdata: %q", f.Name)
		}
		var got = f.Sum
		if got != expected {
			t.Errorf("Expected %q to have sum %q, but got %q", f.Name, expected, got)
		}
	}
}

func TestValidateOneSidedHash(t *testing.T) {
	// Create a new manifest with a hash function and build it
	var m = _m(t)
	m.Hasher = hasher.SHA256()

	var err = m.Build()
	if err != nil {
		t.Fatalf("Unable to build manifest with hash: %s", err)
	}

	// Set a bogus hash value so we know it will fail if both manifests are set
	// to use a checksum
	m.Files[1].Sum = "foobar"

	// Build manifest 2 - no checksum
	var m2 = _m(t)
	err = m2.Build()
	if err != nil {
		t.Fatalf("Unable to build manifest: %s", err)
	}

	// Verify that the manifests are equivalent
	if !m.Equiv(m2) {
		t.Errorf("Manifests should be equivalent when hashing isn't set for one")
	}

	// Give m2 a bogus sum and verify that manifests are not equivalent
	m2.Files[2].Sum = "foobar 2"
	if m.Equiv(m2) {
		t.Errorf("Manifests shouldn't be equivalent when hashing is on for both")
	}
}
