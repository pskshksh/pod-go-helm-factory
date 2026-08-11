package render

import (
	"strconv"
	"strings"
)

// YAML is an Encoder that produces YAML. It is deliberately a line/indent
// builder, not a reflection-based marshaller: the factory emits a fixed, known
// shape and must pass Helm {{ }} templating through untouched — which a generic
// marshaller can't do.
type YAML struct {
	b strings.Builder
}

var _ Encoder = (*YAML)(nil)

// needsQuoting reports whether s must be quoted to round-trip as the string it
// is. Pragmatic, not a full YAML validator.
func needsQuoting(s string) bool {
	if s == "" {
		return true
	}
	if s != strings.TrimSpace(s) {
		return true
	}
	if strings.Contains(s, ": ") || strings.HasSuffix(s, ":") || strings.Contains(s, " #") {
		return true
	}
	switch s[0] {
	case '!', '&', '*', '?', '|', '>', '%', '@', '`', '"', '\'', '[', ']', '{', '}', ',', '#', '-', ':':
		return true
	}
	switch strings.ToLower(s) {
	case "true", "false", "yes", "no", "on", "off", "null", "~":
		return true
	}
	if _, err := strconv.ParseFloat(s, 64); err == nil {
		return true
	}
	return false
}

// scalar renders s as a YAML scalar, double-quoting (with escaping) only when
// necessary.
func scalar(s string) string {
	if needsQuoting(s) {
		return strconv.Quote(s) // Go's double-quote escaping is YAML-compatible
	}
	return s
}

// Field writes "key: value", quoting value only when a plain scalar would be
// misparsed.
func (y *YAML) Field(indent int, key, value string) {
	y.Line(indent, key+": "+scalar(value))
}

// Bool writes "key: true" / "key: false" at the given indent, without quoting —
// unlike Field, which would quote the reserved words true/false via needsQuoting.
func (y *YAML) Bool(indent int, key string, value bool) {
	y.Line(indent, key+": "+strconv.FormatBool(value))
}

// Line writes s at the given indent level (2 spaces per level).
func (y *YAML) Line(indent int, s string) {
	y.b.WriteString(strings.Repeat("  ", indent))
	y.b.WriteString(s)
	y.b.WriteByte('\n')
}

// String returns the accumulated document.
func (y *YAML) String() string { return y.b.String() }
