package assessments_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/modules/assessments"
	authstoretest "github.com/JochiRaider/cartulary/internal/modules/auth/testsupport/storetest"
	entitytest "github.com/JochiRaider/cartulary/internal/modules/entities/testsupport"
	"github.com/JochiRaider/cartulary/internal/modules/records"
	"github.com/JochiRaider/cartulary/internal/platform/postgres"
	"github.com/JochiRaider/cartulary/internal/testutil/appsupport"
)

func TestAssessmentFacadeParticipantFailuresRollbackAndReplayBypassesParticipants(t *testing.T) {
	harness := appsupport.StartStore(t, "assessments-facade-transaction")
	actor := authstoretest.SeedLocalUserRecord(
		t,
		harness.DB,
		"assessment-facade@example.test",
		"Assessment Facade",
		"AssessmentFacadePass1!",
		false,
		false,
		true,
	)
	incident := appsupport.CreateIncidentInStore(
		t,
		harness.DB,
		actor,
		"txn-assessment-facade-incident",
		"IR-ASSESSMENT-FACADE",
		"Assessment facade transaction",
	)
	subjectID := uuid.New()
	entitytest.SeedHostRecord(
		t,
		harness.DB,
		incident.ID,
		actor.ID,
		subjectID,
		"Assessment facade host",
		"assessment-facade-host",
		"",
		"",
	)

	for _, failure := range []string{
		"idempotency_lookup",
		"subject",
		"support_targets",
		"assessor",
		"records",
		"support_links",
		"projection",
		"revisions",
		"idempotency_store",
	} {
		t.Run(failure, func(t *testing.T) {
			ports := &assessmentFacadePorts{failure: failure}
			facade := newAssessmentFacadeForTest(t, harness.DB, ports)
			command := assessmentFacadeCommand(actor.ID, incident.ID, subjectID, "txn-"+failure)
			assessor := actor.ID
			command.Input.Assessor = &assessor

			if _, err := facade.Create(context.Background(), command); err == nil {
				t.Fatalf("expected injected %s failure", failure)
			}
			if got := appsupport.QueryCount(
				t,
				harness.DB,
				`SELECT COUNT(*) FROM assessments WHERE incident_id = $1`,
				incident.ID,
			); got != 0 {
				t.Fatalf("%s failure retained %d assessment rows", failure, got)
			}
			if got := appsupport.QueryCount(
				t,
				harness.DB,
				`SELECT COUNT(*) FROM records WHERE incident_id = $1 AND record_type = 'assessment'`,
				incident.ID,
			); got != 0 {
				t.Fatalf("%s failure retained %d assessment envelopes", failure, got)
			}
		})
	}

	replayedRecordID := uuid.New()
	replayedChangeSetID := uuid.New()
	replayPorts := &assessmentFacadePorts{
		replay: &assessments.CreateIdempotencyRecord{
			RequestHash: []byte("stable-hash"),
			Result: assessments.CreateResult{
				Outcome:      assessments.CreateOutcomeCommitted,
				CanonicalRow: map[string]any{"record_id": replayedRecordID.String(), "row_version": int64(4)},
				RecordID:     replayedRecordID,
				ChangeSetID:  replayedChangeSetID,
				RowVersion:   4,
			},
		},
	}
	facade := newAssessmentFacadeForTest(t, harness.DB, replayPorts)
	result, err := facade.Create(
		context.Background(),
		assessmentFacadeCommand(actor.ID, incident.ID, subjectID, "txn-replay"),
	)
	if err != nil {
		t.Fatalf("replay assessment create: %v", err)
	}
	if result.Outcome != assessments.CreateOutcomeReplayed ||
		result.RecordID != replayedRecordID ||
		result.ChangeSetID != replayedChangeSetID ||
		result.RowVersion != 4 {
		t.Fatalf("unexpected replay result: %#v", result)
	}
	if replayPorts.participantCalls != 0 {
		t.Fatalf("replay called %d transactional participants", replayPorts.participantCalls)
	}
}

type assessmentFacadePorts struct {
	failure          string
	replay           *assessments.CreateIdempotencyRecord
	participantCalls int
}

func (p *assessmentFacadePorts) LookupCreate(
	_ context.Context,
	_ assessments.CreateIdempotencyKey,
) (assessments.CreateIdempotencyRecord, bool, error) {
	if p.failure == "idempotency_lookup" {
		return assessments.CreateIdempotencyRecord{}, false, errors.New("injected idempotency lookup failure")
	}
	if p.replay != nil {
		return *p.replay, true, nil
	}
	return assessments.CreateIdempotencyRecord{}, false, nil
}

func (p *assessmentFacadePorts) StoreCreateTx(
	_ context.Context,
	_ pgx.Tx,
	_ assessments.CreateIdempotencyKey,
	_ assessments.CreateResult,
) error {
	p.participantCalls++
	return p.inject("idempotency_store")
}

