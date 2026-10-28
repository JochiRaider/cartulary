package entities_test

import (
	"testing"

	"github.com/JochiRaider/cartulary/internal/testutil/pgtest"
)

func TestRecordEnvelopeHeadSchema(t *testing.T) {
	db := pgtest.Start(t).BeginRollbackDBT(t, "entity_linking-records-head")

	requireColumns(t, db, "records", "record_id", "incident_id", "record_type", "created_by_user_id", "created_at", "updated_by_user_id", "updated_at", "row_version", "deleted_at", "deleted_by_user_id")
	requireNoColumns(t, db, "record_id", "users", "user_sessions", "bootstrap_tokens", "pending_totp_enrollments", "incident_memberships", "deployment_bootstrap_state")
	requireGenericRecordFKsDoNotTargetTimeline(t, db)
}
