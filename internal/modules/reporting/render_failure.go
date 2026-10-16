package reporting

import (
	"bytes"
	"encoding/json"
	"errors"
	"slices"
	"strings"

	"github.com/JochiRaider/cartulary/internal/platform/renderprocess"
)

// RenderFailure carries only owner-classified diagnostics across the terminal
// transaction boundary. Store binds it to independently retained admission.
type RenderFailure struct {
	Stage       string
	FailureCode string
	ReasonCode  string
	SafeDetails map[string]any
	issues      *renderIssueSet
}

var renderFailureStages = map[string]string{
	"export_model_invalid": "export_model_materialization", "export_model_resource_limit_exceeded": "export_model_materialization",
	"content_class_missing": "export_model_materialization", "support_ref_limit_exceeded": "export_model_materialization",
	"composition_invalid": "composition_validation", "redaction_manifest_invalid": "redaction", "token_manifest_invalid": "token_manifest",
	"graph_projection_unavailable": "diagram_model", "mermaid_invalid": "mermaid_source", "mermaid_render_failed": "mermaid_render",
	"deck_model_invalid": "deck_model", "slidev_source_invalid": "slidev_source", "slidev_export_failed": "slidev_export",
	"undeclared_asset": "bundle_manifest", "remote_asset_reference": "security_sandbox", "bundle_manifest_invalid": "bundle_manifest",
	"bundle_resource_limit_exceeded": "bundle_manifest", "asset_limit_exceeded": "bundle_manifest",
	"nondeterministic_render": "determinism", "security_sandbox_violation": "security_sandbox",
	"render_canceled": "release_state", "render_timeout": "release_state",
}

// These are the current emitted owner reason families. Generic object failures
// retain their stage-specific code; no diagnostic is inferred from error text.
var renderFailureReasons = map[string]string{
	"export_model_invalid":                 "required_reference_pack_unavailable unsupported_derivation_algorithm unsupported_source_family live_query_after_export_model duplicate_object_member duplicate_stable_id duplicate_export_model_path dangling_source_ref dangling_support_ref invalid_support_ref snapshot_ref_mismatch invalid_asset_declaration invalid_generated_identifier blocked_core_dependency duplicate_active_party_assignment timeline_sort_key_missing invalid_timestamp_value deleted_record_not_releasable timeline_overflow_unresolved required_section_empty empty_export_model_sections narrative_slot_unbound invalid_multiline_value",
	"export_model_resource_limit_exceeded": "export_model_canonical_bytes_exceeded sections_count_exceeded blocks_count_exceeded fields_count_exceeded records_count_exceeded relationships_count_exceeded diagrams_count_exceeded",
	"content_class_missing":                "content_class_missing",
	"composition_invalid":                  "composition_digest_mismatch composition_template_mismatch composition_anchor_unresolved composition_anchor_ambiguous composition_drop_invalid_for_external_release composition_duplicate_diagram_insert composition_replacement_target_missing authored_text_not_permitted authored_subject_ref_unresolved authored_title_limit_exceeded authored_text_limit_exceeded diagram_label_override_invalid diagram_layout_invalid diagram_layout_missing_node_position diagram_layout_duplicate_target diagram_layout_unknown_target manual_layout_not_supported_for_output_kind template_static_future_only invalid_mermaid_construct",
	"redaction_manifest_invalid":           "redaction_profile_invalid disclosure_partition_unresolved redaction_action_unresolved invalid_recipient_partition_ref unknown_recipient_partition recipient_partition_profile_mismatch",
	"token_manifest_invalid":               "token_subject_not_stable token_subject_not_unique subject_limit_exceeded",
	"graph_projection_unavailable":         "graph_projection_stale graph_projection_not_completed graph_projection_not_bound graph_projection_ambiguous graph_projection_selection_unresolved graph_projection_digest_mismatch diagram_selection_missing_ref diagram_selection_duplicate_ref",
	"mermaid_invalid":                      "diagram_hard_limit_exceeded invalid_mermaid_construct label_length_exceeded no_diagrams_selected",
	"mermaid_render_failed":                "renderer_process_failed",
	"deck_model_invalid":                   "block_kind_not_allowed deck_title_invalid",
	"slidev_source_invalid":                "slide_count_exceeded slide_block_limit_exceeded speaker_notes_limit_exceeded click_step_limit_exceeded list_nesting_exceeded invalid_speaker_note_sequence invalid_field_value_for_text invalid_slidev_construct",
	"slidev_export_failed":                 "renderer_process_failed click_export_page_count_mismatch",
	"undeclared_asset":                     "undeclared_asset_ref",
	"remote_asset_reference":               "remote_asset_reference_invalid outbound_request_observed",
	"bundle_manifest_invalid":              "invalid_template_binding_pattern invalid_asset_media_type atomic_publish_failed bundle_manifest_self_reference",
	"bundle_resource_limit_exceeded":       "rendered_pdf_bytes_exceeded asset_limit_exceeded",
	"asset_limit_exceeded":                 "asset_limit_exceeded local_asset_count_exceeded local_asset_bytes_exceeded local_assets_total_bytes_exceeded",
	"support_ref_limit_exceeded":           "support_ref_limit_exceeded",
	"nondeterministic_render":              "export_model_mismatch redaction_manifest_mismatch token_manifest_mismatch toolchain_snapshot_mismatch validation_summary_mismatch bundle_manifest_mismatch bundle_file_mismatch",
	"security_sandbox_violation":           "external_reference_in_rendered_output partial_output_cleanup_failed loopback_not_declared sandbox_observation_missing render_runtime_unavailable",
	"render_canceled":                      "render_canceled",
	"render_timeout":                       "render_timeout",
}

