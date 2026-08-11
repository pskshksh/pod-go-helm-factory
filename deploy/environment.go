package deploy

import (
	"context"
	"fmt"

	"github.com/pskshksh/pod-go-helm-factory/charts"
	"github.com/pskshksh/pod-go-helm-factory/sdk/datastructures"
)

// EnvOptions are the deploy-time parameters shared by every service in an
// environment. Per-service value overlays are resolved from EnvsRoot.
type EnvOptions struct {
	ChartVersion string
	ImageTag     string
	EnvsRoot     string
	ExtraValues  datastructures.StringList
}

// release builds the Release for deploying svc into env: it targets the env's
// namespace and resolves the service's overlays from the env's config name
// (deploy/envs/<env.Name>/<service>.yaml), then appends any extra values. Note
// the namespace and the overlay set are separate — env.Name picks the config,
// env.TargetNamespace() picks where it lands.
func release(svc charts.Generator, env charts.Environment, opts EnvOptions) Release {
	values := Overlays(opts.EnvsRoot, env.Name, svc.ChartName())
	values = append(values, opts.ExtraValues...)
	return Release{
		Service:      svc,
		Namespace:    env.TargetNamespace(),
		ChartVersion: opts.ChartVersion,
		ImageTag:     opts.ImageTag,
		Values:       values,
	}
}

// UpAll installs or upgrades every service in env into its target namespace,
// stopping at the first failure.
func (d Deployer) UpAll(ctx context.Context, env charts.Environment, opts EnvOptions) error {
	for _, svc := range env.Services {
		err := d.Up(ctx, release(svc, env, opts))
		if err != nil {
			return fmt.Errorf("deploy %s into %s: %w", svc.ChartName(), env.TargetNamespace(), err)
		}
	}
	return nil
}

// RenderAll renders every service in env (helm template), stopping at the first
// failure. Needs no cluster.
func (d Deployer) RenderAll(ctx context.Context, env charts.Environment, opts EnvOptions) error {
	for _, svc := range env.Services {
		err := d.Render(ctx, release(svc, env, opts))
		if err != nil {
			return fmt.Errorf("render %s: %w", svc.ChartName(), err)
		}
	}
	return nil
}
