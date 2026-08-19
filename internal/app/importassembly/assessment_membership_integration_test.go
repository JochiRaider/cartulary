package importassembly_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/app/assessmentassembly"
	"github.com/JochiRaider/cartulary/internal/modules/assessments"
	"github.com/JochiRaider/cartulary/internal/modules/imports/ownerfacade"
	"github.com/JochiRaider/cartulary/internal/modules/revisions"
	"github.com/JochiRaider/cartulary/internal/modules/workbook"
	"github.com/JochiRaider/cartulary/internal/platform/authn"
	"github.com/JochiRaider/cartulary/internal/testutil/appsupport"
	"github.com/JochiRaider/cartulary/internal/testutil/conflicttest"
)

func TestAssessmentAssessorMembershipContract_Integration(t *testing.T) {
	harness := newTasksDecisionsImportHarness(t, "assessment-membership")
	workbookStore := appsupport.NewWorkbookStore(
		harness.db,
		conflicttest.NewCodec("assessment-membership"),
	)
	hostID := uuid.New()
	seedAssessmentMembershipHost(t, harness, hostID)
	member := seedTasksDecisionsImportUser(t, harness, true, true)
	nonmember := seedTasksDecisionsImportUser(t, harness, true, false)
	inactiveMember := seedTasksDecisionsImportUser(t, harness, false, true)
	foreignMember := seedTasksDecisionsImportUser(t, harness, true, false)
	foreignIncident := appsupport.CreateIncidentInStore(
		t,
		harness.db,
		harness.actor,
		"txn-assessment-membership-foreign-"+uuid.NewString(),
		"IR-AS-MEMBER-"+uuid.NewString()[:8],
		"Assessment foreign membership",
	)
	seedTasksDecisionsMembership(
		t,
		harness.db,
		foreignIncident.ID,
		foreignMember.ID,
		harness.actor.ID,
	)

	t.Run("interactive_default_and_explicit_member_succeed", func(t *testing.T) {
		defaulted, err := createAssessmentMembershipWorkbookRow(
			workbookStore,
			harness.actor,
			harness.incidentID,
			hostID,
			nil,
			"txn-assessment-membership-default",
		)
		if err != nil {
			t.Fatalf("defaulted assessor create: %v", err)
		}
		requireAssessmentMembershipAssessor(t, harness, defaulted.RecordID, harness.actor.ID)

		explicit, err := createAssessmentMembershipWorkbookRow(
			workbookStore,
			harness.actor,
			harness.incidentID,
			hostID,
			&member.ID,
			"txn-assessment-membership-explicit",
		)
		if err != nil {
			t.Fatalf("explicit member assessor create: %v", err)
		}
		requireAssessmentMembershipAssessor(t, harness, explicit.RecordID, member.ID)
	})

	t.Run("interactive_invalid_assessors_fail_without_effects", func(t *testing.T) {
		for name, assessorID := range map[string]uuid.UUID{
			"unknown":             uuid.New(),
			"active_nonmember":    nonmember.ID,
			"inactive_member":     inactiveMember.ID,
			"foreign_only_member": foreignMember.ID,
		} {
			t.Run(name, func(t *testing.T) {
				before := assessmentMembershipEffectCounts(t, harness)
				_, err := createAssessmentMembershipWorkbookRow(
					workbookStore,
					harness.actor,
					harness.incidentID,
					hostID,
					&assessorID,
					"txn-assessment-membership-invalid-"+name,
				)
				requireAssessmentMembershipValidation(t, err)
				after := assessmentMembershipEffectCounts(t, harness)
				if after != before {
					t.Fatalf("invalid interactive assessor changed effects: before=%+v after=%+v", before, after)
				}
			})
		}
	})

	t.Run("import_default_and_invalid_assessors_match_interactive_policy", func(t *testing.T) {
		defaulted := runAssessmentImportCreate(t, harness, assessmentImportCase{
			subjectID: hostID, subjectType: "host", state: "confirmed",
			rationale: "Default import assessor.", commit: true,
		})
		if defaulted.err != nil {
			t.Fatalf("default import assessor: %v", defaulted.err)
		}
		requireAssessmentMembershipAssessor(t, harness, defaulted.response.RecordID, harness.actor.ID)

		for name, assessorID := range map[string]uuid.UUID{
			"unknown":             uuid.New(),
			"active_nonmember":    nonmember.ID,
			"inactive_member":     inactiveMember.ID,
			"foreign_only_member": foreignMember.ID,
		} {
			t.Run(name, func(t *testing.T) {
				before := assessmentMembershipEffectCounts(t, harness)
				result := runAssessmentImportCreate(t, harness, assessmentImportCase{
					subjectID: hostID, subjectType: "host", state: "confirmed",
					rationale: "Invalid import assessor.", assessor: &assessorID,
				})
				requireAssessmentMembershipValidation(t, result.err)
				after := assessmentMembershipEffectCounts(t, harness)
				if after != before {
					t.Fatalf("invalid import assessor changed effects: before=%+v after=%+v", before, after)
				}
			})
		}
	})

	t.Run("omitted_actor_is_materialized_then_validated", func(t *testing.T) {
		departedActor := seedTasksDecisionsImportUser(t, harness, true, true)
		if _, err := harness.db.Exec(
			context.Background(),
			`DELETE FROM incident_memberships WHERE incident_id = $1 AND user_id = $2`,
			harness.incidentID,
			departedActor.ID,
		); err != nil {
			t.Fatalf("remove default assessor membership: %v", err)
		}
		before := assessmentMembershipEffectCounts(t, harness)
		_, err := createAssessmentMembershipWorkbookRow(
			workbookStore,
			departedActor,
			harness.incidentID,
			hostID,
			nil,
			"txn-assessment-membership-departed-default",
		)
		requireAssessmentMembershipValidation(t, err)
		result := runAssessmentImportCreateAsActor(
			t,
			harness,
			hostID,
			departedActor.ID,
		)
		requireAssessmentMembershipValidation(t, result.err)
		after := assessmentMembershipEffectCounts(t, harness)
		if after != before {
			t.Fatalf("invalid omitted assessor changed effects: before=%+v after=%+v", before, after)
		}
	})

	t.Run("exact_replay_preserves_history_after_assessor_departure", func(t *testing.T) {
		departing := seedTasksDecisionsImportUser(t, harness, true, true)
		const clientTxnID = "txn-assessment-membership-replay"
		first, err := createAssessmentMembershipWorkbookRow(
			workbookStore,
			harness.actor,
			harness.incidentID,
			hostID,
			&departing.ID,
			clientTxnID,
		)
		if err != nil {
			t.Fatalf("create assessment before departure: %v", err)
		}
		if _, err := harness.db.Exec(
			context.Background(),
			`DELETE FROM incident_memberships WHERE incident_id = $1 AND user_id = $2`,
			harness.incidentID,
			departing.ID,
		); err != nil {
			t.Fatalf("remove historical assessor membership: %v", err)
		}
		replay, err := createAssessmentMembershipWorkbookRow(
			workbookStore,
			harness.actor,
			harness.incidentID,
			hostID,
			&departing.ID,
			clientTxnID,
		)
		if err != nil {
			t.Fatalf("replay after assessor departure: %v", err)
		}
		if !replay.Replayed || replay.RecordID != first.RecordID || replay.ChangeSetID != first.ChangeSetID {
			t.Fatalf("replay identity drifted: first=%#v replay=%#v", first, replay)
		}
		requireAssessmentMembershipAssessor(t, harness, first.RecordID, departing.ID)
	})

	t.Run("validator_locks_membership_before_user_state", func(t *testing.T) {
		lockedMember := seedTasksDecisionsImportUser(t, harness, true, true)
		validator := assessmentassembly.NewAssessorValidator(harness.db)
		assertAssessmentValidationBlocksConcurrentMutation(
			t,
			harness,
			validator,
			lockedMember.ID,
			`DELETE FROM incident_memberships WHERE incident_id = $1 AND user_id = $2`,
			harness.incidentID,
			lockedMember.ID,
		)
		seedTasksDecisionsMembership(t, harness.db, harness.incidentID, lockedMember.ID, harness.actor.ID)
		assertAssessmentValidationBlocksConcurrentMutation(
			t,
			harness,
			validator,
			lockedMember.ID,
			`UPDATE users SET is_active = false WHERE id = $1`,
			lockedMember.ID,
		)
	})
}

