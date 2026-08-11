package charts

// Environment is a named group of charts deployed together into one namespace
// (the namespace takes the environment's name). Each service is an independent
// Helm release; the deploy tooling loops them, resolving each one's
// per-environment value overlays.
type Environment struct {
	Name     string
	Services []Generator
}
