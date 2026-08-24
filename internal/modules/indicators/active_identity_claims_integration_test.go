package indicators_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	authstoretest "github.com/JochiRaider/cartulary/internal/modules/auth/testsupport/storetest"
	"github.com/JochiRaider/cartulary/internal/modules/indicators"
	indicatortest "github.com/JochiRaider/cartulary/internal/modules/indicators/testsupport"
	"github.com/JochiRaider/cartulary/internal/testutil/appsupport"
	"github.com/JochiRaider/cartulary/internal/testutil/revisionsupport"
)

func TestIndicatorActiveIdentityClaimsFollowRecordsAndRebuild_Integration(t *testing.T) {
	ctx := context.Background()
	harness := appsupport.StartStore(t, "indicators-active-identity-claims")
	application := newIndicatorTestApplication(t, harness.DB, revisionsupport.MustAppender(t))
	actor := authstoretest.SeedLocalUserRecord(
		t, harness.DB, "indicator-claims@example.test", "Indicator Claims",
		"IndicatorClaimsPass1!", false, false, true,
	)
	incident := appsupport.CreateIncidentInStore(
		t, harness.DB, actor, "txn-indicator-claims-incident", "IR-IND-CLAIMS",
		"Indicator identity claims",
	)
	now := time.Date(2026, 8, 3, 20, 0, 0, 0, time.UTC)

	created, err := application.CreateIndicatorRow(ctx, actor.ID, incident.ID, indicators.CreateCommand{
		ClientTxnID:   "txn-indicator-claims-create",
		IndicatorType: "domain_name",
		ValueKind:     "atomic",
		DisplayValue:  "claims.example",
	}, "request-indicator-claims-create")
	if err != nil {
		t.Fatalf("create Indicator: %v", err)
	}
	dedupeKey := indicatortest.CanonicalDedupeKey(t, "domain_name", "atomic", "claims.example")
	requireActiveIdentityClaim(t, harness, incident.ID, "domain_name", dedupeKey, created.RecordID)

	deletedAt := now.Add(time.Minute)
	tx, err := harness.DB.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		t.Fatalf("begin delete bridge transaction: %v", err)
	}
	if _, err := tx.Exec(ctx, `
UPDATE records
   SET deleted_at = $2, deleted_by_user_id = $3, updated_at = $2,
       updated_by_user_id = $3, row_version = row_version + 1
 WHERE record_id = $1
`, created.RecordID, deletedAt, actor.ID); err != nil {
		t.Fatalf("delete Records envelope: %v", err)
	}
	if err := tx.Commit(ctx); err != nil {
		t.Fatalf("commit delete bridge transaction: %v", err)
	}
	requireNoActiveIdentityClaim(t, harness, created.RecordID)

	replacement, err := application.CreateIndicatorRow(ctx, actor.ID, incident.ID, indicators.CreateCommand{
		ClientTxnID:   "txn-indicator-claims-reuse",
		IndicatorType: "domain_name",
		ValueKind:     "atomic",
		DisplayValue:  "claims.example",
	}, "request-indicator-claims-reuse")
	if err != nil {
		t.Fatalf("reuse identity after deletion: %v", err)
	}
	if replacement.RecordID == created.RecordID {
		t.Fatal("identity reuse returned the deleted record")
	}
	requireActiveIdentityClaim(t, harness, incident.ID, "domain_name", dedupeKey, replacement.RecordID)

	conflictTx, err := harness.DB.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		t.Fatalf("begin restore-conflict transaction: %v", err)
	}
	_, conflictErr := conflictTx.Exec(ctx, `
UPDATE records
   SET deleted_at = NULL, deleted_by_user_id = NULL
 WHERE record_id = $1
`, created.RecordID)
	_ = conflictTx.Rollback(ctx)
	if conflictErr == nil {
		t.Fatal("restoring a duplicate active identity unexpectedly succeeded")
	}
	requireActiveIdentityClaim(t, harness, incident.ID, "domain_name", dedupeKey, replacement.RecordID)

	if _, err := harness.DB.Exec(ctx, `DELETE FROM indicator_active_identities WHERE indicator_record_id = $1`, replacement.RecordID); err != nil {
		t.Fatalf("remove rebuildable claim: %v", err)
	}
	if identityClaimsValid(t, harness) {
		t.Fatal("claim validation accepted missing derived state")
	}
	var rebuilt int64
	if err := harness.DB.QueryRow(ctx, `SELECT rebuild_indicator_active_identities()`).Scan(&rebuilt); err != nil {
		t.Fatalf("rebuild active identity claims: %v", err)
	}
	if rebuilt != 1 {
		t.Fatalf("rebuilt claim count = %d, want 1", rebuilt)
	}
	if !identityClaimsValid(t, harness) {
		t.Fatal("claim validation rejected deterministic rebuild")
	}
	requireActiveIdentityClaim(t, harness, incident.ID, "domain_name", dedupeKey, replacement.RecordID)

	// Recovery disables ordinary triggers. ENABLE ALWAYS must still reconstruct a
	// claim as the Records envelope arrives after its subtype row.
	recoveryTx, err := harness.DB.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		t.Fatalf("begin recovery-order transaction: %v", err)
	}
	if _, err := recoveryTx.Exec(ctx, `SET LOCAL session_replication_role = replica`); err != nil {
		t.Fatalf("enter recovery trigger mode: %v", err)
	}
	if _, err := recoveryTx.Exec(ctx, `DELETE FROM indicator_active_identities WHERE indicator_record_id = $1`, replacement.RecordID); err != nil {
		t.Fatalf("clear recovery claim: %v", err)
	}
	if _, err := recoveryTx.Exec(ctx, `
UPDATE records
   SET deleted_at = $2, deleted_by_user_id = $3
 WHERE record_id = $1
`, replacement.RecordID, deletedAt, actor.ID); err != nil {
		t.Fatalf("replay deleted Records envelope: %v", err)
	}
	if _, err := recoveryTx.Exec(ctx, `
UPDATE records
   SET deleted_at = NULL, deleted_by_user_id = NULL
 WHERE record_id = $1
`, replacement.RecordID); err != nil {
		t.Fatalf("replay active Records envelope: %v", err)
	}
	var recoveryClaimCount int
	if err := recoveryTx.QueryRow(ctx, `SELECT count(*) FROM indicator_active_identities WHERE indicator_record_id = $1`, replacement.RecordID).Scan(&recoveryClaimCount); err != nil {
		t.Fatalf("count recovery claim: %v", err)
	}
	if recoveryClaimCount != 1 {
		t.Fatalf("recovery-order claim count = %d, want 1", recoveryClaimCount)
	}
	if err := recoveryTx.Rollback(ctx); err != nil {
		t.Fatalf("roll back recovery-order probe: %v", err)
	}
}

