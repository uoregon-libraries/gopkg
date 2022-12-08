package bagit

import (
	"fmt"
	"hash"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
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
	TagSums   []*FileChecksum
}

// New returns Bag structure for processing the given root path, and sets the
// hasher to the built-in SHA256
func New(root string) *Bag {
	return &Bag{root: root, Hasher: HashSHA256}
}

func parseChecksums(fname string) ([]*FileChecksum, error) {
	var data, err = ioutil.ReadFile(fname)
	if err != nil {
		return nil, err
	}

	var sums []*FileChecksum
	for _, line := range strings.Split(string(data), "\n") {
		// Blank lines are allowed, but skipped
		if line == "" {
			continue
		}

		var parts = strings.Fields(line)
		if len(parts) != 2 {
			return nil, fmt.Errorf("invalid manifest line in %q: %q", fname, line)
		}
		sums = append(sums, &FileChecksum{Checksum: parts[0], Path: parts[1]})
	}

	sort.Slice(sums, func(i, j int) bool {
		return sums[i].Path < sums[j].Path
	})

	return sums, nil
}

// ReadManifests loads "manifest-[hashtype].txt" and, if present,
// "tagmanifest-[hashtype].txt". Data is stored in the Checksums and TagSums
// fields, respectively. It does *not* generate or validate files in the bag.
//
// If an error occurs, it will be returned, and the bag's data may be in an
// incomplete state and should not be relied upon.
//
// Like the Generate... functions, ReadManifests will sort checksum data by
// filepath, allowing for easier comparisons.
func (b *Bag) ReadManifests() error {
	var err error
	b.Checksums = nil
	b.TagSums = nil

	// Manifest file must exist, so all errors are fatal
	b.Checksums, err = parseChecksums(b.manifestFilename())
	if err != nil {
		return fmt.Errorf("unable to read manifest file %q: %w", b.manifestFilename(), err)
	}

	// Tag manifest is optional, so we handle the nonexistence separately from other errors
	b.TagSums, err = parseChecksums(b.tagManifestFilename())
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("unable to read manifest file %q: %w", b.tagManifestFilename(), err)
	}

	return nil
}

// WriteTagFiles traverses all files under the bag's root/data, generates
// hashes for each, and writes out "manifest-[hashtype].txt". Upon completion,
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
		err = b.GenerateTagSums()
	}
	if err == nil {
		err = b.writeTagManifest()
	}

	return
}

// GenerateChecksums iterates over all files in the data path and generates
// each file's checksum in turn, storing the FileChecksums in b.Checksums,
// sorted by file path. The checksum path is always relative to the bag's root.
//
// If there are any errors, relevant error information is returned. b.Checksums
// may be incomplete or incorrect in these cases, and should not be used.
//
// This is typically used internally to generate the manifest file, but can be
// useful for testing, bag validation, or making use of the BagIt data
// structure in cases where checksums need to be stored externally to the data.
func (b *Bag) GenerateChecksums() error {
	var err error
	var realroot string

	realroot, err = filepath.Abs(b.root)
	if err != nil {
		return fmt.Errorf("unable to determine bag's absolute root path from %q: %s", b.root, err)
	}
	b.root = realroot

	var dataPath = filepath.Join(b.root, "data")
	if !fileutil.IsDir(dataPath) {
		return fmt.Errorf("%q is not a directory", dataPath)
	}

	b.Checksums = nil
	err = filepath.Walk(dataPath, func(path string, info os.FileInfo, err error) error {
		// Don't try to proceed if there's already an error!
		if err != nil {
			return err
		}

		if info.Mode().IsRegular() {
			var chksum, err = b.getsum(path)
			if err == nil {
				b.Checksums = append(b.Checksums, chksum)
			}
			return err
		}

		return nil
	})

	sort.Slice(b.Checksums, func(i, j int) bool {
		return b.Checksums[i].Path < b.Checksums[j].Path
	})

	return err
}

func (b *Bag) getsum(path string) (*FileChecksum, error) {
	var err error
	var f *os.File
	var relPath, hexSum string
	var h hash.Hash

	f, err = os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("cannot open %q: %s", path, err)
	}
	defer f.Close()

	h = b.Hasher.Hash()
	_, err = io.Copy(h, f)
	if err != nil {
		return nil, fmt.Errorf("cannot read %q for hashing: %s", path, err)
	}

	relPath, err = filepath.Rel(b.root, path)
	if err != nil {
		return nil, fmt.Errorf("cannot parse %q's relative file path: %s", path, err)
	}

	hexSum = fmt.Sprintf("%x", h.Sum(nil))
	return &FileChecksum{Path: relPath, Checksum: hexSum}, nil
}

func (b *Bag) manifestFilename() string {
	return filepath.Join(b.root, "manifest-"+b.Hasher.Name+".txt")
}

func (b *Bag) tagManifestFilename() string {
	return filepath.Join(b.root, "tagmanifest-"+b.Hasher.Name+".txt")
}

func (b *Bag) writeManifest() error {
	var manifestFile = b.manifestFilename()
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

func (b *Bag) writeBagitFile() error {
	var f = fileutil.NewSafeFile(filepath.Join(b.root, "bagit.txt"))
	f.Write([]byte("BagIt-Version: 0.97\nTag-File-Character-Encoding: UTF-8\n"))
	return f.Close()
}

// GenerateTagSums iterates over all "tag" files (top-level files, not files in
// data/) and generates each file's checksum in turn, storing them in
// b.TagSums, sorted by file path. Files matching "tagmanifest-*.txt" are
// skipped as tag manifests themselves are not "tag" files.
//
// If there are any errors, relevant error information is returned. b.TagSums
// may be incomplete or incorrect in these cases, and should not be used.
//
// This is typically used internally to generate the tag manifest file, but can
// be useful for testing or tag file validation.
func (b *Bag) GenerateTagSums() error {
	var infos, err = ioutil.ReadDir(b.root)
	if err != nil {
		return fmt.Errorf("error reading bag root: %s", err)
	}

	b.TagSums = nil
	for _, info := range infos {
		if !info.Mode().IsRegular() {
			continue
		}

		var path = filepath.Join(b.root, info.Name())
		// Explicitly ignore the error here - if this pattern is broken, the caller
		// has no way to fix it in any case. Better to just keep moving on.
		var match, _ = filepath.Match("tagmanifest-*.txt", info.Name())
		if match {
			continue
		}

		var chksum, err = b.getsum(path)
		if err != nil {
			return fmt.Errorf("error getting %q's checksum: %s", path, err)
		}
		b.TagSums = append(b.TagSums, chksum)
	}

	sort.Slice(b.TagSums, func(i, j int) bool {
		return b.TagSums[i].Path < b.TagSums[j].Path
	})

	return nil
}

func (b *Bag) writeTagManifest() error {
	var manifestFile = b.tagManifestFilename()
	if !fileutil.MustNotExist(manifestFile) {
		return fmt.Errorf("tag manifest file %q must not exist", manifestFile)
	}

	var f = fileutil.NewSafeFile(manifestFile)
	for _, ck := range b.TagSums {
		fmt.Fprintf(f, "%s  %s\n", ck.Checksum, ck.Path)
	}

	var err = f.Close()
	if err != nil {
		return fmt.Errorf("error writing tag manifest file: %s", err)
	}

	return nil
}
