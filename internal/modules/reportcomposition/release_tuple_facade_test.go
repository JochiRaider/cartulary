package reportcomposition

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/testutil/pgtest"
)

func TestResolveReleaseTupleTxReasonMatrix(t *testing.T) {
	db := pgtest.Start(t).BeginRollbackDBT(t, "reportcomposition-release-tuple")
	ctx := context.Background()
	fixture := seedReleaseTupleFixture(t, ctx, db)

	tx, err := db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		t.Fatalf("begin tx: %v", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	resolved, err := ResolveReleaseTupleTx(ctx, tx, fixture.incidentID, "cartulary.report.default", "1", ReleaseTuple{
		CompositionID:      fixture.compositionID,
		CompositionVersion: "v2",
		CompositionSHA256:  fixture.digest,
	})
	if err != nil {
		t.Fatalf("resolve valid tuple: %v", err)
	}
	if resolved.VersionNumber != 2 || string(resolved.CanonicalComposition) != `{"schema_id":"fixture"}` {
		t.Fatalf("resolved tuple = %#v", resolved)
	}

	cases := []struct {
		name            string
		incidentID      uuid.UUID
		templateID      string
		templateVersion string
		tuple           ReleaseTuple
		wantField       string
		wantReason      string
	}{
		{
			name:            "invalid version",
			incidentID:      fixture.incidentID,
			templateID:      "cartulary.report.default",
			templateVersion: "1",
			tuple: ReleaseTuple{
				CompositionID:      fixture.compositionID,
				CompositionVersion: "2",
				CompositionSHA256:  fixture.digest,
			},
			wantField:  "composition_version",
			wantReason: ReleaseTupleReasonInvalidCompositionVersion,
		},
		{
			name:            "missing resource",
			incidentID:      fixture.incidentID,
			templateID:      "cartulary.report.default",
			templateVersion: "1",
			tuple: ReleaseTuple{
				CompositionID:      uuid.New(),
				CompositionVersion: "v2",
				CompositionSHA256:  fixture.digest,
			},
			wantField:  "composition_id",
			wantReason: ReleaseTupleReasonCompositionNotFound,
		},
		{
			name:            "cross incident hidden",
			incidentID:      fixture.otherIncidentID,
			templateID:      "cartulary.report.default",
			templateVersion: "1",
			tuple: ReleaseTuple{
				CompositionID:      fixture.compositionID,
				CompositionVersion: "v2",
				CompositionSHA256:  fixture.digest,
			},
			wantField:  "composition_id",
			wantReason: ReleaseTupleReasonCompositionNotFound,
		},
		{
			name:            "template mismatch",
			incidentID:      fixture.incidentID,
			templateID:      "cartulary.report.other",
			templateVersion: "1",
			tuple: ReleaseTuple{
				CompositionID:      fixture.compositionID,
				CompositionVersion: "v2",
				CompositionSHA256:  fixture.digest,
			},
			wantField:  "composition_id",
			wantReason: ReleaseTupleReasonTemplateMismatch,
		},
		{
			name:            "missing version",
			incidentID:      fixture.incidentID,
			templateID:      "cartulary.report.default",
			templateVersion: "1",
			tuple: ReleaseTuple{
				CompositionID:      fixture.compositionID,
				CompositionVersion: "v3",
				CompositionSHA256:  fixture.digest,
			},
			wantField:  "composition_version",
			wantReason: ReleaseTupleReasonVersionNotFound,
		},
		{
			name:            "digest mismatch",
			incidentID:      fixture.incidentID,
			templateID:      "cartulary.report.default",
			templateVersion: "1",
			tuple: ReleaseTuple{
				CompositionID:      fixture.compositionID,
				CompositionVersion: "v2",
				CompositionSHA256:  strings.Repeat("b", 64),
			},
			wantField:  "composition_sha256",
			wantReason: ReleaseTupleReasonDigestMismatch,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := ResolveReleaseTupleTx(ctx, tx, tc.incidentID, tc.templateID, tc.templateVersion, tc.tuple)
			var tupleErr *ReleaseTupleError
			if !errors.As(err, &tupleErr) {
				t.Fatalf("resolve tuple err = %T %v", err, err)
			}
			if tupleErr.Field != tc.wantField || tupleErr.ReasonCode != tc.wantReason {
				t.Fatalf("tuple error = %#v want field %q reason %q", tupleErr, tc.wantField, tc.wantReason)
			}
		})
	}
}

