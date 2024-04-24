package main

import (
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/uoregon-libraries/gopkg/fileutil/manifest"
	"github.com/uoregon-libraries/gopkg/hasher"
)

func usage(msg string) {
	fmt.Fprintln(os.Stderr, "\033[31;1mError:\033[0m "+msg)
	fmt.Fprintln(os.Stderr)
	fmt.Fprintf(os.Stderr, "usage: %s <operation> <directory>...\n\n", os.Args[0])
	fmt.Fprintln(os.Stderr, `Operation may be either "create" or "verify". "create" may optionally be`)
	fmt.Fprintln(os.Stderr, `suffixed with "-<algo>", where algo is md5, sha256, etc.`)

	os.Exit(-1)
}

func main() {
	if len(os.Args) < 2 {
		usage("No operation specified; no directories specified")
	}
	if len(os.Args) < 3 {
		usage("No directories specified")
	}

	log.Printf("%#v", os.Args)
	var parts = strings.SplitN(os.Args[1], "-", 2)
	var op = parts[0]
	var algo string
	if len(parts) > 1 {
		algo = parts[1]
	}
	switch op {
	case "create":
		create(os.Args[2:], algo)
	case "verify":
		if algo != "" {
			usage(fmt.Sprintf(`Invalid operation: "verify" cannot be given an algorithm`))
		}
		verify(os.Args[2:])
	default:
		usage(fmt.Sprintf("Invalid operation %q", op))
	}
}

func create(dirs []string, algo string) {
	for _, dir := range dirs {
		var start = time.Now()
		var m, err = manifest.Build(dir, hasher.Algo(algo))
		var duration = time.Since(start)
		if err == nil {
			err = m.Write()
		}
		if err != nil {
			log.Printf("Unable to create manifest for %q: %s", dir, err)
			continue
		}
		log.Printf("Built manifest for %q in %s", dir, duration)
	}
}

func verify(dirs []string) {
	for _, dir := range dirs {
		var m, err = manifest.Open(dir)
		if err != nil {
			log.Printf("Unable to read manifest for %q: %s", dir, err)
			continue
		}

		var valid bool
		valid, err = m.Validate()
		if err != nil {
			log.Printf("Unable to validate manifest for %q: %s", dir, err)
			continue
		}
		if valid {
			log.Printf("OK %q", dir)
		} else {
			log.Printf("NOT OK %q", dir)
		}
	}
}
