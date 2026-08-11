// Command deploy generates a catalog service's chart and installs it into a
// namespace with Helm (upgrade --install). With --render it prints the manifests
// instead, which needs no cluster and is handy in CI.
package main

import (
	"context"
	"flag"
	"log"
	"os"
	"os/exec"
	"strings"

	"github.com/pskshksh/pod-go-helm-factory/catalog"
	"github.com/pskshksh/pod-go-helm-factory/deploy"
	"github.com/pskshksh/pod-go-helm-factory/sdk/datastructures"
)

const defaultChartVersion = "0.1.0"

func main() {
	namespaceFlag := flag.String("namespace", "", "target namespace (default: --env)")
	env := flag.String("env", "", "environment: selects deploy/envs/<env> overlays and defaults the namespace")
	release := flag.String("release", "sampleapp", "catalog service / release name")
	tag := flag.String("tag", "", "image tag (default: resolved git sha)")
	version := flag.String("chart-version", defaultChartVersion, "chart version to stamp")
	envsRoot := flag.String("envs-root", deploy.DefaultEnvsRoot, "root dir for env value overlays")
	render := flag.Bool("render", false, "print manifests (helm template) instead of deploying")
	var extraValues datastructures.StringList
	flag.Var(&extraValues, "f", "extra values file, applied after env overlays (repeatable)")
	flag.Parse()

	namespace := *namespaceFlag
	if namespace == "" {
		namespace = *env
	}
	if namespace == "" {
		log.Fatal("deploy: --namespace or --env is required")
	}

	svc, ok := catalog.Find(*release)
	if !ok {
		log.Fatalf("deploy: no service %q in the catalog", *release)
	}

	// Env overlays first (base then env/service), explicit -f files after them.
	values := deploy.Overlays(*envsRoot, *env, svc.ChartName())
	values = append(values, extraValues...)

	rel := deploy.Release{
		Service:      svc,
		Namespace:    namespace,
		ChartVersion: *version,
		ImageTag:     resolveTag(*tag),
		Values:       values,
	}

	ctx := context.Background()
	d := deploy.Deployer{}

	var err error
	if *render {
		err = d.Render(ctx, rel)
	} else {
		err = d.Up(ctx, rel)
	}
	if err != nil {
		log.Fatal(err)
	}
}

// resolveTag picks the image tag: the --tag flag, else the CI-provided
// GITHUB_SHA, else the current git short sha, else "latest".
func resolveTag(flagTag string) string {
	if flagTag != "" {
		return flagTag
	}

	sha := os.Getenv("GITHUB_SHA")
	if sha != "" {
		return shortSHA(sha)
	}

	out, err := exec.Command("git", "rev-parse", "--short", "HEAD").Output()
	if err == nil {
		return strings.TrimSpace(string(out))
	}
	return "latest"
}

func shortSHA(sha string) string {
	if len(sha) > 7 {
		return sha[:7]
	}
	return sha
}
