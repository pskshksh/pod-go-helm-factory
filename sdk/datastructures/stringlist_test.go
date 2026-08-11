package datastructures

import "testing"

func TestStringList(t *testing.T) {
	var s StringList
	_ = s.Set("a.yaml")
	_ = s.Set("b.yaml")

	if len(s) != 2 || s[0] != "a.yaml" || s[1] != "b.yaml" {
		t.Fatalf("collected %v, want [a.yaml b.yaml]", s)
	}
	if s.String() != "a.yaml,b.yaml" {
		t.Errorf("String() = %q, want a.yaml,b.yaml", s.String())
	}
}
