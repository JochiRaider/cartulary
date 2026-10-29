package incidents_test

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	authstoretest "github.com/JochiRaider/cartulary/internal/modules/auth/testsupport/storetest"
	"github.com/JochiRaider/cartulary/internal/modules/incidents"
	"github.com/JochiRaider/cartulary/internal/modules/incidents/testsupport/admissiontest"
	"github.com/JochiRaider/cartulary/internal/modules/incidents/testsupport/storetest"
	workbookstartuppostgres "github.com/JochiRaider/cartulary/internal/modules/workbook/startup/postgres"
	"github.com/JochiRaider/cartulary/internal/platform/postgres"
	"github.com/JochiRaider/cartulary/internal/testutil/pgtest"
)

func TestMembershipVersionHistorySurvivesDeletionAndRollback_Integration(t *testing.T) {
	ctx := t.Context()
	database := pgtest.Start(t).PrepareIsolatedDatabaseT(t, "membership-history")
	pool, err := pgxpool.New(ctx, database.DSN)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(pool.Close)
	newApplication := func(db postgres.DB) *incidents.Application {
		a, err := incidents.NewApplication(incidents.ApplicationDependencies{Postgres: db, PreferenceBootstrap: workbookstartuppostgres.NewWriter(), Now: time.Now})
		if err != nil {
			t.Fatal(err)
		}
		return a
	}
	application := newApplication(pool)
	actor := authstoretest.SeedActor(t, pool, authstoretest.ActorAttributes{Email: "history-admin@example.test", DisplayName: "Admin"})
	target := authstoretest.SeedActor(t, pool, authstoretest.ActorAttributes{Email: "history-target@example.test", DisplayName: "Member"})
	id := storetest.CreateIncidentInStore(t, application, actor, "history-incident", "HISTORY", "Version history").Incident.ID
	request := admissiontest.MembershipCreate(t, admissiontest.MembershipCreateInput{ClientTxnID: "history-create", UserID: &target.ID, Role: "viewer"})
	first, err := application.CreateMembership(ctx, actor, id, request, "first")
	if err != nil || first.Membership.MembershipVersion != 1 {
		t.Fatalf("first membership=%#v err=%v", first, err)
	}
	patch := admissiontest.MembershipPatch(t, admissiontest.MembershipPatchInput{BaseMembershipVersion: 1, Role: "editor"})
	failing := newApplication(auditFailureDB{DB: pool})
	if _, _, err := failing.UpdateMembership(ctx, actor, id, target.ID, patch, "fail-audit"); !errors.Is(err, errAuditUnavailable) {
		t.Fatalf("audit failure=%v", err)
	}
	var high int64
	readHigh := func() int64 {
		if err := pool.QueryRow(ctx, `SELECT last_version FROM incident_membership_versions WHERE incident_id=$1 AND user_id=$2`, id, target.ID).Scan(&high); err != nil {
			t.Fatal(err)
		}
		return high
	}
	if readHigh() != 1 {
		t.Fatal("rolled-back mutation consumed a version")
	}
	changed, _, err := application.UpdateMembership(ctx, actor, id, target.ID, patch, "patch")
	if err != nil || changed.MembershipVersion != 2 {
		t.Fatalf("role change=%#v err=%v", changed, err)
	}
	noop := admissiontest.MembershipPatch(t, admissiontest.MembershipPatchInput{BaseMembershipVersion: 2, Role: "editor"})
	if _, material, err := application.UpdateMembership(ctx, actor, id, target.ID, noop, "noop"); err != nil || material || readHigh() != 2 {
		t.Fatalf("no-op changed history: %v", err)
	}
	if _, err := application.DeleteMembership(ctx, actor, id, target.ID, admissiontest.MembershipDelete(t, 2), "delete"); err != nil {
		t.Fatal(err)
	}
	if readHigh() != 2 {
		t.Fatal("deletion lost version history")
	}
	recreate := admissiontest.MembershipCreate(t, admissiontest.MembershipCreateInput{ClientTxnID: "history-recreate", UserID: &target.ID, Role: "viewer"})
	second, err := application.CreateMembership(ctx, actor, id, recreate, "recreate")
	if err != nil || second.Membership.MembershipVersion != 3 {
		t.Fatalf("recreation=%#v err=%v", second, err)
	}
	if _, _, err := application.UpdateMembership(ctx, actor, id, target.ID, patch, "stale-patch"); !errors.Is(err, incidents.ErrMembershipVersionConflict) {
		t.Fatalf("stale patch=%v", err)
	}
	if _, err := application.DeleteMembership(ctx, actor, id, target.ID, admissiontest.MembershipDelete(t, 1), "stale-delete"); !errors.Is(err, incidents.ErrMembershipVersionConflict) {
		t.Fatalf("stale delete=%v", err)
	}
	if _, err := application.CreateMembership(ctx, actor, id, request, "replay-old"); err != nil || readHigh() != 3 {
		t.Fatalf("replay consumed a version: %v", err)
	}
	if _, err := pool.Exec(ctx, `UPDATE incident_membership_versions SET last_version=9007199254740991 WHERE incident_id=$1 AND user_id=$2`, id, target.ID); err != nil {
		t.Fatal(err)
	}
	if _, _, err := application.UpdateMembership(ctx, actor, id, target.ID, admissiontest.MembershipPatch(t, admissiontest.MembershipPatchInput{BaseMembershipVersion: 3, Role: "editor"}), "exhausted"); err == nil {
		t.Fatal("exhausted allocator accepted a mutation")
	}
	if got := storetest.GetMembership(t, pool, id, target.ID); got.MembershipVersion != 3 || got.Role != "viewer" {
		t.Fatalf("failed allocation changed membership: %#v", got)
	}
}

var errAuditUnavailable = errors.New("test: audit storage unavailable")

type auditFailureDB struct{ postgres.DB }

func (db auditFailureDB) BeginTx(ctx context.Context, options pgx.TxOptions) (pgx.Tx, error) {
	tx, err := db.DB.BeginTx(ctx, options)
	if err != nil {
		return nil, err
	}
	return auditFailureTx{tx}, nil
}

type auditFailureTx struct{ pgx.Tx }

func (tx auditFailureTx) Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error) {
	if strings.Contains(sql, "INSERT INTO deployment_admin_audit_events") {
		return pgconn.CommandTag{}, errAuditUnavailable
	}
	return tx.Tx.Exec(ctx, sql, args...)
}
