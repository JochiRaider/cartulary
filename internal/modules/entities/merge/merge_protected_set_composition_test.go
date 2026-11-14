package merge_test

import (
	"context"
	"errors"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/app/entitymergeassembly"
	"github.com/JochiRaider/cartulary/internal/modules/assessments"
	authstoretest "github.com/JochiRaider/cartulary/internal/modules/auth/testsupport/storetest"
	"github.com/JochiRaider/cartulary/internal/modules/entities/hostidentity"
	"github.com/JochiRaider/cartulary/internal/modules/entities/mentions"
	entitymerge "github.com/JochiRaider/cartulary/internal/modules/entities/merge"
	"github.com/JochiRaider/cartulary/internal/modules/incidents"
	"github.com/JochiRaider/cartulary/internal/modules/incidents/testsupport/admissiontest"
	"github.com/JochiRaider/cartulary/internal/modules/timeline/mentioneffects"
	workbookstartuppostgres "github.com/JochiRaider/cartulary/internal/modules/workbook/startup/postgres"
	"github.com/JochiRaider/cartulary/internal/platform/authn"
	"github.com/JochiRaider/cartulary/internal/platform/postgres/dbport"
	"github.com/JochiRaider/cartulary/internal/testutil/pgtest"
	"github.com/JochiRaider/cartulary/internal/testutil/revisionsupport"
)

