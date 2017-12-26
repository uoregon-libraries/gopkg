// Package bashconf offers parsing of very simple bash-like configuration files
package bashconf

import (
	"io/ioutil"
	"strings"
)

// Config is a simple alias for holding key/value pairs
type Config map[string]string

// New just simplifies making a new Config map to avoid accidental nil map panics
func New() Config {
	return make(Config)
}

// ParseFile reads config from a file and adds to the Config map.  If the file
// can't be read or parsed, an error will be returned.
//
// This function is built to read bash-like environmental variables in order to
// maximize cross-language configurability.  Very minimal processing occurs
// here, leaving that to the caller.  i.e., "FOO=1" will yield a key of "FOO"
// and a value of "1" since bash has no types.  This supports only the simplest
// of bash variable assignment: no arrays, no substitutions of other variables,
// just very basic key/value pairs.
func (c Config) ParseFile(filename string) error {
	var content, err = ioutil.ReadFile(filename)
	if err != nil {
		return err
	}
	c.ParseString(string(content))
	return nil
}

// ParseString reads each line in s and attempts to parse a key/value pair.
// Any line starting with a comment is ignored, and lines that don't split into
// a key/value are ignored.
func (c Config) ParseString(s string) {
	var lines = strings.Split(s, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		if line[0] == '#' {
			continue
		}

		var kvparts = strings.SplitN(line, "=", 2)
		if len(kvparts) != 2 {
			continue
		}
		var key, val = kvparts[0], kvparts[1]

		if key == "" {
			continue
		}

		// Remove surrounding quotes if any exist, but only one level of quotes
		if val[0] == '"' {
			val = strings.Trim(val, `"`)
		} else if val[0] == '\'' {
			val = strings.Trim(val, `'`)
		}
		c[key] = val
	}
}
