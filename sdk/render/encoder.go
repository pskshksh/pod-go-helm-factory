package render

// Encoder accumulates a manifest document field by field and returns it as a
// string. Implementations decide the concrete syntax (YAML, JSON, ...) and any
// escaping the format requires.
//
// indent is a nesting depth: a text format (YAML) renders it as leading
// whitespace; a structural format (JSON) would read it as object depth.
type Encoder interface {
	// Field writes a scalar "key: value", quoting value if the format needs it.
	Field(indent int, key, value string)
	// Bool writes a boolean "key: value" unquoted (true/false). YAML booleans
	// must never be rendered as strings, so this bypasses the scalar-quoting path.
	Bool(indent int, key string, value bool)
	// Item writes a sequence element "- value", quoting value when a plain
	// scalar would be misparsed.
	Item(indent int, value string)
	// Line writes s verbatim at the given depth — for content the encoder must
	// not touch, e.g. Helm {{ }} templating passed straight through to Helm.
	Line(indent int, s string)
	// String returns the document built so far.
	String() string
}
