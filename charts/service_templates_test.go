package charts

import (
	"testing"

	"github.com/pskshksh/pod-go-helm-factory/sdk/render"
)

// securityContextCase renders a single securityContext block from a descriptor
// and compares it to the expected YAML. It mirrors the table + run(t) shape of
// snapshotCase in generator_test.go, so adding a case is one struct literal.
type securityContextCase struct {
	id    string
	svc   Service
	write func(s Service, y *render.YAML) // the pod- or container-level renderer
	want  string
}

// run renders the case's block into a fresh encoder and asserts the output.
func (c securityContextCase) run(t *testing.T) {
	t.Helper()
	var y render.YAML
	c.write(c.svc, &y)
	assertRender(t, y.String(), c.want)
}

func TestSecurityContexts(t *testing.T) {
	pod := func(s Service, y *render.YAML) { s.writePodSecurityContext(y, 0) }
	container := func(s Service, y *render.YAML) { s.writeContainerSecurityContext(y, 0) }

	cases := []securityContextCase{
		{
			id:    "pod_hardened_default",
			svc:   Service{},
			write: pod,
			want: `securityContext:
  runAsNonRoot: true
  runAsUser: 10001
  runAsGroup: 10001
  fsGroup: 10001
  seccompProfile:
    type: RuntimeDefault
`,
		},
		{
			id:    "container_hardened_default",
			svc:   Service{},
			write: container,
			want: `securityContext:
  allowPrivilegeEscalation: false
  readOnlyRootFilesystem: true
  privileged: false
  capabilities:
    drop:
    - ALL
  seccompProfile:
    type: RuntimeDefault
`,
		},
		{
			id:    "container_relaxed",
			svc:   Service{AllowPrivilegeEscalation: true, WritableRootFilesystem: true},
			write: container,
			want: `securityContext:
  allowPrivilegeEscalation: true
  readOnlyRootFilesystem: false
  privileged: false
  capabilities:
    drop:
    - ALL
  seccompProfile:
    type: RuntimeDefault
`,
		},
	}

	for _, c := range cases {
		t.Run(c.id, func(t *testing.T) { c.run(t) })
	}
}

// TestWriteEnv confirms literal env vars render sorted by key, and that an empty
// map emits nothing.
func TestWriteEnv(t *testing.T) {
	cases := []struct {
		id   string
		svc  Service
		want string
	}{
		{"empty", Service{}, ""},
		{
			id:  "sorted_by_key",
			svc: Service{Container: Container{Env: O{"B_KEY": "2", "A_KEY": "1"}}},
			want: `env:
  - name: A_KEY
    value: "1"
  - name: B_KEY
    value: "2"
`,
		},
	}

	for _, c := range cases {
		t.Run(c.id, func(t *testing.T) {
			var y render.YAML
			c.svc.writeEnv(&y, 0)
			assertRender(t, y.String(), c.want)
		})
	}
}

// TestWriteEnvFrom confirms secretRef entries render before configMapRef ones,
// and that no source emits nothing.
func TestWriteEnvFrom(t *testing.T) {
	cases := []struct {
		id   string
		svc  Service
		want string
	}{
		{"empty", Service{}, ""},
		{
			id:  "secrets_then_configmaps",
			svc: Service{Container: Container{Secrets: []string{"s1"}, ConfigMaps: []string{"c1"}}},
			want: `envFrom:
  - secretRef:
      name: s1
  - configMapRef:
      name: c1
`,
		},
	}

	for _, c := range cases {
		t.Run(c.id, func(t *testing.T) {
			var y render.YAML
			c.svc.writeEnvFrom(&y, 0)
			assertRender(t, y.String(), c.want)
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