const renderStages = "route_admission composition_validation export_model_materialization redaction token_manifest diagram_model mermaid_source mermaid_render deck_model slidev_source slidev_export bundle_manifest determinism security_sandbox release_state"
const genericRenderReasons = "invalid_schema_id required_member_missing unknown_object_member invalid_null invalid_type invalid_scalar_value array_order_invalid duplicate_array_item"
const schemaRenderFailureCodes = "export_model_invalid composition_invalid redaction_manifest_invalid token_manifest_invalid mermaid_invalid deck_model_invalid slidev_source_invalid bundle_manifest_invalid"

func renderFailureDetails(cause error) RenderFailure {
	var classified *renderValidationError
	if errors.Is(cause, renderprocess.ErrCleanup) || !errors.As(cause, &classified) {
		classified = classifyRenderRuntimeError(cause)
	}
	f := RenderFailure{Stage: classified.Stage, FailureCode: classified.FailureCode, ReasonCode: classified.ReasonCode, SafeDetails: classified.SafeDetails, issues: classified.issues}
	// A malformed template cannot be admitted normally. If retained input is
	// invalid nonetheless, it is a failed bundle input, not a route rejection.
	if f.FailureCode == "invalid_release_request" {
		f.FailureCode, f.ReasonCode = "bundle_manifest_invalid", "invalid_scalar_value"
	}
	if f.Stage == "" {
		f.Stage = renderFailureStages[f.FailureCode]
	}
	if f.SafeDetails == nil {
		f.SafeDetails = map[string]any{}
	}
	return f
}

func validateRenderFailure(f RenderFailure) error {
	invalid := errors.New("reporting: invalid render failure diagnostic")
	if renderFailureStages[f.FailureCode] == "" || !slices.Contains(strings.Fields(renderStages), f.Stage) ||
		(!slices.Contains(strings.Fields(renderFailureReasons[f.FailureCode]), f.ReasonCode) &&
			!(slices.Contains(strings.Fields(schemaRenderFailureCodes), f.FailureCode) && slices.Contains(strings.Fields(genericRenderReasons), f.ReasonCode))) || f.SafeDetails == nil {
		return invalid
	}
	// This producer emits counts, resource keys and closed timer tokens only.
	// Never copy raw error text or merely filter unknown keys into acceptance.
	for key, value := range f.SafeDetails {
		switch key {
		case "stage":
			if value != f.Stage {
				return invalid
			}
		case "timeout_kind":
			if value != "stage" && value != "total" {
				return invalid
			}
		case "limit_key":
			name, ok := value.(string)
			if !ok {
				return invalid
			}
			if _, ok := renderLimitRules[name]; !ok {
				return invalid
			}
		case "limit", "actual":
			raw, err := json.Marshal(value)
			if err != nil || len(raw) == 0 {
				return invalid
			}
			for _, digit := range raw {
				if digit < '0' || digit > '9' {
					return invalid
				}
			}
		default:
			return invalid
		}
	}
	return nil
}

