//+build linux

package fileutil

import (
	"fmt"
	"os"
	"syscall"
)

const dirEntBuffer = 32768

// Readdir calls low-level Syscall functions in order to avoid the nasty CentOS
// 7 + CIFS bug we're run into
func Readdir(path string) ([]os.FileInfo, error) {
	var d *os.File
	var err error

	d, err = os.Open(path)
	if err != nil {
		return nil, err
	}
	defer d.Close()

	var buf = make([]byte, dirEntBuffer)
	var consumed, n int
	var fnames []string
	for {
		n, err = syscall.ReadDirent(int(d.Fd()), buf)
		if err != nil {
			return nil, err
		}
		if n == 0 {
			break
		}

		consumed, _, fnames = syscall.ParseDirent(buf, -1, fnames)
		if consumed < n {
			return nil, fmt.Errorf("read %d bytes but only consumed %d", n, consumed)
		}
	}

	var items []os.FileInfo
	for _, fname := range fnames {
		if fname == "" {
			continue
		}
		var item, err = os.Stat(path + "/" + fname)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, err
}
