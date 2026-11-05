package extensions

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"
	"time"
)

func verifyStartupDiagnosticProjection(t *testing.T) {
	c := requireGeneratedCoordinator(t)
	t.Run("storage failure is not a callback result", func(t *testing.T) {
		cause := errors.New("private SQL /private/database credential")
		calls := 0
		runtime := newNetworkFlowStateRuntime(t, failedAdmissionStore{cause}, &calls)
		resolution, err := c.ResolveClaims([]string{"import", "network_flow_activity"})
		if err != nil {
			t.Fatal(err)
		}
		err = runtime.ValidateClaimAdmission(t.Context(), c, resolution, "preflight", nil, nil)
		var storage *stateAccessError
		var callback *StartupValidationError
		if !errors.As(err, &storage) || !errors.Is(err, cause) || errors.As(err, &callback) || calls != 0 || strings.Contains(err.Error(), "private") {
			t.Fatalf("storage failure lost its boundary: %v", err)
		}
	})
	_, missing := NewCoordinator(nil)
	for _, tc := range []struct {
		name string
		err  error
		code string
	}{
		{"missing package", missing, "extension_registry_invalid"},
		{"malformed package", invalidArtifact("contracts/extensions/generated/descriptors/example.json", errors.New("private bytes /private/path SELECT secret")), "extension_registry_invalid"},
		{"limit", registryLimit("generator_sources", 1024, 1025), "extension_registry_limit_exceeded"},
		{"missing binding", unavailableBinding("network_flow_activity", 9), "extension_implementation_unavailable"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var failure *StartupValidationError
			var internal *ValidationError
			if !errors.As(tc.err, &failure) || !errors.As(tc.err, &internal) || len(failure.Findings) != 1 {
				t.Fatalf("missing typed diagnostic: %v", tc.err)
			}
			finding := failure.Findings[0]
			if finding.ReasonCode != tc.code || !c.validStartupFinding(finding) {
				t.Fatalf("invalid outward finding: %#v", finding)
			}
			encoded, _ := json.Marshal(failure.Findings)
			if strings.Contains(string(encoded), "private") || strings.Contains(string(encoded), "SELECT") {
				t.Fatal("parser content escaped")
			}
			if tc.name == "limit" && (finding.Details["limit"] != 1024 || finding.Details["actual"] != 1025) {
				t.Fatal("limit facts lost")
			}
		})
	}
	for _, descriptor := range c.Descriptors() {
		if descriptor.Claimable {
			continue
		}
		_, err := c.ResolveClaims([]string{descriptor.ProfileID})
		var failure *StartupValidationError
		if !errors.As(err, &failure) || len(failure.Findings) != 1 || failure.Findings[0].ReasonCode != "extension_profile_not_claimable" || !c.validStartupFinding(failure.Findings[0]) {
			t.Fatalf("unclaimable diagnostic: %#v", failure)
		}
	}
}

type failedAdmissionStore struct{ cause error }

func (s failedAdmissionStore) WithProfileSession(context.Context, string, time.Duration, func(ProfileSession) error) error {
	return s.cause
}
