package hasher

import (
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/sha512"
	"fmt"
	"hash"
	"io"
	"os"
)

// Hasher wraps any hash.Hash implementation with some shortcuts to help with
// common use-cases we have
type Hasher struct {
	hash.Hash
	Name string
}

// MD5 returns a Hasher using crypto/md5
func MD5() *Hasher {
	return &Hasher{md5.New(), "md5"}
}

// SHA1 returns a Hasher using crypto/sha1
func SHA1() *Hasher {
	return &Hasher{sha1.New(), "sha1"}
}

// SHA256 returns a Hasher using crypto/sha256
func SHA256() *Hasher {
	return &Hasher{sha256.New(), "sha256"}
}

// SHA512 returns a Hasher using crypto/sha512
func SHA512() *Hasher {
	return &Hasher{sha512.New(), "sha512"}
}

// Sum resets Hasher's state and generates a hex sum of the given io.Reader
func (h *Hasher) Sum(r io.Reader) string {
	h.Hash.Reset()
	io.Copy(h.Hash, r)
	var data = h.Hash.Sum(nil)
	return fmt.Sprintf("%x", data)
}

// FileSum resets Hasher's state and generates a hex sum of the given file
func (h *Hasher) FileSum(filename string) (string, error) {
	var f, err = os.Open(filename)
	if err != nil {
		return "", fmt.Errorf("opening %q: %w", filename, err)
	}
	defer f.Close()
	return h.Sum(f), nil
}
