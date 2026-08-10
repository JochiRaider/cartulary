package evidence

import (
	"slices"
	"testing"

	"github.com/JochiRaider/cartulary/internal/modules/revisions"
	"github.com/JochiRaider/cartulary/internal/platform/recoverystate"
)

func TestEvidenceSourceOwnerContributionsRemainExact(t *testing.T) {
	t.Parallel()

	t.Run("incident_bundle", func(t *testing.T) {
		t.Parallel()
		descriptor := NewIncidentBundleSourcePort().Descriptor()
		if descriptor.FamilyID != "evidence" ||
			descriptor.ContractMajor != 1 ||
			descriptor.OwnerID != "module.evidence" {
			t.Fatalf("unexpected Evidence incident-bundle descriptor: %#v", descriptor)
		}
		wantPaths := []string{
			"data/evidence_records.ndjson",
			"data/evidence_custody_events.ndjson",
			"data/object_blobs.ndjson",
		}
		gotPaths := make([]string, 0, len(descriptor.Paths))
		for _, path := range descriptor.Paths {
			gotPaths = append(gotPaths, path.LogicalPath)
		}
		if !slices.Equal(gotPaths, wantPaths) {
			t.Fatalf("Evidence incident-bundle paths = %v, want %v", gotPaths, wantPaths)
		}
	})

	t.Run("recovery", func(t *testing.T) {
		t.Parallel()
		contribution := RecoveryStateContribution()
		if contribution.SchemaID != recoverystate.ContributionSchemaID ||
			contribution.OwnerID != "module.evidence" {
			t.Fatalf("unexpected Evidence recovery contribution: %#v", contribution)
		}
		wantTables := []string{
			"evidence",
			"evidence_custody_events",
			"object_blobs",
			"evidence_access_handles",
		}
		gotTables := make([]string, 0, len(contribution.Tables))
		for _, table := range contribution.Tables {
			gotTables = append(gotTables, table.TableName)
		}
		if !slices.Equal(gotTables, wantTables) {
			t.Fatalf("Evidence recovery tables = %v, want %v", gotTables, wantTables)
		}
		if len(contribution.ObjectFamilies) != 1 ||
			contribution.ObjectFamilies[0].ObjectFamilyID != "evidence.blobs" {
			t.Fatalf("unexpected Evidence recovery object families: %#v", contribution.ObjectFamilies)
		}
	})

	t.Run("revisions", func(t *testing.T) {
		t.Parallel()
		contribution := RevisionProviderContribution()
		if contribution.SourceOwnerModule != revisions.SourceOwnerEvidence ||
			len(contribution.Records) != 1 {
			t.Fatalf("unexpected Evidence revision contribution: %#v", contribution)
		}
		record := contribution.Records[0]
		if record.SourceOwnerModule != revisions.SourceOwnerEvidence ||
			record.RecordType != "evidence" ||
			record.DeleteRestoreSource == nil ||
			record.RowRollbackProvider == nil {
			t.Fatalf("unexpected Evidence revision record contribution: %#v", record)
		}
		if len(record.RecordViewRoutes) != 1 ||
			record.RecordViewRoutes[0].ContributionID != "evidence.evidence" ||
			!slices.Equal(record.RecordViewRoutes[0].ViewSchemaIDs, []string{ViewSchemaID}) {
			t.Fatalf("unexpected Evidence revision routes: %#v", record.RecordViewRoutes)
		}
	})
}
