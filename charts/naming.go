package charts

import (
	"fmt"
	"regexp"
)

// nameRE is the RFC 1123 label rule Kubernetes/Helm use for resource names.
var nameRE = regexp.MustCompile(`^[a-z0-9]([-a-z0-9]*[a-z0-9])?$`)

// validateName enforces RFC 1123: lowercase alphanumerics and '-', starting
// and ending alphanumeric, at most 63 characters.
func validateName(name string) error {
	if name == "" {
		return fmt.Errorf("charts: name is required")
	}
	if len(name) > 63 || !nameRE.MatchString(name) {
		return fmt.Errorf("charts: invalid name %q (must be a lowercase RFC 1123 label)", name)
	}
	return nil
}
