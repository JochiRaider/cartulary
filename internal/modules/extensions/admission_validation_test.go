package extensions

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"
	"time"
)

func TestAdmissionAlgorithmResultAndDeadline_Unit(t *testing.T) {
	input := AdmissionValidationContext{Phase: "preflight"}
	for _, test := range []struct {
		name      string
		algorithm AdmissionAlgorithm
	}{
		{"missing", nil},
		{"panic", func(context.Context, AdmissionValidationContext) (AdmissionValidationResult, error) { panic("private") }},
		{"error", func(context.Context, AdmissionValidationContext) (AdmissionValidationResult, error) {
			return AdmissionValidationResult{}, errors.New("private")
		}},
		{"deadline", func(ctx context.Context, _ AdmissionValidationContext) (AdmissionValidationResult, error) {
			<-ctx.Done()
			return AdmissionValidationResult{}, ctx.Err()
		}},
	} {
		t.Run(test.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
			defer cancel()
			if _, err := invokeAdmissionAlgorithm(ctx, test.algorithm, input); err == nil {
				t.Fatal("invalid invocation accepted")
			}
		})
	}
}

func TestClaimAdmissionPreflightBeforeState_ServiceBacked(t *testing.T) {
	fixture := newNetworkFlowStateFixture(t, nil)
	calls := 0
	runtime := newNetworkFlowStateRuntime(t, fixture.store, &calls)
	coordinator := requireGeneratedCoordinator(t)
	resolution, err := coordinator.ResolveClaims([]string{"import", "network_flow_activity"})
	if err != nil {
		t.Fatal(err)
	}
	view, err := coordinator.ConfigurationView("network_flow_activity", []ProfileConfigurationValue{{Key: "network_flow_activity.key_ring_manifest_path", Source: "explicit", Value: map[string]any{"kind": "regular_file_ref", "key": "network_flow_activity.key_ring_manifest_path"}}})
	if err != nil {
		t.Fatal(err)
	}
	views := map[string]ProfileConfigurationView{"network_flow_activity": view}
	algorithmID := "network_flow_activity.saved_graph_cutover_v6"
	invocations := 0
	algorithm := func(_ context.Context, input AdmissionValidationContext) (AdmissionValidationResult, error) {
		invocations++
		raw, _ := json.Marshal(input)
		var members map[string]any
		_ = json.Unmarshal(raw, &members)
		if len(members) != 11 || input.StatePresent || input.StateMetadata != nil || len(input.MigrationDefinitions) != 1 || len(input.Configuration.Values) != 2 || input.Configuration.Values[1].Source != "default" || strings.Contains(string(raw), "/home/") {
			t.Errorf("incomplete or unsafe admission context: %s", raw)
		}
		return ValidAdmissionResult(input, algorithmID), nil
	}
	algorithms := map[string]AdmissionAlgorithm{algorithmID: algorithm}
	if err := runtime.ValidateClaimAdmission(context.Background(), coordinator, resolution.Claims(), "preflight", views, algorithms); err != nil {
		t.Fatal(err)
	}
	if invocations != 1 || calls != 0 {
		t.Fatalf("invocation count=%d state validators=%d", invocations, calls)
	}
	var count int
	if err := fixture.pool.QueryRow(context.Background(), `SELECT count(*) FROM extension_state_metadata`).Scan(&count); err != nil || count != 0 {
		t.Fatalf("preflight initialized state: %d %v", count, err)
	}
	for _, mode := range []string{"wrong phase", "wrong identity", "nil findings", "failure"} {
		t.Run(mode, func(t *testing.T) {
			algorithms[algorithmID] = func(_ context.Context, input AdmissionValidationContext) (AdmissionValidationResult, error) {
				result := ValidAdmissionResult(input, algorithmID)
				switch mode {
				case "wrong phase":
					result.Phase = "post_migration"
				case "wrong identity":
					result.AlgorithmID = "other"
				case "nil findings":
					result.Findings = nil
				case "failure":
					return result, errors.New("private retained data")
				}
				return result, nil
			}
			err := runtime.ValidateClaimAdmission(context.Background(), coordinator, resolution.Claims(), "preflight", views, algorithms)
			var failure *AdmissionValidationError
			if !errors.As(err, &failure) || len(failure.Findings) != 1 || strings.Contains(err.Error(), "private") {
				t.Fatalf("wrong admission failure: %v", err)
			}
			if err := fixture.pool.QueryRow(context.Background(), `SELECT count(*) FROM extension_state_metadata`).Scan(&count); err != nil || count != 0 {
				t.Fatal("failed preflight changed state")
			}
		})
	}
}
