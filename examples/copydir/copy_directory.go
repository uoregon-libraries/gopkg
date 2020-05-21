// This is a quick cli for copying directories to verify that
// fileutil.CopyDirectory and fileutil.LinkDirectory work

package main

import (
	"fmt"
	"os"

	"github.com/uoregon-libraries/gopkg/fileutil"
)

func usageExit(code int) {
	var out = os.Stderr
	if code == 0 {
		out = os.Stdout
	}
	fmt.Fprintf(out, "Usage: %q <source directory> <destination directory> [--link]\n\n", os.Args[0])
	os.Exit(1)
}

func main() {
	if len(os.Args) < 3 || len(os.Args) > 4 {
		usageExit(1)
	}

	var fn = fileutil.CopyDirectory
	if len(os.Args) == 4 {
		if os.Args[3] != "--link" {
			usageExit(1)
		}
		fn = fileutil.LinkDirectory
	}

	var err = fn(os.Args[1], os.Args[2])
	if err != nil {
		fmt.Printf("Fail: %s\n", err)
		os.Exit(1)
	}

	fmt.Println("Success")
}
