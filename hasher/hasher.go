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

// Algo is our enum-like value for supported algorithms we use widely
type Algo string

// The algorithms we support currently
const (
	MD5    Algo = "md5"
	SHA1        = "sha1"
	SHA256      = "sha256"
	SHA512      = "sha512"
)

var fnLookup = map[Algo]func() hash.Hash{
	MD5:    md5.New,
	SHA1:   sha1.New,
	SHA256: sha256.New,
	SHA512: sha512.New,
}

// NewMD5 returns a Hasher using crypto/md5
func NewMD5() *Hasher {
	return New(MD5)
}

// NewSHA1 returns a Hasher using crypto/sha1
func NewSHA1() *Hasher {
	return New(SHA1)
}

// NewSHA256 returns a Hasher using crypto/sha256
func NewSHA256() *Hasher {
	return New(SHA256)
}

// NewSHA512 returns a Hasher using crypto/sha512
func NewSHA512() *Hasher {
	return New(SHA512)
}

// New returns a Hasher for the given algorithm. If you pass in an invalid
// Algo, this will give you a nil Hasher.
func New(a Algo) *Hasher {
	var fn, ok = fnLookup[a]
	if !ok {
		return nil
	}
	return &Hasher{Name: string(a), Hash: fn()}
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
