package tasksdecisions

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"

	"github.com/JochiRaider/cartulary/internal/modules/collaboration"
	"github.com/JochiRaider/cartulary/internal/modules/links"
	"github.com/JochiRaider/cartulary/internal/modules/records"
	"github.com/JochiRaider/cartulary/internal/modules/revisions"
	conflicttokens "github.com/JochiRaider/cartulary/internal/modules/revisions/conflicts"
)

func TestMutationContributionRejectsIncompleteDependencies_Unit(t *testing.T) {
	assertEveryMutationDependencyRequired(t)
	if _, err := NewMutationContribution(nil, conflicttokens.ConflictTokenCodec{}, MutationDependencies{}); err == nil || !strings.Contains(err.Error(), "Postgres is required") {
		t.Fatalf("nil Postgres error = %v", err)
	}
	if _, err := NewIncidentBundleSourcePort(nil, nil, nil); err == nil || !strings.Contains(err.Error(), "Links facts are required") {
		t.Fatalf("missing incident-bundle Links facts error = %v", err)
	}
	if port, err := NewIncidentBundleSourcePort(compositionOperations{}, nil, nil); err == nil || port != nil || !strings.Contains(err.Error(), "retained field references are required") {
		t.Fatalf("missing retained binding reader port=%#v error=%v", port, err)
	}
}

func assertEveryMutationDependencyRequired(t *testing.T) {
	t.Helper()
	for _, tc := range []struct {
		name   string
		remove func(*MutationDependencies)
	}{
		{"Incident admission", func(d *MutationDependencies) { d.IncidentState = nil }},
		{"Route idempotency", func(d *MutationDependencies) { d.Idempotency = nil }},
		{"Record envelopes", func(d *MutationDependencies) { d.RecordEnvelopes = nil }},
		{"Reference targets", func(d *MutationDependencies) { d.ReferenceTargets = nil }},
		{"Links", func(d *MutationDependencies) { d.Links = nil }},
		{"Links facts", func(d *MutationDependencies) { d.LinkFacts = nil }},
		{"Projections", func(d *MutationDependencies) { d.Projections = nil }},
		{"Revisions/history", func(d *MutationDependencies) { d.Revisions = nil }},
		{"Conflict fields", func(d *MutationDependencies) { d.ConflictFields = nil }},
		{"Keep-saved idempotency", func(d *MutationDependencies) { d.KeepSavedIdempotency = nil }},
		{"Collaboration publication", func(d *MutationDependencies) { d.Collaboration = nil }},
	} {
		dependencies := completeMutationDependencies()
		tc.remove(&dependencies)
		facade, err := NewMutationContribution(compositionDB{}, conflicttokens.ConflictTokenCodec{}, dependencies)
		if err == nil || facade != nil || err.Error() != "tasks/decisions mutation dependencies: "+tc.name+" is required" {
			t.Fatalf("missing %s: facade=%#v error=%v", tc.name, facade, err)
		}
	}

}

func TestMutationContributionSharesOneFacade_Unit(t *testing.T) {
	facade, err := NewMutationContribution(
		compositionDB{},
		conflicttokens.ConflictTokenCodec{},
		completeMutationDependencies(),
	)
	if err != nil {
		t.Fatalf("construct mutation contribution: %v", err)
	}
	var create mutationCreateConsumer = facade
	var patch mutationPatchConsumer = facade
	var conflict mutationConflictConsumer = facade
	var supersede mutationSupersedeConsumer = facade
	if create != facade || patch != facade || conflict != facade || supersede != facade {
		t.Fatal("mutation consumers did not retain the one facade instance")
	}
}