func seedAssessmentMembershipHost(t testing.TB, harness tasksDecisionsImportHarness, hostID uuid.UUID) {
	t.Helper()
	if _, err := harness.db.Exec(context.Background(), `
INSERT INTO records (
    record_id, incident_id, record_type, created_by_user_id, updated_by_user_id
) VALUES ($1, $2, 'host', $3, $3)
`, hostID, harness.incidentID, harness.actor.ID); err != nil {
		t.Fatalf("seed membership host envelope: %v", err)
	}
	if _, err := harness.db.Exec(context.Background(), `
INSERT INTO hosts (
    record_id, incident_id, display_name, hostname, host_state,
    created_by_user_id, updated_by_user_id
) VALUES ($1, $2, 'Assessment membership host', 'assessment-membership-host', 'canonical', $3, $3)
`, hostID, harness.incidentID, harness.actor.ID); err != nil {
		t.Fatalf("seed membership host: %v", err)
	}
}

func createAssessmentMembershipWorkbookRow(
	store *workbook.Store,
	actor authn.UserRecord,
	incidentID uuid.UUID,
	subjectID uuid.UUID,
	assessorID *uuid.UUID,
	clientTxnID string,
) (workbook.MutationResult, error) {
	subjectType := "host"
	state := "confirmed"
	rationale := "Assessment membership contract."
	request := workbook.CreateRequest{
		ViewSchemaID: assessments.AssessmentsViewSchemaID,
		ClientTxnID:  clientTxnID,
		Values: map[string]workbook.ValueChange{
			"assessment.subject_ref":      {Kind: "uuid", UUID: &subjectID},
			"assessment.subject_type":     {Kind: "text", Text: &subjectType},
			"assessment.assessment_state": {Kind: "text", Text: &state},
			"assessment.rationale":        {Kind: "text", Text: &rationale},
		},
	}
	if assessorID != nil {
		request.Values["assessment.assessor"] = workbook.ValueChange{Kind: "uuid", UUID: assessorID}
	}
	return store.CreateWorkbookRow(
		context.Background(),
		actor,
		incidentID,
		request,
		workbook.CreateRequestHash(request),
		"req-"+clientTxnID,
		time.Date(2026, time.August, 3, 12, 0, 0, 0, time.UTC),
	)
}

