package manifest

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"time"
)

// Filename is the name used to store the Manifest JSON representation
const Filename = ".manifest"

// FileInfo represents basic information for a single file within a Manifest
type FileInfo struct {
	Name    string
	Size    int64
	Mode    fs.FileMode
	ModTime time.Time
}

// Equal returns true if all fields are *equivalent*. This means normal
// equality checks for all but time, which needs to use time.Equal to handle
// monotonic clocks and potentially different time zones.
func (fi FileInfo) Equal(b FileInfo) bool {
	return fi.Name == b.Name && fi.Size == b.Size && fi.Mode == b.Mode && fi.ModTime.Equal(b.ModTime)
}

func newFileInfo(loc string, e os.DirEntry) (FileInfo, error) {
	var fullpath = filepath.Join(loc, e.Name())
	var fd = FileInfo{Name: e.Name()}
	var info, err = e.Info()
	if err != nil {
		return fd, fmt.Errorf("reading info for %q: %w", fullpath, err)
	}

	fd.Size = info.Size()
	fd.Mode = info.Mode()
	fd.ModTime = info.ModTime()

	return fd, nil
}

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
	path    string
	Created time.Time
	Files   []FileInfo
}

// New returns a Manifest ready for scanning a directory or reading an
// existing manifest file.
func New(location string) *Manifest {
	return &Manifest{path: location, Created: time.Now()}
}

// Build reads all files in the manifest's path and builds our manifest data
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

		var fd, err = newFileInfo(m.path, entry)
		if err != nil {
			return fmt.Errorf("reading dir %q: %w", m.path, err)
		}
		m.Files = append(m.Files, fd)
	}
	return nil
}

// Read replaces all file-level metadata with whatever is in the existing
// manifest file, if anything
func (m *Manifest) Read() error {
	var data, err = ioutil.ReadFile(m.filename())
	if err != nil {
		return err
	}

	err = json.Unmarshal(data, &m)
	if err != nil {
		return err
	}

	return nil
}

func (m *Manifest) filename() string {
	return filepath.Join(m.path, Filename)
}

// Write creates or replaces the manifest file with the current file metadata
func (m *Manifest) Write() error {
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
