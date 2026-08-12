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
	"github.com/pskshksh/pod-go-helm-factory/charts"
	"github.com/pskshksh/pod-go-helm-factory/deploy"
	"github.com/pskshksh/pod-go-helm-factory/sdk/datastructures"
)

const defaultChartVersion = "0.1.0"

func main() {
	env := flag.String("env", "", "config environment: overlay selector, deploy/envs/<env> (required)")
	namespace := flag.String("namespace", "", "target namespace (default: --env)")
	release := flag.String("release", "", "deploy only this service (default: every service in the catalog)")
	tag := flag.String("tag", "", "image tag (default: resolved git sha)")
	version := flag.String("chart-version", defaultChartVersion, "chart version to stamp")
	envsRoot := flag.String("envs-root", deploy.DefaultEnvsRoot, "root dir for env value overlays")
	render := flag.Bool("render", false, "print manifests (helm template) instead of deploying")
	var extraValues datastructures.StringList
	flag.Var(&extraValues, "f", "extra values file, applied after env overlays (repeatable)")
	flag.Parse()

	if *env == "" {
		log.Fatal("deploy: --env is required")
	}

	services := catalog.Services()
	if *release != "" {
		svc, ok := catalog.Find(*release)
		if !ok {
			log.Fatalf("deploy: no service %q in the catalog", *release)
		}
		services = []charts.Generator{svc}
	}

	environment := charts.Environment{Name: *env, Namespace: *namespace, Services: services}
	opts := deploy.EnvOptions{
		ChartVersion: *version,
		ImageTag:     resolveTag(*tag),
		EnvsRoot:     *envsRoot,
		ExtraValues:  extraValues,
	}

	ctx := context.Background()
	d := deploy.Deployer{}

	var err error
	if *render {
		err = d.RenderAll(ctx, environment, opts)
	} else {
		err = d.UpAll(ctx, environment, opts)
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
