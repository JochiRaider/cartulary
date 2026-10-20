//go:build linux

package harnesssession

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/JochiRaider/cartulary/internal/testutil/harnessschema"
)

// cleanupError preserves the native cause for classification while identifying
// the physical boundary that failed. It never grants cleanup authority.
type cleanupError struct {
	stage string
	err   error
}

// Identity context is observational. It never changes the underlying rejection
// or contains a kernel identity, private path or provider error string.
type identityError struct {
	boundary string
	producer string
	err      error
}

func (e *identityError) Error() string { return "runtime process identity unresolved" }
func (e *identityError) Unwrap() error { return e.err }

func atIdentityBoundary(boundary string, err error) error {
	if err == nil {
		return nil
	}
	return &identityError{boundary: boundary, err: err}
}

func atDetailProducer(ordinal int, err error) error {
	if err == nil {
		return nil
	}
	var identity *identityError
	boundary := ""
	if errors.As(err, &identity) {
		boundary = identity.boundary
	}
	return &identityError{boundary: boundary, producer: fmt.Sprintf("producer:%04d", ordinal), err: err}
}

func (e *cleanupError) Error() string { return "runtime cleanup failed" }
func (e *cleanupError) Unwrap() error { return e.err }

// AtCleanupStage preserves the innermost native cleanup boundary and cause.
// Public diagnostics still validate stages against the closed schema.
func AtCleanupStage(stage string, err error) error {
	if err == nil {
		return nil
	}
	var existing *cleanupError
	if errors.As(err, &existing) {
		return err
	}
	return &cleanupError{stage: stage, err: err}
}

// CleanupDiagnostic is a non-authoritative terminal diagnostic. Unlike private
// protocol rejections, it contains no process identity, proof path or payload.
type CleanupDiagnostic struct {
	Schema           string `json:"schema_id"`
	Stage            string `json:"stage"`
	Cause            string `json:"cause"`
	Errno            int    `json:"errno"`
	IdentityBoundary string `json:"identity_boundary,omitempty"`
	ProducerRef      string `json:"producer_ref,omitempty"`
}

const cleanupDiagnosticSchema = "cartulary.harness_cleanup_diagnostic.v4"

func (d *CleanupDiagnostic) Error() string { return d.String() }

// DecodeCleanupDiagnostic accepts only the bounded safe projection, never the
// private runtime rejection (which may contain process identity and proof paths).
func DecodeCleanupDiagnostic(data []byte) *CleanupDiagnostic {
	if len(data) > 4096 {
		return nil
	}
	value, err := harnessschema.Decode(data)
	if err != nil || harnessschema.Validate(cleanupDiagnosticSchema, value) != nil {
		return nil
	}
	var diagnostic CleanupDiagnostic
	if json.Unmarshal(data, &diagnostic) != nil {
		return nil
	}
	return &diagnostic
}

func (d CleanupDiagnostic) String() string {
	return fmt.Sprintf("[RUNTIME-CLEANUP] stage=%s cause=%s errno=%d", d.Stage, d.Cause, d.Errno)
}

func DescribeCleanupFailure(err error) *CleanupDiagnostic {
	if err == nil {
		return nil
	}
	var imported *CleanupDiagnostic
	if errors.As(err, &imported) {
		data, marshalErr := json.Marshal(imported)
		if marshalErr == nil {
			if validated := DecodeCleanupDiagnostic(data); validated != nil {
				return validated
			}
		}
	}
	d := &CleanupDiagnostic{Schema: cleanupDiagnosticSchema, Stage: "runtime_close"}
	var staged *cleanupError
	if errors.As(err, &staged) {
		switch staged.stage {
		case "resource_inventory", "resource_graph", "resource_release", "resident_release", "physical_command", "fixture_publication", "managed_service_settlement", "fixture_release", "fixture_detach", "fixture_acquisition_cleanup", "provider_close", "host_release", "services_close", "recovery_preserve", "detail_producer", "detail_descendants", "detail_remove", "detail_authority", "process_discovery", "process_identity", "process_group_anchor", "process_handle", "process_signal", "process_authority", "process_settlement", "recovery_proof", "recovery_purge":
			d.Stage = staged.stage
		}
	}
	// Copy only the closed fields. The private rejection's PID and proof-file
	// diagnostics must never cross this public boundary.
	r := rejectReason(err)
	d.Cause = r.Reason
	if r.Errno != nil {
		d.Errno = *r.Errno
	}
	var identity *identityError
	if errors.As(err, &identity) {
		switch identity.boundary {
		case "producer_identity", "group_session", "pre_signal_identity", "boot_identity", "process_read", "descendant_membership", "group_anchor":
			d.IdentityBoundary = identity.boundary
		}
		d.ProducerRef = identity.producer
	}
	return d
}
