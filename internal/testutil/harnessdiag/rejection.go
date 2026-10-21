package harnessdiag

import (
	"errors"
	"fmt"
	"slices"
)

type observationRejection struct{ role, stage, reason string }

func (e *observationRejection) Error() string { return "observation rejected" }
func (e *observationRejection) Unwrap() error { return errArtifact }
func rejected(role, stage, reason string) error {
	return &observationRejection{role: role, stage: stage, reason: reason}
}

// Only closed codes cross the command boundary. The JSON projection stays v3.
func rejectionLine(err error) string {
	role, stage, reason := "observation", "projection", "invalid_document"
	var rejection *observationRejection
	if errors.As(err, &rejection) {
		role, stage, reason = rejection.role, rejection.stage, rejection.reason
	}
	if !slices.Contains([]string{"invocation", "manifest", "run_summary", "target_summary", "tool_summary", "live_snapshot", "observation"}, role) {
		role = "observation"
	}
	if !slices.Contains([]string{"opening", "initial_metadata", "reading", "final_metadata", "decoding", "schema", "identity", "projection"}, stage) {
		stage = "projection"
	}
	if !slices.Contains([]string{"directory_unavailable", "invalid_name", "file_unavailable", "stat_unavailable", "not_regular", "foreign_owner", "unsafe_mode", "size_limit", "unsafe_links", "read_failed", "length_changed", "size_changed", "mtime_changed", "ctime_changed", "identity_changed", "mode_changed", "owner_changed", "links_changed", "invalid_document", "invalid_identity", "output_unavailable"}, reason) {
		reason = "invalid_document"
	}
	return fmt.Sprintf("[ARTIFACT] role=%s stage=%s reason=%s\n", role, stage, reason)
}
