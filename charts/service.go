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
	FILE_NETWORKPOLICY  = "templates/networkpolicy.yaml"
	FILE_PDB            = "templates/poddisruptionbudget.yaml"
	FILE_HPA            = "templates/hpa.yaml"
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

	// NetworkPolicy, when non-nil, generates a default-deny NetworkPolicy for
	// the workload (nil = no policy). See the NetworkPolicy type.
	NetworkPolicy *NetworkPolicy

	// PodDisruptionBudget, when non-nil, generates a PDB (nil = no PDB).
	PodDisruptionBudget *PodDisruptionBudget

	// Autoscale, when non-nil, generates a HorizontalPodAutoscaler and drops the
	// Deployment's static replica count so the HPA owns it (nil = no HPA).
	Autoscale *Autoscale

	// More blocks come in later steps: Ingress, ServiceMonitor, and an Extra
	// escape hatch. The struct will grow — that's expected.
}

// NetworkPolicy is an opt-in default-deny NetworkPolicy for the workload. When
// attached to a Service, all ingress and egress is denied by default. Egress to
// the cluster DNS is always allowed — a deny-all that broke name resolution
// would be a footgun — and AllowSameNamespace opens same-namespace traffic both
// directions.
type NetworkPolicy struct {
	// AllowSameNamespace permits ingress from, and egress to, other pods in the
	// same namespace. Cross-namespace and external traffic stay denied.
	AllowSameNamespace bool
}

// PodDisruptionBudget is an opt-in PDB that keeps a minimum number of pods
// running during voluntary disruptions (node drains, rollouts). Set exactly one
// of MinAvailable or MaxUnavailable; when both are empty it defaults to a
// MinAvailable of 1. Values are counts ("2") or percentages ("50%").
type PodDisruptionBudget struct {
	MinAvailable   string
	MaxUnavailable string
}

// validate rejects setting both bounds, which the Kubernetes API forbids.
func (p PodDisruptionBudget) validate() error {
	if p.MinAvailable != "" && p.MaxUnavailable != "" {
		return fmt.Errorf("charts: podDisruptionBudget: set only one of MinAvailable or MaxUnavailable")
	}
	return nil
}

// Autoscale is an opt-in HorizontalPodAutoscaler. MaxReplicas is required;
// MinReplicas defaults to 1 and TargetCPUUtilization to 80%. Set
// TargetMemoryUtilization to add a memory metric. When present, the Deployment
// omits its static replica count so the HPA is the sole owner of scale.
type Autoscale struct {
	MinReplicas             int
	MaxReplicas             int
	TargetCPUUtilization    int // percentage; default 80
	TargetMemoryUtilization int // percentage; 0 = no memory metric
}

// validate ensures the replica bounds form a usable range.
func (a Autoscale) validate() error {
	if a.MaxReplicas <= 0 {
		return fmt.Errorf("charts: autoscale: MaxReplicas must be greater than 0")
	}
	if a.MinReplicas > a.MaxReplicas {
		return fmt.Errorf("charts: autoscale: MinReplicas (%d) exceeds MaxReplicas (%d)", a.MinReplicas, a.MaxReplicas)
	}
	return nil
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
	if s.PodDisruptionBudget != nil {
		err = s.PodDisruptionBudget.validate()
		if err != nil {
			return "", err
		}
	}
	if s.Autoscale != nil {
		err = s.Autoscale.validate()
		if err != nil {
			return "", err
		}
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

	if s.NetworkPolicy != nil {
		err = writeFile(chartDir, FILE_NETWORKPOLICY, s.networkPolicyYAML())
		if err != nil {
			return "", err
		}
	}

	if s.PodDisruptionBudget != nil {
		err = writeFile(chartDir, FILE_PDB, s.podDisruptionBudgetYAML())
		if err != nil {
			return "", err
		}
	}

	if s.Autoscale != nil {
		err = writeFile(chartDir, FILE_HPA, s.autoscaleYAML())
		if err != nil {
			return "", err
		}
	}

	return chartDir, nil
}
