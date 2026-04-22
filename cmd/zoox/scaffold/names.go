package scaffold

import (
	"fmt"
	"regexp"
	"strings"
)

var moduleNameRe = regexp.MustCompile(`^[a-z][a-z0-9]*$`)

// ValidateModuleSegment checks a single-path-segment module name (e.g. "user", "order").
func ValidateModuleSegment(name string) error {
	if name == "" {
		return fmt.Errorf("module name is required")
	}
	if !moduleNameRe.MatchString(name) {
		return fmt.Errorf("module name %q must match [a-z][a-z0-9]*", name)
	}
	return nil
}

// Exported returns the exported Go identifier prefix for the module (e.g. "user" -> "User").
func Exported(name string) string {
	if name == "" {
		return ""
	}
	return strings.ToUpper(name[:1]) + name[1:]
}

// ResourcePath returns a simple REST-style path segment (e.g. "user" -> "users").
func ResourcePath(name string) string {
	if name == "" {
		return ""
	}
	if strings.HasSuffix(name, "s") {
		return name
	}
	return name + "s"
}

// ImportAlias is the Go import alias for an API package (e.g. v1 + user -> v1User).
func ImportAlias(apiVersion, name string) string {
	return apiVersion + Exported(name)
}
