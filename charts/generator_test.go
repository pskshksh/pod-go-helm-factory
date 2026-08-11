package charts

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

// snapshotCase pairs a chart descriptor with the snapshot file that captures
// its expected generated output.
type snapshotCase struct {
	id      string // name of the file in snapshots/ (e.g. "service_api.yaml")
	version string
	gen     Generator // the chart under test
}

// run generates the chart on the fly and compares it to snapshots/<id>.
func (c snapshotCase) run(t *testing.T) {
	t.Helper()
	tmp := t.TempDir()

	chartDir, err := c.gen.Generate(context.Background(), tmp, c.version)
	if err != nil {
		t.Fatalf("generate %s: %v", c.id, err)
	}

	gotPath := filepath.Join(tmp, c.id)
	bundleChart(t, chartDir, gotPath)

	snapPath := filepath.Join("snapshots", c.id)
	if *updateSnapshots {
		data, err := os.ReadFile(gotPath)
		if err != nil {
			t.Fatal(err)
		}
		err = writeFile("snapshots", c.id, string(data))
		if err != nil {
			t.Fatal(err)
		}
		t.Logf("updated snapshot %s", snapPath)
		return
	}

	compareFiles(t, gotPath, snapPath)
}

func TestGenerateSnapshots(t *testing.T) {
	cases := []snapshotCase{
		{
			id:      "service_api.yaml",
			version: "0.1.0",
			gen: Service{
				Name:        "api",
				Description: "Demo API service",
			},
		},
		{
			// Every opt-in block enabled, so one snapshot covers that all the
			// gated files are generated and wired together.
			id:      "service_api_blocks.yaml",
			version: "0.1.0",
			gen: Service{
				Name:                "api",
				Description:         "Demo API service",
				NetworkPolicy:       &NetworkPolicy{AllowSameNamespace: true},
				PodDisruptionBudget: &PodDisruptionBudget{},
			},
		},
	}

	for _, c := range cases {
		t.Run(c.id, func(t *testing.T) { c.run(t) })
	}
}
