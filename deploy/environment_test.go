package deploy

import (
	"context"
	"errors"
	"testing"

	"github.com/pskshksh/pod-go-helm-factory/charts"
)

// twoServices is a staging environment with two services, and a temp (empty)
// overlays root so no -f files are resolved.
func twoServices(t *testing.T) (charts.Environment, EnvOptions) {
	t.Helper()
	env := charts.Environment{
		Name: "staging",
		Services: []charts.Generator{
			charts.Service{Name: "api"},
			charts.Service{Name: "worker"},
		},
	}
	return env, EnvOptions{ChartVersion: "0.1.0", EnvsRoot: t.TempDir()}
}

// TestUpAllDeploysEachService confirms every service is deployed, in order, all
// into the environment's namespace.
func TestUpAllDeploysEachService(t *testing.T) {
	var releases, namespaces []string
	d := Deployer{Runner: func(_ context.Context, _ string, args ...string) error {
		releases = append(releases, args[2]) // upgrade --install <name> ...
		for i, a := range args {
			if a == "--namespace" {
				namespaces = append(namespaces, args[i+1])
			}
		}
		return nil
	}}

	env, opts := twoServices(t)
	err := d.UpAll(context.Background(), env, opts)
	if err != nil {
		t.Fatalf("UpAll: %v", err)
	}

	if len(releases) != 2 || releases[0] != "api" || releases[1] != "worker" {
		t.Errorf("deployed %v, want [api worker]", releases)
	}
	for _, ns := range namespaces {
		if ns != "staging" {
			t.Errorf("namespace %q, want staging", ns)
		}
	}
}

// TestUpAllNamespaceOverride confirms env.Namespace targets a different
// namespace than the config name, while overlays still key off the config name.
func TestUpAllNamespaceOverride(t *testing.T) {
	var namespaces []string
	d := Deployer{Runner: func(_ context.Context, _ string, args ...string) error {
		for i, a := range args {
			if a == "--namespace" {
				namespaces = append(namespaces, args[i+1])
			}
		}
		return nil
	}}

	env := charts.Environment{
		Name:      "prod", // config / overlay selector
		Namespace: "toto", // target namespace
		Services:  []charts.Generator{charts.Service{Name: "api"}},
	}
	err := d.UpAll(context.Background(), env, EnvOptions{ChartVersion: "0.1.0", EnvsRoot: t.TempDir()})
	if err != nil {
		t.Fatalf("UpAll: %v", err)
	}
	if len(namespaces) != 1 || namespaces[0] != "toto" {
		t.Errorf("namespaces = %v, want [toto]", namespaces)
	}
}

// TestUpAllStopsOnError confirms a failure aborts the loop before later services
// are touched.
func TestUpAllStopsOnError(t *testing.T) {
	calls := 0
	d := Deployer{Runner: func(context.Context, string, ...string) error {
		calls++
		return errors.New("boom")
	}}

	env, opts := twoServices(t)
	err := d.UpAll(context.Background(), env, opts)
	if err == nil {
		t.Fatal("expected an error")
	}
	if calls != 1 {
		t.Errorf("ran %d services, want 1 (stop on first failure)", calls)
	}
}
