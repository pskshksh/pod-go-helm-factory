package deploy

import (
	"os"
	"path/filepath"
)

// DefaultEnvsRoot is where per-environment value overlays live.
const DefaultEnvsRoot = "deploy/envs"

// Overlays returns the value files to apply when deploying service into env, in
// precedence order — base.yaml first, then <env>/<service>.yaml — skipping any
// that are absent. Missing overlays are fine: the chart's baked defaults fill
// the gaps, and later files (and --set) override earlier ones. env may be empty,
// in which case only base.yaml is considered.
func Overlays(root, env, service string) []string {
	candidates := []string{filepath.Join(root, "base.yaml")}
	if env != "" {
		candidates = append(candidates, filepath.Join(root, env, service+".yaml"))
	}

	var found []string
	for _, c := range candidates {
		info, err := os.Stat(c)
		if err == nil && !info.IsDir() {
			found = append(found, c)
		}
	}
	return found
}