func (p *assessmentFacadePorts) ValidateAssessmentSubjectTx(
	_ context.Context,
	_ pgx.Tx,
	_ uuid.UUID,
	_ string,
	_ uuid.UUID,
) (bool, error) {
	p.participantCalls++
	if err := p.inject("subject"); err != nil {
		return false, err
	}
	return true, nil
}

func (p *assessmentFacadePorts) ValidateAssessmentAssessorTx(
	_ context.Context,
	_ pgx.Tx,
	_ uuid.UUID,
	_ uuid.UUID,
) (bool, error) {
	p.participantCalls++
	if err := p.inject("assessor"); err != nil {
		return false, err
	}
	return true, nil
}

func (p *assessmentFacadePorts) ValidateAssessmentSupportTargetsTx(
	_ context.Context,
	_ pgx.Tx,
	_ uuid.UUID,
	_ []uuid.UUID,
) (bool, error) {
	p.participantCalls++
	if err := p.inject("support_targets"); err != nil {
		return false, err
	}
	return true, nil
}

func (p *assessmentFacadePorts) CreateAssessmentEnvelopeTx(
	ctx context.Context,
	tx pgx.Tx,
	create assessments.RecordEnvelopeCreate,
) (uuid.UUID, error) {
	p.participantCalls++
	if err := p.inject("records"); err != nil {
		return uuid.Nil, err
	}
	return records.NewStore().InsertTx(ctx, tx, records.InsertParams{
		IncidentID:      create.IncidentID,
		RecordType:      create.RecordType,
		CreatedByUserID: create.ActorID,
		CreatedAt:       create.Now,
		UpdatedByUserID: create.ActorID,
		UpdatedAt:       create.Now,
		RowVersion:      create.RowVersion,
	})
}

func (p *assessmentFacadePorts) ApplyInitialAssessmentSupportLinksTx(
	_ context.Context,
	_ pgx.Tx,
	_ uuid.UUID,
	_ uuid.UUID,
	_ uuid.UUID,
	_ []uuid.UUID,
	_ time.Time,
) error {
	p.participantCalls++
	return p.inject("support_links")
}

func (p *assessmentFacadePorts) AppendAssessmentCreateRevisionTx(
	_ context.Context,
	_ pgx.Tx,
	_ assessments.CreateRevision,
) (uuid.UUID, error) {
	p.participantCalls++
	if err := p.inject("revisions"); err != nil {
		return uuid.Nil, err
	}
	return uuid.New(), nil
}

func (p *assessmentFacadePorts) RefreshAndLoadAssessmentRowTx(
	_ context.Context,
	_ pgx.Tx,
	recordID uuid.UUID,
) (map[string]any, error) {
	p.participantCalls++
	if err := p.inject("projection"); err != nil {
		return nil, err
	}
	return map[string]any{
		"record_id":   recordID.String(),
		"row_version": int64(1),
	}, nil
}

func (p *assessmentFacadePorts) inject(participant string) error {
	if p.failure == participant {
		return errors.New("injected " + participant + " failure")
	}
	return nil
}

func newAssessmentFacadeForTest(
	t *testing.T,
	database postgres.DB,
	ports *assessmentFacadePorts,
) *assessments.Facade {
	t.Helper()
	facade, err := assessments.NewFacade(database, assessments.FacadeDependencies{
		Idempotency:    ports,
		Subjects:       ports,
		Assessors:      ports,
		SupportTargets: ports,
		Records:        ports,
		SupportLinks:   ports,
		Revisions:      ports,
		Projections:    ports,
	})
	if err != nil {
		t.Fatalf("construct assessment facade: %v", err)
	}
	return facade
}

func assessmentFacadeCommand(
	actorID uuid.UUID,
	incidentID uuid.UUID,
	subjectID uuid.UUID,
	clientTxnID string,
) assessments.CreateCommand {
	return assessments.CreateCommand{
		ActorUserID: actorID,
		IncidentID:  incidentID,
		Input: assessments.CreateInput{
			ClientTxnID:     clientTxnID,
			SubjectRef:      subjectID,
			SubjectType:     "host",
			AssessmentState: "confirmed",
			Rationale:       "Assessment facade transaction rationale.",
		},
		Idempotency: assessments.CreateIdempotencyKey{
			RouteKey:    "assessments.rows.create",
			ActorUserID: actorID,
			ScopeKey:    incidentID.String() + ":" + assessments.AssessmentsViewSchemaID,
			ClientTxnID: clientTxnID,
			RequestHash: []byte("stable-hash"),
		},
		RequestID: "req-" + clientTxnID,
		Now:       time.Date(2026, 7, 30, 18, 0, 0, 0, time.UTC),
	}
}
