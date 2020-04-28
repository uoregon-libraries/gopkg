package logger

import (
	"bytes"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestBasic(t *testing.T) {
	var l = Named("foo", Debug, false)

	var byteStream = &bytes.Buffer{}
	var sl = l.Loggable.(*SimpleLogger)
	sl.Output = byteStream
	sl.TimeFormat = "n/a"

	var tests = map[string]struct {
		fn      func(format string, args ...interface{})
		message string
		args    []interface{}
		want    string
	}{
		"info": {
			fn:      l.Infof,
			message: "this is a %q info with the number %d",
			args:    []interface{}{"simple", 5},
			want:    `n/a - foo - INFO - this is a "simple" info with the number 5`,
		},
		"error": {
			fn:      l.Errorf,
			message: "plain error message",
			want:    `n/a - foo - ERROR - plain error message`,
		},
	}

	for _, tc := range tests {
		byteStream.Reset()
		tc.fn(tc.message, tc.args...)
		var got = byteStream.String()
		// We add a newline here because we know the logger prints that for us, but
		// we don't want to litter the test cases with newlines
		var diff = cmp.Diff(tc.want+"\n", got)
		if diff != "" {
			t.Fatalf(diff)
		}
	}
}

func TestStructured(t *testing.T) {
	var l = Named("structured", Debug, true)

	var byteStream = &bytes.Buffer{}
	var sl = l.Loggable.(*SimpleLogger)
	sl.Output = byteStream
	sl.TimeFormat = "n/a"

	var tests = map[string]struct {
		fn      func(format string, args ...interface{})
		message string
		args    []interface{}
		want    string
	}{
		"info": {
			fn:      l.Infof,
			message: "this is a %q info with the number %d",
			args:    []interface{}{"simple", 5},
			want:    `time="n/a" app="structured" level="INFO" message="this is a \"simple\" info with the number 5"`,
		},
		"error": {
			fn:      l.Errorf,
			message: "plain error message",
			want:    `time="n/a" app="structured" level="ERROR" message="plain error message"`,
		},
		"message with newline": {
			fn:      l.Infof,
			message: "this has a \nnewline",
			want:    `time="n/a" app="structured" level="INFO" message="this has a <CR>newline"`,
		},
	}

	for _, tc := range tests {
		byteStream.Reset()
		tc.fn(tc.message, tc.args...)
		var got = byteStream.String()
		// We add a newline here because we know the logger prints that for us, but
		// we don't want to litter the test cases with newlines
		var diff = cmp.Diff(tc.want+"\n", got)
		if diff != "" {
			t.Fatalf(diff)
		}
	}
}