func TestImportContributionRejectsIncompleteDependencies_Unit(t *testing.T) {
	operations := compositionOperations{}
	completeImport := func() ImportDependencies {
		return ImportDependencies{
			RecordEnvelopes: records.NewStore(), ReferenceTargets: operations, Links: links.NewStore(),
			Projections: operations, Revisions: compositionImportRevisions{}, Collaboration: operations,
		}
	}
	if _, err := NewImportContribution(TaskRequestsViewSchemaID, "tasksdecisions.task_request.import_create", completeImport()); err != nil {
		t.Fatalf("complete import dependencies rejected: %v", err)
	}
	for _, tc := range []struct {
		message string
		remove  func(*ImportDependencies)
	}{
		{"Records insert is required", func(d *ImportDependencies) { d.RecordEnvelopes = nil }},
		{"Reference targets are required", func(d *ImportDependencies) { d.ReferenceTargets = nil }},
		{"Links synchronization is required", func(d *ImportDependencies) { d.Links = nil }},
		{"Projection refresh/load is required", func(d *ImportDependencies) { d.Projections = nil }},
		{"Revision finalization is required", func(d *ImportDependencies) { d.Revisions = nil }},
		{"Collaboration publication is required", func(d *ImportDependencies) { d.Collaboration = nil }},
	} {
		dependencies := completeImport()
		tc.remove(&dependencies)
		if _, err := NewImportContribution(TaskRequestsViewSchemaID, "tasksdecisions.task_request.import_create", dependencies); err == nil || err.Error() != "tasks/decisions import dependencies: "+tc.message {
			t.Fatalf("missing import dependency %q: %v", tc.message, err)
		}
	}
}

type mutationCreateConsumer interface {
	Create(context.Context, CreateCommand) (MutationResult, error)
}

type mutationPatchConsumer interface {
	Patch(context.Context, PatchCommand) (MutationResult, error)
}

type mutationConflictConsumer interface {
	ResolveConflict(context.Context, ConflictCommand) (MutationResult, error)
}

type mutationSupersedeConsumer interface {
	SupersedeDecision(context.Context, SupersedeCommand) (SupersedeMutationResult, error)
}

func completeMutationDependencies() MutationDependencies {
	operations := compositionOperations{}
	fieldResolver, err := conflicttokens.NewFieldResolverCatalog(nil)
	if err != nil {
		panic(err)
	}
	return MutationDependencies{
		IncidentState:        operations,
		Idempotency:          compositionIdempotency{},
		RecordEnvelopes:      records.NewStore(),
		ReferenceTargets:     operations,
		Links:                links.NewStore(),
		LinkFacts:            operations,
		Projections:          operations,
		Revisions:            operations,
		ConflictFields:       fieldResolver,
		KeepSavedIdempotency: compositionKeepSavedIdempotency{},
		Collaboration:        operations,
	}
}

type compositionDB struct{}

func (compositionDB) Exec(context.Context, string, ...any) (pgconn.CommandTag, error) {
	return pgconn.NewCommandTag(""), errors.New("unexpected composition DB execution")
}

func (compositionDB) Query(context.Context, string, ...any) (pgx.Rows, error) {
	return nil, errors.New("unexpected composition DB query")
}

func (compositionDB) QueryRow(context.Context, string, ...any) pgx.Row { return compositionRow{} }

func (compositionDB) BeginTx(context.Context, pgx.TxOptions) (pgx.Tx, error) {
	return nil, errors.New("unexpected composition transaction")
}

type compositionRow struct{}

func (compositionRow) Scan(...any) error { return errors.New("unexpected composition row scan") }

type compositionIdempotency struct{}

func (compositionIdempotency) Get(context.Context, IdempotencyKey, []byte) (IdempotencyRecord, error) {
	return IdempotencyRecord{}, ErrIdempotencyNotFound
}

func (compositionIdempotency) PutTx(context.Context, pgx.Tx, IdempotencyKey, []byte, StoredMutationResult) error {
	return nil
}

type compositionKeepSavedIdempotency struct{}

func (compositionKeepSavedIdempotency) Get(context.Context, conflicttokens.IdempotencyKey, []byte) (conflicttokens.IdempotencyRecord, error) {
	return conflicttokens.IdempotencyRecord{}, conflicttokens.ErrIdempotencyNotFound
}

