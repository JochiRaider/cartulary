package importassembly_test

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/app/importassembly"
	"github.com/JochiRaider/cartulary/internal/app/indicatorassembly"
	"github.com/JochiRaider/cartulary/internal/app/projectionassembly"
	"github.com/JochiRaider/cartulary/internal/app/timelineassembly"
	authstoretest "github.com/JochiRaider/cartulary/internal/modules/auth/testsupport/storetest"
	"github.com/JochiRaider/cartulary/internal/modules/evidence"
	"github.com/JochiRaider/cartulary/internal/modules/imports/ownerfacade"
	"github.com/JochiRaider/cartulary/internal/modules/indicators"
	"github.com/JochiRaider/cartulary/internal/modules/revisions"
	"github.com/JochiRaider/cartulary/internal/modules/tasksdecisions"
	"github.com/JochiRaider/cartulary/internal/platform/authn"
	"github.com/JochiRaider/cartulary/internal/platform/postgres"
	"github.com/JochiRaider/cartulary/internal/testutil/appsupport"
	"github.com/JochiRaider/cartulary/internal/testutil/conflicttest"
	"github.com/JochiRaider/cartulary/internal/testutil/revisionsupport"
)

type tasksDecisionsImportHarness struct {
	db         postgres.DB
	actor      authn.UserRecord
	incidentID uuid.UUID
	appender   *revisions.Appender
	registry   *ownerfacade.ImportOwnerCreateRegistry
}

func TestTasksDecisionsImportOwnerReferenceMatrix(t *testing.T) {
	harness := newTasksDecisionsImportHarness(t, "owner-matrix")
	currentMember := seedTasksDecisionsImportUser(t, harness, true, true)
	nonmember := seedTasksDecisionsImportUser(t, harness, true, false)
	inactiveMember := seedTasksDecisionsImportUser(t, harness, false, true)
	foreignMember := seedTasksDecisionsImportUser(t, harness, true, false)
	foreignIncident := appsupport.CreateIncidentInStore(
		t,
		harness.db,
		harness.actor,
		"txn-tasks-decisions-import-foreign-"+uuid.NewString(),
		"IR-TD-IMPORT-FOREIGN-"+uuid.NewString()[:8],
		"Tasks decisions import foreign membership",
	)
	seedTasksDecisionsMembership(t, harness.db, foreignIncident.ID, foreignMember.ID, harness.actor.ID)

	for _, viewSchemaID := range []string{
		tasksdecisions.TaskRequestsViewSchemaID,
		tasksdecisions.DecisionsViewSchemaID,
	} {
		t.Run(viewSchemaID+"_current_member", func(t *testing.T) {
			response, ownerID, err := runTasksDecisionsImportCreate(t, harness, viewSchemaID, &currentMember.ID, false, nil)
			if err != nil {
				t.Fatalf("current member import failed: %v", err)
			}
			if response.RecordID == uuid.Nil || ownerID != currentMember.ID {
				t.Fatalf("current member import response=%#v owner=%s", response, ownerID)
			}
		})

		t.Run(viewSchemaID+"_omitted_defaults_actor", func(t *testing.T) {
			response, ownerID, err := runTasksDecisionsImportCreate(t, harness, viewSchemaID, nil, false, nil)
			if err != nil {
				t.Fatalf("omitted owner import failed: %v", err)
			}
			if response.RecordID == uuid.Nil || ownerID != harness.actor.ID {
				t.Fatalf("omitted owner response=%#v owner=%s want actor=%s", response, ownerID, harness.actor.ID)
			}
		})

		for name, userID := range map[string]uuid.UUID{
			"active_nonmember":    nonmember.ID,
			"foreign_only_member": foreignMember.ID,
			"inactive_member":     inactiveMember.ID,
		} {
			t.Run(viewSchemaID+"_"+name, func(t *testing.T) {
				_, _, err := runTasksDecisionsImportCreate(t, harness, viewSchemaID, &userID, false, nil)
				requireTasksDecisionsOwnerValidation(t, err)
			})
		}

		t.Run(viewSchemaID+"_explicit_null", func(t *testing.T) {
			_, _, err := runTasksDecisionsImportCreate(t, harness, viewSchemaID, nil, true, nil)
			requireTasksDecisionsOwnerValidation(t, err)
		})
	}
}

