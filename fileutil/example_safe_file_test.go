package fileutil_test

import (
	"fmt"
	"io/ioutil"

	"github.com/uoregon-libraries/gopkg/fileutil"
)

func makeWrite(fname string) *fileutil.SafeFile {
	var f = fileutil.NewSafeFile(fname)
	f.Write([]byte("This is a test.\n\nA what?\n\nA test.\n\nA what?\n\nA test.\n\nOh, a test.\n"))
	return f
}

func Example_minimal() {
	var fname = "/tmp/blah.txt"
	var f = makeWrite(fname)

	// Closing twice shouldn't cause problems
	f.Close()
	f.Close()

	var data, err = ioutil.ReadFile(fname)
	if err != nil {
		fmt.Printf("Error trying to read %q: %s", fname, err)
	}
	fmt.Println(string(data))

	f = makeWrite(fname)
	f.Cancel()
	data, err = ioutil.ReadFile(fname)
	fmt.Println(err)

	// Output:
	// This is a test.
	//
	// A what?
	//
	// A test.
	//
	// A what?
	//
	// A test.
	//
	// Oh, a test.
	//
	// open /tmp/blah.txt: no such file or directory
}
