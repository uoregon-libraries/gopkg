package fileutil

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"sync"
)

// WriteCancelCloser is an io.WriteCloser which also exposes a cancel method so
// a writer can (attempt to) clean up anything it may have left behind in the
// writing process when errors occurred.
type WriteCancelCloser interface {
	io.WriteCloser
	Cancel()
}

// SafeFile wraps the process of creating a temporary file, writing to it,
// copying it to the final file once all writing is successful, and then
// deleting the temp file.  This reduces the opportunities for errors
// (file-writing or program logic) to leave behind remnants of files.
//
// Errors are returned from Write and Close methods, but are also stored
// internally to allow the SafeFile to automatically skip certain methods and
// clean up after itself for simpler use-cases.  This means that when errors
// occur which aren't related to the caller's logic (e.g., an error is returned
// by Write), calling Cancel is unnecessary.
type SafeFile struct {
	sync.Mutex
	temp      *os.File
	tempName  string
	finalPath string
	Err       error
	closed    bool
}

// NewSafeFile returns a new SafeFile construct, wrapping the given path
func NewSafeFile(path string) *SafeFile {
	var f, err = ioutil.TempFile("", "")
	var sf = &SafeFile{temp: f, finalPath: path, Err: err}
	if err == nil {
		sf.tempName = f.Name()
	} else {
		sf.Err = fmt.Errorf("couldn't generate temp file: %w", err)
	}

	return sf
}

// Write delegates to the temporary file handle
func (f *SafeFile) Write(p []byte) (n int, err error) {
	if f.Err != nil {
		return 0, fmt.Errorf("cannot write to SafeFile with errors")
	}

	n, f.Err = f.temp.Write(p)
	return n, f.Err
}

// WriteAt delegates to the temporary file handle so we can use a SafeFile when
// WriteAt is used.  Since we check and store internal state (errors), this is
// a mutex-locked operation even if writes could otherwise be concurrent.
func (f *SafeFile) WriteAt(p []byte, off int64) (n int, err error) {
	f.Lock()
	defer f.Unlock()

	if f.Err != nil {
		return 0, fmt.Errorf("cannot write to SafeFile with errors")
	}

	n, f.Err = f.temp.WriteAt(p, off)
	return n, f.Err
}

// Close closes the temporary file, copies its data to the final location, and
// removes the temp file from disk.  If an error occurs, an attempt is made to
// remove all files from disk to avoid broken files, and the error is returned.
func (f *SafeFile) Close() error {
	// Closing a file twice isn't great, but it should simply return the error
	// from f.temp.Close, not cancel the otherwise successful operation
	if f.closed {
		return f.temp.Close()
	}

	if f.Err != nil {
		f.Cancel()
		return f.Err
	}

	var err = f.temp.Close()
	if err != nil {
		f.Cancel()
		f.Err = fmt.Errorf("unable to close temp file: %s", f.Err)
		return f.Err
	}

	err = CopyVerify(f.tempName, f.finalPath)
	if err != nil {
		f.Cancel()
		f.Err = fmt.Errorf("unable to copy temp file: %s", f.Err)
		return f.Err
	}

	f.closed = true
	return os.Remove(f.tempName)
}

// Cancel attempts to close and delete all files
func (f *SafeFile) Cancel() {
	f.temp.Close()
	os.Remove(f.tempName)
	os.Remove(f.finalPath)
}
