package fileutil

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

// SyncDirectory syncs files from srcPath to dstPath, copying any which are
// missing or different.  Files are different if they're a different size or
// checksum (MD5).  Notes:
// - Anything that isn't a file or a directory returns an error; this includes symlinks for now
// - The operation stops on the first error, and the partial copy is left in place
// - Basic permissions (file mode) will by preserved, though owner, group, ACLs, and other metadata will not
// - Files in dstPath which are not in srcPath will not be removed
func SyncDirectory(srcPath, dstPath string) error {
	var err error

	// Figure out absolute paths for clarity
	srcPath, err = filepath.Abs(srcPath)
	if err != nil {
		return fmt.Errorf("source %q error: %s", srcPath, err)
	}
	dstPath, err = filepath.Abs(dstPath)
	if err != nil {
		return fmt.Errorf("destination %q error: %s", dstPath, err)
	}

	// Validate source exists
	if !Exists(srcPath) {
		return fmt.Errorf("source %q does not exist", srcPath)
	}

	// Destination parent must already exist
	if !IsDir(filepath.Dir(dstPath)) {
		return fmt.Errorf("destination's parent %q does not exist", dstPath)
	}

	// Get source path info and validate it's a directory
	var srcInfo os.FileInfo
	srcInfo, err = os.Stat(srcPath)
	if err != nil {
		return fmt.Errorf("source %q error: %s", srcPath, err)
	}
	if !srcInfo.IsDir() {
		return fmt.Errorf("source %q is not a directory", srcPath)
	}

	return syncRecursive(srcPath, dstPath)
}

// syncRecursive is the actual file-syncing function which SyncDirectory uses
func syncRecursive(srcPath, dstPath string) error {
	if strings.HasPrefix(dstPath, srcPath) {
		return fmt.Errorf("cannot have destination under source")
	}

	var dirInfo, err = os.Stat(srcPath)
	if err != nil {
		return fmt.Errorf("unable to stat source directory %q: %s", srcPath, err)
	}
	var mode = dirInfo.Mode() & os.ModePerm

	err = os.MkdirAll(dstPath, 0700)
	if err != nil {
		return fmt.Errorf("unable to create directory %q: %s", dstPath, err)
	}
	os.Chmod(dstPath, mode)

	var infos []os.FileInfo
	infos, err = ioutil.ReadDir(srcPath)
	if err != nil {
		return fmt.Errorf("unable to read source directory %q: %s", srcPath, err)
	}

	for _, info := range infos {
		var srcFull = filepath.Join(srcPath, info.Name())
		var dstFull = filepath.Join(dstPath, info.Name())

		var file = InfoToFile(info)
		switch {
		case file.IsDir():
			err = syncRecursive(srcFull, dstFull)
			if err != nil {
				return err
			}

		case file.IsRegular():
			err = syncFile(srcFull, dstFull)
			if err != nil {
				return err
			}
			os.Chmod(dstFull, info.Mode()&os.ModePerm)

		default:
			return fmt.Errorf("unable to copy special file %q", srcFull)
		}
	}

	return nil
}

// syncFile checks the two files to see if they differ, and copies src to dest
// if so.  Files are considered different if (a) dst doesn't exist, (b) dst
// isn't the same size as src, or (c) dst doesn't have the same SHA256 sum as
// src.
func syncFile(src, dst string) error {
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
