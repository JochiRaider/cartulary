package reporting

import "time"

const fixtureSnapshotID = "00000000-0000-0000-0000-000000000402"

func fixtureRenderIdentity() RenderIdentity {
	id := "00000000-0000-0000-0000-000000000403"
	return RenderIdentity{ReleaseID: &id, RenderAdmittedAt: "2026-05-23T12:00:00.000000Z"}
}

func bindFixtureRenderModel(content SnapshotContent) ExportModel {
	if content.SnapshotID == "" {
		content.SnapshotID = fixtureSnapshotID
	}
	if content.SnapshotAt.IsZero() {
		content.SnapshotAt = time.Date(2026, 5, 23, 12, 0, 0, 0, time.UTC)
	}
	content.DerivationVersion = DerivationVersion
	id, err := snapshotModelID(content.SnapshotID, content.DerivationVersion, content.SnapshotAt)
	if err != nil {
		panic(err)
	}
	model, _, err := bindRenderModel(SnapshotModel{SchemaID: SnapshotModelSchemaID, SnapshotModelID: id, SnapshotContent: content}, fixtureRenderIdentity(), ReleaseScopeInternalReview, nil)
	if err != nil {
		panic(err)
	}
	return model
}

func buildStructuredTestExportModel(incidentID, snapshotID string, at time.Time, watermark, scope string, partitions []string, fields []ExportField) (ExportModel, string, error) {
	snapshot, _, err := buildStructuredSnapshotModel(incidentID, snapshotID, at, watermark, fields)
	if err != nil {
		return ExportModel{}, "", err
	}
	return bindRenderModel(snapshot, fixtureRenderIdentity(), scope, partitions)
}
