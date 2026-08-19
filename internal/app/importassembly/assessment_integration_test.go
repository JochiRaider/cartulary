package importassembly_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/modules/assessments"
	entitytest "github.com/JochiRaider/cartulary/internal/modules/entities/testsupport"
	"github.com/JochiRaider/cartulary/internal/modules/imports/ownerfacade"
	"github.com/JochiRaider/cartulary/internal/modules/revisions"
	"github.com/JochiRaider/cartulary/internal/testutil/appsupport"
)

func TestAssessmentImportCreateFacadeContract_Integration(t *testing.T) {
	harness := newTasksDecisionsImportHarness(t, "assessment-contract")
	hostID := uuid.New()
	identityID := uuid.New()
	entitytest.SeedHostRecord(
		t, harness.db, harness.incidentID, harness.actor.ID, hostID,
		"Imported host", "imported-host", "imported-host.example.test", "",
	)
	entitytest.SeedIdentityRecord(
		t, harness.db, harness.incidentID, harness.actor.ID, identityID,
		"Imported identity", "imported@example.test", "imported@example.test", "imported",
	)

	t.Run("explicit_values_publish_source_projection_and_revision", func(t *testing.T) {
		score := int64(70)
		assessedAt := time.Date(2026, time.August, 1, 15, 45, 0, 0, time.UTC)
		result := runAssessmentImportCreate(t, harness, assessmentImportCase{
			subjectID: hostID, subjectType: "host", state: "confirmed",
			rationale: "Imported assessment rationale.", score: &score,
			assessor: &harness.actor.ID, assessedAt: &assessedAt, commit: true,
		})
		if result.err != nil {
			t.Fatalf("create imported assessment: %v", result.err)
		}
		if result.response.RecordID == uuid.Nil || result.response.RowVersion != 1 ||
			result.response.CreatedOrReused != "created" || result.response.OwnerResultCode != "created" {
			t.Fatalf("unexpected owner response: %#v", result.response)
		}
		if got := assessmentImportCellValue(t, result.response.RowRefresh, "assessment.confidence_band"); got != "high" {
			t.Fatalf("confidence band = %#v, want high", got)
		}
		requireAssessmentImportStored(t, harness, result.response.RecordID, assessmentStoredWant{
			subjectID: hostID, subjectType: "host", state: "confirmed", score: &score,
			rationale: "Imported assessment rationale.", assessor: harness.actor.ID,
			assessedAt: assessedAt,
		})
	})

	t.Run("omitted_values_use_owner_defaults", func(t *testing.T) {
		now := time.Date(2026, time.August, 2, 9, 30, 0, 0, time.UTC)
		result := runAssessmentImportCreate(t, harness, assessmentImportCase{
			subjectID: identityID, subjectType: "identity", state: "unknown",
			rationale: "Defaulted imported assessment.", now: now, commit: true,
		})
		if result.err != nil {
			t.Fatalf("create defaulted imported assessment: %v", result.err)
		}
		if got := assessmentImportCellValue(t, result.response.RowRefresh, "assessment.confidence_band"); got != "unset" {
			t.Fatalf("default confidence band = %#v, want unset", got)
		}
		requireAssessmentImportStored(t, harness, result.response.RecordID, assessmentStoredWant{
			subjectID: identityID, subjectType: "identity", state: "unknown",
			rationale: "Defaulted imported assessment.", assessor: harness.actor.ID,
			assessedAt: now,
		})
	})

	t.Run("invalid_shape_and_subjects_have_no_effects", func(t *testing.T) {
		foreignIncident := appsupport.CreateIncidentInStore(
			t, harness.db, harness.actor, "txn-assessment-import-foreign-"+uuid.NewString(),
			"IR-AS-IMPORT-"+uuid.NewString()[:8], "Assessment import foreign incident",
		)
		foreignHostID := uuid.New()
		entitytest.SeedHostRecord(
			t, harness.db, foreignIncident.ID, harness.actor.ID, foreignHostID,
			"Foreign host", "foreign-host", "", "",
		)
		before := assessmentImportEffectCounts(t, harness, harness.incidentID)
		badScore := int64(101)
		cases := []assessmentImportCase{
			{subjectID: uuid.Nil, subjectType: "host", state: "confirmed", rationale: "Missing subject."},
			{subjectID: hostID, subjectType: "party", state: "confirmed", rationale: "Wrong type."},
			{subjectID: hostID, subjectType: "identity", state: "confirmed", rationale: "Mismatched type."},
			{subjectID: hostID, subjectType: "host", state: "invalid", rationale: "Wrong state."},
			{subjectID: hostID, subjectType: "host", state: "confirmed", rationale: "", score: nil},
			{subjectID: hostID, subjectType: "host", state: "confirmed", rationale: "Wrong score.", score: &badScore},
			{subjectID: foreignHostID, subjectType: "host", state: "confirmed", rationale: "Foreign subject."},
		}
		for index, tc := range cases {
			result := runAssessmentImportCreate(t, harness, tc)
			var validation *assessments.CreateValidationError
			if !errors.As(result.err, &validation) {
				t.Fatalf("case %d error = %T %v, want assessment validation", index, result.err, result.err)
			}
		}
		after := assessmentImportEffectCounts(t, harness, harness.incidentID)
		if after != before {
			t.Fatalf("rejected imports changed effects: before=%+v after=%+v", before, after)
		}
	})

	t.Run("collection_fields_stay_at_the_imports_boundary", func(t *testing.T) {
		facade := assessmentImportFacade(t, harness)
		_, _, err := facade.NormalizeImportField(
			"assessment.support_refs", uuid.NewString(), "omit_field",
		)
		var ownerErr *ownerfacade.ImportOwnerCreateError
		if !errors.As(err, &ownerErr) || ownerErr.ReasonCode != "collection_owner_support_required" {
			t.Fatalf("support_refs normalization error = %T %v", err, err)
		}
	})

	t.Run("caller_rollback_is_atomic_and_finalization_is_not_owned", func(t *testing.T) {
		before := assessmentImportEffectCounts(t, harness, harness.incidentID)
		result := runAssessmentImportCreate(t, harness, assessmentImportCase{
			subjectID: hostID, subjectType: "host", state: "suspected",
			rationale: "Caller rollback.", commit: false,
		})
		if result.err != nil || result.response.RecordID == uuid.Nil {
			t.Fatalf("owner create before rollback: response=%#v err=%v", result.response, result.err)
		}
		after := assessmentImportEffectCounts(t, harness, harness.incidentID)
		if after != before {
			t.Fatalf("caller rollback changed effects: before=%+v after=%+v", before, after)
		}
		for _, table := range []string{"import_apply_journal", "import_unit_apply_outcomes"} {
			var count int
			if err := harness.db.QueryRow(
				context.Background(), "SELECT count(*) FROM "+table+" WHERE import_unit_id = $1", result.importUnitID,
			).Scan(&count); err != nil {
				t.Fatalf("query %s: %v", table, err)
			}
			if count != 0 {
				t.Fatalf("assessment owner wrote %d %s rows", count, table)
			}
		}
	})
}

