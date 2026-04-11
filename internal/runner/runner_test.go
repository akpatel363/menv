package runner

import (
	"runtime"
	"testing"
)

func TestShellQuote_POSIX(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("POSIX quoting only")
	}
	cases := []struct {
		in, want string
	}{
		{"foo", "foo"},
		{"", "''"},
		{"hello world", "'hello world'"},
		{"a'b", `'a'\''b'`},
		{"$VAR", "'$VAR'"},
		{"--flag=value", "--flag=value"},
		{"path/to/file", "path/to/file"},
		{"semi;colon", "'semi;colon'"},
	}
	for _, c := range cases {
		if got := shellQuote(c.in); got != c.want {
			t.Errorf("shellQuote(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}
