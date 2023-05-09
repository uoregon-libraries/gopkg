package fileutil

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// SyncDirectory syncs files from srcPath to dstPath, copying any which are
// missing or different.  Files are different if they're a different size or
// checksum (MD5).  Notes:
// - Anything that isn't a file or a directory returns an error; this includes symlinks for now
// - The operation stops on the first error, and the partial copy is left in place
// - Basic permissions (file mode) will by preserved, though owner, group, ACLs, and other metadata will not
// - Files in dstPath which are not in srcPath will not be removed
func SyncDirectory(srcPath, dstPath string) error {
	return SyncDirectoryExcluding(srcPath, dstPath, nil)
}

// SyncDirectoryExcluding syncs files from srcPath to dstPath excluding files
// which match any of the given patterns. Other than the exclusions, this is
// precisely the same as SyncDirectory.
func SyncDirectoryExcluding(srcPath, dstPath string, exclusionPatterns []string) error {
	var err error

	srcPath, dstPath, err = getAbsPaths(srcPath, dstPath)
	if err != nil {
		return err
	}

	err = validateCopyDirs(srcPath, dstPath, false)
	if err != nil {
		return err
	}

	var tolerance = 5
	var copyFn = func(src, dst string) error {
		return syncFileTolerant(src, dst, tolerance)
	}

	if len(exclusionPatterns) > 0 {
		copyFn = func(src, dst string) error {
			for _, pattern := range exclusionPatterns {
				var basename = filepath.Base(src)
				var match, err = filepath.Match(pattern, basename)
				if err != nil {
					return fmt.Errorf("SyncDirectoryExcluding: invalid pattern %q", pattern)
				}
				if match {
					return nil
				}
			}
			return syncFileTolerant(src, dst, tolerance)
		}
	}

	return copyRecursive(srcPath, dstPath, copyFn)
}

// syncFileTolerant calls syncFile up to n times, returning the last error if
// retries are exceeded. Each error forces a short delay to allow for very
// brief I/O "hiccups".
func syncFileTolerant(src, dst string, tolerance int) error {
	var err error
	for i := 0; i < tolerance; i++ {
		err = syncOneFile(src, dst)
		if err == nil {
			return nil
		}
		var delay = 1 << i
		if delay > 60 {
			delay = 60
		}
		time.Sleep(time.Second * time.Duration(delay))
	}

	return err
}

// syncOneFile checks the two files to see if they differ, and copies src to dest
// if so.  Files are considered different if (a) dst doesn't exist, (b) dst
// isn't the same size as src, or (c) dst doesn't have the same SHA256 sum as
// src.
func syncOneFile(src, dst string) error {
	var doCopy, err = needSync(src, dst)
	if err != nil {
		return err
	}

	if doCopy {
		return CopyVerify(src, dst)
	}
	return nil
}

// needSync determines if src and dst are different
func needSync(src, dst string) (bool, error) {
	// Easiest case: dst doesn't exist, so we just copy it
	if MustNotExist(dst) {
		return true, nil
	}

	// Case 2: files differ by size
	var err error
	var si, di os.FileInfo
	si, err = os.Stat(src)
	if err != nil {
		return false, err
	}
	di, err = os.Stat(dst)
	if err != nil {
		return false, err
	}
	if si.Size() != di.Size() {
		return true, nil
	}

	// Case 3: files are the same size, so we do a full SHA256 of both files to
	// be 100% certain they're the same.  Slow, but safe.
	var sumSrc, sumDst []byte
	sumSrc, err = SHA256(src)
	if err != nil {
		return false, err
	}
	sumDst, err = SHA256(dst)
	if err != nil {
		return false, err
	}

	if bytes.Equal(sumSrc, sumDst) {
		return false, nil
	}

	return true, nil
}
