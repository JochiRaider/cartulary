package suiteservices

import (
	"errors"
	"fmt"
	"sort"
	"strings"
	"testing"

	"github.com/JochiRaider/cartulary/internal/testutil/testfailure"
)

var knownServiceDependencies = map[string]struct{}{
	"object_store": {},
	"postgres":     {},
}

type ServiceDependencyError struct {
	Service string
	Reason  string
}

func (e *ServiceDependencyError) Error() string {
	return fmt.Sprintf("harness service dependency declaration rejected: service=%s reason=%s", e.Service, e.Reason)
}

// CheckServiceDependencies validates a canonical graph assignment before any
// managed-service acquisition. An absent variable identifies a noncanonical
// direct developer invocation; the helper call's required list is then the
// explicit local declaration and no scheduler claim is inferred.
func CheckServiceDependencies(env map[string]string, required ...string) error {
	declaredValue, canonical := LookupEnv(env, HarnessServiceDependenciesEnv)
	if !canonical {
		return validateDependencyList(required, "required")
	}
	declared := []string{}
	if declaredValue != "" {
		declared = strings.Split(declaredValue, ",")
	}
	if err := validateDependencyList(declared, "declared"); err != nil {
		return err
	}
	if err := validateDependencyList(required, "required"); err != nil {
		return err
	}
	available := make(map[string]struct{}, len(declared))
	for _, service := range declared {
		available[service] = struct{}{}
	}
	for _, service := range required {
		if _, ok := available[service]; !ok {
			return &ServiceDependencyError{Service: service, Reason: "omitted"}
		}
	}
	return nil
}

func RequireServiceDependencies(t testing.TB, setupSource string, required ...string) {
	t.Helper()
	if err := CheckServiceDependencies(nil, required...); err != nil {
		service := required[0]
		var dependencyErr *ServiceDependencyError
		if errors.As(err, &dependencyErr) && dependencyErr.Service != "" {
			service = dependencyErr.Service
		}
		testfailure.Fail(t, testfailure.NewEnvelope(
			"harness",
			"fixture_error",
			setupSource,
			service,
			"dependency_guard",
			0,
			"not_required",
		))
	}
}

func validateDependencyList(values []string, label string) error {
	if !sort.StringsAreSorted(values) {
		return &ServiceDependencyError{Service: firstService(values), Reason: label + "_unsorted"}
	}
	seen := make(map[string]struct{}, len(values))
	for _, service := range values {
		if _, ok := knownServiceDependencies[service]; !ok {
			return &ServiceDependencyError{Service: firstService(values), Reason: label + "_unknown"}
		}
		if _, duplicate := seen[service]; duplicate {
			return &ServiceDependencyError{Service: service, Reason: label + "_duplicate"}
		}
		seen[service] = struct{}{}
	}
	return nil
}

func firstService(values []string) string {
	if len(values) == 0 {
		return "postgres"
	}
	if _, ok := knownServiceDependencies[values[0]]; ok {
		return values[0]
	}
	return "postgres"
}
