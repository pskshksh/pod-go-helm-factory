// Package datastructures holds small, reusable data types shared across the
// repo's commands and libraries.
package datastructures

import "strings"

// StringList is a flag.Value that accumulates a repeated string flag, e.g.
// -f a.yaml -f b.yaml collects ["a.yaml", "b.yaml"].
type StringList []string

// String renders the collected values, satisfying flag.Value.
func (s *StringList) String() string { return strings.Join(*s, ",") }

// Set appends one occurrence of the flag, satisfying flag.Value.
func (s *StringList) Set(v string) error {
	*s = append(*s, v)
	return nil
}
