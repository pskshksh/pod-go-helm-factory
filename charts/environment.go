package charts

// Environment is a named group of charts deployed together into one namespace.
// Each service is an independent Helm release; the deploy tooling loops them,
// resolving each one's per-environment value overlays.
//
// Name selects the value-overlay set (deploy/envs/<Name>/…), e.g. "prod" or
// "stage". Namespace is the target namespace; when empty it defaults to Name, so
// the common case (namespace == environment) needs only Name. Set Namespace to
// deploy a given config into a differently-named namespace — e.g. the "prod"
// config into namespace "toto".
type Environment struct {
	Name      string
	Namespace string
	Services  []Generator
}

// TargetNamespace is the namespace to deploy into: Namespace, or Name when
// Namespace is unset.
func (e Environment) TargetNamespace() string {
	if e.Namespace != "" {
		return e.Namespace
	}
	return e.Name
}
