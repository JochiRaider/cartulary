package recoveryassembly

import (
	"errors"
	"slices"
	"testing"

	"github.com/JochiRaider/cartulary/internal/modules/recovery/restorecontract"
	"github.com/JochiRaider/cartulary/internal/platform/recoverystate"
)

func TestRecoveryStateCatalogClassifiesEveryAuthoredUnitAndRejectsDrift_Unit(t *testing.T) {
	contributions, err := CurrentRecoveryStateContributions()
	if err != nil {
		t.Fatalf("construct current recovery state contributions: %v", err)
	}
	catalog, err := recoverystate.Build(contributions...)
	if err != nil {
		t.Fatalf("build current recovery state catalog: %v", err)
	}
	if err := catalog.ValidateFrozen(); err != nil {
		t.Fatalf("validate frozen recovery state catalog: %v", err)
	}
	document := catalog.Document()
	if got := len(document.ContributionDigests); got != recoverystate.ContributionCount {
		t.Fatalf("owner contribution count = %d, want %d", got, recoverystate.ContributionCount)
	}
	if got := len(document.Tables); got != recoverystate.AuthoredTableCount {
		t.Fatalf("authored table count = %d, want %d", got, recoverystate.AuthoredTableCount)
	}
	if got := len(catalog.RequiredTableNames()); got != recoverystate.RequiredTableCount {
		t.Fatalf("required table count = %d, want %d", got, recoverystate.RequiredTableCount)
	}
	if got := len(document.ObjectFamilies); got != recoverystate.ObjectFamilyCount {
		t.Fatalf("object family count = %d, want %d", got, recoverystate.ObjectFamilyCount)
	}
	preferenceOwners := map[string]string{}
	assessmentRecovery := map[string]recoverystate.CatalogTable{}
	for _, table := range document.Tables {
		if table.TableName == "incident_workbook_preferences" || table.TableName == "user_workbook_preferences" {
			preferenceOwners[table.TableName] = table.OwnerID
		}
		if table.TableName == "assessments" || table.TableName == "assessment_grid_projection" {
			assessmentRecovery[table.TableName] = table
		}
	}
	if !slices.Equal(
		[]string{preferenceOwners["incident_workbook_preferences"], preferenceOwners["user_workbook_preferences"]},
		[]string{"module.workbook", "module.workbook"},
	) {
		t.Fatalf("Workbook preference Recovery ownership = %#v", preferenceOwners)
	}
	if source := assessmentRecovery["assessments"]; source.OwnerID != "module.assessments" ||
		source.StateClass != recoverystate.StateAuthoritative ||
		source.BackupInclusion != recoverystate.InclusionRequired ||
		source.RestoreAction != recoverystate.RestoreState {
		t.Fatalf("assessment source Recovery classification = %#v", source)
	}
	if projection := assessmentRecovery["assessment_grid_projection"]; projection.OwnerID != "module.projections" ||
		projection.StateClass != recoverystate.StateDerived ||
		projection.BackupInclusion != recoverystate.InclusionRebuildable ||
		projection.RestoreAction != recoverystate.RebuildState {
		t.Fatalf("assessment projection Recovery classification = %#v", projection)
	}
	if _, err := CurrentVNextObjectInventoryCatalog(nil); err != nil {
		t.Fatalf("exact six-family vNext inventory registration: %v", err)
	}
	if catalog.DigestSHA256() == "" {
		t.Fatal("catalog digest is empty")
	}
	if got, want := catalog.DigestSHA256(), "96ab9cac942a3729afcefa47a02bbfe910a2c09af0fb25ee32f7b610b6352055"; got != want {
		t.Fatalf("pre-Workbook-ownership catalog digest = %s, want compatibility identity %s", got, want)
	}
	if want := restorecontract.CurrentGraphProjectionImplementationBinding().Binding.RecoveryStateCatalogSHA256; catalog.DigestSHA256() != want {
		t.Fatalf("current Graph restore binding recovery catalog digest = %s, want %s", want, catalog.DigestSHA256())
	}

	reversed := cloneRecoveryStateContributions(contributions)
	slices.Reverse(reversed)
	reversedCatalog, err := recoverystate.Build(reversed...)
	if err != nil {
		t.Fatalf("build reversed recovery state catalog: %v", err)
	}
	if reversedCatalog.DigestSHA256() != catalog.DigestSHA256() {
		t.Fatal("catalog digest depends on contribution registration order")
	}

	legacyTables := catalog.RequiredTableNames()
	legacyTables = append(legacyTables,
		"collaboration_event_intents",
		"collaboration_incident_stream_cursors",
		"collaboration_replay_events",
		"collaboration_resume_tokens",
		"enterprise_auth_transactions",
	)
	if err := catalog.ValidateLegacyShadowTables(legacyTables); err != nil {
		t.Fatalf("validate exact transitional legacy table set: %v", err)
	}
	if err := catalog.ValidateLegacyShadowTables(legacyTables[1:]); !errors.Is(err, recoverystate.ErrInvalidCatalog) {
		t.Fatalf("missing legacy table error = %v, want ErrInvalidCatalog", err)
	}
	if err := catalog.ValidateLegacyShadowTables(append(legacyTables, "unclassified_future_table")); !errors.Is(err, recoverystate.ErrInvalidCatalog) {
		t.Fatalf("unknown legacy table error = %v, want ErrInvalidCatalog", err)
	}
	databaseTables := append(catalog.AuthoredTableNames(), "goose_db_version")
	if err := catalog.ValidateDatabaseTableNames(databaseTables); err != nil {
		t.Fatalf("validate exact database table coverage: %v", err)
	}
	if err := catalog.ValidateDatabaseTableNames(databaseTables[1:]); !errors.Is(err, recoverystate.ErrInvalidCatalog) {
		t.Fatalf("missing database table error = %v, want ErrInvalidCatalog", err)
	}
	if err := catalog.ValidateDatabaseTableNames(append(databaseTables, "unclassified_future_table")); !errors.Is(err, recoverystate.ErrInvalidCatalog) {
		t.Fatalf("unknown database table error = %v, want ErrInvalidCatalog", err)
	}

	tests := []struct {
		name   string
		mutate func([]recoverystate.Contribution) []recoverystate.Contribution
	}{
		{
			name: "missing owner",
			mutate: func(values []recoverystate.Contribution) []recoverystate.Contribution {
				return values[1:]
			},
		},
		{
			name: "duplicate owner",
			mutate: func(values []recoverystate.Contribution) []recoverystate.Contribution {
				return append(values, values[0])
			},
		},
		{
			name: "duplicate table",
			mutate: func(values []recoverystate.Contribution) []recoverystate.Contribution {
				values[1].Tables = append(values[1].Tables, values[0].Tables[0])
				return values
			},
		},
		{
			name: "unknown table",
			mutate: func(values []recoverystate.Contribution) []recoverystate.Contribution {
				values[0].Tables[0].TableName = "unclassified_future_table"
				return values
			},
		},
		{
			name: "missing object family",
			mutate: func(values []recoverystate.Contribution) []recoverystate.Contribution {
				for index := range values {
					if len(values[index].ObjectFamilies) > 0 {
						values[index].ObjectFamilies = nil
						break
					}
				}
				return values
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			mutated := test.mutate(cloneRecoveryStateContributions(contributions))
			if _, err := recoverystate.Build(mutated...); !errors.Is(err, recoverystate.ErrInvalidCatalog) {
				t.Fatalf("Build() error = %v, want ErrInvalidCatalog", err)
			}
		})
	}

	external := catalog.Document()
	external.Tables[0].TableName = "mutated"
	if err := catalog.ValidateFrozen(); err != nil {
		t.Fatalf("caller mutated frozen catalog through a cloned view: %v", err)
	}
}

func cloneRecoveryStateContributions(values []recoverystate.Contribution) []recoverystate.Contribution {
	cloned := make([]recoverystate.Contribution, len(values))
	for index, value := range values {
		cloned[index] = value
		cloned[index].Tables = append([]recoverystate.Table(nil), value.Tables...)
		cloned[index].ObjectFamilies = append([]recoverystate.ObjectFamily(nil), value.ObjectFamilies...)
	}
	return cloned
}
