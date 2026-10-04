package reporting

import (
	"errors"
	"github.com/google/uuid"
	"time"
)

const SnapshotModelSchemaID = "cartulary.reporting_snapshot_model.v1"
const reportingTimestampLayout = "2006-01-02T15:04:05.000000Z"

// RenderIdentity separates an admitted preview from a release. Exactly one
// identity is present; neither a snapshot nor a retry invents a release ID.
type RenderIdentity struct {
	ReleaseID        *string `json:"release_id"`
	PreviewAttemptID *string `json:"preview_attempt_id"`
	RenderAdmittedAt string  `json:"render_admitted_at"`
}

func normalizedReportingTimestamp(t time.Time) (string, error) {
	t = t.UTC()
	if t.IsZero() || t.Year() < 1 || t.Year() > 9999 || t.Nanosecond()%1000 != 0 {
		return "", newRenderValidationError("export_model_invalid", "invalid_timestamp_value")
	}
	return t.UTC().Format(reportingTimestampLayout), nil
}

func (r RenderIdentity) validate() error {
	if (r.ReleaseID == nil) == (r.PreviewAttemptID == nil) {
		return errors.New("reporting: invalid render identity")
	}
	id := r.ReleaseID
	if id == nil {
		id = r.PreviewAttemptID
	}
	parsed, err := uuid.Parse(*id)
	if err != nil || parsed == uuid.Nil || parsed.String() != *id {
		return errors.New("reporting: invalid render identity")
	}
	t, err := time.Parse(reportingTimestampLayout, r.RenderAdmittedAt)
	if err != nil || t.IsZero() || t.Format(reportingTimestampLayout) != r.RenderAdmittedAt {
		return newRenderValidationError("export_model_invalid", "invalid_timestamp_value")
	}
	return nil
}

func generatedReportingID(schema, prefix string, tuple map[string]any) (string, error) {
	tuple["schema_id"] = schema
	b, err := canonicalJSON(tuple)
	if err != nil {
		return "", err
	}
	return prefix + reportingObjectDigest(schema, b), nil
}

func snapshotModelID(snapshotID, derivation string, snapshotAt time.Time) (string, error) {
	stamp, err := normalizedReportingTimestamp(snapshotAt)
	if err != nil {
		return "", err
	}
	return generatedReportingID("cartulary.reporting_snapshot_model_id.v1", "snapm_", map[string]any{
		"snapshot_id": snapshotID, "derivation_version": derivation, "snapshot_at": stamp,
	})
}

func (r RenderIdentity) generatedID(kind, prefix string, tuple map[string]any) (string, error) {
	if err := r.validate(); err != nil {
		return "", err
	}
	schema := "cartulary.reporting_" + kind + "_id.v1"
	if r.ReleaseID != nil {
		tuple["release_id"] = *r.ReleaseID
	} else {
		schema = "cartulary.reporting_preview_" + kind + "_id.v1"
		tuple["preview_attempt_id"] = *r.PreviewAttemptID
	}
	return generatedReportingID(schema, prefix, tuple)
}

func bindRenderModel(snapshot SnapshotModel, identity RenderIdentity, scope string, partitions []string) (ExportModel, string, error) {
	if err := validateSnapshotModelIdentity(snapshot); err != nil {
		return ExportModel{}, "", err
	}
	id, err := identity.generatedID("export_model", "expm_", map[string]any{
		"snapshot_id": snapshot.SnapshotID, "derivation_version": snapshot.DerivationVersion, "render_admitted_at": identity.RenderAdmittedAt,
	})
	if err != nil {
		return ExportModel{}, "", err
	}
	// The common immutable source content, including its exact pack binding, is
	// copied into the render input. Subsequent derivation must never write it back.
	model := ExportModel{
		SchemaID: ExportModelSchemaID, SnapshotModelID: snapshot.SnapshotModelID, ExportModelID: id,
		SnapshotContent: snapshot.SnapshotContent, RenderIdentity: identity,
		ExportModelCreatedAt: identity.RenderAdmittedAt, ReleaseScope: scope, RecipientPartitionRefs: cloneStrings(partitions),
	}
	if err := validateRenderModelIdentity(model); err != nil {
		return ExportModel{}, "", err
	}
	b, err := canonicalJSON(model)
	if err != nil {
		return ExportModel{}, "", err
	}
	return model, reportingObjectDigest(model.SchemaID, b), nil
}

