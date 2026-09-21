package revisionassembly

import (
	"encoding/json"
	"errors"
	"github.com/JochiRaider/cartulary/internal/testutil/historytest"
	"strings"
	"testing"

	"github.com/JochiRaider/cartulary/internal/modules/revisions"
	"github.com/JochiRaider/cartulary/internal/modules/revisions/historycontract"
)

const historyFixtureID = "20000000-0000-4000-8000-000000000001"

func TestSemanticHistorySourceContributions(t *testing.T) {
	contributions, err := CurrentProviderContributions()
	if err != nil {
		t.Fatal(err)
	}
	if _, err := Build(contributions...); err != nil {
		t.Fatal(err)
	}
	count := 0
	for _, owner := range contributions {
		for _, record := range owner.Records {
			variants := historytest.Variants(record.RecordType)
			for _, variant := range variants {
				t.Run(record.RecordType+variant, func(t *testing.T) {
					snapshot := historytest.Snapshot(record.RecordType, record.SnapshotSchemaID, variant)
					facts := historycontract.Facts{RecordID: historyFixtureID, Operation: "create", After: snapshot}
					assertHistoryProjection(t, record.HistoryProjector, facts)
					missing := record
					missing.HistoryProjector = nil
					altered := cloneProviderContributions(contributions)
					for i := range altered {
						for j := range altered[i].Records {
							if altered[i].Records[j].RecordType == record.RecordType {
								altered[i].Records[j] = missing
							}
						}
					}
					if _, err := Build(altered...); !errors.Is(err, revisions.ErrMissingProviderContribution) {
						t.Fatalf("missing projector admitted: %v", err)
					}
					count++
				})
			}
		}
		for _, target := range owner.NonRowTargets {
			t.Run(target.TargetKind, func(t *testing.T) {
				source := historytest.NonRowFacts(target.TargetKind)
				facts := historycontract.Facts{RecordID: historyFixtureID, Operation: "create", After: source}
				assertHistoryProjection(t, target.HistoryProjector, facts)
				altered := cloneProviderContributions(contributions)
				for i := range altered {
					for j := range altered[i].NonRowTargets {
						if altered[i].NonRowTargets[j].TargetKind == target.TargetKind {
							altered[i].NonRowTargets[j].HistoryProjector = nil
						}
					}
				}
				if _, err := Build(altered...); !errors.Is(err, revisions.ErrMissingProviderContribution) {
					t.Fatalf("missing projector admitted: %v", err)
				}
				count++
			})
		}
	}
	if count != 24 {
		t.Fatalf("covered %d source configurations, want 17 rows and 7 collection targets", count)
	}
}

func assertHistoryProjection(t *testing.T, project historycontract.Projector, facts historycontract.Facts) {
	t.Helper()
	units, err := project(facts)
	if err != nil {
		t.Fatal(err)
	}
	summary, err := historycontract.Summarize(units)
	if err != nil {
		t.Fatal(err)
	}
	first, err := json.Marshal(summary)
	if err != nil {
		t.Fatal(err)
	}
	repeated, err := project(facts)
	if err != nil {
		t.Fatal(err)
	}
	again, err := historycontract.Summarize(repeated)
	if err != nil {
		t.Fatal(err)
	}
	second, _ := json.Marshal(again)
	if string(first) != string(second) {
		t.Fatal("nondeterministic projection")
	}
	for _, secret := range []string{"private-upload-token", "private-credential", "snapshot_schema_id", "target_id", "object_blob_id"} {
		if strings.Contains(string(first), secret) {
			t.Fatalf("private retained fact leaked: %s", secret)
		}
	}
	if len(summary.Units) == 0 {
		t.Fatal("required detail missing")
	}
}
