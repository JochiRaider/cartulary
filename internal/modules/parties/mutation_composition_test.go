package parties

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"

	"github.com/JochiRaider/cartulary/internal/modules/collaboration"
	"github.com/JochiRaider/cartulary/internal/modules/records"
	"github.com/JochiRaider/cartulary/internal/modules/revisions"
	conflicttokens "github.com/JochiRaider/cartulary/internal/modules/revisions/conflicts"
)

func TestMutationContributionRejectsIncompleteDependencies_Unit(t *testing.T) {
	if _, err := NewMutationContribution(nil, conflicttokens.ConflictTokenCodec{}, MutationDependencies{}); err == nil || !strings.Contains(err.Error(), "Postgres is required") {
		t.Fatalf("nil Postgres error = %v", err)
	}
	var typedNilPostgres *compositionDB
	if _, err := NewMutationContribution(typedNilPostgres, conflicttokens.ConflictTokenCodec{}, completeMutationDependencies()); err == nil || !strings.Contains(err.Error(), "Postgres is required") {
		t.Fatalf("typed-nil Postgres error = %v", err)
	}
	tests := []struct {
		name string
		want string
		drop func(*MutationDependencies)
	}{
		{name: "incident state", want: "Incident admission", drop: func(d *MutationDependencies) { d.IncidentState = nil }},
		{name: "idempotency", want: "Route idempotency", drop: func(d *MutationDependencies) { d.Idempotency = nil }},
		{name: "records", want: "Record envelopes", drop: func(d *MutationDependencies) { d.RecordEnvelopes = nil }},
		{name: "projections", want: "Projections", drop: func(d *MutationDependencies) { d.Projections = nil }},
		{name: "revisions", want: "Revisions/history", drop: func(d *MutationDependencies) { d.Revisions = nil }},
		{name: "conflict fields", want: "Conflict fields", drop: func(d *MutationDependencies) { d.ConflictFields = nil }},
		{name: "keep saved", want: "Keep-saved resolution", drop: func(d *MutationDependencies) { d.KeepSaved = nil }},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			dependencies := completeMutationDependencies()
			test.drop(&dependencies)
			if _, err := NewMutationContribution(compositionDB{}, conflicttokens.ConflictTokenCodec{}, dependencies); err == nil || !strings.Contains(err.Error(), test.want) {
				t.Fatalf("missing %s error = %v", test.name, err)
			}
		})
	}

	typedNilOperations := (*compositionOperations)(nil)
	typedNilIdempotency := (*compositionIdempotency)(nil)
	typedNilKeepSaved := (*compositionKeepSaved)(nil)
	var typedNilConflictFields *typedNilConflictFieldResolver
	typedNilTests := []struct {
		name string
		want string
		drop func(*MutationDependencies)
	}{
		{name: "incident state", want: "Incident admission", drop: func(d *MutationDependencies) { d.IncidentState = typedNilOperations }},
		{name: "idempotency", want: "Route idempotency", drop: func(d *MutationDependencies) { d.Idempotency = typedNilIdempotency }},
		{name: "records", want: "Record envelopes", drop: func(d *MutationDependencies) { d.RecordEnvelopes = typedNilOperations }},
		{name: "projections", want: "Projections", drop: func(d *MutationDependencies) { d.Projections = typedNilOperations }},
		{name: "revisions", want: "Revisions/history", drop: func(d *MutationDependencies) { d.Revisions = typedNilOperations }},
		{name: "conflict fields", want: "Conflict fields", drop: func(d *MutationDependencies) { d.ConflictFields = typedNilConflictFields }},
		{name: "keep saved", want: "Keep-saved resolution", drop: func(d *MutationDependencies) { d.KeepSaved = typedNilKeepSaved }},
	}
	for _, test := range typedNilTests {
		t.Run("typed nil "+test.name, func(t *testing.T) {
			dependencies := completeMutationDependencies()
			test.drop(&dependencies)
			if _, err := NewMutationContribution(compositionDB{}, conflicttokens.ConflictTokenCodec{}, dependencies); err == nil || !strings.Contains(err.Error(), test.want) {
				t.Fatalf("typed-nil %s error = %v", test.name, err)
			}
		})
	}
}

type typedNilConflictFieldResolver struct{ conflicttokens.FieldResolver }

func TestMutationContributionSharesOneFacade_Unit(t *testing.T) {
	facade, err := NewMutationContribution(
		compositionDB{},
		conflicttokens.ConflictTokenCodec{},
		completeMutationDependencies(),
	)
	if err != nil {
		t.Fatalf("construct mutation contribution: %v", err)
	}
	var create interface {
		Create(context.Context, CreateCommand) (MutationResult, error)
	} = facade
	var patch interface {
		Patch(context.Context, PatchCommand) (MutationResult, error)
	} = facade
	var conflict interface {
		ResolveConflict(context.Context, ConflictCommand) (MutationResult, error)
	} = facade
	if create != facade || patch != facade || conflict != facade {
		t.Fatal("mutation consumers did not retain the one facade instance")
	}
}

