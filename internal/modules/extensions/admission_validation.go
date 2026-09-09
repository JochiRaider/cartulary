package extensions

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"time"
)

// These closed logical values cross the admission boundary. Executable readers
// and resolved configuration material stay in the composition-owned callback.
type ProfileConfigurationValue struct {
	Key    string `json:"key"`
	Source string `json:"source"`
	Value  any    `json:"value"`
}
type ProfileConfigurationView struct {
	SchemaID                    string                      `json:"schema_id"`
	ProfileID                   string                      `json:"profile_id"`
	ConfigurationContractSHA256 string                      `json:"configuration_contract_sha256"`
	Values                      []ProfileConfigurationValue `json:"values"`
}
type AdmissionStateMetadata struct {
	SchemaID           string  `json:"schema_id"`
	ProfileID          string  `json:"profile_id"`
	MigrationLineageID string  `json:"migration_lineage_id"`
	StateVersion       int     `json:"state_version"`
	LastMigrationID    *string `json:"last_migration_id"`
	MetadataVersion    int     `json:"metadata_version"`
}
type AdmissionMigrationDigest struct {
	MigrationID string `json:"migration_id"`
	SHA256      string `json:"migration_definition_sha256"`
}
type AdmissionValidationContext struct {
	SchemaID             string                     `json:"schema_id"`
	Phase                string                     `json:"phase"`
	ProfileID            string                     `json:"profile_id"`
	DescriptorSHA256     string                     `json:"descriptor_sha256"`
	BindingSHA256        string                     `json:"binding_sha256"`
	Configuration        ProfileConfigurationView   `json:"profile_configuration_view"`
	StatePresent         bool                       `json:"state_present"`
	StateMetadata        *AdmissionStateMetadata    `json:"state_metadata"`
	MigrationDefinitions []AdmissionMigrationDigest `json:"migration_definition_digests"`
	ProbeID              *string                    `json:"probe_id"`
	TimeoutSeconds       int64                      `json:"timeout_seconds"`
}
type AdmissionFinding struct {
	Path       string         `json:"path"`
	ReasonCode string         `json:"reason_code"`
	Message    string         `json:"message"`
	Details    map[string]any `json:"details"`
}
type AdmissionValidationResult struct {
	SchemaID    string             `json:"schema_id"`
	Phase       string             `json:"phase"`
	AlgorithmID string             `json:"algorithm_id"`
	Findings    []AdmissionFinding `json:"findings"`
}
type AdmissionAlgorithm func(context.Context, AdmissionValidationContext) (AdmissionValidationResult, error)

type AdmissionValidationError struct{ Findings []AdmissionFinding }

func (e *AdmissionValidationError) Error() string { return "extension_admission_validation_failed" }

func ValidAdmissionResult(input AdmissionValidationContext, algorithm string) AdmissionValidationResult {
	return AdmissionValidationResult{SchemaID: "cartulary.extension_admission_validation_result.v1", Phase: input.Phase, AlgorithmID: algorithm, Findings: []AdmissionFinding{}}
}

