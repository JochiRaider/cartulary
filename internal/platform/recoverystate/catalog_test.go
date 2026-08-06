package recoverystate

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/JochiRaider/cartulary/internal/gen/contractrecovery"
)

func TestRecoveryCatalogRejectsEveryExactSetAndClassificationMutation(t *testing.T) {
	baseline := generatedRecoveryContributions(t)
	if _, err := Build(baseline...); err != nil {
		t.Fatalf("canonical generated Recovery contributions: %v", err)
	}

	tests := []struct {
		name   string
		mutate func([]Contribution)
	}{
		{
			name: "extra table",
			mutate: func(contributions []Contribution) {
				contributions[0].Tables = append(contributions[0].Tables, AuthoritativeTables("unexpected_extra_table")...)
			},
		},
		{
			name: "missing table",
			mutate: func(contributions []Contribution) {
				contributions[0].Tables = contributions[0].Tables[1:]
			},
		},
		{
			name: "duplicate table",
			mutate: func(contributions []Contribution) {
				contributions[0].Tables = append(contributions[0].Tables, contributions[0].Tables[0])
			},
		},
		{
			name: "misowned table",
			mutate: func(contributions []Contribution) {
				ownerIndex, tableIndex := findRecoveryTable(t, contributions, "extension_migration_ledger")
				table := contributions[ownerIndex].Tables[tableIndex]
				contributions[ownerIndex].Tables = append(contributions[ownerIndex].Tables[:tableIndex], contributions[ownerIndex].Tables[tableIndex+1:]...)
				targetIndex := (ownerIndex + 1) % len(contributions)
				contributions[targetIndex].Tables = append(contributions[targetIndex].Tables, table)
			},
		},
		{
			name: "misclassified table",
			mutate: func(contributions []Contribution) {
				ownerIndex, tableIndex := findRecoveryTable(t, contributions, "collaboration_event_intents")
				table := &contributions[ownerIndex].Tables[tableIndex]
				table.StateClass = StateTransient
				table.BackupInclusion = InclusionTransient
				table.RestoreAction = IgnoreState
				table.AlgorithmID = nil
			},
		},
		{
			name: "wrong restore action",
			mutate: func(contributions []Contribution) {
				ownerIndex, tableIndex := findRecoveryTable(t, contributions, "extension_migration_ledger")
				contributions[ownerIndex].Tables[tableIndex].RestoreAction = IgnoreState
			},
		},
		{
			name: "wrong codec",
			mutate: func(contributions []Contribution) {
				ownerIndex, tableIndex := findRecoveryTable(t, contributions, "extension_migration_ledger")
				wrong := "cartulary.wrong_codec.v1"
				contributions[ownerIndex].Tables[tableIndex].CodecID = &wrong
			},
		},
		{
			name: "contribution owner mismatch",
			mutate: func(contributions []Contribution) {
				contributions[0].OwnerID = "module.wrong_owner"
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			contributions := generatedRecoveryContributions(t)
			test.mutate(contributions)
			if _, err := Build(contributions...); !errors.Is(err, ErrInvalidCatalog) {
				t.Fatalf("mutated Recovery catalog error = %v; want ErrInvalidCatalog", err)
			}
		})
	}
}

func TestRecoveryCatalogRequiresExactlyOneSeparateSyntheticGooseTable(t *testing.T) {
	catalog, err := Build(generatedRecoveryContributions(t)...)
	if err != nil {
		t.Fatalf("build canonical Recovery catalog: %v", err)
	}
	canonical := append(catalog.AuthoredTableNames(), "goose_db_version")
	if err := catalog.ValidateDatabaseTableNames(canonical); err != nil {
		t.Fatalf("canonical authored plus Goose database coverage: %v", err)
	}
	if err := catalog.ValidateDatabaseTableNames(catalog.AuthoredTableNames()); !errors.Is(err, ErrInvalidCatalog) {
		t.Fatalf("missing Goose error = %v; want ErrInvalidCatalog", err)
	}
	secondSynthetic := append(append([]string(nil), canonical...), "second_synthetic_table")
	if err := catalog.ValidateDatabaseTableNames(secondSynthetic); !errors.Is(err, ErrInvalidCatalog) {
		t.Fatalf("second synthetic error = %v; want ErrInvalidCatalog", err)
	}
	duplicateGoose := append(append([]string(nil), canonical...), "goose_db_version")
	if err := catalog.ValidateDatabaseTableNames(duplicateGoose); !errors.Is(err, ErrInvalidCatalog) {
		t.Fatalf("duplicate Goose error = %v; want ErrInvalidCatalog", err)
	}
}

func generatedRecoveryContributions(t testing.TB) []Contribution {
	t.Helper()
	artifact, ok := contractrecovery.Index[catalogFixturePath]
	if !ok {
		t.Fatal("generated Recovery catalog fixture is unavailable")
	}
	var document Document
	if err := json.Unmarshal([]byte(artifact.JSON), &document); err != nil {
		t.Fatalf("decode generated Recovery catalog fixture: %v", err)
	}
	contributions := make([]Contribution, len(document.ContributionDigests))
	ownerIndexes := make(map[string]int, len(contributions))
	for index, digest := range document.ContributionDigests {
		ownerIndexes[digest.OwnerID] = index
		contributions[index] = NewContribution(digest.OwnerID, nil)
	}
	for _, table := range document.Tables {
		index, ok := ownerIndexes[table.OwnerID]
		if !ok {
			t.Fatalf("generated table %s has unknown owner %s", table.TableName, table.OwnerID)
		}
		contributions[index].Tables = append(contributions[index].Tables, Table{
			TableName:       table.TableName,
			StateClass:      table.StateClass,
			BackupInclusion: table.BackupInclusion,
			RestoreAction:   table.RestoreAction,
			CodecID:         cloneStringPointer(table.CodecID),
			AlgorithmID:     cloneStringPointer(table.AlgorithmID),
		})
	}
	for _, family := range document.ObjectFamilies {
		index, ok := ownerIndexes[family.OwnerID]
		if !ok {
			t.Fatalf("generated object family %s has unknown owner %s", family.ObjectFamilyID, family.OwnerID)
		}
		contributions[index].ObjectFamilies = append(contributions[index].ObjectFamilies, ObjectFamily{
			ObjectFamilyID:        family.ObjectFamilyID,
			StateClass:            family.StateClass,
			BackupInclusion:       family.BackupInclusion,
			RestoreAction:         family.RestoreAction,
			InventoryAlgorithmID:  cloneStringPointer(family.InventoryAlgorithmID),
			ValidationAlgorithmID: cloneStringPointer(family.ValidationAlgorithmID),
			RestoreAlgorithmID:    cloneStringPointer(family.RestoreAlgorithmID),
		})
	}
	return contributions
}

func findRecoveryTable(t testing.TB, contributions []Contribution, name string) (int, int) {
	t.Helper()
	for ownerIndex := range contributions {
		for tableIndex := range contributions[ownerIndex].Tables {
			if contributions[ownerIndex].Tables[tableIndex].TableName == name {
				return ownerIndex, tableIndex
			}
		}
	}
	t.Fatalf("Recovery table %s is unavailable", name)
	return 0, 0
}
