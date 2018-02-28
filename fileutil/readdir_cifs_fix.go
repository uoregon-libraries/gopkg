//+build linux

package fileutil

import (
	"os"
	"syscall"
)

const dirEntBuffer = 32768

// Readdir calls low-level Syscall functions in order to avoid the nasty CentOS
// 7 + CIFS bug we're run into
func Readdir(path string) ([]os.FileInfo, error) {
	var d *os.File
	var err error
	var bufp, nbuf int

	d, err = os.Open(path)
	if err != nil {
		return nil, err
	}
	defer d.Close()

	var buf = make([]byte, dirEntBuffer)
	var names = make([]string, 0, 100)
	for {
		if bufp >= nbuf {
			bufp = 0
			nbuf, err = syscall.ReadDirent(int(d.Fd()), buf)
			if err != nil {
				return nil, err
			}
		}

		// Check for EOF
		if nbuf <= 0 {
			break
		}

		var nb int
		nb, _, names = syscall.ParseDirent(buf[bufp:nbuf], -1, names)
		bufp += nb
	}

	var items []os.FileInfo
	for _, fname := range names {
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