// ConfigurationView binds a structurally admitted owner projection to the exact
// packaged configuration contract. A reference is a key-addressed handle, never
// a resolved path, secret, file, or connection value.
func (c *Coordinator) ConfigurationView(profile string, values []ProfileConfigurationValue) (ProfileConfigurationView, error) {
	record, ok := c.profiles[profile]
	if !ok {
		return ProfileConfigurationView{}, ErrStateIncomplete
	}
	keys, ok := objectSlice(record.configurationContract["keys"])
	if !ok {
		return ProfileConfigurationView{}, ErrStateIncomplete
	}
	indexed := map[string]ProfileConfigurationValue{}
	for _, value := range values {
		if _, duplicate := indexed[value.Key]; duplicate || (value.Source != "explicit" && value.Source != "default") {
			return ProfileConfigurationView{}, ErrStateIncomplete
		}
		indexed[value.Key] = value
	}
	normalized := make([]ProfileConfigurationValue, 0, len(keys))
	for _, key := range keys {
		name := stringValue(key["key"])
		value, present := indexed[name]
		omission, _ := key["omission_policy"].(map[string]any)
		if !present {
			if omission["kind"] == "absent" {
				continue
			}
			if omission["kind"] != "default" {
				return ProfileConfigurationView{}, ErrStateIncomplete
			}
			value = ProfileConfigurationValue{Key: name, Source: "default", Value: omission["value"]}
		}
		if value.Source == "default" && (omission["kind"] != "default" || !equalJSONValue(value.Value, omission["value"])) {
			return ProfileConfigurationView{}, ErrStateIncomplete
		}
		if kind := stringValue(key["resolution_kind"]); kind != "plain" {
			ref, ok := value.Value.(map[string]any)
			if !ok || len(ref) != 2 || ref["kind"] != kind || ref["key"] != name {
				return ProfileConfigurationView{}, ErrStateIncomplete
			}
		}
		normalized = append(normalized, value)
		delete(indexed, name)
	}
	if len(indexed) != 0 {
		return ProfileConfigurationView{}, ErrStateIncomplete
	}
	sort.Slice(normalized, func(i, j int) bool { return normalized[i].Key < normalized[j].Key })
	return ProfileConfigurationView{SchemaID: "cartulary.extension_profile_configuration_view.v1", ProfileID: profile, ConfigurationContractSHA256: record.configurationContractSHA256, Values: normalized}, nil
}
func equalJSONValue(a, b any) bool {
	left, e1 := json.Marshal(a)
	right, e2 := json.Marshal(b)
	return e1 == nil && e2 == nil && string(left) == string(right)
}

