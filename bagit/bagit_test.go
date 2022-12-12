package bagit

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/uoregon-libraries/gopkg/assert"
)

func TestGenerateChecksums(t *testing.T) {
	var wd, err = os.Getwd()
	if err != nil {
		panic(err)
	}

	var path = filepath.Join(wd, "testdata")
	var b = New(path)
	err = b.GenerateChecksums()
	assert.NilError(err, fmt.Sprintf("generating checksums in %q", b.root), t)

	var expectedChecksums = []string{
		"60fa80b948a0acc557a6ba7523f4040a7b452736723df20f118d0aacb5c1901b", // another.txt's "sha256sum" value
		"55f8718109829bf506b09d8af615b9f107a266e19f7a311039d1035f180b22d4", // test.txt's "sha256sum" value
	}

	assert.Equal(len(expectedChecksums), len(b.Checksums), "checksum list length", t)

	for i, ck := range b.Checksums {
		assert.Equal(expectedChecksums[i], ck.Checksum, "checksum for "+ck.Path, t)
	}
}

func TestWriteTagFiles(t *testing.T) {
	var wd, err = os.Getwd()
	if err != nil {
		panic(err)
	}

	var path = filepath.Join(wd, "testdata")
	os.Remove(filepath.Join(path, "manifest-sha256.txt"))
	os.Remove(filepath.Join(path, "tagmanifest-sha256.txt"))
	os.Remove(filepath.Join(path, "bagit.txt"))
	var b = New(path)
	err = b.WriteTagFiles()
	if err != nil {
		t.Fatalf("error generating checksums in %q: %s", b.root, err)
	}

	var fname = "manifest-sha256.txt"
	var raw []byte
	raw, err = ioutil.ReadFile(filepath.Join(path, fname))
	if err != nil {
		t.Fatalf("error reading %q: %s", fname, err)
	}
	var got = string(raw)
	var expected = `60fa80b948a0acc557a6ba7523f4040a7b452736723df20f118d0aacb5c1901b  data/another.txt
55f8718109829bf506b09d8af615b9f107a266e19f7a311039d1035f180b22d4  data/test.txt
`
	if expected != got {
		t.Fatalf("Expected %q to be %q, but got %q", fname, expected, raw)
	}

	fname = "tagmanifest-sha256.txt"
	raw, err = ioutil.ReadFile(filepath.Join(path, fname))
	if err != nil {
		t.Fatalf("error reading %q: %s", fname, err)
	}
	got = string(raw)
	expected = `157add7a6600f47a8149b9eab2b35370300f54a73475ded76694078eec5a77df  .gitignore
e91f941be5973ff71f1dccbdd1a32d598881893a7f21be516aca743da38b1689  bagit.txt
e24a952af486ce42a2119d89bec8c7a8c42c2ae9e6302efce5833cf381775594  manifest-sha256.txt
`
	if expected != got {
		t.Fatalf("Expected %q to be %q, but got %q", fname, expected, raw)
	}
}

func TestReadManifests(t *testing.T) {
	var wd, err = os.Getwd()
	if err != nil {
		panic(err)
	}

	var path = filepath.Join(wd, "testdata")
	os.Remove(filepath.Join(path, "manifest-sha256.txt"))
	os.Remove(filepath.Join(path, "tagmanifest-sha256.txt"))
	os.Remove(filepath.Join(path, "bagit.txt"))
	var b = New(path)
	err = b.WriteTagFiles()
	if err != nil {
		t.Fatalf("Error writing tag files: %s", err)
	}

	var b2 = New(path)
	err = b2.ReadManifests()
	if err != nil {
		t.Fatalf("Error reading tag files: %s", err)
	}

	if len(b.Checksums) != len(b2.Checksums) {
		t.Fatalf("b.Checksums: %d elements; b2.Checksums: %d elements", len(b.Checksums), len(b2.Checksums))
	}
	for i := range b.Checksums {
		if *b.Checksums[i] != *b2.Checksums[i] {
			t.Fatalf("b.Checksums[%d]: %#v; b2.Checksums[%d]: %#v", i, b.Checksums[i], i, b2.Checksums[i])
		}
	}

	if len(b.TagSums) != len(b2.TagSums) {
		t.Fatalf("b.TagSums: %d elements; b2.TagSums: %d elements", len(b.TagSums), len(b2.TagSums))
	}
	for i := range b.TagSums {
		if *b.TagSums[i] != *b2.TagSums[i] {
			t.Fatalf("b.TagSums[%d]: %#v; b2.TagSums[%d]: %#v", i, b.TagSums[i], i, b2.TagSums[i])
		}
	}

	// It should be fine without a tag manifest; it just won't have that data
	os.Remove(filepath.Join(path, "tagmanifest-sha256.txt"))
	err = b2.ReadManifests()
	if err != nil {
		t.Fatalf("Lack of a tag manifest shouldn't get an error, but we got %s", err)
	}

	if len(b.Checksums) != len(b2.Checksums) {
		t.Fatalf("b.Checksums: %d elements; b2.Checksums: %d elements", len(b.Checksums), len(b2.Checksums))
	}
	for i := range b.Checksums {
		if *b.Checksums[i] != *b2.Checksums[i] {
			t.Fatalf("b.Checksums[%d]: %#v; b2.Checksums[%d]: %#v", i, b.Checksums[i], i, b2.Checksums[i])
		}
	}

	if len(b2.TagSums) != 0 {
		t.Fatalf("TagSums should be empty")
	}

	// This should puke - manifest is required
	os.Remove(filepath.Join(path, "manifest-sha256.txt"))
	err = b2.ReadManifests()
	if err == nil {
		t.Fatalf("Lack of a manifest should get an error, but we didn't get one")
	}
}
