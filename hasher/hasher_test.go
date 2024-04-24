package hasher

import (
	"strings"
	"testing"
)

func TestSum(t *testing.T) {
	var tests = []struct {
		name     string
		hasher   *Hasher
		input    string
		expected string
	}{
		{"MD5", NewMD5(), "test", "098f6bcd4621d373cade4e832627b4f6"},
		{"SHA1", NewSHA1(), "test", "a94a8fe5ccb19ba61c4c0873d391e987982fbbd3"},
		{"SHA256", NewSHA256(), "test", "9f86d081884c7d659a2feaa0c55ad015a3bf4f1b2b0b822cd15d6c15b0f00a08"},
		{"SHA512", NewSHA512(), "test", "ee26b0dd4af7e749aa1a8ee3c10ae9923f618980772e473f8819a5d4940e0db27ac185f8a0e1d5f84f88bc887fd67b143732c304cc5fa9ad8e6f57f50028a8ff"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var r = strings.NewReader(tt.input)
			var got = tt.hasher.Sum(r)
			if got != tt.expected {
				t.Errorf("Sum() = %v, want %v", got, tt.expected)
			}

			// Verify hash is reset between calls - this also requires resetting the
			// string reader
			r.Reset(tt.input)
			got = tt.hasher.Sum(r)
			if got != tt.expected {
				t.Errorf("Sum() after reset = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestFileSum(t *testing.T) {
	var tests = []struct {
		name     string
		hasher   *Hasher
		filename string
		expected string
		wantErr  bool
	}{

		{"MD5 test.txt", NewMD5(), "testdata/test.txt", "2490a3d39b0004e4afeb517ef0ddbe2d", false},
		{"SHA1 test.txt", NewSHA1(), "testdata/test.txt", "b54e43082887d1e7cdb10b7a21fe4a1e56b44b5a", false},
		{"SHA256 test.txt", NewSHA256(), "testdata/test.txt", "3cd203ac11340842055a6de561c9d69ca4493e912bd4c3c440c80711e16d5aee", false},
		{"MD5 test2.bin", NewMD5(), "testdata/test2.bin", "84f053268f5341dfbd8c7373d9155992", false},
		{"SHA512 missing", NewSHA512(), "testdata/missing.txt", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got, err = tt.hasher.FileSum(tt.filename)
			if (err != nil) != tt.wantErr {
				t.Errorf("FileSum() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.expected {
				t.Errorf("FileSum() = %v, want %v", got, tt.expected)
			}

			// Verify hash is reset between calls
			got, err = tt.hasher.FileSum(tt.filename)
			if (err != nil) != tt.wantErr {
				t.Errorf("FileSum() after reset error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.expected {
				t.Errorf("FileSum() after reset = %v, want %v", got, tt.expected)
			}
		})
	}
}