func TestMergeProtectedRecordIDsIncludesAssessmentSubjects(t *testing.T) {
	db := pgtest.Start(t).BeginRollbackDBT(t, "merge-protected-set-composition")
	actor := seedComposedMergeUser(t, db)
	incident := createComposedMergeIncident(t, db, actor, "")
	survivorID := uuid.New()
	loserID := uuid.New()
	assessmentID := uuid.New()
	seedComposedMergeHost(t, db, incident.ID, actor.ID, survivorID, "Survivor host", "survivor-host")
	seedComposedMergeHost(t, db, incident.ID, actor.ID, loserID, "Loser host", "loser-host")
	seedComposedMergeAssessment(t, db, incident.ID, actor.ID, assessmentID, loserID)

	composition := revisionsupport.MustComposition(t)
	appender := composition.Runtime.Appender()
	assessmentEffects, err := assessments.NewMergeEffects(composedMergeAssessmentProjection{}, appender)
	if err != nil {
		t.Fatalf("construct assessment merge effects: %v", err)
	}
	entityAssessmentEffects, err := entitymergeassembly.NewAssessmentEffects(assessmentEffects)
	if err != nil {
		t.Fatalf("adapt assessment merge effects: %v", err)
	}
	t.Run("assessment adapter translates protected-set drift", func(t *testing.T) {
		tx, err := db.BeginTx(context.Background(), pgx.TxOptions{})
		if err != nil {
			t.Fatalf("begin assessment adapter translation transaction: %v", err)
		}
		defer func() { _ = tx.Rollback(context.Background()) }()
		protectedRecordIDs := []uuid.UUID{survivorID, loserID}
		slices.SortFunc(protectedRecordIDs, func(left, right uuid.UUID) int {
			return strings.Compare(left.String(), right.String())
		})
		_, err = entityAssessmentEffects.RepointTx(context.Background(), tx, entitymerge.AssessmentRepointCommand{
			IncidentID:         incident.ID,
			RecordType:         "host",
			SurvivorRecordID:   survivorID,
			LoserRecordID:      loserID,
			ProtectedRecordIDs: protectedRecordIDs,
			Now:                time.Now().UTC(),
		})
		var changed *entitymerge.AssessmentProtectedSetChangedError
		if !errors.As(err, &changed) || changed.RecordID != assessmentID {
			t.Fatalf("translated protected-set error = %T %[1]v, want assessment %s", err, assessmentID)
		}
	})
	t.Run("assessment adapter owns mutable command and result values", func(t *testing.T) {
		tx, err := db.BeginTx(context.Background(), pgx.TxOptions{})
		if err != nil {
			t.Fatalf("begin assessment adapter ownership transaction: %v", err)
		}
		defer func() { _ = tx.Rollback(context.Background()) }()
		protectedRecordIDs := []uuid.UUID{survivorID, loserID, assessmentID}
		slices.SortFunc(protectedRecordIDs, func(left, right uuid.UUID) int {
			return strings.Compare(left.String(), right.String())
		})
		originalProtectedRecordIDs := append([]uuid.UUID(nil), protectedRecordIDs...)
		result, err := entityAssessmentEffects.RepointTx(context.Background(), tx, entitymerge.AssessmentRepointCommand{
			IncidentID:         incident.ID,
			RecordType:         "host",
			SurvivorRecordID:   survivorID,
			LoserRecordID:      loserID,
			ProtectedRecordIDs: protectedRecordIDs,
			Now:                time.Now().UTC(),
		})
		if err != nil {
			t.Fatalf("repoint through assessment adapter: %v", err)
		}
		if !slices.Equal(protectedRecordIDs, originalProtectedRecordIDs) {
			t.Fatalf("adapter mutated protected record IDs: got %v want %v", protectedRecordIDs, originalProtectedRecordIDs)
		}
		if result.RepointedCount != 1 || len(result.Mutations) != 1 {
			t.Fatalf("assessment adapter result = %#v, want one repoint", result)
		}
		before := result.Mutations[0].BeforeValue.(map[string]any)
		after := result.Mutations[0].AfterValue.(map[string]any)
		before["subject_record_id"] = "mutated-by-consumer"
		if after["subject_record_id"] != survivorID.String() {
			t.Fatalf("assessment mutation values alias: before=%#v after=%#v", before, after)
		}
		if result.Mutations[0].BeforeSnapshot == result.Mutations[0].AfterSnapshot {
			t.Fatal("assessment mutation snapshots alias")
		}
	})
	dependencies := entitymerge.StoreDependencies{
		Postgres:       db,
		Revisions:      appender,
		HostIdentity:   hostidentity.NewMergeCapability(),
		Assessments:    entityAssessmentEffects,
		Mentions:       composedMentionStore{},
		Links:          composedMergeLinkEffects{},
		Timeline:       composedMergeTimelineEffects{},
		Projections:    composedMergeWorkbookProjection{},
		Collaboration:  composition.RecordChanges,
		ReferenceScope: entitymergeassembly.NewReferenceScope(),
	}

	extraID := uuid.New()
	seedComposedMergeHost(t, db, incident.ID, actor.ID, extraID, "Unchanged peer", "unchanged-peer")
	foreignIncident := createComposedMergeIncident(t, db, actor, "-FOREIGN")
	foreignID := uuid.New()
	seedComposedMergeHost(t, db, foreignIncident.ID, actor.ID, foreignID, "Foreign peer", "foreign-peer")
	scopeFailure := errors.New("reference owner unavailable")
	for _, mode := range []string{"changed set", "owner error", "zero identity", "missing identity", "foreign incident"} {
		t.Run(mode, func(t *testing.T) {
			calls := 0
			configured := dependencies
			configured.ReferenceScope = composedReferenceScopeFunc(func(_ context.Context, _ pgx.Tx, id uuid.UUID, changed []uuid.UUID) ([]uuid.UUID, error) {
				calls++
				want := []uuid.UUID{survivorID, loserID, assessmentID}
				less := func(a, b uuid.UUID) int { return strings.Compare(a.String(), b.String()) }
				slices.SortFunc(want, less)
				slices.SortFunc(changed, less)
				if id != incident.ID || !slices.Equal(changed, want) {
					t.Fatalf("dependency scope incident=%s records=%v want=%v", id, changed, want)
				}
				switch mode {
				case "changed set":
					if calls == 1 {
						return nil, nil
					}
					return []uuid.UUID{extraID}, nil
				case "owner error":
					return nil, scopeFailure
				case "zero identity":
					return []uuid.UUID{uuid.Nil}, nil
				case "missing identity":
					return []uuid.UUID{uuid.New()}, nil
				default:
					return []uuid.UUID{foreignID}, nil
				}
			})
			candidate, err := entitymerge.NewStore(configured)
			if err != nil {
				t.Fatal(err)
			}
			_, err = candidate.MergeEntity(t.Context(), actor, survivorID, entitymerge.MergeRequest{LoserRecordID: loserID, SurvivorBaseRowVersion: 1, LoserBaseRowVersion: 1, ClientTxnID: "scope-" + mode}, []byte(mode), "scope-request", time.Now().UTC())
			if err == nil {
				t.Fatal("invalid dependency scope was admitted")
			}
			wantCalls := 1
			if mode == "changed set" {
				wantCalls = 2
				var precondition *entitymerge.MergePreconditionError
				if !errors.As(err, &precondition) || precondition.ReasonCode != "protected_set_changed" {
					t.Fatalf("scope drift error=%v", err)
				}
			} else if mode == "owner error" && !errors.Is(err, scopeFailure) {
				t.Fatalf("owner cause lost: %v", err)
			}
			if calls != wantCalls {
				t.Fatalf("dependency reads=%d want=%d", calls, wantCalls)
			}
			var unchanged bool
			if err := db.QueryRow(t.Context(), `SELECT
 (SELECT count(*) FROM records WHERE record_id=ANY($1::uuid[]) AND row_version=1)=3
 AND NOT EXISTS(SELECT 1 FROM record_revisions WHERE record_id=ANY($1::uuid[]))
 AND NOT EXISTS(SELECT 1 FROM change_sets WHERE client_txn_id=$2)
 AND NOT EXISTS(SELECT 1 FROM route_idempotency WHERE client_txn_id=$2)
 AND (SELECT subject_record_id=$3 FROM assessments WHERE record_id=$4)`, []uuid.UUID{survivorID, loserID, assessmentID}, "scope-"+mode, loserID, assessmentID).Scan(&unchanged); err != nil {
				t.Fatal(err)
			}
			if !unchanged {
				t.Fatal("scope rejection left partial mutation")
			}
		})
	}
	store, err := entitymerge.NewStore(dependencies)
	if err != nil {
		t.Fatalf("compose Merge store: %v", err)
	}

	result, err := store.MergeEntity(
		context.Background(),
		actor,
		survivorID,
		entitymerge.MergeRequest{
			LoserRecordID:          loserID,
			SurvivorBaseRowVersion: 1,
			LoserBaseRowVersion:    1,
			ClientTxnID:            "txn-merge-protected-set-composition",
		},
		[]byte("merge-protected-set-composition"),
		"req-merge-protected-set-composition",
		time.Now().UTC(),
	)
	if err != nil {
		t.Fatalf("merge with protected assessment: %v", err)
	}
	if result.MergeSummary.RepointedAssessmentCount != 1 {
		t.Fatalf("repointed assessment count = %d, want 1", result.MergeSummary.RepointedAssessmentCount)
	}
	var subjectRecordID uuid.UUID
	if err := db.QueryRow(context.Background(), `SELECT subject_record_id FROM assessments WHERE record_id = $1`, assessmentID).Scan(&subjectRecordID); err != nil {
		t.Fatalf("load repointed assessment: %v", err)
	}
	if subjectRecordID != survivorID {
		t.Fatalf("assessment subject = %s, want survivor %s", subjectRecordID, survivorID)
	}
}

