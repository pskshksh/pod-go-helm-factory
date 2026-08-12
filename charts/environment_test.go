package charts

import "testing"

// TestTargetNamespace confirms Namespace overrides Name, and Name is the default.
func TestTargetNamespace(t *testing.T) {
	cases := []struct {
		id   string
		env  Environment
		want string
	}{
		{"defaults_to_name", Environment{Name: "prod"}, "prod"},
		{"namespace_overrides", Environment{Name: "prod", Namespace: "toto"}, "toto"},
	}

	for _, c := range cases {
		t.Run(c.id, func(t *testing.T) {
			got := c.env.TargetNamespace()
			if got != c.want {
				t.Errorf("TargetNamespace() = %q, want %q", got, c.want)
			}
		})
	}
}
