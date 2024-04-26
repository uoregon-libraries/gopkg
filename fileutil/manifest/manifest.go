package manifest

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"time"

	"github.com/uoregon-libraries/gopkg/hasher"
)

// Filename is the name used to store the Manifest JSON representation
const Filename = ".manifest"

// A Manifest is a somewhat special-case representation of a filesystem
// directory's state. It only works with very simple directories: no subdirs,
// no special files, etc. Hidden files are ignored from the Manifest by design
// to allow for the common cases without getting problems from things like a
// .gitignore file for instance.
//
// The data stored can be useful to determine if a directory changes
// purposefully: filesize, file modes (permissions) and file's modification
// times. Additionally, the manifest stores its own creation time in order to
// effectively know when a directory was first seen, even if the files are all
// very old (this can happen when moving a directory).
type Manifest struct {
	path     string
	Created  time.Time
	Files    []FileInfo
	HashAlgo string
	Hasher   *hasher.Hasher `json:"-"`
}

// New returns a Manifest ready for scanning a directory or reading an existing
// manifest file. This should generally not be needed: Build and Open are
// easier for typical use-cases.
func New(location string) *Manifest {
	return &Manifest{path: location, Created: time.Now()}
}

// Build reads files in the given location, builds a Manifest, and returns it
// (or nil and an error)
func Build(location string) (*Manifest, error) {
	return BuildHashed(location, nil)
}

// BuildHashed tries to build a manifest for the given location, hashing files
// using the given hasher.Hasher (e.g., hasher.NewMD5())
func BuildHashed(location string, h *hasher.Hasher) (*Manifest, error) {
	var m = New(location)
	m.Hasher = h
	if h != nil {
		m.HashAlgo = h.Name
	}
	var err = m.Build()
	return m, err
}

// Open looks for a manifest file in the given location, and returns a Manifest
// or an error (e.g., no manifest file existed)
func Open(location string) (*Manifest, error) {
	var m = &Manifest{path: location}
	var data, err = ioutil.ReadFile(m.filename())
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(data, &m)
	if err != nil {
		return nil, err
	}

	var h = hasher.FromString(m.HashAlgo)
	if m.HashAlgo != "" && h == nil {
		return nil, fmt.Errorf("reading %q: invalid hash algorithm (%q)", m.filename(), m.HashAlgo)
	}
	m.Hasher = h

	return m, nil
}

// Build reads all files in the manifest's path and builds our manifest data.
func (m *Manifest) Build() error {
	var entries, err = os.ReadDir(m.path)
	if err != nil {
		return fmt.Errorf("reading dir %q: %w", m.path, err)
	}

	for _, entry := range entries {
		if !entry.Type().IsRegular() {
			return fmt.Errorf("reading dir %q: one or more entries are not a regular file", m.path)
		}

		// Skip the manifest as well as any hidden files - we explicitly check for
		// the manifest in case we change the constant string to not be hidden for
		// some reason.
		if entry.Name()[0] == '.' || entry.Name() == Filename {
			continue
		}

		var fd, err = newFileInfo(m.path, entry, m.Hasher)
		if err != nil {
			return fmt.Errorf("reading dir %q: %w", m.path, err)
		}
		m.Files = append(m.Files, fd)
	}
	return nil
}

func (m *Manifest) filename() string {
	return filepath.Join(m.path, Filename)
}

// Write creates or replaces the manifest file with the current file metadata
func (m *Manifest) Write() error {
	// Ensure HashAlgo is set to the right value
	m.HashAlgo = ""
	if m.Hasher != nil {
		m.HashAlgo = m.Hasher.Name
	}

	var data, err = json.Marshal(m)
	if err != nil {
		return err
	}
	return ioutil.WriteFile(m.filename(), data, 0600)
}

func (m *Manifest) sortFiles() {
	sort.Slice(m.Files, func(i, j int) bool {
		return m.Files[i].Name < m.Files[j].Name
	})
}

// Validate returns true if the current manifest matches what's actually in the
// directory. Behind the scenes this just builds a new manifest with the same
// path and hashing algorithm as m.
//
// This can return an error for the same reasons Build can: particularly if the
// path is not valid or there are non-file directory entries in the path.
func (m *Manifest) Validate() (bool, error) {
	var m2, err = BuildHashed(m.path, hasher.FromString(m.HashAlgo))
	if err != nil {
		return false, err
	}
	return m.Equiv(m2), nil
}

// Equiv returns true if m and m2 have the *exact* same file lists.
// Struct requires manual comparison as ModTime values must use Equal
// to handle monotonic clock values. (Ref: https://pkg.go.dev/time)
func (m *Manifest) Equiv(m2 *Manifest) bool {
	if len(m.Files) != len(m2.Files) {
		return false
	}
	m.sortFiles()
	m2.sortFiles()

	for i := range m.Files {
		var f1, f2 = m.Files[i], m2.Files[i]
		if !f1.Equal(f2) {
			return false
		}
	}

	return true
}