func TestTasksDecisionsImportMembershipAndAuthorizationRaces(t *testing.T) {
	harness := newTasksDecisionsImportHarness(t, "membership-races")
	member := seedTasksDecisionsImportUser(t, harness, true, true)
	for _, viewSchemaID := range []string{
		tasksdecisions.TaskRequestsViewSchemaID,
		tasksdecisions.DecisionsViewSchemaID,
	} {
		t.Run(viewSchemaID, func(t *testing.T) {
			_, _, err := runTasksDecisionsImportCreate(
				t,
				harness,
				viewSchemaID,
				&member.ID,
				false,
				func(ctx context.Context, tx pgx.Tx) {
					if _, deleteErr := tx.Exec(ctx, `DELETE FROM incident_memberships WHERE incident_id = $1 AND user_id = $2`, harness.incidentID, member.ID); deleteErr != nil {
						t.Fatalf("remove membership before owner validation: %v", deleteErr)
					}
				},
			)
			requireTasksDecisionsOwnerValidation(t, err)
		})

		t.Run(viewSchemaID+"_actor_authorization_removed", func(t *testing.T) {
			_, _, err := runTasksDecisionsImportCreate(
				t,
				harness,
				viewSchemaID,
				nil,
				false,
				func(ctx context.Context, tx pgx.Tx) {
					if _, deleteErr := tx.Exec(ctx, `DELETE FROM incident_memberships WHERE incident_id = $1 AND user_id = $2`, harness.incidentID, harness.actor.ID); deleteErr != nil {
						t.Fatalf("remove actor authorization before owner validation: %v", deleteErr)
					}
				},
			)
			requireTasksDecisionsOwnerValidation(t, err)
		})
	}
}

func TestTasksDecisionsImportReplayAndAtomicity(t *testing.T) {
	harness := newTasksDecisionsImportHarness(t, "atomicity")
	nonmember := seedTasksDecisionsImportUser(t, harness, true, false)
	before := tasksDecisionsImportEffectCounts(t, harness.db, harness.incidentID)
	for _, viewSchemaID := range []string{
		tasksdecisions.TaskRequestsViewSchemaID,
		tasksdecisions.DecisionsViewSchemaID,
	} {
		_, _, err := runTasksDecisionsImportCreate(t, harness, viewSchemaID, &nonmember.ID, false, nil)
		requireTasksDecisionsOwnerValidation(t, err)
	}
	after := tasksDecisionsImportEffectCounts(t, harness.db, harness.incidentID)
	if after != before {
		t.Fatalf("rejected imports changed durable effects: before=%+v after=%+v", before, after)
	}

	response, _, err := runTasksDecisionsImportCreate(
		t,
		harness,
		tasksdecisions.TaskRequestsViewSchemaID,
		nil,
		false,
		nil,
	)
	if err != nil {
		t.Fatalf("valid caller-owned import transaction failed: %v", err)
	}
	if response.RecordID == uuid.Nil {
		t.Fatal("valid caller-owned import returned no record")
	}
	if durable := tasksDecisionsImportEffectCounts(t, harness.db, harness.incidentID); durable != before {
		t.Fatalf("rolled-back import changed durable effects: before=%+v after=%+v", before, durable)
	}
}

