package reporting

import (
	"bytes"
	"encoding/json"
	"errors"
)

// A passing artifact is a closed, admission-bound statement. A matching digest
// cannot make a missing identity, an unknown diagnostic, or a failed summary
// eligible for publication. Failure summaries are retained outside the bundle.
func validatePassedRenderSummary(raw []byte, manifest RenderBundleManifest) error {
	invalid := errors.New("reporting: invalid passing render validation summary")
	var summary reportingRenderValidationSummary
	canonical, err := canonicalReportingBytes(raw)
	if err != nil || !bytes.Equal(canonical, raw) || json.Unmarshal(raw, &summary) != nil {
		return invalid
	}
	encoded, err := canonicalJSON(summary)
	if err != nil || !bytes.Equal(encoded, raw) || summary.SchemaID != "cartulary.reporting_render_validation_summary.v1" ||
		reportingObjectDigest(summary.SchemaID, raw) != manifest.ValidationSummarySHA256 ||
		!equalOptionalString(summary.ReleaseID, manifest.ReleaseID) || summary.SnapshotID == nil || *summary.SnapshotID != manifest.SnapshotID ||
		summary.CreatedAt != manifest.RenderAdmittedAt || summary.Result != "passed" || summary.TerminalStage != "release_state" ||
		summary.FailureCode != nil || summary.ReasonCode != nil || summary.FirstFailure != nil || summary.IssueCount != 0 || summary.Issues == nil || len(summary.Issues) != 0 {
		return invalid
	}
	expected := map[string]any{
		"output_kind":                         manifest.OutputKind,
		"release_scope":                       manifest.ReleaseScope,
		"network_policy_id":                   "render_sandbox_policy_v1",
		"non_loopback_outbound_attempt_count": 0,
	}
	if manifest.ReleaseID == nil {
		expected["pre_release"] = true
	}
	if !sameReportingValue(summary.SafeDetails, expected) {
		return invalid
	}
	return nil
}