func failedRenderSummary(input renderInput, failure RenderFailure) ([]byte, error) {
	if err := validateRenderFailure(failure); err != nil {
		return nil, err
	}
	identity := input.ExportModel.RenderIdentity
	if err := identity.validate(); err != nil {
		return nil, err
	}
	issue := reportingRenderValidationIssue{
		Stage: failure.Stage, Severity: "error", FailureCode: &failure.FailureCode,
		ReasonCode: &failure.ReasonCode, SafeDetails: failure.SafeDetails,
	}
	issues, count := []reportingRenderValidationIssue{issue}, 1
	if failure.issues != nil {
		if err := failure.issues.validate(failure); err != nil {
			return nil, err
		}
		issues, count = failure.issues.retained, failure.issues.count
	}
	details := map[string]any{}
	for key, value := range failure.SafeDetails {
		details[key] = value
	}
	if identity.ReleaseID == nil {
		details["pre_release"] = true
	}
	summary := reportingRenderValidationSummary{
		SchemaID: "cartulary.reporting_render_validation_summary.v1", ReleaseID: identity.ReleaseID, SnapshotID: &input.SnapshotID,
		CreatedAt: identity.RenderAdmittedAt, Result: "failed", TerminalStage: failure.Stage,
		FailureCode: &failure.FailureCode, ReasonCode: &failure.ReasonCode, SafeDetails: details,
		IssueCount: count, Issues: issues,
		FirstFailure: &reportingRenderFirstFailure{SchemaID: "first_failure.v1", reportingRenderValidationIssue: issue},
	}
	return canonicalJSON(summary)
}

// Reconstruct only the shape and admission binding from retained diagnostics.
// This is validation of existing bytes, never inference of missing history.
func validateFailedRenderSummary(raw []byte, input renderInput, reason string) error {
	invalid := errors.New("reporting: retained render failure differs from admission")
	var summary reportingRenderValidationSummary
	if len(raw) == 0 || len(raw) > 1<<20 || json.Unmarshal(raw, &summary) != nil || len(summary.Issues) == 0 || summary.ReasonCode == nil || summary.FailureCode == nil || *summary.ReasonCode != reason {
		return invalid
	}
	issue := summary.Issues[0]
	failure := RenderFailure{Stage: summary.TerminalStage, FailureCode: *summary.FailureCode, ReasonCode: *summary.ReasonCode, SafeDetails: issue.SafeDetails}
	failure.issues = &renderIssueSet{count: summary.IssueCount, retained: summary.Issues}
	expected, err := failedRenderSummary(input, failure)
	if err != nil || !bytes.Equal(expected, raw) {
		return invalid
	}
	return nil
}

// Internal errors never supply public diagnostic text. Preserve an existing
// typed owner failure, otherwise classify at the boundary that rejected the
// input or result. The original cause remains available only inside the worker.
func renderFailureAt(err error, stage, failureCode, reasonCode string) *renderValidationError {
	var existing *renderValidationError
	if errors.As(err, &existing) {
		copy := *existing
		if copy.Stage == "" {
			copy.Stage = stage
		}
		return &copy
	}
	return &renderValidationError{Stage: stage, FailureCode: failureCode, ReasonCode: reasonCode, cause: err}
}

func classifyBundleFailure(err error) *renderValidationError {
	switch {
	case errors.Is(err, ErrRemoteRuntimeAsset):
		return renderFailureAt(err, "security_sandbox", "remote_asset_reference", "remote_asset_reference_invalid")
	case errors.Is(err, ErrUndeclaredTemplateBinding):
		return renderFailureAt(err, "deck_model", "bundle_manifest_invalid", "invalid_template_binding_pattern")
	case errors.Is(err, ErrMissingRequiredField):
		return renderFailureAt(err, "deck_model", "export_model_invalid", "required_member_missing")
	default:
		var failure *renderValidationError
		if errors.As(err, &failure) {
			return renderFailureAt(err, renderFailureStages[failure.FailureCode], "bundle_manifest_invalid", "invalid_type")
		}
		return renderFailureAt(err, "bundle_manifest", "bundle_manifest_invalid", "invalid_type")
	}
}

// Reference Data's closed missing-pin outcome survives background preparation.
func renderFailureJobCode(reason string) string {
	if reason == "required_reference_pack_unavailable" {
		return reason
	}
	return reportingJobRenderFailedCode
}