func newTasksDecisionsImportHarness(t testing.TB, suffix string) tasksDecisionsImportHarness {
	t.Helper()
	storeHarness := appsupport.StartStore(t, "tasks-decisions-import-"+suffix)
	actor := authstoretest.SeedLocalUserRecord(
		t,
		storeHarness.DB,
		"tasks-decisions-import-"+uuid.NewString()+"@example.test",
		"Tasks Decisions Import",
		"TasksDecisionsImport1!",
		false,
		false,
		true,
	)
	incident := appsupport.CreateIncidentInStore(
		t,
		storeHarness.DB,
		actor,
		"txn-tasks-decisions-import-"+uuid.NewString(),
		"IR-TD-IMPORT-"+uuid.NewString()[:8],
		"Tasks decisions import "+suffix,
	)
	revisionComposition := revisionsupport.MustComposition(t)
	appender := revisionComposition.Runtime.Appender()
	intents := revisionComposition.Intents
	projections, err := projectionassembly.Build(storeHarness.DB)
	if err != nil {
		t.Fatalf("compose Projections: %v", err)
	}
	timelineBundle, err := timelineassembly.NewBundle(timelineassembly.Dependencies{
		Postgres:            storeHarness.DB,
		ConflictTokens:      conflicttest.NewCodec("tasks-decisions-import"),
		Revisions:           appender,
		Collaboration:       intents,
		EvidenceAttachments: evidence.NewTimelineAttachmentContribution(storeHarness.DB),
		TimelineProjection:  projections.TimelinePorts().Writer,
		EntityProjection:    projections.EntityPorts().Writer,
		AssessmentRows:      projections.AssessmentPorts().Rows,
	})
	if err != nil {
		t.Fatalf("compose Timeline: %v", err)
	}
	indicatorOwner, err := indicators.NewStore(indicators.StoreDependencies{
		Postgres:    storeHarness.DB,
		Revisions:   appender,
		Projections: projections.IndicatorPorts().Rows,
		SourceText:  indicatorassembly.NewSourceTextPort(projections.SourceTextRows()),
	})
	if err != nil {
		t.Fatalf("compose Indicators owner: %v", err)
	}
	registry, err := importassembly.NewOwnerCreateRegistry(importassembly.OwnerRegistryDependencies{
		Postgres:                storeHarness.DB,
		RevisionAppender:        appender,
		Intents:                 intents,
		Timeline:                timelineBundle.Facade,
		EntityProjections:       projections.EntityPorts().Writer,
		AssessmentProjections:   projections.AssessmentPorts().Rows,
		ArtifactProjections:     projections.ArtifactPorts().Rows,
		EvidenceProjections:     projections.EvidencePorts().Rows,
		PartyProjections:        projections.PartyPorts().Rows,
		TaskDecisionProjections: projections.TaskDecisionPorts().Rows,
		Indicators:              indicatorOwner,
	})
	if err != nil {
		t.Fatalf("compose application import owner registry: %v", err)
	}
	return tasksDecisionsImportHarness{
		db: storeHarness.DB, actor: actor, incidentID: incident.ID,
		appender: appender, registry: registry,
	}
}

func seedTasksDecisionsImportUser(t testing.TB, harness tasksDecisionsImportHarness, active bool, member bool) authn.UserRecord {
	t.Helper()
	user := authstoretest.SeedLocalUserRecord(
		t,
		harness.db,
		"tasks-decisions-import-ref-"+uuid.NewString()+"@example.test",
		"Tasks Decisions Import Reference",
		"TasksDecisionsImportReference1!",
		false,
		false,
		true,
	)
	if member {
		seedTasksDecisionsMembership(t, harness.db, harness.incidentID, user.ID, harness.actor.ID)
	}
	if !active {
		if _, err := harness.db.Exec(context.Background(), `UPDATE users SET is_active = false WHERE id = $1`, user.ID); err != nil {
			t.Fatalf("deactivate imported owner: %v", err)
		}
	}
	return user
}

