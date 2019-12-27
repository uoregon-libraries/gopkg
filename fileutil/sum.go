package fileutil

import (
	"crypto/sha256"
	"fmt"
	"hash/crc32"
	"io"
	"os"
)

// CRC32 returns the checksum of the given file.  This is intended for
// verifying file copies immediately after the copy happens.  It should not be
// relied upon to detect malicious file changes.
func CRC32(file string) (string, error) {
	var f, err = os.Open(file)
	if err != nil {
		return "", err
	}
	defer f.Close()

	var h = crc32.NewIEEE()
	_, err = io.Copy(h, f)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%x", h.Sum(nil)), nil
}

// SHA256 returns the checksum of the given file
func SHA256(file string) ([]byte, error) {
	var f, err = os.Open(file)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var h = sha256.New()
	_, err = io.Copy(h, f)
	if err != nil {
		return nil, err
	}

	return h.Sum(nil), nil
}