func (compositionKeepSavedIdempotency) PutTx(context.Context, pgx.Tx, conflicttokens.IdempotencyKey, []byte, conflicttokens.StoredTarget) error {
	return nil
}

type compositionOperations struct{}

func (compositionOperations) LoadRecordLinkFactsTx(context.Context, pgx.Tx, uuid.UUID, uuid.UUID) ([]LinkFact, error) {
	return []LinkFact{}, nil
}

func (compositionOperations) RequireOpenTx(context.Context, pgx.Tx, uuid.UUID) error { return nil }

func (compositionOperations) RefreshTaskRequestTx(context.Context, pgx.Tx, uuid.UUID) error {
	return nil
}

func (compositionOperations) RefreshDecisionTx(context.Context, pgx.Tx, uuid.UUID) error { return nil }

func (compositionOperations) LoadTaskRequestTx(context.Context, pgx.Tx, uuid.UUID) (map[string]any, error) {
	return map[string]any{}, nil
}

func (compositionOperations) LoadDecisionTx(context.Context, pgx.Tx, uuid.UUID) (map[string]any, error) {
	return map[string]any{}, nil
}

func (compositionOperations) RecordSummaryLabelsTx(context.Context, pgx.Tx, uuid.UUID, []uuid.UUID) (map[uuid.UUID]string, error) {
	return map[uuid.UUID]string{}, nil
}

func (compositionOperations) CaptureRecordSnapshotTx(context.Context, pgx.Tx, uuid.UUID) (revisions.RecordSnapshot, error) {
	return revisions.RecordSnapshot{}, nil
}

func (compositionOperations) AppendChangeSetTx(context.Context, pgx.Tx, revisions.AppendChangeSetParams) (uuid.UUID, error) {
	return uuid.New(), nil
}

func (compositionOperations) AppendMutationTx(context.Context, pgx.Tx, revisions.AppendNonRowMutationParams) error {
	return nil
}

func (compositionOperations) AppendRecordMutationTx(context.Context, pgx.Tx, revisions.AppendRecordMutationParams) error {
	return nil
}

func (compositionOperations) AppendLiveRevisionTx(context.Context, pgx.Tx, revisions.LiveRevisionInput) error {
	return nil
}

func (compositionOperations) AppendRecordChangedTx(context.Context, pgx.Tx, collaboration.RecordChangeIntentInput) error {
	return nil
}

func (compositionOperations) LoadRevisionWindowTx(context.Context, pgx.Tx, uuid.UUID, int64, int64) ([]conflicttokens.RevisionWindowRow, error) {
	return nil, nil
}

func (compositionOperations) LockEnvelopeFactTx(context.Context, pgx.Tx, uuid.UUID, string, bool) (EnvelopeFact, bool, error) {
	return EnvelopeFact{}, false, errors.New("unexpected reference envelope lock")
}
func (compositionOperations) LockPartySourceTx(context.Context, pgx.Tx, uuid.UUID, uuid.UUID, bool) (bool, error) {
	return false, errors.New("unexpected reference Party lock")
}

// compositionImportRevisions satisfies the import revision capability for
// construction tests only; any call is unexpected.
type compositionImportRevisions struct{}

func (compositionImportRevisions) CaptureRecordSnapshotTx(context.Context, pgx.Tx, uuid.UUID) (revisions.RecordSnapshot, error) {
	return revisions.RecordSnapshot{}, errors.New("unexpected composition revision capture")
}

func (compositionImportRevisions) AppendRecordMutationTx(context.Context, pgx.Tx, revisions.AppendRecordMutationParams) error {
	return errors.New("unexpected composition revision append")
}

func (compositionImportRevisions) AppendLiveRevisionTx(context.Context, pgx.Tx, revisions.LiveRevisionInput) error {
	return errors.New("unexpected composition live revision")
}

func (compositionImportRevisions) AppendNonRowMutationTx(context.Context, pgx.Tx, revisions.AppendNonRowMutationParams) error {
	return errors.New("unexpected composition non-row mutation")
}