type composedMergeLinkEffects struct{}

type composedMentionStore struct{}

func (composedMentionStore) RepointMergedMentionsTx(context.Context, pgx.Tx, mentions.RepointMergedMentionsCommand) (mentions.RepointMergedMentionsResult, error) {
	return mentions.RepointMergedMentionsResult{
		Mutations:             []mentions.MergeMutation{},
		TimelineInvalidations: map[uuid.UUID][]string{},
	}, nil
}

func (composedMergeLinkEffects) RepointLinksTx(
	context.Context,
	pgx.Tx,
	entitymerge.RepointLinksCommand,
) (entitymerge.RepointLinksResult, error) {
	return entitymerge.RepointLinksResult{
		Mutations:                 []entitymerge.LinkEffectMutation{},
		LinkTypesBySourceRecordID: map[uuid.UUID][]string{},
	}, nil
}

func (composedMergeLinkEffects) RepointTagsTx(
	context.Context,
	pgx.Tx,
	entitymerge.RepointTagsCommand,
) (entitymerge.RepointTagsResult, error) {
	return entitymerge.RepointTagsResult{Mutations: []entitymerge.LinkEffectMutation{}}, nil
}

type composedMergeAssessmentProjection struct{}

func (composedMergeAssessmentProjection) RefreshAssessmentProjectionTx(context.Context, pgx.Tx, uuid.UUID) error {
	return nil
}

type composedMergeWorkbookProjection struct{}

func (composedMergeWorkbookProjection) RefreshHostTx(context.Context, pgx.Tx, uuid.UUID) error {
	return nil
}

func (composedMergeWorkbookProjection) RefreshIdentityTx(context.Context, pgx.Tx, uuid.UUID) error {
	return nil
}

func (composedMergeWorkbookProjection) DeleteHostTx(context.Context, pgx.Tx, uuid.UUID) error {
	return nil
}

