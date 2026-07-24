package extensionassembly

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/JochiRaider/cartulary/internal/modules/crossownertransaction"
	"github.com/JochiRaider/cartulary/internal/testutil/pgtest"
)

func TestCrossOwnerTransaction_Integration_CommitAndRollbackAtomicity(t *testing.T) {
	ctx := context.Background()
	pool := newCrossOwnerIntegrationPool(t, "cross-owner-atomicity")
	if _, err := pool.Exec(ctx, `CREATE TABLE cross_owner_protocol_test (value text PRIMARY KEY)`); err != nil {
		t.Fatal(err)
	}
	provider := &crossOwnerCapabilityProviderFake{}
	backend, err := NewCrossOwnerBackend(pool, provider)
	if err != nil {
		t.Fatal(err)
	}
	coordinator, err := crossownertransaction.New(crossownertransaction.Options{
		Backend: backend, Catalog: []crossownertransaction.Descriptor{crossOwnerTestDescriptor()},
		Timeout: time.Second, FatalSink: func(err error) { t.Fatalf("unexpected fatal: %v", err) },
	})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := coordinator.Execute(ctx, crossownertransaction.Operation{
		OperationID: "committed", NormalizedRequestSHA256: strings.Repeat("a", 64),
		Participants: []crossownertransaction.Participant{&crossOwnerParticipantFake{value: "committed"}},
	}); err != nil {
		t.Fatal(err)
	}
	if countCrossOwnerRows(t, pool) != 1 {
		t.Fatal("proven commit did not publish exactly one row")
	}
	_, err = coordinator.Execute(ctx, crossownertransaction.Operation{
		OperationID: "rolled-back", NormalizedRequestSHA256: strings.Repeat("b", 64),
		Participants: []crossownertransaction.Participant{&crossOwnerParticipantFake{value: "must-not-persist", failAfterWrite: true}},
	})
	if err == nil || countCrossOwnerRows(t, pool) != 1 {
		t.Fatalf("write fault = %v, rows=%d", err, countCrossOwnerRows(t, pool))
	}
}

func TestCrossOwnerTransaction_Integration_OrderedAdvisoryLocks(t *testing.T) {
	ctx := context.Background()
	pool := newCrossOwnerIntegrationPool(t, "cross-owner-locks")
	provider := &crossOwnerCapabilityProviderFake{}
	backend, err := NewCrossOwnerBackend(pool, provider)
	if err != nil {
		t.Fatal(err)
	}
	descriptors := []crossownertransaction.Descriptor{crossOwnerTestDescriptor()}
	first, err := backend.Begin(ctx, descriptors)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _, _ = first.Rollback(context.Background()) })
	key := crossownertransaction.OrderedSerializationKey{
		ParticipantID: "test.participant", KeyKind: "test.key", Key: "shared",
	}
	if err := first.AcquireSerializationLock(ctx, key); err != nil {
		t.Fatal(err)
	}
	second, err := backend.Begin(ctx, descriptors)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _, _ = second.Rollback(context.Background()) })
	waitCtx, cancel := context.WithTimeout(ctx, 50*time.Millisecond)
	defer cancel()
	if err := second.AcquireSerializationLock(waitCtx, key); !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("same serialization key did not block: %v", err)
	}
	third, err := backend.Begin(ctx, descriptors)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _, _ = third.Rollback(context.Background()) }()
	distinct := key
	distinct.Key = "distinct"
	if err := third.AcquireSerializationLock(ctx, distinct); err != nil {
		t.Fatalf("distinct serialization key blocked: %v", err)
	}
}

func newCrossOwnerIntegrationPool(t *testing.T, name string) *pgxpool.Pool {
	t.Helper()
	harness := pgtest.Start(t)
	testDB := harness.PrepareIsolatedDatabaseT(t, name)
	pool, err := pgxpool.New(context.Background(), testDB.DSN)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(pool.Close)
	return pool
}

func crossOwnerTestDescriptor() crossownertransaction.Descriptor {
	return crossownertransaction.Descriptor{
		ParticipantID: "test.participant", OwnerProfileID: "test", ContractSHA256: "digest",
		InputSchemaID: "test.input", PrepareAlgorithmID: "test.prepare",
		ValidationAlgorithmID: "test.validate", WriteAlgorithmID: "test.write",
		SerializationKeyKinds: []string{"test.key"}, OwnedStateFamilyIDs: []string{"test.rows"},
	}
}

type crossOwnerCapabilityProviderFake struct{}

func (*crossOwnerCapabilityProviderFake) TransactionCapabilities(participantID string, tx pgx.Tx) (crossownertransaction.ReadCapability, crossownertransaction.WriteCapability, error) {
	capability := &crossOwnerCapabilityFake{id: participantID, tx: tx}
	return capability, capability, nil
}

type crossOwnerCapabilityFake struct {
	id string
	tx pgx.Tx
}

func (c *crossOwnerCapabilityFake) ParticipantScope() string { return c.id }
func (c *crossOwnerCapabilityFake) insert(ctx context.Context, value string) error {
	_, err := c.tx.Exec(ctx, `INSERT INTO cross_owner_protocol_test (value) VALUES ($1)`, value)
	return err
}

type crossOwnerParticipantFake struct {
	value          string
	failAfterWrite bool
}

func (*crossOwnerParticipantFake) ID() string { return "test.participant" }
func (p *crossOwnerParticipantFake) BuildInput(context.Context, crossownertransaction.OperationContext) (crossownertransaction.Input, error) {
	return crossownertransaction.Input{SchemaID: "test.input", CanonicalBytes: []byte(`{"value":"` + p.value + `"}`)}, nil
}
func (*crossOwnerParticipantFake) Prepare(context.Context, crossownertransaction.Invocation) (crossownertransaction.PrepareResult, error) {
	return crossownertransaction.PrepareResult{SerializationKeys: []crossownertransaction.SerializationKey{{KeyKind: "test.key", Key: "shared"}}}, nil
}
func (*crossOwnerParticipantFake) Validate(context.Context, crossownertransaction.Invocation) (crossownertransaction.ValidationResult, error) {
	return crossownertransaction.Valid(), nil
}
func (p *crossOwnerParticipantFake) Write(ctx context.Context, invocation crossownertransaction.Invocation) (crossownertransaction.WriteResult, error) {
	capability, ok := invocation.WriteAccess.(*crossOwnerCapabilityFake)
	if !ok {
		return crossownertransaction.WriteResult{}, errors.New("write capability unavailable")
	}
	if err := capability.insert(ctx, p.value); err != nil {
		return crossownertransaction.WriteResult{}, err
	}
	if p.failAfterWrite {
		return crossownertransaction.WriteResult{}, errors.New("injected write failure")
	}
	return crossownertransaction.Written(p.value), nil
}

func countCrossOwnerRows(t *testing.T, pool *pgxpool.Pool) int {
	t.Helper()
	var count int
	if err := pool.QueryRow(context.Background(), `SELECT count(*) FROM cross_owner_protocol_test`).Scan(&count); err != nil {
		t.Fatal(err)
	}
	return count
}
