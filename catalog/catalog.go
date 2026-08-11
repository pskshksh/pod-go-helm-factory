// Package catalog is the source of truth for the services this repo builds and
// deploys. Both cmd/factory (generate charts) and cmd/deploy (generate + deploy)
// read it, so the two never drift.
package catalog

import "github.com/pskshksh/pod-go-helm-factory/charts"

// Services returns every chart descriptor the repo manages.
func Services() []charts.Generator {
	return []charts.Generator{
		charts.Service{
			Name:        "sampleapp",
			Description: "Sample HTTP service",
			Container: charts.Container{
				Image: "ghcr.io/pskshksh/pod-go-helm-factory/sampleapp",
			},
			NetworkPolicy:       &charts.NetworkPolicy{AllowSameNamespace: true},
			PodDisruptionBudget: &charts.PodDisruptionBudget{},
		},
	}
}

// Find returns the service whose chart name matches name, or ok=false.
func Find(name string) (charts.Generator, bool) {
	for _, s := range Services() {
		if s.ChartName() == name {
			return s, true
		}
	}
	return nil, false
}