func TestBindReleaseTupleTxRecordsReleaseBoundVersion(t *testing.T) {
	db := pgtest.Start(t).BeginRollbackDBT(t, "reportcomposition-release-bind")
	ctx := context.Background()
	fixture := seedReleaseTupleFixture(t, ctx, db)
	releaseID := seedReportingReleaseForTuple(t, ctx, db, fixture)

	tx, err := db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		t.Fatalf("begin tx: %v", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	tuple := ReleaseTuple{
		CompositionID:      fixture.compositionID,
		CompositionVersion: "v2",
		CompositionSHA256:  fixture.digest,
	}
	if err := BindReleaseTupleTx(ctx, tx, releaseID, tuple, ReleaseScopeInternalDraft, fixture.now); err != nil {
		t.Fatalf("bind release tuple: %v", err)
	}
	if err := BindReleaseTupleTx(ctx, tx, releaseID, tuple, ReleaseScopeInternalDraft, fixture.now); err != nil {
		t.Fatalf("rebind release tuple: %v", err)
	}
	bound, err := releaseBoundVersionsTx(ctx, tx, fixture.compositionID)
	if err != nil {
		t.Fatalf("load bound versions: %v", err)
	}
	if len(bound) != 1 || bound[0] != 2 {
		t.Fatalf("bound versions = %#v want [2]", bound)
	}
}

type releaseTupleFixture struct {
	now             time.Time
	userID          uuid.UUID
	incidentID      uuid.UUID
	otherIncidentID uuid.UUID
	compositionID   uuid.UUID
	digest          string
}

func seedReleaseTupleFixture(t testing.TB, ctx context.Context, db *pgtest.RollbackDB) releaseTupleFixture {
	t.Helper()
	now := time.Date(2026, 7, 7, 12, 0, 0, 0, time.UTC)
	userID := uuid.New()
	incidentID := uuid.New()
	otherIncidentID := uuid.New()
	compositionID := uuid.New()
	digest := strings.Repeat("a", 64)
	if _, err := db.Exec(ctx, `
INSERT INTO users (id, email, display_name, password_hash, mfa_required, is_active, is_deployment_admin, created_at, updated_at)
VALUES ($1, 'tuple-owner@example.test', 'Tuple Owner', 'hash', false, true, false, $2, $2)
`, userID, now); err != nil {
		t.Fatalf("seed user: %v", err)
	}
	for _, incident := range []struct {
		id  uuid.UUID
		key string
	}{
		{id: incidentID, key: "IR-TUPLE"},
		{id: otherIncidentID, key: "IR-TUPLE-OTHER"},
	} {
		if _, err := db.Exec(ctx, `
INSERT INTO incidents (id, incident_key, incident_key_canonical, title, status, created_by_user_id, updated_by_user_id, created_at, updated_at)
VALUES ($1, $2, $2, $3, 'active', $4, $4, $5, $5)
`, incident.id, incident.key, incident.key+" title", userID, now); err != nil {
			t.Fatalf("seed incident %s: %v", incident.key, err)
		}
	}
	if _, err := db.Exec(ctx, `
INSERT INTO report_compositions (
    composition_id, incident_id, created_by_user_id, client_txn_id, template_id, template_version,
    draft_version, deck_ops, diagram_decls, authored_texts, latest_composition_version, created_at, updated_at
)
VALUES ($1, $2, $3, 'txn-tuple', 'cartulary.report.default', '1', 1, '[]'::jsonb, '[]'::jsonb, '[]'::jsonb, 2, $4, $4)
`, compositionID, incidentID, userID, now); err != nil {
		t.Fatalf("seed composition: %v", err)
	}
	if _, err := db.Exec(ctx, `
INSERT INTO report_composition_versions (
    composition_id, composition_version, composition_sha256, canonical_composition, canonical_composition_bytes, created_by_user_id, created_at
)
VALUES ($1, 2, $2, '{"schema_id":"fixture"}'::jsonb, $3, $4, $5)
`, compositionID, digest, []byte(`{"schema_id":"fixture"}`), userID, now); err != nil {
		t.Fatalf("seed composition version: %v", err)
	}
	return releaseTupleFixture{
		now:             now,
		userID:          userID,
		incidentID:      incidentID,
		otherIncidentID: otherIncidentID,
		compositionID:   compositionID,
		digest:          digest,
	}
}

func seedReportingReleaseForTuple(t testing.TB, ctx context.Context, db *pgtest.RollbackDB, fixture releaseTupleFixture) uuid.UUID {
	t.Helper()
	jobID := uuid.New()
	snapshotID := uuid.New()
	releaseID := uuid.New()
	if _, err := db.Exec(ctx, `
INSERT INTO jobs (
    job_id, scope_kind, incident_id, status, cancelable, submitted_by_user_id, submitted_at, updated_at, progress_completed
)
VALUES ($1, 'incident', $2, 'queued', true, $3, $4, $4, 0)
`, jobID, fixture.incidentID, fixture.userID, fixture.now); err != nil {
		t.Fatalf("seed job: %v", err)
	}
	if _, err := db.Exec(ctx, `
INSERT INTO reporting_snapshots (
    snapshot_id, incident_id, created_by_user_id, client_txn_id, snapshot_at, source_change_set_high_watermark,
    source_boundary_json, derivation_version, export_model_sha256, export_model_json, create_job_id, created_at
)
VALUES ($1, $2, $3, 'txn-snapshot', $4, 'cartulary.source_boundary.v1:test',
        '{}'::jsonb, 'cartulary.snapshot_export_model.v3', $5, '{}'::jsonb, $6, $4)
`, snapshotID, fixture.incidentID, fixture.userID, fixture.now, strings.Repeat("c", 64), jobID); err != nil {
		t.Fatalf("seed snapshot: %v", err)
	}
	if _, err := db.Exec(ctx, `
INSERT INTO reporting_releases (
    release_id, incident_id, snapshot_id, created_by_user_id, client_txn_id, release_scope, release_state,
    snapshot_at, source_change_set_high_watermark, derivation_version, export_model_sha256,
    template_id, template_version, redaction_profile_id, redaction_profile_version, redaction_profile_sha256,
    output_kind, output_media_type, output_sha256, redaction_manifest_sha256, redaction_manifest_json,
    rendered_output, create_job_id, recipient_partition_refs, output_options, graph_projection_refs,
    render_admitted_at, created_at, updated_at
)
VALUES ($1, $2, $3, $4, 'txn-release', 'internal_draft', 'approved',
        $5, 'cartulary.source_boundary.v1:test', 'cartulary.snapshot_export_model.v3', $6,
        'cartulary.report.default', '1', 'cartulary.redaction.internal', '1', $7,
        'slidev', 'text/html', $8, $9, '{}'::jsonb,
        'rendered', $10, '[]'::jsonb, '{}'::jsonb, '[]'::jsonb,
        $5, $5, $5)
`, releaseID, fixture.incidentID, snapshotID, fixture.userID, fixture.now, strings.Repeat("c", 64), strings.Repeat("d", 64), strings.Repeat("e", 64), strings.Repeat("f", 64), jobID); err != nil {
		t.Fatalf("seed release: %v", err)
	}
	return releaseID
}
