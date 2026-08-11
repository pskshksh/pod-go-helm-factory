package deploy

import (
	"context"
	"reflect"
	"strings"
	"testing"

	"github.com/pskshksh/pod-go-helm-factory/charts"
	"github.com/pskshksh/pod-go-helm-factory/sdk/datastructures"
)

// deployCase is one deployment scenario: a Release plus the helm upgrade command
// it should produce. chartDir is fixed so the assertion is deterministic (the
// real Deployer generates it into a temp dir).
type deployCase struct {
	id       string
	chartDir string
	release  Release
	want     []string
}

// run builds the upgrade args for the case's release and checks them.
func (c deployCase) run(t *testing.T) {
	t.Helper()
	got := upgradeArgs(c.chartDir, c.release)
	if !reflect.DeepEqual(got, c.want) {
		t.Errorf("%s upgradeArgs =\n%v\nwant\n%v", c.id, got, c.want)
	}
}

func TestUpgradeArgs(t *testing.T) {
	cases := []deployCase{
		{
			id:       "full",
			chartDir: "dist/sampleapp",
			release: Release{
				Service:      charts.Service{Name: "sampleapp"},
				Namespace:    "staging",
				ChartVersion: "0.1.0",
				ImageTag:     "abc123",
				Values:       datastructures.StringList{"envs/staging.yaml"},
			},
			want: []string{
				"upgrade", "--install", "sampleapp", "dist/sampleapp",
				"--namespace", "staging", "--create-namespace",
				"-f", "envs/staging.yaml",
				"--set", "image.tag=abc123",
			},
		},
		{
			id:       "minimal",
			chartDir: "c",
			release: Release{
				Service:      charts.Service{Name: "app"},
				Namespace:    "prod",
				ChartVersion: "0.1.0",
			},
			want: []string{
				"upgrade", "--install", "app", "c",
				"--namespace", "prod", "--create-namespace",
			},
		},
	}

	for _, c := range cases {
		t.Run(c.id, func(t *testing.T) { c.run(t) })
	}
}

func TestTemplateArgs(t *testing.T) {
	r := Release{Service: charts.Service{Name: "app"}, Namespace: "prod", ImageTag: "v1"}

	want := []string{"template", "app", "c", "--namespace", "prod", "--set", "image.tag=v1"}
	got := templateArgs("c", r)
	if !reflect.DeepEqual(got, want) {
		t.Errorf("templateArgs =\n%v\nwant\n%v", got, want)
	}
}

// TestUpGeneratesAndRunsHelm confirms Up generates the chart and invokes helm
// upgrade --install for it, through the injectable runner.
func TestUpGeneratesAndRunsHelm(t *testing.T) {
	var gotName string
	var gotArgs []string
	d := Deployer{
		Runner: func(_ context.Context, name string, args ...string) error {
			gotName = name
			gotArgs = args
			return nil
		},
	}

	rel := Release{Service: charts.Service{Name: "app"}, Namespace: "ns", ChartVersion: "0.1.0"}
	err := d.Up(context.Background(), rel)
	if err != nil {
		t.Fatalf("Up: %v", err)
	}
	if gotName != "helm" {
		t.Errorf("ran %q, want helm", gotName)
	}

	joined := strings.Join(gotArgs, " ")
	if !strings.HasPrefix(joined, "upgrade --install app ") {
		t.Errorf("args = %v", gotArgs)
	}
	if !strings.Contains(joined, "--namespace ns --create-namespace") {
		t.Errorf("args = %v", gotArgs)
	}
}

// TestUpValidates confirms an incomplete release is rejected before helm runs.
func TestUpValidates(t *testing.T) {
	ran := false
	d := Deployer{Runner: func(context.Context, string, ...string) error { ran = true; return nil }}

	err := d.Up(context.Background(), Release{Namespace: "ns", ChartVersion: "0.1.0"}) // no Service
	if err == nil {
		t.Fatal("expected validation error for missing service")
	}
	if ran {
		t.Error("runner should not run when validation fails")
	}
}