func requireActiveIdentityClaim(
	t testing.TB,
	harness *appsupport.StoreHarness,
	incidentID uuid.UUID,
	indicatorType string,
	dedupeKey string,
	recordID uuid.UUID,
) {
	t.Helper()
	var claimed uuid.UUID
	if err := harness.DB.QueryRow(context.Background(), `
SELECT indicator_record_id
  FROM indicator_active_identities
 WHERE incident_id = $1
   AND indicator_type = $2
   AND dedupe_key = $3
`, incidentID, indicatorType, dedupeKey).Scan(&claimed); err != nil {
		t.Fatalf("load active identity claim: %v", err)
	}
	if claimed != recordID {
		t.Fatalf("active identity claim = %s, want %s", claimed, recordID)
	}
}

func requireNoActiveIdentityClaim(t testing.TB, harness *appsupport.StoreHarness, recordID uuid.UUID) {
	t.Helper()
	var count int
	if err := harness.DB.QueryRow(context.Background(), `
SELECT count(*) FROM indicator_active_identities WHERE indicator_record_id = $1
`, recordID).Scan(&count); err != nil {
		t.Fatalf("count active identity claim: %v", err)
	}
	if count != 0 {
		t.Fatalf("active identity claim count = %d, want 0", count)
	}
}

func identityClaimsValid(t testing.TB, harness *appsupport.StoreHarness) bool {
	t.Helper()
	var valid bool
	if err := harness.DB.QueryRow(context.Background(), `SELECT indicator_active_identities_are_valid()`).Scan(&valid); err != nil {
		t.Fatalf("validate active identity claims: %v", err)
	}
	return valid
}
