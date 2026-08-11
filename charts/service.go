package charts

import (
	"context"
	"fmt"
)

type O = map[string]string

const (
	FILE_CHART_YAML     = "Chart.yaml"
	FILE_VALUES_YAML    = "values.yaml"
	FILE_HELPERS_TPL    = "templates/_helpers.tpl"
	FILE_DEPLOYMENT     = "templates/deployment.yaml"
	FILE_SERVICE        = "templates/service.yaml"
	FILE_SERVICEACCOUNT = "templates/serviceaccount.yaml"
)

// Workload selects the Kubernetes controller a Service chart generates.
// The zero value is Deployment — the most common case — so a descriptor that
// omits Kind gets a Deployment for free.
type Workload int

const (
	Deployment  Workload = iota // stateless, rolling updates (default)
	StatefulSet                 // stable identity + ordered rollout
	DaemonSet                   // one pod per node
	CronJob                     // scheduled batch
	Job                         // one-shot batch
)

// Volume is an extra volume mounted into the container. If Secret is set it
// mounts that Secret; otherwise an emptyDir is used (e.g. a writable scratch
// path when the root filesystem is read-only).
type Volume struct {
	Name      string
	MountPath string
	Secret    string
	ReadOnly  bool
}

// Container describes the main application container
type Container struct {
	Image         string   // default image repository; tag comes from the deploy version
	Port          int      // container port (default 8080)
	Command       []string // overrides the image ENTRYPOINT
	Args          []string // overrides the image CMD
	Env           O        // literal environment variables
	Secrets       []string // Secret names exposed via envFrom
	ConfigMaps    []string // ConfigMap names exposed via envFrom
	Volumes       []Volume // extra volumes/mounts
	LivenessPath  string   // HTTP liveness path (default /healthz)
	ReadinessPath string   // HTTP readiness path (default /readyz)
}

// Service is the descriptor for a workload chart. Its zero value is hardened:
// read-only root filesystem, no privilege escalation, no mounted service-account
// token, all Linux capabilities dropped, runs as non-root. Set a field only to
// relax that baseline or add a feature.
type Service struct {
	Name        string
	Description string

	Kind                  Workload
	Replicas              int    // desired replicas (default 1, ignored for DaemonSet)
	RecreateDuringRollout bool   // Recreate/OnDelete instead of a rolling update
	Schedule              string // cron expression (CronJob only)

	Container Container

	// Security. Field names are chosen so the zero value is the hardened
	// default — set these true only to deliberately relax a control.
	WritableRootFilesystem   bool
	AllowPrivilegeEscalation bool
	MountServiceAccountToken bool

	// LegacySelectorLabels keeps old-style selector labels for adopting the
	// factory on top of a release that the hand-written chart already deployed
	// (a Deployment's selector is immutable, so it can't just change).
	LegacySelectorLabels bool

	// Blocks come in later steps: Ingress, Autoscale, PodDisruptionBudget,
	// NetworkPolicy, ServiceMonitor, and an Extra escape hatch. The struct
	// will grow — that's expected.
}

func (s Service) ChartName() string {
	return s.Name
}

func (s Service) Generate(ctx context.Context, dir, version string) (string, error) {
	name := s.Name
	err := validateName(name)
	if err != nil {
		return "", err
	}
	if version == "" {
		return "", fmt.Errorf("charts: '%s': version is required", name)
	}

	chartDir, err := createChartDir(dir, name)
	if err != nil {
		return "", fmt.Errorf("charts: %s: create dirs: %w", name, err)
	}

	err = writeFile(chartDir, FILE_CHART_YAML, s.chartYAML(version))
	if err != nil {
		return "", err
	}

	err = writeFile(chartDir, FILE_VALUES_YAML, s.valuesYAML())
	if err != nil {
		return "", err
	}

	err = writeFile(chartDir, FILE_HELPERS_TPL, s.helpersTPL())
	if err != nil {
		return "", err
	}

	err = writeFile(chartDir, FILE_DEPLOYMENT, s.deploymentYAML())
	if err != nil {
		return "", err
	}

	err = writeFile(chartDir, FILE_SERVICE, s.serviceYAML())
	if err != nil {
		return "", err
	}

	err = writeFile(chartDir, FILE_SERVICEACCOUNT, s.serviceaccountYAML())
	if err != nil {
		return "", err
	}

	return chartDir, nil
}
