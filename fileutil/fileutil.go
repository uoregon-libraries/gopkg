// Package fileutil holds various general utilities for working with the
// filesystem more easily
package fileutil

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
)

// IsDir returns true if the given path exists and is a directory
func IsDir(path string) bool {
	info, err := os.Stat(path)
	if err != nil && os.IsNotExist(err) {
		return false
	}

	// This means something weird happened that we probably want to report (often
	// a permissions issue), but the function's purpose is simplicity, so we
	// consider this a non-file.
	if err != nil {
		return false
	}

	return info.IsDir()
}

// IsFile returns true if the given path exists and is a regular file
func IsFile(path string) bool {
	info, err := os.Stat(path)
	if err != nil && os.IsNotExist(err) {
		return false
	}

	// This means something weird happened that we probably want to report (often
	// a permissions issue), but the function's purpose is simplicity, so we
	// consider this a non-file.
	if err != nil {
		return false
	}

	return info.Mode().IsRegular()
}

// Exists returns true if the given path exists and has no errors.  All errors
// are treated as the path not existing in order to avoid trying to determine
// what to do to handle the unknown errors which may be returned.
func Exists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// MustNotExist is used when we need to be absolutely certain a path doesn't
// exist, such as when a directory's existence means a duplicate operation
// occurred.
func MustNotExist(path string) bool {
	_, err := os.Stat(path)
	return err != nil && os.IsNotExist(err)
}

// ReaddirSorted calls ioutil.ReadDir and sorts the results
func ReaddirSorted(path string) ([]os.FileInfo, error) {
	var fi, err = ioutil.ReadDir(path)
	if err == nil {
		sort.Sort(byName(fi))
	}

	return fi, err
}

// ReaddirSortedNumeric returns the results of ioutil.ReadDir sorted in a
// "human-friendly" way such that, e.g., 1.pdf is followed by 2.pdf, etc., and
// then later on 10.pdf.  Similar to `sort -n`.
func ReaddirSortedNumeric(path string) ([]os.FileInfo, error) {
	var list, err = ioutil.ReadDir(path)
	if err != nil {
		return list, err
	}

	sortFileInfosNumerically(list)

	return list, err
}

// byName implements sort.Interface for sorting os.FileInfo data by name
type byName []os.FileInfo

func (n byName) Len() int           { return len(n) }
func (n byName) Swap(i, j int)      { n[i], n[j] = n[j], n[i] }
func (n byName) Less(i, j int) bool { return n[i].Name() < n[j].Name() }

// SortFileInfos sorts a slice of os.FileInfo data by the underlying filename
func SortFileInfos(list []os.FileInfo) {
	sort.Sort(byName(list))
}

// sortFileInfosNumerically sorts a slice of os.FileInfo data by the underlying filename
func sortFileInfosNumerically(list []os.FileInfo) {
	sort.Slice(list, numericInfoSortFn(list))
}

// numberify stripts preceding zeros and everything after the first non-numeric
// byte in order to make a string into a valid int
//
// This is basically a stripped-down Atoi that has no error cases and allows
// things like "002312dafdsa.pdf" to return 2312.
func numberify(s string) int {
	s0 := s
	if s[0] == '-' || s[0] == '+' {
		s = s[1:]
		if len(s) < 1 {
			return 0
		}
	}

	// This excludes many valid ints, but it's easier this way
	if len(s) > 9 {
		return 0
	}

	n := 0
	for _, ch := range []byte(s) {
		ch -= '0'
		if ch > 9 {
			break
		}
		n = n*10 + int(ch)
	}
	if s0[0] == '-' {
		n = -n
	}
	return n
}

func numericInfoSortFn(infos []os.FileInfo) func(i, j int) bool {
	return func(i, j int) bool {
		var iName = infos[i].Name()
		var jName = infos[j].Name()
		var iVal = numberify(iName)
		var jVal = numberify(jName)

		if iVal == jVal || iVal == 0 || jVal == 0 {
			return iName < jName
		}
		return iVal < jVal
	}
}

// FindIf iterates over all directory entries in the given path, running the
// given selector on each, and returning a list of those for which the selector
// returned true.
//
// Symlinks are resolved to their real file for the selector function, but the
// path added to the return will be a path to the symlink, not its target.
//
// Filesystem errors, including permission errors, will cause FindIf to halt
// and return an empty list and the error.
func FindIf(path string, selector func(i os.FileInfo) bool) ([]string, error) {
	var results []string
	var items, err = ReaddirSorted(path)
	if err != nil {
		return nil, err
	}

	for _, i := range items {
		var fName = i.Name()
		var path = filepath.Join(path, fName)
		var realPath = path
		if i.Mode()&os.ModeSymlink != 0 {
			realPath, err = os.Readlink(path)
			if err != nil {
				return nil, err
			}
			// Symlinks kind of suck - they can be absolute or relative, and if
			// they're relative we have to make them absolute....
			if !filepath.IsAbs(realPath) {
				realPath = filepath.Join(path, realPath)
			}

			i, err = os.Stat(realPath)
			if err != nil {
				return nil, err
			}
		}
		realPath = filepath.Clean(realPath)

		// See if the selector allows this file to be put in the list
		if !selector(i) {
			continue
		}

		results = append(results, path)
	}

	return results, nil
}

// FindFiles returns a list of all entries in a given path which are *not*
// directories or symlinks to directories.  For the purpose of this function,
// we define "files" as "things from which we can directly read data".
func FindFiles(path string) ([]string, error) {
	return FindIf(path, func(i os.FileInfo) bool {
		return !i.IsDir()
	})
}

// FindDirectories returns a list of all directories or symlinks to directories
// within the given path
func FindDirectories(path string) ([]string, error) {
	return FindIf(path, func(i os.FileInfo) bool {
		return i.IsDir()
	})
}

// Find traverses the filesystem to the given depth, returning only the items
// that are found at that depth.  Traverses symlinks if they are directories.
// Returns the first error found if any occur.
func Find(root string, depth int) ([]string, error) {
	var paths = []string{root}
	var newPaths []string
	for depth > 0 {
		for _, p := range paths {
			var appendList, err = FindDirectories(p)
			if err != nil {
				return nil, err
			}
			newPaths = append(newPaths, appendList...)
		}
		paths = newPaths
		newPaths = nil
		depth--
	}

	return paths, nil
}
