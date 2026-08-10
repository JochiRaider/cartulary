package tasksdecisions

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"

	"github.com/JochiRaider/cartulary/internal/modules/links"
	"github.com/JochiRaider/cartulary/internal/modules/records"
	"github.com/JochiRaider/cartulary/internal/modules/revisions"
	conflicttokens "github.com/JochiRaider/cartulary/internal/modules/revisions/conflicts"
)

func TestMutationContributionRejectsIncompleteDependencies_Unit(t *testing.T) {
	if _, err := NewMutationContribution(nil, conflicttokens.ConflictTokenCodec{}, MutationDependencies{}); err == nil || !strings.Contains(err.Error(), "Postgres is required") {
		t.Fatalf("nil Postgres error = %v", err)
	}
	if _, err := NewMutationContribution(compositionDB{}, conflicttokens.ConflictTokenCodec{}, MutationDependencies{}); err == nil || !strings.Contains(err.Error(), "Incident admission is required") {
		t.Fatalf("incomplete dependency error = %v", err)
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
	_, err := NewImportContribution(
		TaskRequestsViewSchemaID,
		"tasksdecisions.task_request.import_create",
		ImportDependencies{},
	)
	if err == nil || !strings.Contains(err.Error(), "Records insert is required") {
		t.Fatalf("incomplete import dependency error = %v", err)
	}
}

type mutationCreateConsumer interface {
	Create(context.Context, WorkbookCreateCommand) (WorkbookMutationResult, error)
}

type mutationPatchConsumer interface {
	Patch(context.Context, WorkbookPatchCommand) (WorkbookMutationResult, error)
}

type mutationConflictConsumer interface {
	ResolveConflict(context.Context, WorkbookConflictCommand) (WorkbookMutationResult, error)
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
		MemberReferences:     operations,
		Idempotency:          compositionIdempotency{},
		RecordEnvelopes:      records.NewStore(),
		Links:                links.NewStore(),
		Projections:          operations,
		Revisions:            operations,
		ConflictFields:       fieldResolver,
		KeepSavedIdempotency: compositionKeepSavedIdempotency{},
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

func (compositionOperations) EnsureOpenTx(context.Context, pgx.Tx, uuid.UUID) error { return nil }

func (compositionOperations) ValidateIncidentMemberUserTx(context.Context, pgx.Tx, uuid.UUID, uuid.UUID, string) error {
	return nil
}

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

func (compositionOperations) RebuildTaskRequestsTx(context.Context, pgx.Tx, uuid.UUID) error {
	return nil
}

func (compositionOperations) RebuildDecisionsTx(context.Context, pgx.Tx, uuid.UUID) error { return nil }

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

func (compositionOperations) AppendRecordRevisionAndIntentTx(context.Context, pgx.Tx, revisions.AppendRecordRevisionParams) error {
	return nil
}

func (compositionOperations) LoadRevisionWindowTx(context.Context, pgx.Tx, uuid.UUID, int64, int64) ([]conflicttokens.RevisionWindowRow, error) {
	return nil, nil
}
