package database_migrations

import (
	"errors"
	"testing"
	"testing/fstest"
)

func TestClassifyMigrationStateMatrix(t *testing.T) {
	source, err := buildSource(fstest.MapFS{
		"00001_one.sql":   &fstest.MapFile{Data: []byte(validMigrationBody)},
		"00002_two.sql":   &fstest.MapFile{Data: []byte(validMigrationBody)},
		"00003_three.sql": &fstest.MapFile{Data: []byte(validMigrationBody)},
	}, ".", "expected.lineage.v1", "expected_lineage_v1")
	if err != nil {
		t.Fatalf("construct classifier source: %v", err)
	}

	rows := func(versions ...int64) []migrationLedgerRow {
		result := []migrationLedgerRow{{Version: 0, IsApplied: true}}
		for _, version := range versions {
			result = append(result, migrationLedgerRow{Version: version, IsApplied: true})
		}
		return result
	}
	lineage := func(ids ...string) migrationStateSnapshot {
		return migrationStateSnapshot{LineageTablePresent: true, LineageIDs: ids}
	}

	tests := []struct {
		name            string
		source          *Source
		snapshot        migrationStateSnapshot
		wantState       migrationState
		wantReason      string
		wantRemediation bool
	}{
		{name: "pristine", source: source, snapshot: migrationStateSnapshot{}, wantState: migrationStatePristine},
		{name: "zero only", source: source, snapshot: migrationStateSnapshot{LedgerTablePresent: true, LedgerRows: rows()}, wantState: migrationStatePristine},
		{name: "behind", source: source, snapshot: mergeSnapshot(migrationStateSnapshot{LedgerTablePresent: true, LedgerRows: rows(1)}, lineage("expected.lineage.v1")), wantState: migrationStateBehind},
		{name: "current", source: source, snapshot: mergeSnapshot(migrationStateSnapshot{LedgerTablePresent: true, LedgerRows: rows(1, 2, 3)}, lineage("expected.lineage.v1")), wantState: migrationStateCurrent},
		{name: "ahead", source: source, snapshot: mergeSnapshot(migrationStateSnapshot{LedgerTablePresent: true, LedgerRows: rows(1, 2, 3, 4)}, lineage("expected.lineage.v1")), wantState: migrationStateAhead},
		{name: "duplicate", source: source, snapshot: migrationStateSnapshot{LedgerTablePresent: true, LedgerRows: rows(1, 2, 2)}, wantReason: reasonSchemaMigrationHistoryInvalid},
		{name: "false", source: source, snapshot: migrationStateSnapshot{LedgerTablePresent: true, LedgerRows: []migrationLedgerRow{{Version: 0, IsApplied: true}, {Version: 1, IsApplied: false}}}, wantReason: reasonSchemaMigrationHistoryInvalid},
		{name: "negative", source: source, snapshot: migrationStateSnapshot{LedgerTablePresent: true, LedgerRows: []migrationLedgerRow{{Version: -1, IsApplied: true}}}, wantReason: reasonSchemaMigrationHistoryInvalid},
		{name: "repeated zero", source: source, snapshot: migrationStateSnapshot{LedgerTablePresent: true, LedgerRows: rows(0)}, wantReason: reasonSchemaMigrationHistoryInvalid},
		{name: "out of order", source: source, snapshot: migrationStateSnapshot{LedgerTablePresent: true, LedgerRows: rows(2, 1)}, wantReason: reasonSchemaMigrationHistoryInvalid},
		{name: "gap", source: source, snapshot: migrationStateSnapshot{LedgerTablePresent: true, LedgerRows: rows(1, 3)}, wantReason: reasonSchemaMigrationHistoryInvalid},
		{name: "missing lineage table", source: source, snapshot: migrationStateSnapshot{LedgerTablePresent: true, LedgerRows: rows(1)}, wantRemediation: true},
		{name: "empty lineage set", source: source, snapshot: mergeSnapshot(migrationStateSnapshot{LedgerTablePresent: true, LedgerRows: rows(1)}, lineage()), wantRemediation: true},
		{name: "wrong lineage", source: source, snapshot: mergeSnapshot(migrationStateSnapshot{LedgerTablePresent: true, LedgerRows: rows(1)}, lineage("wrong.lineage.v1")), wantRemediation: true},
		{name: "mixed lineage", source: source, snapshot: mergeSnapshot(migrationStateSnapshot{LedgerTablePresent: true, LedgerRows: rows(1)}, lineage("expected.lineage.v1", "wrong.lineage.v1")), wantRemediation: true},
		{name: "corruption wins over lineage", source: source, snapshot: mergeSnapshot(migrationStateSnapshot{LedgerTablePresent: true, LedgerRows: rows(1, 3)}, lineage("wrong.lineage.v1")), wantReason: reasonSchemaMigrationHistoryInvalid},
		{name: "lineage without history", source: source, snapshot: lineage("expected.lineage.v1"), wantReason: reasonSchemaMigrationHistoryInvalid},
		{name: "zero only with lineage", source: source, snapshot: mergeSnapshot(migrationStateSnapshot{LedgerTablePresent: true, LedgerRows: rows()}, lineage("expected.lineage.v1")), wantReason: reasonSchemaMigrationHistoryInvalid},
		{name: "zero source", snapshot: migrationStateSnapshot{}, wantReason: reasonMigrationSourceInvalid},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			classification, classifyErr := classifyMigrationState(test.source, test.snapshot)
			switch {
			case test.wantRemediation:
				var reporter RemediationReporter
				if !errors.As(classifyErr, &reporter) || reporter.ReasonCode() != reasonHistoricalMigrationLineage {
					t.Fatalf("expected remediation reporter, got %T: %v", classifyErr, classifyErr)
				}
			case test.wantReason != "":
				requireMigrationFailureReason(t, classifyErr, test.wantReason)
			default:
				if classifyErr != nil {
					t.Fatalf("classify state: %v", classifyErr)
				}
				if classification.State != test.wantState {
					t.Fatalf("classification state = %d, want %d", classification.State, test.wantState)
				}
			}
		})
	}
}

func mergeSnapshot(left migrationStateSnapshot, right migrationStateSnapshot) migrationStateSnapshot {
	left.LineageTablePresent = right.LineageTablePresent
	left.LineageIDs = right.LineageIDs
	return left
}