type assessmentImportCase struct {
	subjectID   uuid.UUID
	subjectType string
	state       string
	rationale   string
	score       *int64
	assessor    *uuid.UUID
	assessedAt  *time.Time
	now         time.Time
	commit      bool
}

type assessmentImportResult struct {
	response     ownerfacade.ImportOwnerCreateResponse
	importUnitID uuid.UUID
	err          error
}

func runAssessmentImportCreate(
	t testing.TB,
	harness tasksDecisionsImportHarness,
	tc assessmentImportCase,
) assessmentImportResult {
	t.Helper()
	ctx := context.Background()
	tx, err := harness.db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		t.Fatalf("begin assessment import transaction: %v", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()
	now := tc.now
	if now.IsZero() {
		now = time.Date(2026, time.August, 1, 14, 30, 0, 0, time.UTC)
	}
	changeSetID, err := harness.appender.AppendChangeSetTx(ctx, tx, revisions.AppendChangeSetParams{
		IncidentID: harness.incidentID, ActorUserID: harness.actor.ID,
		Source: "imports.unit.apply", CreatedAt: now,
	})
	if err != nil {
		t.Fatalf("append assessment import change set: %v", err)
	}
	importUnitID := uuid.New()
	response, createErr := assessmentImportFacade(t, harness).CreateImportRowTx(
		ctx,
		tx,
		ownerfacade.ImportOwnerCreateCommand{
			Request: ownerfacade.ImportOwnerCreateRequest{
				IncidentID: harness.incidentID, ActorUserID: harness.actor.ID,
				TargetViewSchemaID: assessments.AssessmentsViewSchemaID,
				ImportSessionID:    uuid.New(), ImportUnitID: importUnitID,
				MappingFingerprint: "assessment-import-mapping", SourceFileKind: "csv",
				ParserProfileID: "synthetic", ParserVersion: "1", LocatorKind: "row",
				Locator: "1", ClientTxnID: "txn-" + uuid.NewString(),
				FieldValues: assessmentImportFields(tc),
			},
			ChangeSetID: changeSetID, SequenceNo: 1, Now: now,
		},
	)
	if createErr == nil && tc.commit {
		if err := tx.Commit(ctx); err != nil {
			t.Fatalf("commit assessment import transaction: %v", err)
		}
	}
	return assessmentImportResult{response: response, importUnitID: importUnitID, err: createErr}
}

func assessmentImportFacade(t testing.TB, harness tasksDecisionsImportHarness) ownerfacade.ImportOwnerCreateFacade {
	t.Helper()
	if facade, ok := harness.registry.Resolve(assessments.AssessmentsViewSchemaID, "assessments.import_create"); ok {
		return facade
	}
	t.Fatal("application import registry missing assessments facade")
	return nil
}

func assessmentImportCellValue(t testing.TB, row map[string]any, fieldKey string) any {
	t.Helper()
	cells, ok := row["cells"].(map[string]any)
	if !ok {
		t.Fatalf("assessment import row has no cells: %#v", row)
	}
	cell, ok := cells[fieldKey].(map[string]any)
	if !ok {
		t.Fatalf("assessment import row has no %s cell: %#v", fieldKey, row)
	}
	return cell["value"]
}

func assessmentImportFields(tc assessmentImportCase) []ownerfacade.ImportFieldValue {
	text := func(key string, value string) ownerfacade.ImportFieldValue {
		return ownerfacade.ImportFieldValue{FieldKey: key, NormalizedValue: ownerfacade.ImportScalarValue{Kind: "text", Text: &value}}
	}
	fields := []ownerfacade.ImportFieldValue{
		{FieldKey: "assessment.subject_ref", NormalizedValue: ownerfacade.ImportScalarValue{Kind: "uuid", UUID: &tc.subjectID}},
		text("assessment.subject_type", tc.subjectType),
		text("assessment.assessment_state", tc.state),
		text("assessment.rationale", tc.rationale),
	}
	if tc.score != nil {
		fields = append(fields, ownerfacade.ImportFieldValue{
			FieldKey: "assessment.confidence_score", NormalizedValue: ownerfacade.ImportScalarValue{Kind: "number", Number: tc.score},
		})
	}
	if tc.assessor != nil {
		fields = append(fields, ownerfacade.ImportFieldValue{
			FieldKey: "assessment.assessor", NormalizedValue: ownerfacade.ImportScalarValue{Kind: "uuid", UUID: tc.assessor},
		})
	}
	if tc.assessedAt != nil {
		fields = append(fields, ownerfacade.ImportFieldValue{
			FieldKey: "assessment.assessed_at", NormalizedValue: ownerfacade.ImportScalarValue{Kind: "timestamp", Timestamp: tc.assessedAt},
		})
	}
	return fields
}

type assessmentStoredWant struct {
	subjectID   uuid.UUID
	subjectType string
	state       string
	score       *int64
	rationale   string
	assessor    uuid.UUID
	assessedAt  time.Time
}

func requireAssessmentImportStored(
	t testing.TB,
	harness tasksDecisionsImportHarness,
	recordID uuid.UUID,
	want assessmentStoredWant,
) {
	t.Helper()
	var (
		subjectID                     uuid.UUID
		subjectType, state, rationale string
		score                         *int64
		assessor                      uuid.UUID
		assessedAt                    time.Time
	)
	if err := harness.db.QueryRow(context.Background(), `
SELECT subject_record_id, subject_type, assessment_state, confidence_score,
       rationale, assessor_user_id, assessed_at
  FROM assessments
 WHERE record_id = $1
`, recordID).Scan(&subjectID, &subjectType, &state, &score, &rationale, &assessor, &assessedAt); err != nil {
		t.Fatalf("query imported assessment source: %v", err)
	}
	if subjectID != want.subjectID || subjectType != want.subjectType || state != want.state ||
		rationale != want.rationale || assessor != want.assessor || !assessedAt.Equal(want.assessedAt) ||
		!equalOptionalInt64(score, want.score) {
		t.Fatalf("stored assessment mismatch: subject=%s/%s state=%s score=%v rationale=%q assessor=%s assessed_at=%s want=%+v",
			subjectID, subjectType, state, score, rationale, assessor, assessedAt, want)
	}
	for label, query := range map[string]string{
		"projection": `SELECT count(*) FROM assessment_grid_projection WHERE record_id = $1`,
		"mutation":   `SELECT count(*) FROM change_set_mutations WHERE target_id = $1`,
		"revision":   `SELECT count(*) FROM record_revisions WHERE record_id = $1`,
	} {
		var count int
		if err := harness.db.QueryRow(context.Background(), query, recordID).Scan(&count); err != nil {
			t.Fatalf("query imported assessment %s: %v", label, err)
		}
		if count != 1 {
			t.Fatalf("imported assessment %s rows = %d, want 1", label, count)
		}
	}
}

func equalOptionalInt64(left *int64, right *int64) bool {
	return left == nil && right == nil || left != nil && right != nil && *left == *right
}

type assessmentImportCounts struct {
	records, assessments, projections, changeSets, mutations, revisions int
}

func assessmentImportEffectCounts(
	t testing.TB,
	harness tasksDecisionsImportHarness,
	incidentID uuid.UUID,
) assessmentImportCounts {
	t.Helper()
	queries := []string{
		`SELECT count(*) FROM records WHERE incident_id = $1`,
		`SELECT count(*) FROM assessments WHERE incident_id = $1`,
		`SELECT count(*) FROM assessment_grid_projection WHERE incident_id = $1`,
		`SELECT count(*) FROM change_sets WHERE incident_id = $1`,
		`SELECT count(*) FROM change_set_mutations mutation JOIN change_sets change_set USING (change_set_id) WHERE change_set.incident_id = $1`,
		`SELECT count(*) FROM record_revisions revision JOIN records record USING (record_id) WHERE record.incident_id = $1`,
	}
	counts := assessmentImportCounts{}
	values := []*int{&counts.records, &counts.assessments, &counts.projections, &counts.changeSets, &counts.mutations, &counts.revisions}
	for index, query := range queries {
		if err := harness.db.QueryRow(context.Background(), query, incidentID).Scan(values[index]); err != nil {
			t.Fatalf("query assessment import effect %d: %v", index, err)
		}
	}
	return counts
}
