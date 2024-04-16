package manifest

import (
	"fmt"
	"hash"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"time"
)

// FileInfo represents basic information for a single file within a Manifest
type FileInfo struct {
	Name    string
	Sum     string
	Size    int64
	Mode    fs.FileMode
	ModTime time.Time
}

// Equal returns true if all fields are *equivalent*. This means normal
// equality checks for all but time, which needs to use time.Equal to handle
// monotonic clocks and potentially different time zones.
func (fi FileInfo) Equal(b FileInfo) bool {
	if fi.Name != b.Name {
		return false
	}
	if fi.Size != b.Size {
		return false
	}
	if fi.Mode != b.Mode {
		return false
	}
	if !fi.ModTime.Equal(b.ModTime) {
		return false
	}
	if fi.Sum != b.Sum && fi.Sum != "" && b.Sum != "" {
		return false
	}

	return true
}

func newFileInfo(loc string, e os.DirEntry, hsh hash.Hash) (FileInfo, error) {
	var fullpath = filepath.Join(loc, e.Name())
	var fd = FileInfo{Name: e.Name()}
	var info, err = e.Info()
	if err != nil {
		return fd, fmt.Errorf("reading info for %q: %w", fullpath, err)
	}
	if hsh != nil {
		var f, err = os.Open(fullpath)
		if err != nil {
			return fd, fmt.Errorf("opening %q: %w", fullpath, err)
		}
		defer f.Close()

		hsh.Reset()
		_, err = io.Copy(hsh, f)
		if err != nil {
			return fd, fmt.Errorf("hashing %q: %w", fullpath, err)
		}
		fd.Sum = fmt.Sprintf("%x", hsh.Sum(nil))
	}

	fd.Size = info.Size()
	fd.Mode = info.Mode()
	fd.ModTime = info.ModTime()

	return fd, nil
}
