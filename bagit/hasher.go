package bagit

import (
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/sha512"
	"hash"
)

// A Hasher represents a hash implementation for generating tag files
type Hasher struct {
	Hash func() hash.Hash
	Name string
}

// HashName is an enum-like int for simplifying bag hasher lookups
type HashName int

// Built-in hash lookup names
const (
	MD5 HashName = iota
	SHA1
	SHA256
	SHA512
)

var hasherLookup = map[HashName]*Hasher{
	MD5:    &Hasher{md5.New, "md5"},
	SHA1:   &Hasher{sha1.New, "sha1"},
	SHA256: &Hasher{sha256.New, "sha256"},
	SHA512: &Hasher{sha512.New, "sha512"},
}

// Hash returns a known Hasher for the given name. A nil hasher will be
// returned if the name is unknown.
func Hash(name HashName) *Hasher {
	return hasherLookup[name]
}
