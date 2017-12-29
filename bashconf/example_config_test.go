package bashconf_test

import (
	"fmt"
	"os"

	"github.com/uoregon-libraries/gopkg/bashconf"
)

var configString = `
# This is a comment.  It is ignored.
This is not a value key/value pair, so it, too, is ignored

# These are both treated as strings; right now we don't support even
# simple data types
SOMEARG=5
SOMEARG2="6"

# This is treated as if it were "foo bar" even though bash wouldn't see it that
# way.  Repeat after me: this is *not* a bash parser, just a very naive config
# reader.
VALUE=foo bar

# This is a valid URL
URL_VALUE=https://uoregon.edu

# This is technically valid per the URL RFC, but our parser requires a scheme
# of http or https
BAD_URL=ftp://uoregon.edu

# Okay, now we support ints, woo!
NUMBER="75"
BAD_NUMBER="x"
`

func Example() {
	var c = bashconf.New()
	os.Setenv("ENVONLY", "foo")
	c.ParseString(configString)
	fmt.Printf("SOMEARG is %#v; SOMEARG2 is %#v\n", c.Get("SOMEARG"), c.Get("SOMEARG2"))

	var st struct {
		Value     string `setting:"VALUE"`
		URLValue  string `setting:"URL_VALUE" type:"url"`
		BadURL    string `setting:"BAD_URL" type:"url"`
		Number    int    `setting:"NUMBER" type:"int"`
		BadNumber int    `setting:"BAD_NUMBER" type:"int"`
		EnvVar    string `setting:"ENVONLY"`
	}
	var err = c.Store(&st)
	fmt.Printf("Errors: %s\n", err)
	fmt.Printf("st.Value: %#v\n", st.Value)
	fmt.Printf("st.URLValue: %#v\n", st.URLValue)
	fmt.Printf("st.BadURL: %#v\n", st.BadURL)
	fmt.Printf("st.Number: %d\n", st.Number)
	fmt.Printf("st.BadNumber: %d\n", st.BadNumber)
	fmt.Printf("st.EnvVar: %q\n", st.EnvVar)

	c.EnvironmentOverrides(true)
	c.Store(&st)
	fmt.Printf("st.EnvVar after allowing overrides: %q\n", st.EnvVar)

	// Output:
	// SOMEARG is "5"; SOMEARG2 is "6"
	// Errors: invalid configuration: "BAD_URL" ("ftp://uoregon.edu") is not a valid URL: must be http(s), "BAD_NUMBER" ("x") is not a valid integer
	// st.Value: "foo bar"
	// st.URLValue: "https://uoregon.edu"
	// st.BadURL: "ftp://uoregon.edu"
	// st.Number: 75
	// st.BadNumber: 0
	// st.EnvVar: ""
	// st.EnvVar after allowing overrides: "foo"
}
