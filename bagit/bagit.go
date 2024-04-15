package bagit

import (
	"fmt"
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
	root              string
	Hasher            *Hasher
	Cache             Cacher
	ActualChecksums   []*FileChecksum // Checksums for everything in data/
	ActualTagSums     []*FileChecksum // Checksums for all tag files
	ManifestChecksums []*FileChecksum // Parsed checksum data from manifest-*.txt
	ManifestTagSums   []*FileChecksum // Parsed checksum data from tagmanifest-*.txt
}

// New returns Bag structure for processing the given root path, and sets the
// hasher to the built-in SHA256
func New(root string) *Bag {
	return &Bag{
		root:   root,
		Hasher: HashSHA256,
		Cache:  noopCache{},
	}
}

func readSums(fname string) ([]*FileChecksum, error) {
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
// "tagmanifest-[hashtype].txt". Data is stored in the ManifestChecksums and
// ManifestTagSums fields, respectively. It does *not* generate or validate
// files in the bag.
//
// If an error occurs, it will be returned, and the bag's data may be in an
// incomplete state and should not be relied upon.
//
// Like the Generate... functions, ReadManifests will sort checksum data by
// filepath, allowing for predictable manual comparisons if necessary.
func (b *Bag) ReadManifests() error {
	var err error
	b.ManifestChecksums = nil
	b.ManifestTagSums = nil

	// Manifest file must exist, so all errors are fatal
	b.ManifestChecksums, err = readSums(b.manifestFilename())
	if err != nil {
		return fmt.Errorf("unable to read manifest file %q: %w", b.manifestFilename(), err)
	}

	// Tag manifest is optional, so we handle the nonexistence separately from other errors
	b.ManifestTagSums, err = readSums(b.tagManifestFilename())
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
// each file's checksum in turn, storing the FileChecksums in
// b.ActualChecksums, sorted by file path. The checksum path is always relative
// to the bag's root, which means it should always start with "data/".
//
// If there are any errors, relevant error information is returned. b.ActualChecksums
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
		return fmt.Errorf(`%q is not a bag: missing or invalid "data" directory`, b.root)
	}

	b.ActualChecksums = nil
	err = filepath.Walk(dataPath, func(path string, info os.FileInfo, err error) error {
		// Don't try to proceed if there's already an error!
		if err != nil {
			return err
		}

		if info.Mode().IsRegular() {
			var chksum, err = b.getsum(path)
			if err == nil {
				b.ActualChecksums = append(b.ActualChecksums, chksum)
			}
			return err
		}

		return nil
	})

	sort.Slice(b.ActualChecksums, func(i, j int) bool {
		return b.ActualChecksums[i].Path < b.ActualChecksums[j].Path
	})

	return err
}

func (b *Bag) getsum(path string) (*FileChecksum, error) {
	var relPath, err = filepath.Rel(b.root, path)
	if err != nil {
		return nil, fmt.Errorf("cannot parse %q's relative file path: %s", path, err)
	}

	var sum, exists = b.Cache.GetSum(relPath)
	if !exists {
		sum, err = b.compute(path)
		if err != nil {
			return nil, err
		}
	}
	b.Cache.SetSum(path, sum)

	return &FileChecksum{Path: relPath, Checksum: sum}, nil
}