func TestRowVersionConsumptionRequiresExactInt64_Unit(t *testing.T) {
	for _, test := range []struct {
		name string
		row  map[string]any
	}{
		{name: "missing", row: map[string]any{}},
		{name: "int", row: map[string]any{"row_version": int(1)}},
		{name: "int32", row: map[string]any{"row_version": int32(1)}},
		{name: "float64", row: map[string]any{"row_version": float64(1)}},
		{name: "zero", row: map[string]any{"row_version": int64(0)}},
	} {
		t.Run(test.name, func(t *testing.T) {
			if _, err := rowVersionFromGenericRow(test.row); err == nil {
				t.Fatalf("row version %#v was admitted", test.row["row_version"])
			}
		})
	}
	if value, err := rowVersionFromGenericRow(map[string]any{"row_version": int64(7)}); err != nil || value != 7 {
		t.Fatalf("exact int64 row version = %d, err %v", value, err)
	}
}

func TestImportContributionRejectsIncompleteDependencies_Unit(t *testing.T) {
	if _, err := NewImportContribution("cartulary.view.unknown.v1", "parties.unknown", completeImportDependencies()); err == nil || !strings.Contains(err.Error(), "not mapped") {
		t.Fatalf("unknown import surface error = %v", err)
	}
	tests := []struct {
		name string
		want string
		drop func(*ImportDependencies)
	}{
		{name: "records", want: "Records insert", drop: func(d *ImportDependencies) { d.RecordEnvelopes = nil }},
		{name: "projections", want: "Projection refresh/load", drop: func(d *ImportDependencies) { d.Projections = nil }},
		{name: "revisions", want: "Revision finalization", drop: func(d *ImportDependencies) { d.Revisions = nil }},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			dependencies := completeImportDependencies()
			test.drop(&dependencies)
			if _, err := NewImportContribution(ViewSchemaID, "parties.create", dependencies); err == nil || !strings.Contains(err.Error(), test.want) {
				t.Fatalf("missing %s error = %v", test.name, err)
			}
		})
	}
	typedNilOperations := (*compositionOperations)(nil)
	typedNilTests := []struct {
		name string
		want string
		drop func(*ImportDependencies)
	}{
		{name: "records", want: "Records insert", drop: func(d *ImportDependencies) { d.RecordEnvelopes = typedNilOperations }},
		{name: "projections", want: "Projection refresh/load", drop: func(d *ImportDependencies) { d.Projections = typedNilOperations }},
		{name: "revisions", want: "Revision finalization", drop: func(d *ImportDependencies) { d.Revisions = typedNilOperations }},
	}
	for _, test := range typedNilTests {
		t.Run("typed nil "+test.name, func(t *testing.T) {
			dependencies := completeImportDependencies()
			test.drop(&dependencies)
			if _, err := NewImportContribution(ViewSchemaID, "parties.create", dependencies); err == nil || !strings.Contains(err.Error(), test.want) {
				t.Fatalf("typed-nil %s error = %v", test.name, err)
			}
		})
	}
}

func completeMutationDependencies() MutationDependencies {
	operations := compositionOperations{}
	resolver, err := conflicttokens.NewFieldResolverCatalog(nil)
	if err != nil {
		panic(err)
	}
	return MutationDependencies{
		IncidentState: operations, Idempotency: compositionIdempotency{},
		RecordEnvelopes: operations, Projections: operations, Revisions: operations,
		ConflictFields: resolver, KeepSaved: compositionKeepSaved{}, Collaboration: operations,
	}
}

func completeImportDependencies() ImportDependencies {
	operations := compositionOperations{}
	return ImportDependencies{RecordEnvelopes: operations, Projections: operations, Revisions: operations, Collaboration: operations}
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

func (compositionIdempotency) Get(context.Context, IdempotencyKey, []byte) (StoredMutationResult, bool, error) {
	return StoredMutationResult{}, false, nil
}
func (compositionIdempotency) PutTx(context.Context, pgx.Tx, IdempotencyKey, []byte, StoredMutationResult) error {
	return nil
}

type compositionKeepSaved struct{}

func (compositionKeepSaved) KeepSaved(
	context.Context,
	conflicttokens.TransactionRunner,
	conflicttokens.Command,
	conflicttokens.TargetLoader,
) (KeepSavedResult, error) {
	return KeepSavedResult{}, nil
}

type compositionOperations struct{}

func (compositionOperations) RequireOpenTx(context.Context, pgx.Tx, uuid.UUID) error { return nil }
func (compositionOperations) InsertTx(context.Context, pgx.Tx, records.InsertParams) (uuid.UUID, error) {
	return uuid.New(), nil
}
func (compositionOperations) AdvanceVersionTx(context.Context, pgx.Tx, uuid.UUID, uuid.UUID, time.Time) (int64, error) {
	return 1, nil
}
func (compositionOperations) LoadEnvelope(context.Context, uuid.UUID) (records.Envelope, error) {
	return records.Envelope{}, nil
}
func (compositionOperations) RefreshPartyTx(context.Context, pgx.Tx, uuid.UUID) error { return nil }
func (compositionOperations) LoadPartyTx(context.Context, pgx.Tx, uuid.UUID) (map[string]any, error) {
	return map[string]any{}, nil
}
func (compositionOperations) CaptureRecordSnapshotTx(context.Context, pgx.Tx, uuid.UUID) (revisions.RecordSnapshot, error) {
	return revisions.RecordSnapshot{}, nil
}
func (compositionOperations) AppendChangeSetTx(context.Context, pgx.Tx, revisions.AppendChangeSetParams) (uuid.UUID, error) {
	return uuid.New(), nil
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
