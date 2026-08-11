package database_migrations

import (
	"encoding/json"
)

const (
	migrationRemediationSchemaID   = "cartulary.migration_remediation_report.v1"
	historicalMigrationLineageHint = "Reset this database, or move data through an explicit owner-approved export/import path before applying the production DDL rebaseline."
)

type migrationRemediationReport struct {
	SchemaID    string                        `json:"schema_id"`
	Boundary    string                        `json:"boundary"`
	FromVersion int64                         `json:"from_version"`
	ToVersion   int64                         `json:"to_version"`
	Findings    []migrationRemediationFinding `json:"findings"`
}

type migrationRemediationFinding struct {
	Field           string                    `json:"field"`
	RawValue        *string                   `json:"raw_value,omitempty"`
	RawValuePair    migrationRemediationFacts `json:"raw_value_pair"`
	ReasonCode      string                    `json:"reason_code"`
	RemediationHint string                    `json:"remediation_hint"`
}

type migrationRemediationFacts struct {
	CurrentVersion        int64    `json:"current_version"`
	ExpectedLineageID     string   `json:"expected_lineage_id"`
	LineageTablePresent   bool     `json:"lineage_table_present"`
	ObservedLineageIDs    []string `json:"observed_lineage_ids"`
	RepositoryHeadVersion int64    `json:"repository_head_version"`
	TargetVersion         int64    `json:"target_version"`
}

type migrationRemediationError struct {
	report migrationRemediationReport
}

func (err *migrationRemediationError) Error() string {
	return err.RemediationReportJSON()
}

func (err *migrationRemediationError) ReasonCode() string {
	return reasonHistoricalMigrationLineage
}

func (err *migrationRemediationError) RemediationReportJSON() string {
	encoded, _ := json.Marshal(err.report)
	return string(encoded)
}

type migrationLineageState struct {
	TablePresent bool
	ObservedIDs  []string
}

func (state migrationLineageState) HasExactExpected(lineageID string) bool {
	return state.TablePresent && len(state.ObservedIDs) == 1 && state.ObservedIDs[0] == lineageID
}

func newMigrationLineageRemediationError(
	source *Source,
	state migrationLineageState,
	currentVersion int64,
	targetVersion int64,
	repositoryHeadVersion int64,
) error {
	return &migrationRemediationError{
		report: migrationLineageRemediationReport(source, state, currentVersion, targetVersion, repositoryHeadVersion),
	}
}

func migrationLineageRemediationReport(source *Source, state migrationLineageState, currentVersion int64, targetVersion int64, repositoryHeadVersion int64) migrationRemediationReport {
	var rawValue *string
	if len(state.ObservedIDs) > 0 {
		raw := state.ObservedIDs[0]
		rawValue = &raw
	}
	observedIDs := append([]string{}, state.ObservedIDs...)

	return migrationRemediationReport{
		SchemaID:    migrationRemediationSchemaID,
		Boundary:    sourceLineageBoundary(source),
		FromVersion: currentVersion,
		ToVersion:   targetVersion,
		Findings: []migrationRemediationFinding{
			{
				Field:    "schema_migration_lineage",
				RawValue: rawValue,
				RawValuePair: migrationRemediationFacts{
					CurrentVersion:        currentVersion,
					ExpectedLineageID:     sourceLineageID(source),
					LineageTablePresent:   state.TablePresent,
					ObservedLineageIDs:    observedIDs,
					RepositoryHeadVersion: repositoryHeadVersion,
					TargetVersion:         targetVersion,
				},
				ReasonCode:      reasonHistoricalMigrationLineage,
				RemediationHint: historicalMigrationLineageHint,
			},
		},
	}
}

var _ RemediationReporter = (*migrationRemediationError)(nil)
