package render

import "testing"

// Bool must emit an unquoted YAML boolean. Field would quote the reserved words
// true/false (needsQuoting flags them), turning a bool into a string — this
// guards against a regression back to that path.
func TestBoolUnquoted(t *testing.T) {
	cases := []struct {
		value bool
		want  string
	}{
		{true, "readOnlyRootFilesystem: true\n"},
		{false, "allowPrivilegeEscalation: false\n"},
	}
	for _, c := range cases {
		var y YAML
		key := "readOnlyRootFilesystem"
		if !c.value {
			key = "allowPrivilegeEscalation"
		}
		y.Bool(0, key, c.value)

		got := y.String()
		if got != c.want {
			t.Errorf("Bool(%v) = %q, want %q", c.value, got, c.want)
		}
	}
}

// Item must quote a scalar only when a plain one would be misparsed — a leading
// '-' (like a CLI flag) needs quoting; a plain path does not.
func TestItemQuoting(t *testing.T) {
	cases := []struct {
		value string
		want  string
	}{
		{"/bin/sh", "- /bin/sh\n"},
		{"-c", "- \"-c\"\n"},
	}
	for _, c := range cases {
		var y YAML
		y.Item(0, c.value)

		got := y.String()
		if got != c.want {
			t.Errorf("Item(%q) = %q, want %q", c.value, got, c.want)
		}
	}
}
