package fileutil

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

// getAbsPaths figures out absolute paths for a recursive copy operation
func getAbsPaths(srcPath, dstPath string) (string, string, error) {
	var err error
	srcPath, err = filepath.Abs(srcPath)
	if err == nil {
		dstPath, err = filepath.Abs(dstPath)
	}
	return srcPath, dstPath, err
}

// validateCopyDirs centralizes the common logic of ensuring a source and
// destination path are at least semi-valid: the source exists and is a
// directory, that the destination's parent dir exists, and optionally that the
// destination does not exist.
func validateCopyDirs(srcPath, dstPath string, failOnDestinationExists bool) error {
	var err error

	// Validate source exists and destination does not
	if !Exists(srcPath) {
		return fmt.Errorf("source %q does not exist", srcPath)
	}
	if failOnDestinationExists && !MustNotExist(dstPath) {
		return fmt.Errorf("destination %q already exists", dstPath)
	}

	if strings.HasPrefix(dstPath, srcPath) {
		return fmt.Errorf("cannot have destination under source")
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

	return nil
}

// CopyDirectory attempts to copy all files from srcPath to dstPath
// recursively.  dstPath must not exist.  Anything that isn't a file or a
// directory returns an error.  This includes symlinks for now.  The operation
// stops on the first error, and the partial copy is left in place.
func CopyDirectory(srcPath, dstPath string) error {
	var err error

	srcPath, dstPath, err = getAbsPaths(srcPath, dstPath)
	if err != nil {
		return err
	}

	err = validateCopyDirs(srcPath, dstPath, true)
	if err != nil {
		return err
	}

	return copyRecursive(srcPath, dstPath, CopyVerify)
}

// LinkDirectory attempts to hard-link all files from srcPath to dstPath
// recursively.  dstPath must not exist.  Anything that isn't a file or a
// directory returns an error.  This includes symlinks for now.  The operation
// stops on the first error, and the partial copy is left in place.
func LinkDirectory(srcPath, dstPath string) error {
	var err error

	srcPath, dstPath, err = getAbsPaths(srcPath, dstPath)
	if err != nil {
		return err
	}

	err = validateCopyDirs(srcPath, dstPath, true)
	if err != nil {
		return err
	}

	return copyRecursive(srcPath, dstPath, os.Link)
}

// copyFunc takes a source and destination (absolute paths), does something to
// copy them (i.e., copy data, hard-link them, eventually maybe symlink), and
// returns any errors which occur.
type copyFunc func(string, string) error

// copyRecursive does the actual work of copying files, using a callback to
// allow custom copying behavior
func copyRecursive(srcPath, dstPath string, cpFunc copyFunc) error {
	var dirInfo, err = os.Stat(srcPath)
	if err != nil {
		return fmt.Errorf("unable to stat source directory %q: %s", srcPath, err)
	}
	var mode = dirInfo.Mode() & os.ModePerm

	err = os.MkdirAll(dstPath, mode)
	if err != nil {
		return fmt.Errorf("unable to create directory %q: %s", dstPath, err)
	}

	// If the dir wasn't created, make sure we still set its mode
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
			err = copyRecursive(srcFull, dstFull, cpFunc)
			if err != nil {
				return err
			}

		case file.IsRegular():
			err = cpFunc(srcFull, dstFull)
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

// CopyFile attempts to copy the bytes from src into dst, returning an error if
// applicable. Does not use [os.Link] regardless of where the two files reside,
// as that can cause massive confusion when copying a file in order to back it
// up while writing out to the original.  The destination file permissions
// aren't set here, and must be managed externally.
func CopyFile(src, dst string) error {
	var err error
	var srcInfo os.FileInfo

	srcInfo, err = os.Stat(src)
	if err != nil {
		return fmt.Errorf("cannot stat %#v: %s", src, err)
	}
	if !srcInfo.Mode().IsRegular() {
		return fmt.Errorf("cannot copy non-regular file %#v: %s", src, err)
	}

	_, err = os.Stat(dst)
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("cannot stat %#v: %s", dst, err)
	}

	return copyFileContents(src, dst)
}

// CopyVerify copies the bytes from src into dst using [CopyFile], then
// verifies the two files have the same CRC32, giving a small measure of
// certainty that the copy succeeded.
func CopyVerify(src, dst string) error {
	var err = CopyFile(src, dst)
	if err != nil {
		return err
	}

	var srcChecksum, dstChecksum string
	srcChecksum, err = CRC32(src)
	if err != nil {
		return fmt.Errorf("unable to get source file's checksum: %s", err)
	}
	dstChecksum, err = CRC32(dst)
	if err != nil {
		return fmt.Errorf("unable to get destination file's checksum: %s", err)
	}
	if srcChecksum != dstChecksum {
		return fmt.Errorf("checksum failure")
	}

	return nil
}

// copyFileContents actually copies bytes from src to dst.  On any error, an
// attempt is made to clean up the state of the filesystem (though this is not
// guaranteed) and the first error encountered is returned.  i.e., if there's a
// failure in the [io.Copy] call, the caller will get that error, not the
// potentially meaningless error in the call to close the destination file.
func copyFileContents(src, dst string) error {
	var srcFile, dstFile *os.File
	var err error

	// Open source file or exit
	srcFile, err = os.Open(src)
	if err != nil {
		return fmt.Errorf("unable to read %#v: %s", src, err)
	}
	defer srcFile.Close()

	// Create destination file or exit
	dstFile, err = os.Create(dst)
	if err != nil {
		return fmt.Errorf("unable to create %#v: %s", dst, err)
	}

	// Attempt to copy, and if the operation fails, attempt to clean up, then exit
	_, err = io.Copy(dstFile, srcFile)
	if err != nil {
		err = fmt.Errorf("unable to copy data from %#v to %#v: %s", src, dst, err)
		dstFile.Close()
		os.Remove(dst)
		return err
	}

	// Attempt to sync the destination file
	err = dstFile.Sync()
	if err != nil {
		dstFile.Close()
		return fmt.Errorf("error syncing %#v: %s", dst, err)
	}

	// Attempt to close the destination file
	err = dstFile.Close()
	if err != nil {
		return fmt.Errorf("errro closing %#v: %s", dst, err)
	}

	return nil
}
