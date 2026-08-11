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
	FILE_STATEFULSET    = "templates/statefulset.yaml"
	FILE_DAEMONSET      = "templates/daemonset.yaml"
	FILE_JOB            = "templates/job.yaml"
	FILE_CRONJOB        = "templates/cronjob.yaml"
	FILE_SERVICE        = "templates/service.yaml"
	FILE_SERVICEACCOUNT = "templates/serviceaccount.yaml"
	FILE_NETWORKPOLICY  = "templates/networkpolicy.yaml"
	FILE_PDB            = "templates/poddisruptionbudget.yaml"
	FILE_HPA            = "templates/hpa.yaml"
	FILE_INGRESS        = "templates/ingress.yaml"
	FILE_SERVICEMONITOR = "templates/servicemonitor.yaml"
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

	// Ingress, when non-nil, exposes the Service externally (nil = no route).
	Ingress *Ingress

	// ServiceMonitor, when non-nil, generates a Prometheus Operator
	// ServiceMonitor scraping the Service (nil = no ServiceMonitor).
	ServiceMonitor *ServiceMonitor

	// An Extra escape hatch comes in a later step. The struct will grow — that's
	// expected.
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

// IngressMode selects which ingress mechanism the Ingress block generates.
type IngressMode int

const (
	IngressNginx IngressMode = iota // networking.k8s.io/v1 Ingress, nginx class (default)
	IngressEnvoy                    // Gateway API HTTPRoute (Envoy Gateway)
)

// Ingress is an opt-in external route to the workload's Service. In Nginx mode
// it generates a networking.k8s.io/v1 Ingress (with optional TLS); in Envoy mode
// a Gateway API HTTPRoute attached to the Gateway named by ClassName.
type Ingress struct {
	Mode      IngressMode
	Host      string // required — the external hostname
	Path      string // default "/"
	ClassName string // Nginx: ingressClassName (default "nginx"). Envoy: parent Gateway name (required).
	TLSSecret string // Nginx only: terminate TLS using this Secret.
}

// validate ensures a routable host and, for Envoy mode, a parent Gateway.
func (i Ingress) validate() error {
	if i.Host == "" {
		return fmt.Errorf("charts: ingress: Host is required")
	}
	if i.Mode == IngressEnvoy && i.ClassName == "" {
		return fmt.Errorf("charts: ingress: Envoy mode requires ClassName (the parent Gateway name)")
	}
	return nil
}

// ServiceMonitor is an opt-in Prometheus Operator ServiceMonitor that scrapes
// the workload's Service. All fields default: Port "http", Path "/metrics",
// Interval "30s".
type ServiceMonitor struct {
	Port     string // the Service port name to scrape
	Path     string // the metrics path
	Interval string // the scrape interval
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
	if s.Kind == CronJob && s.Schedule == "" {
		return "", fmt.Errorf("charts: '%s': CronJob requires Schedule", name)
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
	if s.Ingress != nil {
		err = s.Ingress.validate()
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

	err = writeFile(chartDir, s.workloadFile(), s.workloadYAML())
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

	if s.Ingress != nil {
		err = writeFile(chartDir, FILE_INGRESS, s.ingressYAML())
		if err != nil {
			return "", err
		}
	}

	if s.ServiceMonitor != nil {
		err = writeFile(chartDir, FILE_SERVICEMONITOR, s.serviceMonitorYAML())
		if err != nil {
			return "", err
		}
	}

	return chartDir, nil
}
