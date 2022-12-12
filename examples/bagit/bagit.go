package main

import (
	"fmt"
	"os"

	"github.com/uoregon-libraries/gopkg/bagit"
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

	perrf("Usage: %s <operation> <path to bag directory>", os.Args[0])
	perr("")
	perr(`<operation> must be either "write" or "validate"`)
	os.Exit(1)
}

func main() {
	if len(os.Args) != 3 {
		usage("invalid arguments")
	}

	var op, fname = os.Args[1], os.Args[2]

	switch op {
	case "write":
		write(fname)
	case "validate":
		fmt.Println("Validating bag at ", fname)
		validate(fname)
		fmt.Println("Valid")
	default:
		usage("invalid operation: " + op)
	}
}

func write(path string) {
	var b = bagit.New(path)
	var err = b.WriteTagFiles()
	if err != nil {
		perrf("Error generating tag files for %q: %s", path, err)
	}
}

func validate(path string) {
	var b = bagit.New(path)
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