func (composedMergeWorkbookProjection) DeleteIdentityTx(context.Context, pgx.Tx, uuid.UUID) error {
	return nil
}

type composedMergeTimelineEffects struct{}

func (composedMergeTimelineEffects) LoadTimelineInvalidationsTx(context.Context, pgx.Tx, map[uuid.UUID][]string) ([]mentioneffects.TimelineInvalidation, error) {
	return nil, nil
}

func (composedMergeTimelineEffects) LoadRelationshipInvalidationsTx(context.Context, pgx.Tx, map[uuid.UUID][]string) ([]mentioneffects.TimelineInvalidation, error) {
	return nil, nil
}

func (composedMergeTimelineEffects) RefreshTimelineProjectionRowsTx(context.Context, pgx.Tx, []uuid.UUID) error {
	return nil
}

func seedComposedMergeUser(t testing.TB, db dbport.DB) authn.UserRecord {
	t.Helper()
	return authstoretest.SeedActor(t, db, authstoretest.ActorAttributes{Email: "merge-protected-composition@example.test", DisplayName: "Merge Protected", DeploymentAdmin: true})
}

func createComposedMergeIncident(t testing.TB, db dbport.DB, actor authn.UserRecord, suffix string) incidents.IncidentRecord {
	t.Helper()
	application, err := incidents.NewApplication(incidents.ApplicationDependencies{
		Postgres:            db,
		PreferenceBootstrap: workbookstartuppostgres.NewWriter(),
		Now:                 time.Now,
	})
	if err != nil {
		t.Fatalf("construct Incidents application: %v", err)
	}
	result, err := application.CreateIncident(
		context.Background(),
		actor,
		admissiontest.IncidentCreate(t, admissiontest.IncidentCreateInput{
			ClientTxnID: "txn-merge-protected-incident-composition" + suffix,
			IncidentKey: "IR-MERGE-PROTECTED-COMPOSITION" + suffix,
			Title:       "Merge protected set composition",
		}),
		"req-merge-protected-incident-composition",
	)
	if err != nil {
		t.Fatalf("create merge protected-set incident: %v", err)
	}
	return result.Incident
}

func seedComposedMergeHost(t testing.TB, db dbport.DB, incidentID uuid.UUID, actorID uuid.UUID, recordID uuid.UUID, displayName string, hostname string) {
	t.Helper()
	seedComposedMergeRecord(t, db, incidentID, actorID, recordID, "host")
	if _, err := db.Exec(context.Background(), `
INSERT INTO hosts (record_id, incident_id, display_name, hostname, host_state, created_by_user_id, updated_by_user_id)
VALUES ($1, $2, $3, $4, 'canonical', $5, $5)
`, recordID, incidentID, displayName, hostname, actorID); err != nil {
		t.Fatalf("seed merge protected-set host: %v", err)
	}
}

func seedComposedMergeAssessment(t testing.TB, db dbport.DB, incidentID uuid.UUID, actorID uuid.UUID, assessmentID uuid.UUID, subjectID uuid.UUID) {
	t.Helper()
	seedComposedMergeRecord(t, db, incidentID, actorID, assessmentID, "assessment")
	if _, err := db.Exec(context.Background(), `
INSERT INTO assessments (record_id, incident_id, subject_record_id, subject_type, assessment_state, rationale, assessor_user_id)
VALUES ($1, $2, $3, 'host', 'confirmed', 'Seeded merge protected set assessment.', $4)
`, assessmentID, incidentID, subjectID, actorID); err != nil {
		t.Fatalf("seed merge protected-set assessment: %v", err)
	}
}

func seedComposedMergeRecord(t testing.TB, db dbport.DB, incidentID uuid.UUID, actorID uuid.UUID, recordID uuid.UUID, recordType string) {
	t.Helper()
	if _, err := db.Exec(context.Background(), `
INSERT INTO records (record_id, incident_id, record_type, created_by_user_id, updated_by_user_id)
VALUES ($1, $2, $3, $4, $4)
`, recordID, incidentID, recordType, actorID); err != nil {
		t.Fatalf("seed merge protected-set record: %v", err)
	}
}

type composedReferenceScopeFunc func(context.Context, pgx.Tx, uuid.UUID, []uuid.UUID) ([]uuid.UUID, error)

func (f composedReferenceScopeFunc) DependentsTx(ctx context.Context, tx pgx.Tx, incident uuid.UUID, changed []uuid.UUID) ([]uuid.UUID, error) {
	return f(ctx, tx, incident, changed)
}