func (b *Bag) compute(path string) (string, error) {
	var f, err = os.Open(path)
	if err != nil {
		return "", fmt.Errorf("cannot open %q: %s", path, err)
	}
	defer f.Close()

	var h = b.Hasher.Hash()
	_, err = io.Copy(h, f)
	if err != nil {
		return "", fmt.Errorf("cannot read %q for hashing: %s", path, err)
	}

	return fmt.Sprintf("%x", h.Sum(nil)), nil
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
	for _, ck := range b.ActualChecksums {
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
// b.ActualTagSums, sorted by file path. Files matching "tagmanifest-*.txt" are
// skipped as tag manifests themselves are not "tag" files.
//
// If there are any errors, relevant error information is returned.
// b.ActualTagSums may be incomplete or incorrect in these cases, and should
// not be used.
//
// This is typically used internally to generate the tag manifest file, but can
// be useful for testing or tag file validation.
func (b *Bag) GenerateTagSums() error {
	var infos, err = ioutil.ReadDir(b.root)
	if err != nil {
		return fmt.Errorf("error reading bag root: %s", err)
	}

	b.ActualTagSums = nil
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
		b.ActualTagSums = append(b.ActualTagSums, chksum)
	}

	sort.Slice(b.ActualTagSums, func(i, j int) bool {
		return b.ActualTagSums[i].Path < b.ActualTagSums[j].Path
	})

	return nil
}

func (b *Bag) writeTagManifest() error {
	var manifestFile = b.tagManifestFilename()
	if !fileutil.MustNotExist(manifestFile) {
		return fmt.Errorf("tag manifest file %q must not exist", manifestFile)
	}

	var f = fileutil.NewSafeFile(manifestFile)
	for _, ck := range b.ActualTagSums {
		fmt.Fprintf(f, "%s  %s\n", ck.Checksum, ck.Path)
	}

	var err = f.Close()
	if err != nil {
		return fmt.Errorf("error writing tag manifest file: %s", err)
	}

	return nil
}

// Validate reads all manifest files (standard manifest plus the optional tag
// manifest), generates fresh checksums, and compares what the manifest claims
// we should have to what's actually on disk. The return will contain any
// discrepancies in a human-readable format.
//
// If something fails, as opposed to there being incorrect data or manifests,
// an error will be returned and discrepancies will be empty. This can happen
// if there are no manifest files, if there is no "data" directory, if files
// are unreadable, etc.
//
// If a tag manifest is present, it is validated first, and the rest of the bag
// *is not validated* if the tag manifest has discrepancies. This avoids
// unnecessary work when there are easily-identified top-level bag problems.
func (b *Bag) Validate() (discrepancies []string, err error) {
	err = b.ReadManifests()
	if err != nil {
		return nil, err
	}

	if len(b.ManifestChecksums) == 0 {
		return nil, fmt.Errorf("%s contains no data", b.manifestFilename())
	}

	if len(b.ManifestTagSums) > 0 {
		err = b.GenerateTagSums()
		if err != nil {
			return nil, err
		}

		discrepancies = Compare("tag manifest", b.ManifestTagSums, b.ActualTagSums)
		if len(discrepancies) > 0 {
			return discrepancies, nil
		}
	}

	err = b.GenerateChecksums()
	if err != nil {
		return nil, err
	}
	discrepancies = Compare("manifest", b.ManifestChecksums, b.ActualChecksums)

	return discrepancies, nil
}

// Compare validates a manifest file's checksums against the actual checksums
// from the filesystem. The return will contain any discrepancies in a
// human-readable format.
//
// This is normally not meant to be used externally, but it can be handy to
// validate two separate bags or do a simple compare just of tag manifests or
// something.
//
// manifestType should describe the kind of manifest - internally we only use
// "tag manifest" and "manifest". This string is used when reporting
// discrepancies, e.g.: "tag manifest lists the file, but blah blah blah".
func Compare(manifestType string, manifest, actual []*FileChecksum) []string {
	var manifestMap = mapify(manifest)
	var actualMap = mapify(actual)
	var errs []string

	// Step 1: everything in the manifest should have a corresponding (and equal)
	// item in the generated list
	for path := range manifestMap {
		var mchk, achk = manifestMap[path], actualMap[path]
		if achk == "" {
			errs = append(errs, fmt.Sprintf("missing file: %q (%s lists the file, but it is not present on disk)", path, manifestType))
		} else if achk != mchk {
			errs = append(errs, fmt.Sprintf("corrupt file: %q (%s checksum was %q, actual checksum was %q", path, manifestType, mchk, achk))
		}
	}

	// Step 2: check for anything on the filesystem that wasn't in the manifest
	// at all; we do not re-validate items that are in both lists here
	for path := range actualMap {
		if manifestMap[path] == "" {
			errs = append(errs, fmt.Sprintf("extra file: %q (%s does not list the file, but it is present on disk)", path, manifestType))
		}
	}

	return errs
}

// mapify turns a checksum slice into a map of path-to-checksum data to allow
// for easier comparing of two checksum lists
func mapify(src []*FileChecksum) map[string]string {
	var m = make(map[string]string, len(src))
	for _, chksum := range src {
		m[chksum.Path] = chksum.Checksum
	}

	return m
}
