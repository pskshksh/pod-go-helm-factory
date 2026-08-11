package deploy

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func mkOverlay(t *testing.T, path string) {
	t.Helper()
	err := os.MkdirAll(filepath.Dir(path), 0o755)
	if err != nil {
		t.Fatal(err)
	}
	err = os.WriteFile(path, []byte("x: 1\n"), 0o644)
	if err != nil {
		t.Fatal(err)
	}
}

func TestOverlays(t *testing.T) {
	root := t.TempDir()
	base := filepath.Join(root, "base.yaml")
	prod := filepath.Join(root, "prod", "sampleapp.yaml")
	// base.yaml and prod/sampleapp.yaml exist; staging has no overlay.
	mkOverlay(t, base)
	mkOverlay(t, prod)

	cases := []struct {
		id      string
		env     string
		service string
		want    []string
	}{
		{"base_and_env", "prod", "sampleapp", []string{base, prod}},
		{"only_base_when_env_overlay_missing", "staging", "sampleapp", []string{base}},
		{"only_base_when_no_env", "", "sampleapp", []string{base}},
	}

	for _, c := range cases {
		t.Run(c.id, func(t *testing.T) {
			got := Overlays(root, c.env, c.service)
			if !reflect.DeepEqual(got, c.want) {
				t.Errorf("Overlays(%q, %q) = %v, want %v", c.env, c.service, got, c.want)
			}
		})
	}
}
