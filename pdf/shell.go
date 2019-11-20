package pdf

import (
	"bytes"
	"fmt"
	"os/exec"
)

// ShellError holds error context for command-line executions so this package
// can return information rather than logging to stderr
type ShellError struct {
	e       error
	Command string
	Output  []string
}

// Error returns a simple error string to fulfill the basic error interface
func (se ShellError) Error() string {
	return fmt.Sprintf("gopkg/pdf: executing %s: %s", se.Command, se.e.Error())
}

var pdfImagesBinary = "pdfimages"

// pdfimagesLines returns the unparsed from "pdfimages -list <path>"
func pdfimagesLines(path string) (output []string, err error) {
	var cmd = exec.Command(pdfImagesBinary, "-list", path)
	var stdout, stderr bytes.Buffer
	cmd.Stderr = &stderr
	cmd.Stdout = &stdout

	err = cmd.Run()
	if err != nil {
		var se = &ShellError{e: err, Command: fmt.Sprintf("pdfimage -list %s", path)}
		for _, line := range bytes.Split(stderr.Bytes(), []byte{'\n'}) {
			se.Output = append(se.Output, string(line))
		}
		return nil, se
	}

	for i, line := range bytes.Split(stdout.Bytes(), []byte{'\n'}) {
		// Lines 1 and 2 are always headers
		if i == 0 && bytes.HasPrefix(line, []byte("page")) {
			continue
		}
		if i == 1 && bytes.HasPrefix(line, []byte("------")) {
			continue
		}
		// The last line appears to always be blank
		if len(line) == 0 {
			continue
		}
		output = append(output, string(line))
	}

	return output, nil
}