func validateSnapshotModelIdentity(model SnapshotModel) error {
	id, err := uuid.Parse(model.SnapshotID)
	if err != nil || id == uuid.Nil || id.String() != model.SnapshotID || model.SchemaID != SnapshotModelSchemaID || model.DerivationVersion != DerivationVersion {
		return errors.New("reporting: invalid snapshot model identity")
	}
	expected, err := snapshotModelID(model.SnapshotID, model.DerivationVersion, model.SnapshotAt)
	if err != nil || model.SnapshotModelID != expected {
		return errors.New("reporting: invalid snapshot model identity")
	}
	for _, r := range model.Records {
		if r.SourceRecordRef.SourceSnapshotID != model.SnapshotID {
			return errors.New("reporting: invalid source snapshot identity")
		}
	}
	for _, r := range model.Relationships {
		if r.SrcRecordRef.SourceRecordRef.SourceSnapshotID != model.SnapshotID || r.DstRecordRef.SourceRecordRef.SourceSnapshotID != model.SnapshotID {
			return errors.New("reporting: invalid source snapshot identity")
		}
	}
	for _, r := range model.TimelineEvents {
		if r.SourceRecordRef.SourceSnapshotID != model.SnapshotID {
			return errors.New("reporting: invalid source snapshot identity")
		}
	}
	for _, r := range model.Subjects {
		if r.SourceRecordRef != nil && r.SourceRecordRef.SourceSnapshotID != model.SnapshotID {
			return errors.New("reporting: invalid source snapshot identity")
		}
	}
	for _, r := range model.SupportIndex {
		if r.SourceSnapshotID != model.SnapshotID {
			return errors.New("reporting: invalid source snapshot identity")
		}
	}
	return nil
}

func validateRenderModelIdentity(model ExportModel) error {
	if model.SchemaID != ExportModelSchemaID || model.DerivationVersion != DerivationVersion || model.ExportModelCreatedAt != model.RenderAdmittedAt {
		return errors.New("reporting: invalid render model identity")
	}
	if err := validateSnapshotModelIdentity(SnapshotModel{SchemaID: SnapshotModelSchemaID, SnapshotModelID: model.SnapshotModelID, SnapshotContent: model.SnapshotContent}); err != nil {
		return err
	}
	id, err := model.RenderIdentity.generatedID("export_model", "expm_", map[string]any{
		"snapshot_id": model.SnapshotID, "derivation_version": model.DerivationVersion, "render_admitted_at": model.RenderAdmittedAt,
	})
	if err != nil || id != model.ExportModelID {
		return errors.New("reporting: invalid render model identity")
	}
	if model.PreviewAttemptID != nil && model.ReleaseScope != ReleaseScopeInternalDraft {
		return errors.New("reporting: preview cannot authorize a release")
	}
	return nil
}

func validateReleasePayloadIdentity(payload releaseCreateJobPayload, jobID uuid.UUID) error {
	expected := reportingResourceID(jobID, "release").String()
	if payload.ExportModel.ReleaseID == nil || *payload.ExportModel.ReleaseID != expected || payload.ExportModel.PreviewAttemptID != nil || payload.ExportModel.SnapshotID != payload.SnapshotID || payload.ExportModel.IncidentID != payload.IncidentID || payload.ExportModel.ReleaseScope != payload.ReleaseScope {
		return errors.New("reporting: invalid frozen release identity")
	}
	if err := validateRenderModelIdentity(payload.ExportModel); err != nil {
		return err
	}
	stamp, err := normalizedReportingTimestamp(payload.RenderAdmittedAt)
	if err != nil || stamp != payload.ExportModel.RenderAdmittedAt {
		return errors.New("reporting: invalid frozen release timestamp")
	}
	data, err := canonicalJSON(payload.ExportModel)
	if err != nil || reportingObjectDigest(payload.ExportModel.SchemaID, data) != payload.ExportModelSHA256 {
		return errors.New("reporting: altered frozen release model")
	}
	return nil
}

func validateBundleIdentity(model RenderBundleManifest) error {
	if model.BundleCreatedAt != model.RenderAdmittedAt || !sha256HexPattern.MatchString(model.ExportModelSHA256) {
		return errors.New("reporting: invalid bundle identity")
	}
	expected, err := model.RenderIdentity.generatedID("export_model", "expm_", map[string]any{"snapshot_id": model.SnapshotID, "derivation_version": DerivationVersion, "render_admitted_at": model.RenderAdmittedAt})
	if err != nil || expected != model.ExportModelID {
		return errors.New("reporting: invalid bundle identity")
	}
	return nil
}
