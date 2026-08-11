package charts

import (
	"fmt"
	"os"
	"path/filepath"
)

// ensureDir creates path and all necessary parent directories. This is the one
// place the directory permission is defined.
func ensureDir(path string) error {
	err := os.MkdirAll(path, 0o755)
	if err != nil {
		return fmt.Errorf("charts: create dir '%s': %w", path, err)
	}
	return nil
}

// createChartDir creates the chart directory dir/<name>/ and its templates/
// subdirectory, returning the chart directory path.
func createChartDir(dir, name string) (string, error) {
	chartDir := filepath.Join(dir, name)
	err := ensureDir(filepath.Join(chartDir, "templates"))
	if err != nil {
		return "", err
	}
	return chartDir, nil
}

// writeFile writes content to dir/name, creating the file's parent directory
// as needed.
func writeFile(dir, name, content string) error {
	path := filepath.Join(dir, name)
	err := ensureDir(filepath.Dir(path))
	if err != nil {
		return err
	}
	err = os.WriteFile(path, []byte(content), 0o644)
	if err != nil {
		return fmt.Errorf("charts: write %s: %w", path, err)
	}
	return nil
}
