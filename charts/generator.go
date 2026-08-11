package charts

import (
	"context"
)

// Generator is implemented by every chart descriptor (Service, and later
// Database). Generate writes a complete Helm chart for the descriptor into
// dir/<ChartName>/, stamped with version, and returns the chart directory path.
type Generator interface {
	// ChartName is the chart's name and its default Helm release name.
	ChartName() string
	// Generate writes the chart directory under dir and returns its path.
	Generate(ctx context.Context, dir, version string) (string, error)
}
