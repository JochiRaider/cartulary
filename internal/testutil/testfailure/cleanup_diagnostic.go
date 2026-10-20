package testfailure

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"reflect"

	"github.com/JochiRaider/cartulary/internal/testutil/postgrescleanup"
)

// CleanupDiagnostic intentionally excludes error text and resource identity.
// Stage, cause and optional wait details are checked again when encoding.
type CleanupDiagnostic struct {
	Stage    string                       `json:"stage"`
	Cause    string                       `json:"cause"`
	Postgres *postgrescleanup.DropAttempt `json:"postgres,omitempty"`
}

func validCleanupStage(stage string) bool {
	return member(stage, "unknown", "catalog_open", "catalog_connection", "catalog_lock", "catalog_unlock", "catalog_close", "forced_drop")
}

func validCleanupCause(cause string) bool {
	return member(cause, "unknown", "cancelled", "deadline_exceeded", "connection_closed", "database_in_use", "connection_failure", "lock_unavailable", "query_cancelled", "permission_denied", "server_shutdown", "diagnostic_limit")
}

func describeCleanupFailure(err error) []CleanupDiagnostic {
	var result []CleanupDiagnostic
	visited := 0
	limited := false
	appendDiagnostic := func(stage, cause string, observation *postgrescleanup.DropAttempt) {
		diagnostic := CleanupDiagnostic{Stage: stage, Cause: cause, Postgres: observation}
		for _, existing := range result {
			if reflect.DeepEqual(existing, diagnostic) {
				return
			}
		}
		if len(result) == 8 {
			limited = true
			return
		}
		result = append(result, diagnostic)
	}
	var visit func(error, string, *postgrescleanup.DropAttempt)
	visit = func(current error, stage string, observation *postgrescleanup.DropAttempt) {
		visited++
		if visited > 32 {
			limited = true
			return
		}
		if staged, ok := current.(interface{ CleanupStage() string }); ok {
			stage = staged.CleanupStage()
			if !validCleanupStage(stage) {
				stage = "unknown"
			}
		}
		if observed, ok := current.(interface {
			PostgresCleanup() postgrescleanup.DropAttempt
		}); ok {
			value := observed.PostgresCleanup()
			if value.Validate() != nil {
				value = postgrescleanup.UnavailableAttempt()
			}
			observation = &value
		}
		// Inspect typed causes before unwrapping: driver errors may carry a
		// normalized SQLSTATE and also wrap a lower-level transport error.
		cause := cleanupCause(current)
		if cause != "unknown" {
			appendDiagnostic(stage, cause, observation)
			return
		}
		switch wrapped := current.(type) {
		case interface{ Unwrap() []error }:
			for _, child := range wrapped.Unwrap() {
				if child != nil {
					visit(child, stage, observation)
				}
				if limited {
					break
				}
			}
		case interface{ Unwrap() error }:
			visit(wrapped.Unwrap(), stage, observation)
		default:
			appendDiagnostic(stage, "unknown", observation)
		}
	}
	visit(err, "unknown", nil)
	if limited {
		if len(result) == 8 {
			result = result[:7]
		}
		result = append(result, CleanupDiagnostic{Stage: "unknown", Cause: "diagnostic_limit"})
	}
	if len(result) == 0 {
		result = append(result, CleanupDiagnostic{Stage: "unknown", Cause: "unknown"})
	}
	return result
}

func cleanupCause(err error) string {
	// Equality avoids unbounded traversal of an untrusted error chain. The
	// caller performs its own bounded walk and never calls Error().
	switch err {
	case context.Canceled:
		return "cancelled"
	case context.DeadlineExceeded:
		return "deadline_exceeded"
	case sql.ErrConnDone, driver.ErrBadConn:
		return "connection_closed"
	}
	if state, ok := err.(interface{ SQLState() string }); ok {
		switch state.SQLState() {
		case "55006":
			return "database_in_use"
		case "08000", "08001", "08003", "08004", "08006", "08007", "08P01":
			return "connection_failure"
		case "55P03":
			return "lock_unavailable"
		case "57014":
			return "query_cancelled"
		case "42501":
			return "permission_denied"
		case "57P01", "57P02", "57P03":
			return "server_shutdown"
		}
	}
	return "unknown"
}
