package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/uoregon-libraries/gopkg/fileutil/manifest"
)

func usage(msg string) {
	fmt.Fprintln(os.Stderr, "\033[31;1mError:\033[0m " + msg)
	fmt.Fprintln(os.Stderr)
	fmt.Fprintf(os.Stderr, "usage: %s <operation> <directory>...\n\n", os.Args[0])
	fmt.Fprintln(os.Stderr, `Operation may be either "create" or "verify"`)
}

func main() {
	if len(os.Args) < 2 {
		usage("No operation specified; no directories specified")
		os.Exit(-1)
	}
	if len(os.Args) < 3 {
		usage("No directories specified")
		os.Exit(-1)
	}

	var op = os.Args[1]
	switch op {
	case "create":
		create(os.Args[2:])
	case "verify":
		verify(os.Args[2:])
	default:
		usage(fmt.Sprintf("Invalid operation %q", op))
	}
}

func build(dirs []string) map[string]*manifest.Manifest {
	var manifests = make(map[string]*manifest.Manifest)
	for _, dir := range dirs {
		var start = time.Now()
		var m = manifest.New(dir)
		var err = m.Build()
		if err != nil {
			log.Printf("Skipping %q due to error: %s", dir, err)
			continue
		}
		var duration = time.Since(start)
		log.Printf("Built manifest for %q in %s", dir, duration)
		manifests[dir] = m
	}

	return manifests
}

func create(dirs []string) {
	var manifests = build(dirs)
	for dir, m := range manifests {
		var err = m.Write()
		if err != nil {
			log.Printf("Unable to create manifest for %q: %s", dir, err)
			continue
		}
	}
}

func verify(dirs []string) {
	var manifests = build(dirs)
	for dir, m := range manifests {
		var exist = manifest.New(dir)
		var err = exist.Read()
		if err != nil {
			log.Printf("Unable to read manifest for %q: %s", dir, err)
			continue
		}

		if m.Equiv(exist) {
			log.Printf("OK %q", dir)
		} else {
			log.Printf("NOT OK %q", dir)
		}
	}
}
