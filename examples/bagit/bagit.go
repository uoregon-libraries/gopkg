package main

import (
	"fmt"
	"os"

	"github.com/uoregon-libraries/gopkg/bagit"
	"github.com/uoregon-libraries/gopkg/hasher"
)

func perr(msg string) {
	fmt.Fprintln(os.Stderr, msg)
}

func perrf(msg string, args ...any) {
	fmt.Fprintf(os.Stderr, msg+"\n", args...)
}

func usage(err string) {
	if err != "" {
		perrf("Error: %s", err)
		perr("")
	}

	perrf("Usage: %s <operation> <algorithm> <path to bag directory>", os.Args[0])
	perr("")
	perr(`<operation> must be either "write" or "validate"`)
	perr(`<algorithm> must be one of: "md5", "sha1", "sha256", "sha512"`)
	os.Exit(1)
}

func main() {
	if len(os.Args) != 4 {
		usage("invalid arguments")
	}

	var op, algo, fname = os.Args[1], os.Args[2], os.Args[3]

	var h *hasher.Hasher
	switch algo {
	case "md5":
		h = hasher.MD5()
	case "sha1":
		h = hasher.SHA1()
	case "sha256":
		h = hasher.SHA256()
	case "sha512":
		h = hasher.SHA512()
	default:
		usage("invalid algorithm: " + algo)
	}

	switch op {
	case "write":
		write(fname, h)
	case "validate":
		fmt.Println("Validating bag at ", fname)
		validate(fname, h)
		fmt.Println("Valid")
	default:
		usage("invalid operation: " + op)
	}
}

func write(path string, h *hasher.Hasher) {
	var b = bagit.New(path, h)
	var err = b.WriteTagFiles()
	if err != nil {
		perrf("Error generating tag files for %q: %s", path, err)
	}
}

func validate(path string, h *hasher.Hasher) {
	var b = bagit.New(path, h)
	var discrepancies, err = b.Validate()
	if err != nil {
		perrf("Error trying to validate %q: %s", path, err)
		os.Exit(255)
	}

	if len(discrepancies) > 0 {
		perr("Bag is invalid:")
		for _, txt := range discrepancies {
			perrf("  - %s", txt)
		}
		os.Exit(1)
	}
}
