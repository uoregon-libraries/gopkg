package bagit

import (
	"fmt"
	"os"
	"sort"
	"testing"

	"github.com/uoregon-libraries/gopkg/assert"
)

func TestGenerateChecksums(t *testing.T) {
	var err error
	var path string
	var b *Bag

	path, err = os.Getwd()
	if err != nil {
		panic(err)
	}

	b = New(path)
	err = b.GenerateChecksums()
	assert.NilError(err, fmt.Sprintf("generating checksums in %q", b.root), t)

	// Sort the list for easier comparison - filepath.Walk should already do
	// this, but I can't stand tests that stop passing just because of sorting
	sort.Slice(b.Checksums, func(i, j int) bool {
		return b.Checksums[i].Path < b.Checksums[j].Path
	})

	var expectedChecksums = []string{
		"60fa80b948a0acc557a6ba7523f4040a7b452736723df20f118d0aacb5c1901b", // another.txt's "sha256sum" value
		"55f8718109829bf506b09d8af615b9f107a266e19f7a311039d1035f180b22d4", // test.txt's "sha256sum" value
	}

	assert.Equal(len(expectedChecksums), len(b.Checksums), "checksum list length", t)

	for i, ck := range b.Checksums {
		assert.Equal(expectedChecksums[i], ck.Checksum, "checksum for "+ck.Path, t)
	}
}
