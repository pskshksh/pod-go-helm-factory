package charts

import (
	"context"
	"testing"

	"github.com/pskshksh/pod-go-helm-factory/sdk/render"
)

// securityContextCase renders a single securityContext block from a descriptor
// and compares it to a golden fragment file. It mirrors the table + run(t) shape
// of snapshotCase in generator_test.go, so adding a case is one struct literal.
type securityContextCase struct {
	file  string // fragment file under snapshots/fragments/
	svc   Service
	write func(s Service, y *render.YAML) // the pod- or container-level renderer
}

// run renders the case's block into a fresh encoder and asserts the output.
func (c securityContextCase) run(t *testing.T) {
	t.Helper()
	var y render.YAML
	c.write(c.svc, &y)
	assertFragment(t, y.String(), c.file)
}

func TestSecurityContexts(t *testing.T) {
	pod := func(s Service, y *render.YAML) { s.writePodSecurityContext(y, 0) }
	container := func(s Service, y *render.YAML) { s.writeContainerSecurityContext(y, 0) }

	cases := []securityContextCase{
		{file: "securitycontext_pod.yaml", svc: Service{}, write: pod},
		{file: "securitycontext_container_hardened.yaml", svc: Service{}, write: container},
		{
			file:  "securitycontext_container_relaxed.yaml",
			svc:   Service{AllowPrivilegeEscalation: true, WritableRootFilesystem: true},
			write: container,
		},
	}

	for _, c := range cases {
		t.Run(c.file, func(t *testing.T) { c.run(t) })
	}
}

// TestWriteEnv confirms literal env vars render sorted by key, and that an empty
// map emits nothing.
func TestWriteEnv(t *testing.T) {
	t.Run("empty", func(t *testing.T) {
		var y render.YAML
		Service{}.writeEnv(&y, 0)
		assertRender(t, y.String(), "")
	})

	t.Run("sorted_by_key", func(t *testing.T) {
		var y render.YAML
		Service{Container: Container{Env: O{"B_KEY": "2", "A_KEY": "1"}}}.writeEnv(&y, 0)
		assertFragment(t, y.String(), "env_sorted.yaml")
	})
}

// TestWriteEnvFrom confirms secretRef entries render before configMapRef ones,
// and that no source emits nothing.
func TestWriteEnvFrom(t *testing.T) {
	t.Run("empty", func(t *testing.T) {
		var y render.YAML
		Service{}.writeEnvFrom(&y, 0)
		assertRender(t, y.String(), "")
	})

	t.Run("secrets_then_configmaps", func(t *testing.T) {
		var y render.YAML
		Service{Container: Container{Secrets: []string{"s1"}, ConfigMaps: []string{"c1"}}}.writeEnvFrom(&y, 0)
		assertFragment(t, y.String(), "envfrom_secrets_configmaps.yaml")
	})
}

// TestNetworkPolicyYAML pins the default-deny variant (no ingress block, DNS-only
// egress); the same-namespace variant is covered by the blocks snapshot.
func TestNetworkPolicyYAML(t *testing.T) {
	svc := Service{Name: "api", NetworkPolicy: &NetworkPolicy{}}
	assertFragment(t, svc.networkPolicyYAML(), "networkpolicy_default.yaml")
}

// TestPodDisruptionBudgetYAML pins which single bound is emitted: default
// (minAvailable 1), an explicit percentage, and maxUnavailable.
func TestPodDisruptionBudgetYAML(t *testing.T) {
	cases := []struct {
		id   string
		file string
		pdb  PodDisruptionBudget
	}{
		{"default_min_available", "pdb_min_available.yaml", PodDisruptionBudget{}},
		{"percentage", "pdb_percentage.yaml", PodDisruptionBudget{MinAvailable: "50%"}},
		{"max_unavailable", "pdb_max_unavailable.yaml", PodDisruptionBudget{MaxUnavailable: "1"}},
	}

	for _, c := range cases {
		t.Run(c.id, func(t *testing.T) {
			svc := Service{Name: "api", PodDisruptionBudget: &c.pdb}
			assertFragment(t, svc.podDisruptionBudgetYAML(), c.file)
		})
	}
}

// TestPodDisruptionBudgetValidate rejects setting both bounds (forbidden by the
// Kubernetes API) via Generate.
func TestPodDisruptionBudgetValidate(t *testing.T) {
	svc := Service{
		Name:                "api",
		PodDisruptionBudget: &PodDisruptionBudget{MinAvailable: "1", MaxUnavailable: "1"},
	}

	_, err := svc.Generate(context.Background(), t.TempDir(), "0.1.0")
	if err == nil {
		t.Fatal("expected an error when both MinAvailable and MaxUnavailable are set")
	}
}

