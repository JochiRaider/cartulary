package artifacts

import (
	"strings"
	"testing"

	artifactprojection "github.com/JochiRaider/cartulary/internal/modules/artifacts/workbookprojection"
	"github.com/JochiRaider/cartulary/internal/modules/imports/ownerfacade"
	conflicttokens "github.com/JochiRaider/cartulary/internal/modules/revisions/conflicts"
	"github.com/JochiRaider/cartulary/internal/platform/postgres"
)

type artifactCompositionDB struct{ postgres.DB }
type artifactIncidentStateStub struct{ IncidentStateCapability }
type artifactMemberReferencesStub struct{ MemberReferenceCapability }
type artifactIdempotencyStub struct{ IdempotencyCapability }
type artifactRecordEnvelopesStub struct{ RecordEnvelopeCapability }
type artifactLinksStub struct{ LinkCapability }
type artifactProjectionRowsStub struct{ artifactprojection.Rows }
type artifactRevisionsStub struct{ RevisionCapability }
type artifactConflictFieldsStub struct{ conflicttokens.FieldResolver }
type artifactKeepSavedStub struct{ conflicttokens.IdempotencyPort }
type artifactImportRecordStub struct{ recordEnvelopeInserter }
type artifactImportActiveUserStub struct{ activeUserLookup }
type artifactImportProjectionStub struct{ artifactProjectionRows }
type artifactImportRevisionStub struct {
	ownerfacade.LiveRecordRevisionAppender
}

func completeArtifactMutationDependencies() MutationDependencies {
	return MutationDependencies{
		IncidentState:        artifactIncidentStateStub{},
		MemberReferences:     artifactMemberReferencesStub{},
		Idempotency:          artifactIdempotencyStub{},
		RecordEnvelopes:      artifactRecordEnvelopesStub{},
		Links:                artifactLinksStub{},
		Projections:          artifactProjectionRowsStub{},
		Revisions:            artifactRevisionsStub{},
		ConflictFields:       artifactConflictFieldsStub{},
		KeepSavedIdempotency: artifactKeepSavedStub{},
	}
}

func completeArtifactImportDependencies() ImportDependencies {
	return ImportDependencies{
		RecordEnvelopes: artifactImportRecordStub{},
		ActiveUsers:     artifactImportActiveUserStub{},
		Projections:     artifactImportProjectionStub{},
		Revisions:       artifactImportRevisionStub{},
	}
}

func TestArtifactMutationContributionRejectsMissingDependencies(t *testing.T) {
	t.Parallel()
	if _, err := NewMutationContribution(nil, conflicttokens.ConflictTokenCodec{}, MutationDependencies{}); err == nil || !strings.Contains(err.Error(), "Postgres is required") {
		t.Fatalf("nil Postgres error = %v", err)
	}
	tests := []struct {
		name string
		want string
		drop func(*MutationDependencies)
	}{
		{name: "incident state", want: "Incident admission", drop: func(d *MutationDependencies) { d.IncidentState = nil }},
		{name: "member references", want: "Member validation", drop: func(d *MutationDependencies) { d.MemberReferences = nil }},
		{name: "idempotency", want: "Route idempotency", drop: func(d *MutationDependencies) { d.Idempotency = nil }},
		{name: "records", want: "Record envelopes", drop: func(d *MutationDependencies) { d.RecordEnvelopes = nil }},
		{name: "links", want: "Links", drop: func(d *MutationDependencies) { d.Links = nil }},
		{name: "projections", want: "Projections", drop: func(d *MutationDependencies) { d.Projections = nil }},
		{name: "revisions", want: "Revisions/history", drop: func(d *MutationDependencies) { d.Revisions = nil }},
		{name: "conflict fields", want: "Conflict fields", drop: func(d *MutationDependencies) { d.ConflictFields = nil }},
		{name: "keep saved", want: "Keep-saved idempotency", drop: func(d *MutationDependencies) { d.KeepSavedIdempotency = nil }},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			dependencies := completeArtifactMutationDependencies()
			tc.drop(&dependencies)
			if _, err := NewMutationContribution(&artifactCompositionDB{}, conflicttokens.ConflictTokenCodec{}, dependencies); err == nil || !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("missing %s error = %v", tc.name, err)
			}
		})
	}
}

func TestArtifactImportContributionRejectsMissingDependencies(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		want string
		drop func(*ImportDependencies)
	}{
		{name: "records insertion", want: "Records insertion", drop: func(d *ImportDependencies) { d.RecordEnvelopes = nil }},
		{name: "active-user lookup", want: "Active-user lookup", drop: func(d *ImportDependencies) { d.ActiveUsers = nil }},
		{name: "projection refresh/load", want: "Projection refresh/load", drop: func(d *ImportDependencies) { d.Projections = nil }},
		{name: "revision and intent appender", want: "Revision and intent appender", drop: func(d *ImportDependencies) { d.Revisions = nil }},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			dependencies := completeArtifactImportDependencies()
			tc.drop(&dependencies)
			if _, err := NewImportContribution(NotesViewSchemaID, "artifacts.notes.create", dependencies); err == nil || !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("missing %s error = %v", tc.name, err)
			}
		})
	}
	if _, err := NewImportContribution("cartulary.view.unknown.v1", "artifacts.unknown.create", completeArtifactImportDependencies()); err == nil || !strings.Contains(err.Error(), "not mapped") {
		t.Fatalf("unknown import surface error = %v", err)
	}
}
