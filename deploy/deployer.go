// Package deploy installs generated charts into a cluster with Helm. A Release
// is a declarative request — which service, at which version, into which
// namespace — and the Deployer generates the chart and runs helm on it. Helm is
// invoked through an injectable Runner and the argument construction is pure
// (upgradeArgs/templateArgs), so deploys are unit-testable without a cluster.
package deploy

import (
	"context"
	"fmt"
	"os"
	"os/exec"

	"github.com/pskshksh/pod-go-helm-factory/charts"
	"github.com/pskshksh/pod-go-helm-factory/sdk/datastructures"
)

// Release is a declarative request to deploy one service: which chart, at which
// version, into which namespace, with which image tag and value overlays.
type Release struct {
	Service      charts.Generator          // the chart descriptor to deploy
	Namespace    string                    // target namespace
	ChartVersion string                    // version stamped into Chart.yaml
	ImageTag     string                    // becomes --set image.tag (empty keeps the chart default)
	Values       datastructures.StringList // -f overlays, merged in order (later wins)
}

func (r Release) validate() error {
	if r.Service == nil {
		return fmt.Errorf("deploy: release service is required")
	}
	if r.Namespace == "" {
		return fmt.Errorf("deploy: namespace is required")
	}
	if r.ChartVersion == "" {
		return fmt.Errorf("deploy: chart version is required")
	}
	return nil
}

// Deployer runs Helm to install or upgrade releases. The zero value uses the
// "helm" binary on PATH and execs for real; set Runner in tests to capture the
// command instead.
type Deployer struct {
	Helm   string                                                       // helm binary (default "helm")
	Runner func(ctx context.Context, name string, args ...string) error // injectable for tests
}

// Up generates r's chart and installs or upgrades it idempotently — safe to run
// on every push.
func (d Deployer) Up(ctx context.Context, r Release) error {
	chartDir, cleanup, err := d.generate(ctx, r)
	if err != nil {
		return err
	}
	defer cleanup()
	return d.run(ctx, upgradeArgs(chartDir, r)...)
}

// Render generates r's chart and prints the manifests it would apply (helm
// template). Needs no cluster.
func (d Deployer) Render(ctx context.Context, r Release) error {
	chartDir, cleanup, err := d.generate(ctx, r)
	if err != nil {
		return err
	}
	defer cleanup()
	return d.run(ctx, templateArgs(chartDir, r)...)
}

// generate validates r and writes its chart to a temp dir, returning the chart
// path and a cleanup func to remove it.
func (d Deployer) generate(ctx context.Context, r Release) (string, func(), error) {
	noop := func() {}

	err := r.validate()
	if err != nil {
		return "", noop, err
	}

	tmp, err := os.MkdirTemp("", "deploy-*")
	if err != nil {
		return "", noop, err
	}
	cleanup := func() { _ = os.RemoveAll(tmp) }

	chartDir, err := r.Service.Generate(ctx, tmp, r.ChartVersion)
	if err != nil {
		cleanup()
		return "", noop, err
	}
	return chartDir, cleanup, nil
}

// upgradeArgs builds the idempotent `helm upgrade --install` argument list.
func upgradeArgs(chartDir string, r Release) []string {
	return append([]string{"upgrade", "--install", r.Service.ChartName(), chartDir,
		"--namespace", r.Namespace, "--create-namespace"}, valueArgs(r)...)
}

// templateArgs builds a `helm template` argument list mirroring the deploy, so a
// render validates exactly what an install would apply — no cluster required.
func templateArgs(chartDir string, r Release) []string {
	return append([]string{"template", r.Service.ChartName(), chartDir,
		"--namespace", r.Namespace}, valueArgs(r)...)
}

// valueArgs is the shared -f/--set tail of both commands.
func valueArgs(r Release) []string {
	var args []string
	for _, f := range r.Values {
		args = append(args, "-f", f)
	}
	if r.ImageTag != "" {
		args = append(args, "--set", "image.tag="+r.ImageTag)
	}
	return args
}

func (d Deployer) helm() string {
	if d.Helm != "" {
		return d.Helm
	}
	return "helm"
}

func (d Deployer) run(ctx context.Context, args ...string) error {
	if d.Runner != nil {
		return d.Runner(ctx, d.helm(), args...)
	}

	cmd := exec.CommandContext(ctx, d.helm(), args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