// ValidateClaimAdmission completes the all-profile phase before any migration.
// State is re-read again by Admit under the migration lock before any write.
func (r *StateRuntime) ValidateClaimAdmission(ctx context.Context, c *Coordinator, claims ResolvedClaimSet, phase string, views map[string]ProfileConfigurationView, algorithms map[string]AdmissionAlgorithm) error {
	if phase != "preflight" && phase != "post_migration" {
		return ErrStateIncomplete
	}
	failures := []AdmissionFinding{}
	for _, profile := range claims.ProfileIDs() {
		record := c.profiles[profile]
		algorithmKey := "preflight_algorithm_id"
		if phase == "post_migration" {
			algorithmKey = "post_migration_algorithm_id"
		}
		algorithmID := stringValue(record.bindingObject[algorithmKey])
		// Profiles without an additional owner algorithm keep their existing owner
		// configuration admission; extension-versioned state still preflights here.
		plan, hasState := c.StatePlan(profile)
		if !hasState && algorithmID == "" {
			continue
		}
		input := AdmissionValidationContext{SchemaID: "cartulary.extension_admission_validation_context.v1", Phase: phase, ProfileID: profile, DescriptorSHA256: record.descriptorSHA256, BindingSHA256: record.bindingSHA256, MigrationDefinitions: []AdmissionMigrationDigest{}, TimeoutSeconds: int64(r.validationTimeout / time.Second)}
		invocationCtx, cancel := context.WithTimeout(ctx, r.validationTimeout)
		err := func() error {
			if hasState {
				if err := r.preflightPackagedPlan(plan); err != nil {
					return err
				}
				families, err := stateFamilies(plan)
				if err != nil {
					return err
				}
				err = r.store.WithProfileSession(invocationCtx, profile, r.lockTimeout, func(session ProfileSession) error {
					snapshot, err := session.Snapshot(invocationCtx, profile, plan.MigrationLineageID, families)
					if err != nil {
						return err
					}
					present, err := authoritativeStatePresent(snapshot.FamilyCounts, families)
					if err != nil {
						return err
					}
					input.StatePresent = present
					decision, err := DecideStatePresence(plan.EmptyStatePolicy, snapshot.Metadata != nil, present)
					if err != nil {
						return err
					}
					if snapshot.Metadata != nil {
						m := snapshot.Metadata
						input.StateMetadata = &AdmissionStateMetadata{SchemaID: "cartulary.extension_state_metadata.v1", ProfileID: m.ProfileID, MigrationLineageID: m.MigrationLineageID, StateVersion: m.StateVersion, LastMigrationID: m.LastMigrationID, MetadataVersion: m.MetadataVersion}
					}
					if decision == StateValidate {
						if err := preflightStoredState(plan, snapshot); err != nil {
							return err
						}
						_, err = requiredMigrations(plan, snapshot.Metadata.StateVersion)
						return err
					}
					return nil
				})
				if err != nil {
					return err
				}
				for _, migration := range plan.MigrationDefinitions {
					input.MigrationDefinitions = append(input.MigrationDefinitions, AdmissionMigrationDigest{MigrationID: migration.MigrationID, SHA256: migration.DefinitionSHA256})
				}
				sort.Slice(input.MigrationDefinitions, func(i, j int) bool {
					return input.MigrationDefinitions[i].MigrationID < input.MigrationDefinitions[j].MigrationID
				})
			}
			if algorithmID == "" {
				return nil
			}
			view, ok := views[profile]
			if !ok {
				return ErrStateIncomplete
			}
			checked, err := c.ConfigurationView(profile, view.Values)
			if err != nil || !equalJSONValue(view, checked) {
				return ErrStateIncomplete
			}
			input.Configuration = checked
			result, err := invokeAdmissionAlgorithm(invocationCtx, algorithms[algorithmID], input)
			if err != nil {
				return err
			}
			if result.SchemaID != "cartulary.extension_admission_validation_result.v1" || result.Phase != phase || result.AlgorithmID != algorithmID || result.Findings == nil || len(result.Findings) > 4096 {
				return ErrStateValidationFailed
			}
			expectedFinding := admissionFailure(input, algorithmID, false)
			for _, finding := range result.Findings {
				if !equalJSONValue(finding, expectedFinding) {
					return ErrStateValidationFailed
				}
			}

			failures = append(failures, result.Findings...)
			return nil
		}()
		timedOut := errors.Is(invocationCtx.Err(), context.DeadlineExceeded)
		cancel()
		if err != nil {
			failures = append(failures, admissionFailure(input, algorithmID, timedOut))
		}
		if phase == "post_migration" && len(failures) > 0 {
			break
		}
	}
	if len(failures) > 0 {
		sort.Slice(failures, func(i, j int) bool {
			left, _ := canonicalJSON(failures[i].Details, false)
			right, _ := canonicalJSON(failures[j].Details, false)
			return string(left) < string(right)
		})
		return &AdmissionValidationError{Findings: failures}
	}
	return nil
}

func invokeAdmissionAlgorithm(ctx context.Context, algorithm AdmissionAlgorithm, input AdmissionValidationContext) (AdmissionValidationResult, error) {
	if algorithm == nil {
		return AdmissionValidationResult{}, ErrStateValidationFailed
	}
	type outcome struct {
		result AdmissionValidationResult
		err    error
	}
	done := make(chan outcome, 1)
	go func() {
		var result AdmissionValidationResult
		var err error
		defer func() {
			if recover() != nil {
				err = fmt.Errorf("admission algorithm failed")
			}
			done <- outcome{result, err}
		}()
		result, err = algorithm(ctx, input)
	}()
	select {
	case <-ctx.Done():
		return AdmissionValidationResult{}, ctx.Err()
	case out := <-done:
		return out.result, out.err
	}
}

func admissionFailure(input AdmissionValidationContext, algorithmID string, timedOut bool) AdmissionFinding {
	phase := "profile_preflight"
	if input.Phase == "post_migration" {
		phase = "post_migration_validation"
	}
	return AdmissionFinding{Path: "$", ReasonCode: "extension_admission_validation_failed", Message: "Extension admission validation failed.", Details: map[string]any{"profile_id": input.ProfileID, "phase": phase, "algorithm_id": algorithmID, "timed_out": timedOut, "timeout_seconds": input.TimeoutSeconds}}
}
