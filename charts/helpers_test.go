package charts

import (
	"bytes"
	"flag"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"
)

// updateSnapshots rewrites snapshot files instead of comparing. Run:
//
//	go test ./charts -run TestName -update
var updateSnapshots = flag.Bool("update", false, "rewrite snapshots instead of comparing")

// compareFiles deeply compares two files. It establishes inequality as cheaply
// as possible (a byte compare) and only pays for the human-readable line diff
// when the contents actually differ.
func compareFiles(t *testing.T, gotPath, wantPath string) {
	t.Helper()
	got, err := os.ReadFile(gotPath)
	if err != nil {
		t.Fatalf("read generated file %s: %v", gotPath, err)
	}
	want, err := os.ReadFile(wantPath)
	if err != nil {
		t.Fatalf("read snapshot %s (create it with -update): %v", wantPath, err)
	}

	if bytes.Equal(got, want) { // fast path: no allocation, no diff work
		return
	}

	f, err := os.CreateTemp("", "*-"+filepath.Base(wantPath))
	if err != nil {
		t.Fatalf("create temp file for generated output: %v", err)
	}
	_, err = f.Write(got)
	if err != nil {
		t.Fatalf("write generated output to %s: %v", f.Name(), err)
	}

	err = f.Close()
	if err != nil {
		t.Fatalf("close %s: %v", f.Name(), err)
	}

	wantAbs, err := filepath.Abs(wantPath)
	if err != nil {
		t.Fatalf("resolve snapshot path %s: %v", wantPath, err)
	}
	t.Errorf("generated output does not match snapshot\n  snapshot:  %s\n  generated: %s",
		wantAbs, f.Name())
}

// assertRender compares a rendered string against an expected block and reports
// both in full on mismatch. It is the in-memory counterpart to compareFiles for
// tests that exercise a single renderer rather than a whole chart directory.
func assertRender(t *testing.T, got, want string) {
	t.Helper()
	if got != want {
		t.Errorf("rendered output mismatch:\n--- got ---\n%s\n--- want ---\n%s", got, want)
	}
}

// bundleChart writes every file in the generated chart directory into a single
// deterministic YAML blob at dst, each file introduced by a "# file: <path>"
// header. Files are sorted so output is stable regardless of walk order.
func bundleChart(t *testing.T, chartDir, dst string) {
	t.Helper()
	files := readTree(t, chartDir)

	names := make([]string, 0, len(files))
	for name := range files {
		names = append(names, name)
	}
	sort.Strings(names)

	var b strings.Builder
	for i, name := range names {
		if i > 0 {
			b.WriteByte('\n')
		}
		b.WriteString("# file: " + name + "\n")
		content := files[name]
		b.WriteString(content)
		if !strings.HasSuffix(content, "\n") {
			b.WriteByte('\n')
		}
	}
	err := writeFile(filepath.Dir(dst), filepath.Base(dst), b.String())
	if err != nil {
		t.Fatalf("write bundle %s: %v", dst, err)
	}
}

// readTree maps relative slash-paths to file contents for every file under root.
func readTree(t *testing.T, root string) map[string]string {
	t.Helper()
	files := map[string]string{}
	err := filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		rel, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		b, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		files[filepath.ToSlash(rel)] = string(b)
		return nil
	})
	if err != nil {
		t.Fatalf("read tree %s: %v", root, err)
	}
	return files
}
