package main

import (
	"encoding/json"
	"errors"
	"os"
	"time"

	"github.com/JochiRaider/cartulary/internal/platform/harnessruntime"
)

const databaseResetDiagnosticSchemaID = "cartulary.test.database_reset_diagnostic.v2"

type databaseResetSettlementFailure struct {
	Stage    string  `json:"stage"`
	SQLState *string `json:"sqlstate"`
	TimedOut bool    `json:"timed_out"`
}

type databaseResetDiagnostic struct {
	SettlementFailures         []databaseResetSettlementFailure `json:"settlement_failures"`
	TransactionSettlement      string                           `json:"transaction_settlement"`
	SchemaID                   string                           `json:"schema_id"`
	ResetID                    string                           `json:"reset_id"`
	Status                     string                           `json:"status"`
	Attempt                    int                              `json:"attempt"`
	Stage                      string                           `json:"stage"`
	SQLState                   *string                          `json:"sqlstate"`
	TimedOut                   bool                             `json:"timed_out"`
	DurationMS                 int64                            `json:"duration_ms"`
	TablesReset                []string                         `json:"tables_reset"`
	TableCounts                []databaseResetTableCount        `json:"table_counts"`
	MutableTableCount          int                              `json:"mutable_table_count"`
	MigrationMetadataPreserved bool                             `json:"migration_metadata_preserved"`
	BootstrapAdminRestored     bool                             `json:"bootstrap_admin_restored"`
	PostResetCounts            databaseResetDiagnosticCounts    `json:"post_reset_counts"`
	FailureClass               *string                          `json:"failure_class"`
	FailureReason              *string                          `json:"failure_reason"`
}

type databaseResetTableCount struct {
	Table string `json:"table"`
	Rows  int    `json:"rows"`
}

type databaseResetDiagnosticCounts struct {
	ActiveDeploymentAdmins int `json:"active_deployment_admins"`
	BootstrapMarkers       int `json:"bootstrap_markers"`
	Incidents              int `json:"incidents"`
	Records                int `json:"records"`
	UserSessions           int `json:"user_sessions"`
	RouteIdempotency       int `json:"route_idempotency"`
}

func newDatabaseResetDiagnostic(resetID string, result harnessruntime.DatabaseResetResult, resetErr error, duration time.Duration) databaseResetDiagnostic {
	diagnostic := databaseResetDiagnostic{
		SchemaID:                   databaseResetDiagnosticSchemaID,
		SettlementFailures:         []databaseResetSettlementFailure{},
		TransactionSettlement:      result.TransactionSettlement,
		ResetID:                    resetID,
		Status:                     "pass",
		Attempt:                    1,
		Stage:                      "complete",
		DurationMS:                 max(duration.Milliseconds(), 0),
		TablesReset:                append([]string{}, result.TablesReset...),
		TableCounts:                make([]databaseResetTableCount, 0, len(result.TableCounts)),
		MutableTableCount:          result.MutableTableCount,
		MigrationMetadataPreserved: result.MigrationMetadataPreserved,
		BootstrapAdminRestored:     result.BootstrapAdminRestored,
		PostResetCounts: databaseResetDiagnosticCounts{
			ActiveDeploymentAdmins: result.PostResetCounts.ActiveDeploymentAdmins,
			BootstrapMarkers:       result.PostResetCounts.BootstrapMarkers,
			Incidents:              result.PostResetCounts.Incidents,
			Records:                result.PostResetCounts.Records,
			UserSessions:           result.PostResetCounts.UserSessions,
			RouteIdempotency:       result.PostResetCounts.RouteIdempotency,
		},
	}
	if diagnostic.TransactionSettlement == "" {
		diagnostic.TransactionSettlement = "not_started"
	}
	for _, failure := range result.SettlementFailures {
		entry := databaseResetSettlementFailure{Stage: failure.Stage(), TimedOut: failure.TimedOut()}
		if state := failure.SQLState(); state != "" {
			entry.SQLState = &state
		}
		diagnostic.SettlementFailures = append(diagnostic.SettlementFailures, entry)
	}
	for _, count := range result.TableCounts {
		diagnostic.TableCounts = append(diagnostic.TableCounts, databaseResetTableCount{
			Table: count.Table,
			Rows:  count.Rows,
		})
	}
	if resetErr == nil {
		return diagnostic
	}
	diagnostic.Status = "fail"
	diagnostic.Stage = "unknown"
	diagnostic.FailureClass = stringPointer("harness")
	diagnostic.FailureReason = stringPointer("fixture_error")
	var failure *harnessruntime.DatabaseResetFailure
	if errors.As(resetErr, &failure) {
		diagnostic.Stage = failure.Stage()
		diagnostic.TimedOut = failure.TimedOut()
		if sqlState := failure.SQLState(); sqlState != "" {
			diagnostic.SQLState = &sqlState
		}
	}
	if diagnostic.TimedOut {
		diagnostic.FailureClass = stringPointer("timing")
		diagnostic.FailureReason = stringPointer("timeout_failure")
	}
	return diagnostic
}

func writeDatabaseResetDiagnostic(path string, diagnostic databaseResetDiagnostic) error {
	payload, err := json.MarshalIndent(diagnostic, "", "  ")
	if err != nil {
		return err
	}
	if err := writeFileAtomic(path, append(payload, '\n'), 0o600); err != nil {
		return err
	}
	return os.Chmod(path, 0o600)
}

func stringPointer(value string) *string {
	return &value
}