func seedTasksDecisionsMembership(t testing.TB, db postgres.DB, incidentID uuid.UUID, userID uuid.UUID, actorID uuid.UUID) {
	t.Helper()
	if _, err := db.Exec(context.Background(), `
INSERT INTO incident_memberships (
    incident_id, user_id, role, joined_at, added_by_user_id,
    updated_at, updated_by_user_id, membership_version
) VALUES ($1, $2, 'editor', $3, $4, $3, $4, 1)
`, incidentID, userID, time.Date(2026, time.July, 31, 14, 0, 0, 0, time.UTC), actorID); err != nil {
		t.Fatalf("seed imported-owner membership: %v", err)
	}
}

func runTasksDecisionsImportCreate(
	t testing.TB,
	harness tasksDecisionsImportHarness,
	viewSchemaID string,
	ownerID *uuid.UUID,
	explicitNull bool,
	beforeCreate func(context.Context, pgx.Tx),
) (ownerfacade.ImportOwnerCreateResponse, uuid.UUID, error) {
	t.Helper()
	ctx := context.Background()
	facade := tasksDecisionsImportFacade(t, harness, viewSchemaID)
	tx, err := harness.db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		t.Fatalf("begin task/decision import transaction: %v", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()
	if beforeCreate != nil {
		beforeCreate(ctx, tx)
	}
	now := time.Date(2026, time.July, 31, 14, 30, 0, 0, time.UTC)
	changeSetID, err := harness.appender.AppendChangeSetTx(ctx, tx, revisions.AppendChangeSetParams{
		IncidentID: harness.incidentID, ActorUserID: harness.actor.ID,
		Source: "imports.unit.apply", CreatedAt: now,
	})
	if err != nil {
		t.Fatalf("append task/decision import change set: %v", err)
	}
	fields := tasksDecisionsImportFields(viewSchemaID, ownerID, explicitNull)
	response, createErr := facade.CreateImportRowTx(ctx, tx, ownerfacade.ImportOwnerCreateCommand{
		Request: ownerfacade.ImportOwnerCreateRequest{
			IncidentID: harness.incidentID, ActorUserID: harness.actor.ID,
			TargetViewSchemaID: viewSchemaID,
			ImportSessionID:    uuid.New(), ImportUnitID: uuid.New(),
			MappingFingerprint: "tasks-decisions-import-mapping",
			SourceFileKind:     "csv", ParserProfileID: "synthetic", ParserVersion: "1",
			LocatorKind: "row", Locator: "1", ClientTxnID: "txn-" + uuid.NewString(),
			FieldValues: fields,
		},
		ChangeSetID: changeSetID, SequenceNo: 1, Now: now,
	})
	if createErr != nil {
		return ownerfacade.ImportOwnerCreateResponse{}, uuid.Nil, createErr
	}
	var storedOwner uuid.UUID
	table := "task_requests"
	if viewSchemaID == tasksdecisions.DecisionsViewSchemaID {
		table = "decisions"
	}
	if queryErr := tx.QueryRow(ctx, fmt.Sprintf(`SELECT owner_user_id FROM %s WHERE record_id = $1`, table), response.RecordID).Scan(&storedOwner); queryErr != nil {
		t.Fatalf("query imported %s owner: %v", viewSchemaID, queryErr)
	}
	return response, storedOwner, nil
}

func tasksDecisionsImportFacade(t testing.TB, harness tasksDecisionsImportHarness, viewSchemaID string) ownerfacade.ImportOwnerCreateFacade {
	t.Helper()
	facadeID := "tasksdecisions.task_request.import_create"
	if viewSchemaID == tasksdecisions.DecisionsViewSchemaID {
		facadeID = "tasksdecisions.decision.import_create"
	}
	if facade, ok := harness.registry.Resolve(viewSchemaID, facadeID); ok {
		return facade
	}
	t.Fatalf("application import registry missing %s", viewSchemaID)
	return nil
}

func tasksDecisionsImportFields(viewSchemaID string, ownerID *uuid.UUID, explicitNull bool) []ownerfacade.ImportFieldValue {
	text := func(field string, value string) ownerfacade.ImportFieldValue {
		return ownerfacade.ImportFieldValue{FieldKey: field, NormalizedValue: ownerfacade.ImportScalarValue{Kind: "text", Text: &value}}
	}
	var fields []ownerfacade.ImportFieldValue
	ownerField := "task.owner_user_id"
	if viewSchemaID == tasksdecisions.TaskRequestsViewSchemaID {
		fields = []ownerfacade.ImportFieldValue{text("task.title", "Imported task"), text("task.task_kind", "follow_up")}
	} else {
		ownerField = "decision.owner_user_id"
		fields = []ownerfacade.ImportFieldValue{
			text("decision.summary", "Imported decision"),
			text("decision.decision_type", "scope"),
			text("decision.rationale", "Imported rationale"),
		}
	}
	if ownerID != nil {
		fields = append(fields, ownerfacade.ImportFieldValue{
			FieldKey:        ownerField,
			NormalizedValue: ownerfacade.ImportScalarValue{Kind: "uuid", UUID: ownerID},
		})
	} else if explicitNull {
		fields = append(fields, ownerfacade.ImportFieldValue{
			FieldKey:        ownerField,
			NormalizedValue: ownerfacade.ImportScalarValue{Kind: "null"},
		})
	}
	return fields
}

func requireTasksDecisionsOwnerValidation(t testing.TB, err error) {
	t.Helper()
	if err == nil {
		t.Fatal("invalid imported owner unexpectedly succeeded")
	}
	var validation *tasksdecisions.ValidationError
	var ownerValidation *ownerfacade.ImportOwnerCreateError
	if !errors.As(err, &validation) && !errors.As(err, &ownerValidation) {
		t.Fatalf("imported owner error = %T %v; want safe owner validation", err, err)
	}
	if errors.As(err, &ownerValidation) && ownerValidation.OwnerCode != ownerfacade.ImportOwnerCreateValidationFailed {
		t.Fatalf("owner error code = %q", ownerValidation.OwnerCode)
	}
}

type tasksDecisionsImportCounts struct {
	records     int
	tasks       int
	decisions   int
	changeSets  int
	mutations   int
	revisions   int
	projections int
	intents     int
}

func tasksDecisionsImportEffectCounts(t testing.TB, db postgres.DB, incidentID uuid.UUID) tasksDecisionsImportCounts {
	t.Helper()
	ctx := context.Background()
	queries := []string{
		`SELECT count(*) FROM records WHERE incident_id = $1`,
		`SELECT count(*) FROM task_requests WHERE incident_id = $1`,
		`SELECT count(*) FROM decisions WHERE incident_id = $1`,
		`SELECT count(*) FROM change_sets WHERE incident_id = $1`,
		`SELECT count(*) FROM change_set_mutations mutation JOIN change_sets change_set USING (change_set_id) WHERE change_set.incident_id = $1`,
		`SELECT count(*) FROM record_revisions revision JOIN records record USING (record_id) WHERE record.incident_id = $1`,
		`SELECT (SELECT count(*) FROM task_request_grid_projection WHERE incident_id = $1) + (SELECT count(*) FROM decision_grid_projection WHERE incident_id = $1)`,
		`SELECT count(*) FROM collaboration_event_intents WHERE incident_id = $1`,
	}
	values := []*int{}
	counts := tasksDecisionsImportCounts{}
	values = append(values, &counts.records, &counts.tasks, &counts.decisions, &counts.changeSets, &counts.mutations, &counts.revisions, &counts.projections, &counts.intents)
	for index, query := range queries {
		if err := db.QueryRow(ctx, query, incidentID).Scan(values[index]); err != nil {
			t.Fatalf("query task/decision import effect %d: %v", index, err)
		}
	}
	return counts
}