func runAssessmentImportCreateAsActor(
	t testing.TB,
	harness tasksDecisionsImportHarness,
	subjectID uuid.UUID,
	actorID uuid.UUID,
) assessmentImportResult {
	t.Helper()
	ctx := context.Background()
	tx, err := harness.db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		t.Fatalf("begin default-assessor import transaction: %v", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()
	now := time.Date(2026, time.August, 3, 13, 0, 0, 0, time.UTC)
	changeSetID, err := harness.appender.AppendChangeSetTx(ctx, tx, revisions.AppendChangeSetParams{
		IncidentID:  harness.incidentID,
		ActorUserID: actorID,
		Source:      "imports.unit.apply",
		CreatedAt:   now,
	})
	if err != nil {
		t.Fatalf("append default-assessor import change set: %v", err)
	}
	importUnitID := uuid.New()
	request := assessmentImportCase{
		subjectID: subjectID, subjectType: "host", state: "confirmed",
		rationale: "Default import assessor without membership.",
	}
	response, createErr := assessmentImportFacade(t, harness).CreateImportRowTx(
		ctx,
		tx,
		ownerfacade.ImportOwnerCreateCommand{
			Request: ownerfacade.ImportOwnerCreateRequest{
				IncidentID: harness.incidentID, ActorUserID: actorID,
				TargetViewSchemaID: assessments.AssessmentsViewSchemaID,
				ImportSessionID:    uuid.New(), ImportUnitID: importUnitID,
				MappingFingerprint: "assessment-membership-import", SourceFileKind: "csv",
				ParserProfileID: "synthetic", ParserVersion: "1", LocatorKind: "row",
				Locator: "1", ClientTxnID: "txn-" + uuid.NewString(),
				FieldValues: assessmentImportFields(request),
			},
			ChangeSetID: changeSetID, SequenceNo: 1, Now: now,
		},
	)
	return assessmentImportResult{response: response, importUnitID: importUnitID, err: createErr}
}

func requireAssessmentMembershipValidation(t testing.TB, err error) {
	t.Helper()
	if err == nil {
		t.Fatal("invalid assessment assessor unexpectedly succeeded")
	}
	var validation *workbook.MutationValidationError
	var assessmentValidation *assessments.CreateValidationError
	if errors.As(err, &validation) {
		if validation.Field != "assessment.assessor" || validation.ReasonCode != "invalid_value" {
			t.Fatalf("workbook validation = %#v", validation)
		}
		return
	}
	if errors.As(err, &assessmentValidation) {
		if assessmentValidation.Field != "assessment.assessor" || assessmentValidation.ReasonCode != "invalid_value" {
			t.Fatalf("assessment validation = %#v", assessmentValidation)
		}
		return
	}
	t.Fatalf("assessment assessor error = %T %v", err, err)
}

func requireAssessmentMembershipAssessor(
	t testing.TB,
	harness tasksDecisionsImportHarness,
	recordID uuid.UUID,
	want uuid.UUID,
) {
	t.Helper()
	var got uuid.UUID
	if err := harness.db.QueryRow(
		context.Background(),
		`SELECT assessor_user_id FROM assessments WHERE record_id = $1`,
		recordID,
	).Scan(&got); err != nil {
		t.Fatalf("query assessment assessor: %v", err)
	}
	if got != want {
		t.Fatalf("assessment assessor = %s, want %s", got, want)
	}
}

type assessmentMembershipCounts struct {
	records, assessments, projections, links, changeSets, mutations, revisions, intents, idempotency int
}

func assessmentMembershipEffectCounts(
	t testing.TB,
	harness tasksDecisionsImportHarness,
) assessmentMembershipCounts {
	t.Helper()
	queries := []string{
		`SELECT count(*) FROM records WHERE incident_id = $1 AND record_type = 'assessment'`,
		`SELECT count(*) FROM assessments WHERE incident_id = $1`,
		`SELECT count(*) FROM assessment_grid_projection WHERE incident_id = $1`,
		`SELECT count(*) FROM record_links WHERE incident_id = $1`,
		`SELECT count(*) FROM change_sets WHERE incident_id = $1`,
		`SELECT count(*) FROM change_set_mutations mutation JOIN change_sets change_set USING (change_set_id) WHERE change_set.incident_id = $1`,
		`SELECT count(*) FROM record_revisions revision JOIN records record USING (record_id) WHERE record.incident_id = $1`,
		`SELECT count(*) FROM collaboration_event_intents WHERE incident_id = $1`,
		`SELECT count(*) FROM route_idempotency WHERE route_key = 'assessments.rows.create' AND scope_key = $1`,
	}
	counts := assessmentMembershipCounts{}
	values := []*int{
		&counts.records, &counts.assessments, &counts.projections, &counts.links,
		&counts.changeSets, &counts.mutations, &counts.revisions, &counts.intents,
		&counts.idempotency,
	}
	arguments := []any{harness.incidentID}
	for index, query := range queries {
		if index == len(queries)-1 {
			arguments = []any{harness.incidentID.String() + ":" + assessments.AssessmentsViewSchemaID}
		}
		if err := harness.db.QueryRow(context.Background(), query, arguments...).Scan(values[index]); err != nil {
			t.Fatalf("query assessment membership effect %d: %v", index, err)
		}
	}
	return counts
}

func assertAssessmentValidationBlocksConcurrentMutation(
	t testing.TB,
	harness tasksDecisionsImportHarness,
	validator assessments.AssessorValidator,
	userID uuid.UUID,
	mutationSQL string,
	mutationArgs ...any,
) {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	validationTx, err := harness.db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		t.Fatalf("begin assessor validation transaction: %v", err)
	}
	valid, err := validator.ValidateAssessmentAssessorTx(
		ctx,
		validationTx,
		harness.incidentID,
		userID,
	)
	if err != nil || !valid {
		_ = validationTx.Rollback(ctx)
		t.Fatalf("lock valid assessment assessor: valid=%v err=%v", valid, err)
	}
	done := make(chan error, 1)
	go func() {
		_, mutationErr := harness.db.Exec(ctx, mutationSQL, mutationArgs...)
		done <- mutationErr
	}()
	select {
	case mutationErr := <-done:
		_ = validationTx.Rollback(ctx)
		t.Fatalf("concurrent membership/user mutation bypassed validation lock: %v", mutationErr)
	case <-time.After(100 * time.Millisecond):
	}
	if err := validationTx.Rollback(ctx); err != nil && !errors.Is(err, pgx.ErrTxClosed) {
		t.Fatalf("release assessor validation locks: %v", err)
	}
	select {
	case mutationErr := <-done:
		if mutationErr != nil {
			t.Fatalf("concurrent membership/user mutation after release: %v", mutationErr)
		}
	case <-ctx.Done():
		t.Fatalf("concurrent membership/user mutation remained blocked: %v", ctx.Err())
	}
}
