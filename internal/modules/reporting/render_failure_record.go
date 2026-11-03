package reporting

import (
	"bytes"
	"encoding/json"
	"errors"

	"github.com/JochiRaider/cartulary/internal/platform/jobs"
)

const renderFailureRecordSchema = "cartulary.reporting_render_failure_record.v1"
const maxRenderFailureRecordBytes = 2 << 20

type renderFailureRecord struct {
	SchemaID      string                     `json:"schema_id"`
	JobID         string                     `json:"job_id"`
	JobKind       string                     `json:"job_kind"`
	IncidentID    string                     `json:"incident_id"`
	ActorUserID   string                     `json:"actor_user_id"`
	AdmittedAt    string                     `json:"admitted_at"`
	RequestSHA256 string                     `json:"request_sha256"`
	InputBinding  *renderFailureInputBinding `json:"input_binding"`
	Summary       json.RawMessage            `json:"summary"`
}

type renderFailureInputBinding struct {
	AdmissionSHA256   string  `json:"admission_sha256"`
	PreparationSHA256 *string `json:"preparation_sha256"`
}

// The envelope is rooted in Jobs admission, not in the input being rejected.
// Public-response expiry never reconstructs or replaces the original envelope.
func bindRenderFailureRecord(retained jobs.RetainedExtensionJob, summary []byte) (renderFailureRecord, error) {
	resource := retained.Resource
	invalid := errors.New("reporting: failure record lacks independently retained admission")
	if retained.OwnerProfileID != ProfileID || retained.HandlerName != JobWorkerKind ||
		(retained.JobKind != ReleaseCreateJobKind && retained.JobKind != CompositionPreviewJobKind) ||
		resource.AuthPolicy != jobs.AuthPolicyIncidentMembership || resource.Scope.Kind != jobs.ScopeKindIncident ||
		resource.Scope.IncidentID == nil || resource.SubmittedByUserID == nil ||
		!canonicalUUID(resource.JobID) || !canonicalUUID(resource.Scope.IncidentID.String()) || !canonicalUUID(*resource.SubmittedByUserID) ||
		!sha256HexPattern.MatchString(retained.RequestSHA256) || retained.ExpiredAt != nil {
		return renderFailureRecord{}, invalid
	}
	at, err := normalizedReportingTimestamp(resource.SubmittedAt)
	if err != nil {
		return renderFailureRecord{}, invalid
	}
	return renderFailureRecord{SchemaID: renderFailureRecordSchema, JobID: resource.JobID, JobKind: retained.JobKind,
		IncidentID: resource.Scope.IncidentID.String(), ActorUserID: *resource.SubmittedByUserID,
		AdmittedAt: at, RequestSHA256: retained.RequestSHA256, Summary: append(json.RawMessage(nil), summary...)}, nil
}

func encodeRenderFailureRecord(record renderFailureRecord) ([]byte, string, error) {
	raw, err := canonicalJSON(record)
	if err != nil {
		return nil, "", err
	}
	if len(raw) == 0 || len(raw) > maxRenderFailureRecordBytes {
		return nil, "", errors.New("reporting: failure record exceeds its byte contract")
	}
	return raw, reportingObjectDigest(renderFailureRecordSchema, raw), nil
}

// Validate closed canonical bytes and the surviving independent admission. A
// valid diagnostic does not confer validity on the rejected private input.
func decodeRenderFailureRecord(raw []byte, digest string, retained jobs.RetainedExtensionJob) (renderFailureRecord, reportingRenderValidationSummary, error) {
	var record renderFailureRecord
	var summary reportingRenderValidationSummary
	invalid := errors.New("reporting: invalid independent failure record")
	if len(raw) == 0 || len(raw) > maxRenderFailureRecordBytes || reportingObjectDigest(renderFailureRecordSchema, raw) != digest {
		return record, summary, invalid
	}
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.DisallowUnknownFields()
	if decoder.Decode(&record) != nil {
		return record, summary, invalid
	}
	encoded, _, err := encodeRenderFailureRecord(record)
	if err != nil || !bytes.Equal(encoded, raw) {
		return record, summary, invalid
	}
	if err := validateRenderFailureRecordBinding(record, retained); err != nil {
		return record, summary, err
	}
	if record.InputBinding != nil && (!sha256HexPattern.MatchString(record.InputBinding.AdmissionSHA256) || record.InputBinding.PreparationSHA256 != nil && !sha256HexPattern.MatchString(*record.InputBinding.PreparationSHA256)) {
		return record, summary, invalid
	}
	canonical, err := canonicalReportingBytes(record.Summary)
	if err != nil || len(canonical) > 1<<20 || json.Unmarshal(canonical, &summary) != nil ||
		summary.Result != "failed" || summary.FailureCode == nil || summary.ReasonCode == nil || len(summary.Issues) == 0 ||
		summary.CreatedAt != record.AdmittedAt ||
		summary.SnapshotID != nil && !canonicalUUID(*summary.SnapshotID) ||
		summary.ReleaseID != nil && (summary.SnapshotID == nil || !canonicalUUID(*summary.ReleaseID) || record.JobKind == CompositionPreviewJobKind) {
		return record, summary, invalid
	}
	if (record.InputBinding == nil) != (summary.SnapshotID == nil) || record.InputBinding == nil && summary.ReleaseID != nil {
		return record, summary, invalid
	}
	if *summary.FailureCode == "retained_input_invalid" && (summary.ReleaseID != nil || summary.SnapshotID != nil) {
		return record, summary, invalid
	}
	issue := summary.Issues[0]
	failure := RenderFailure{Stage: summary.TerminalStage, FailureCode: *summary.FailureCode, ReasonCode: *summary.ReasonCode, SafeDetails: issue.SafeDetails,
		issues: &renderIssueSet{count: summary.IssueCount, retained: summary.Issues}}
	expected, err := encodeFailedRenderSummary(summary.ReleaseID, summary.SnapshotID, record.AdmittedAt, failure)
	if err != nil || !bytes.Equal(expected, canonical) {
		return record, summary, invalid
	}
	record.Summary = canonical
	return record, summary, nil
}

// Compare only bounded independent admission facts inside a terminal transaction.
func validateRenderFailureRecordBinding(record renderFailureRecord, retained jobs.RetainedExtensionJob) error {
	resource := retained.Resource
	at, err := normalizedReportingTimestamp(resource.SubmittedAt)
	if err != nil || record.SchemaID != renderFailureRecordSchema || record.JobID != resource.JobID || !canonicalUUID(record.JobID) ||
		record.JobKind != retained.JobKind || (record.JobKind != ReleaseCreateJobKind && record.JobKind != CompositionPreviewJobKind) ||
		retained.OwnerProfileID != ProfileID || retained.HandlerName != JobWorkerKind || resource.AuthPolicy != jobs.AuthPolicyIncidentMembership ||
		resource.Scope.Kind != jobs.ScopeKindIncident || resource.Scope.IncidentID == nil || record.IncidentID != resource.Scope.IncidentID.String() ||
		!canonicalUUID(record.IncidentID) || resource.SubmittedByUserID == nil || record.ActorUserID != *resource.SubmittedByUserID ||
		!canonicalUUID(record.ActorUserID) || record.AdmittedAt != at || !sha256HexPattern.MatchString(record.RequestSHA256) ||
		retained.ExpiredAt == nil && record.RequestSHA256 != retained.RequestSHA256 {
		return errors.New("reporting: failure record differs from independent admission")
	}
	return nil
}
