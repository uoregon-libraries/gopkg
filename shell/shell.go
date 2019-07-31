// Package shell centralizes common exec.Cmd functionality
package shell

import (
	"bytes"
	"os/exec"
	"strings"
	"syscall"

	"github.com/uoregon-libraries/gopkg/logger"
)

// Cmd extends os/exec's Cmd with a Logger for easier debugging or logging to
// custom sources (without having to inspect the various fields of Cmd).  On
// success, a debug-level message will be emitted; on failure, there will also
// be a failure notice at error level, and the entire command's output, line by
// line, as warning-level logs.
type Cmd struct {
	*exec.Cmd
	Logger *logger.Logger
}

// Command returns a generic Cmd which logs to stderr
func Command(path string, args ...string) *Cmd {
	return &Cmd{Cmd: exec.Command(path, args...), Logger: logger.Named("gopkg/pdf.ImageDPIs", logger.Debug)}
}

// Exec runs the command, logging output on failure
func (c *Cmd) Exec() (ok bool) {
	var cstr = strings.Join(c.Args, " ")
	c.Logger.Debugf("Running %q", cstr)
	var output, err = c.CombinedOutput()
	if err != nil {
		c.Logger.Errorf(`Failed to run %q: %s`, cstr, err)
		for _, line := range bytes.Split(output, []byte("\n")) {
			c.Logger.Debugf("--> %s", line)
		}

		return false
	}

	return true
}

// Exec attempts to run the given command, using the default logger
func Exec(path string, args ...string) (ok bool) {
	return Command(path, args...).Exec()
}

// ExecSubgroup is just like Exec, but sets the process to run in its own group
// so it doesn't get killed on CTRL+C
func ExecSubgroup(path string, args ...string) (ok bool) {
	var cmd = Command(path, args...)
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	return cmd.Exec()
}
