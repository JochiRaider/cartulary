package recovery

import (
	"strings"
	"testing"
)

func TestTargetControlServingAdmission(t *testing.T) {
	registry, err := loadVNextRecoveryGenerationRegistry()
	if err != nil {
		t.Fatal(err)
	}
	text := func(s string) *string { return &s }
	bindings := TargetBindingDigests{strings.Repeat("1", 64), strings.Repeat("2", 64), strings.Repeat("3", 64), strings.Repeat("4", 64)}
	complete := TargetControl{
		SchemaID: TargetStateSchemaID, State: TargetComplete, Revision: 8,
		TargetGenerationID: text("10000000-0000-4000-8000-000000000001"), Purpose: text("restore_target"),
		Operation: text("restore_latest"), OperationID: text("20000000-0000-4000-8000-000000000001"),
		BackupSetID:                text("30000000-0000-4000-8000-000000000001"),
		RecoveryStateCatalogSHA256: text(registry.current.stateCatalog.DigestSHA256()), CodecRegistrySHA256: text(registry.current.codecRegistrySHA256),
		BindingDigests: &bindings, CompletionEvidenceSHA256: text(strings.Repeat("5", 64)),
	}
	for _, tc := range []struct {
		name    string
		change  func(*TargetControl)
		valid   bool
		serving bool
	}{
		{"completed ordinary", func(*TargetControl) {}, true, true},
		{"portable ordinary", func(c *TargetControl) { c.Operation = text("restore_bundle") }, true, true},
		{"initialized", func(c *TargetControl) {
			*c = TargetControl{SchemaID: TargetStateSchemaID, State: TargetInitialized, Revision: 1}
		}, true, true},
		{"unproven", func(c *TargetControl) {
			*c = TargetControl{SchemaID: TargetStateSchemaID, State: TargetUnproven, Revision: 1}
		}, true, false},
		{"restoring", func(c *TargetControl) { c.State = TargetRestoring; c.CompletionEvidenceSHA256 = nil }, true, false},
		{"completed verification", func(c *TargetControl) {
			c.Purpose = text("restore_verification_target")
			c.Operation = text("restore_verify_latest")
		}, true, false},
		{"completed due", func(c *TargetControl) {
			c.Purpose = text("restore_verification_target")
			c.Operation = text("restore_verify_due")
			c.AttemptID = text("attempt.1")
		}, true, false},
		{"ready verification", func(c *TargetControl) {
			*c = TargetControl{SchemaID: TargetStateSchemaID, State: TargetVerificationReady, Revision: 7, TargetGenerationID: c.TargetGenerationID, Purpose: text("restore_verification_target"), BindingDigests: &bindings}
		}, true, false},
		{"resetting verification", func(c *TargetControl) {
			*c = TargetControl{SchemaID: TargetStateSchemaID, State: TargetResetting, Revision: 7, TargetGenerationID: c.TargetGenerationID, Purpose: text("restore_verification_target"), BindingDigests: &bindings}
		}, true, false},
		{"retired schema", func(c *TargetControl) { c.SchemaID = "cartulary.recovery_target_state.v0" }, false, false},
		{"unknown state", func(c *TargetControl) { c.State = "healthy" }, false, false},
		{"invalid revision", func(c *TargetControl) { c.Revision = 0 }, false, false},
		{"unproven with evidence", func(c *TargetControl) { c.State = TargetUnproven }, false, false},
		{"initialized with evidence", func(c *TargetControl) { c.State = TargetInitialized }, false, false},
		{"ready with operation", func(c *TargetControl) { c.State = TargetVerificationReady }, false, false},
		{"resetting with completion", func(c *TargetControl) { c.State = TargetResetting }, false, false},
		{"missing generation", func(c *TargetControl) { c.TargetGenerationID = nil }, false, false},
		{"zero generation", func(c *TargetControl) { c.TargetGenerationID = text("00000000-0000-0000-0000-000000000000") }, false, false},
		{"noncanonical generation", func(c *TargetControl) { c.TargetGenerationID = text("{10000000-0000-4000-8000-000000000001}") }, false, false},
		{"missing operation", func(c *TargetControl) { c.Operation = nil }, false, false},
		{"unsupported operation", func(c *TargetControl) { c.Operation = text("backup") }, false, false},
		{"operation purpose mismatch", func(c *TargetControl) { c.Purpose = text("restore_verification_target") }, false, false},
		{"missing operation id", func(c *TargetControl) { c.OperationID = nil }, false, false},
		{"missing backup id", func(c *TargetControl) { c.BackupSetID = nil }, false, false},
		{"unexpected attempt", func(c *TargetControl) { c.AttemptID = text("attempt.1") }, false, false},
		{"missing due attempt", func(c *TargetControl) {
			c.Purpose = text("restore_verification_target")
			c.Operation = text("restore_verify_due")
		}, false, false},
		{"invalid due attempt", func(c *TargetControl) {
			c.Purpose = text("restore_verification_target")
			c.Operation = text("restore_verify_due")
			c.AttemptID = text("unsafe/attempt")
		}, false, false},
		{"missing catalog", func(c *TargetControl) { c.RecoveryStateCatalogSHA256 = nil }, false, false},
		{"missing codec", func(c *TargetControl) { c.CodecRegistrySHA256 = nil }, false, false},
		{"missing bindings", func(c *TargetControl) { c.BindingDigests = nil }, false, false},
		{"malformed bindings", func(c *TargetControl) { b := bindings; b.DatabaseSHA256 = "invalid"; c.BindingDigests = &b }, false, false},
		{"missing completion", func(c *TargetControl) { c.CompletionEvidenceSHA256 = nil }, false, false},
		{"noncanonical completion", func(c *TargetControl) { c.CompletionEvidenceSHA256 = text(strings.Repeat("A", 64)) }, false, false},
		{"restoring with completion", func(c *TargetControl) { c.State = TargetRestoring }, false, false},
		{"stale catalog", func(c *TargetControl) { c.RecoveryStateCatalogSHA256 = text(strings.Repeat("6", 64)) }, true, false},
		{"stale codec", func(c *TargetControl) { c.CodecRegistrySHA256 = text(strings.Repeat("7", 64)) }, true, false},
		{"wrong physical binding", func(c *TargetControl) {
			b := bindings
			b.DatabaseSHA256 = strings.Repeat("8", 64)
			c.BindingDigests = &b
		}, true, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			control := complete
			tc.change(&control)
			if err := control.Validate(); (err == nil) != tc.valid {
				t.Fatalf("control validity: %v", err)
			}
			if err := control.RequireServing(bindings); (err == nil) != tc.serving {
				t.Fatalf("serving admission: %v", err)
			}
		})
	}
}
