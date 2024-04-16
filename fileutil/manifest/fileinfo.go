package manifest

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"time"
)

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

