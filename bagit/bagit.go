package bagit

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/uoregon-libraries/gopkg/fileutil"
)

// FileChecksum holds a path to a file and its checksum
type FileChecksum struct {
	Path     string
	Checksum string
}

// Bag holds state for the generation of bag manifest and other tag files
type Bag struct {
	root      string
	Hasher    *Hasher
	Checksums []*FileChecksum
}

// New returns Bag structure for processing the given root path, and sets the
// hasher to the built-in SHA256
func New(root string) *Bag {
	return &Bag{root: root, Hasher: HashSHA256}
}

// WriteTagFiles traverses all files under the bag's root/data, generates
// hashes for each, and writes out "manifest-[hashtype].txt".  Upon completion,
// bagit.txt and tagmanifest-[hashtype].txt are then written.
//
// This is not parallelized as it seems unlikely any advantage would be gained
// since file IO is likely to be the main cost, not CPU.
func (b *Bag) WriteTagFiles() (err error) {
	err = b.GenerateChecksums()
	if err == nil {
		err = b.writeManifest()
	}
	if err == nil {
		err = b.writeBagitFile()
	}
	if err == nil {
		err = b.writeTagManifest()
	}

	return
}

// GenerateChecksums iterates over all files in the data path and generates
// each file's checksum in turn, returning the resulting slice of
// FileChecksums.  The checksum path is always relative to the bag's root.
//
// If there are any errors, GenerateChecksums returns an empty slice along with
// relevant error information.
//
// This is typically used internally to generate the manifest file, but can be
// useful for testing, bag validation, or making use of the BagIt data
// structure in cases where checksums need to be stored externally to the data.
func (b *Bag) GenerateChecksums() error {
	b.Checksums = nil
	var dataPath = filepath.Join(b.root, "data")
	if !fileutil.IsDir(dataPath) {
		return fmt.Errorf("%q is not a directory", dataPath)
	}

	var err = filepath.Walk(dataPath, func(path string, info os.FileInfo, err error) error {
		if info.Mode().IsRegular() {
			var f, err = os.Open(path)
			if err != nil {
				return fmt.Errorf("cannot open %q: %s", path, err)
			}
			defer f.Close()

			var hash = b.Hasher.Hash()
			_, err = io.Copy(hash, f)
			if err != nil {
				return fmt.Errorf("cannot read %q for hashing: %s", path, err)
			}

			b.Checksums = append(b.Checksums, &FileChecksum{
				Path:     strings.Replace(path, b.root+"/", "", 1),
				Checksum: fmt.Sprintf("%x", hash.Sum(nil)),
			})
		}

		return nil
	})

	return err
}

func (b *Bag) writeManifest() error {
	var manifestFile = filepath.Join(b.root, "manifest-"+b.Hasher.Name+".txt")
	if !fileutil.MustNotExist(manifestFile) {
		return fmt.Errorf("manifest file %q must not exist", manifestFile)
	}

	var f = fileutil.NewSafeFile(manifestFile)
	for _, ck := range b.Checksums {
		fmt.Fprintf(f, "%s  %s\n", ck.Checksum, ck.Path)
	}

	var err = f.Close()
	if err != nil {
		return fmt.Errorf("error writing manifest file: %s", err)
	}

	return nil
}

func (b *Bag) writeTagManifest() error {
	return fmt.Errorf("not implemented")
}

func (b *Bag) writeBagitFile() error {
	return fmt.Errorf("not implemented")
}