// TestAutoscaleYAML pins the HPA: defaults (min 1, cpu 80, no memory metric) and
// the memory-metric variant.
func TestAutoscaleYAML(t *testing.T) {
	cases := []struct {
		id   string
		file string
		auto Autoscale
	}{
		{"cpu_defaults", "hpa_cpu.yaml", Autoscale{MaxReplicas: 10}},
		{"cpu_and_memory", "hpa_cpu_memory.yaml", Autoscale{MaxReplicas: 10, TargetMemoryUtilization: 75}},
	}

	for _, c := range cases {
		t.Run(c.id, func(t *testing.T) {
			svc := Service{Name: "api", Autoscale: &c.auto}
			assertFragment(t, svc.autoscaleYAML(), c.file)
		})
	}
}

// TestAutoscaleValidate rejects a missing MaxReplicas and an inverted range.
func TestAutoscaleValidate(t *testing.T) {
	cases := []struct {
		id   string
		auto Autoscale
	}{
		{"max_missing", Autoscale{}},
		{"min_exceeds_max", Autoscale{MinReplicas: 5, MaxReplicas: 3}},
	}

	for _, c := range cases {
		t.Run(c.id, func(t *testing.T) {
			svc := Service{Name: "api", Autoscale: &c.auto}
			_, err := svc.Generate(context.Background(), t.TempDir(), "0.1.0")
			if err == nil {
				t.Fatal("expected a validation error")
			}
		})
	}
}

// TestIngressYAML pins both ingress modes: nginx (with and without TLS) and the
// Gateway API HTTPRoute.
func TestIngressYAML(t *testing.T) {
	cases := []struct {
		id   string
		file string
		ing  Ingress
	}{
		{"nginx", "ingress_nginx.yaml", Ingress{Host: "api.example.com"}},
		{"nginx_tls", "ingress_nginx_tls.yaml", Ingress{Host: "api.example.com", TLSSecret: "api-tls"}},
		{"httproute", "ingress_httproute.yaml", Ingress{Mode: IngressEnvoy, Host: "api.example.com", ClassName: "web-gateway"}},
	}

	for _, c := range cases {
		t.Run(c.id, func(t *testing.T) {
			svc := Service{Name: "api", Ingress: &c.ing}
			assertFragment(t, svc.ingressYAML(), c.file)
		})
	}
}

// TestIngressValidate rejects a missing host and Envoy mode without a Gateway.
func TestIngressValidate(t *testing.T) {
	cases := []struct {
		id  string
		ing Ingress
	}{
		{"host_missing", Ingress{}},
		{"envoy_without_gateway", Ingress{Mode: IngressEnvoy, Host: "api.example.com"}},
	}

	for _, c := range cases {
		t.Run(c.id, func(t *testing.T) {
			svc := Service{Name: "api", Ingress: &c.ing}
			_, err := svc.Generate(context.Background(), t.TempDir(), "0.1.0")
			if err == nil {
				t.Fatal("expected a validation error")
			}
		})
	}
}

// TestSecurityAccessors confirms the zero value is hardened and each field
// relaxes only its own control.
func TestSecurityAccessors(t *testing.T) {
	cases := []struct {
		id           string
		svc          Service
		readOnlyRoot bool
		automountSA  bool
	}{
		{"zero_value_hardened", Service{}, true, false},
		{"writable_root_fs", Service{WritableRootFilesystem: true}, false, false},
		{"mount_sa_token", Service{MountServiceAccountToken: true}, true, true},
	}

	for _, c := range cases {
		t.Run(c.id, func(t *testing.T) {
			readOnly := c.svc.readOnlyRootFilesystem()
			if readOnly != c.readOnlyRoot {
				t.Errorf("readOnlyRootFilesystem = %v, want %v", readOnly, c.readOnlyRoot)
			}

			automount := c.svc.automountSAToken()
			if automount != c.automountSA {
				t.Errorf("automountSAToken = %v, want %v", automount, c.automountSA)
			}

			uid := c.svc.runAsUser()
			if uid != DEFAULT_RUN_AS_UID {
				t.Errorf("runAsUser = %d, want %d", uid, DEFAULT_RUN_AS_UID)
			}
		})
	}
}
