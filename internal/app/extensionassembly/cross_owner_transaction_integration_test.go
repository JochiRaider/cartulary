package extensionassembly

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/JochiRaider/cartulary/internal/modules/crossownertransaction"
	"github.com/JochiRaider/cartulary/internal/modules/extensions"
	"github.com/JochiRaider/cartulary/internal/modules/networkflow"
	"github.com/JochiRaider/cartulary/internal/platform/extensionstore"
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
	registry, registryErr := extensions.NewGeneratedCoordinator()
	if registryErr != nil {
		t.Fatal(registryErr)
	}
	assertAdmittedValidationPolicy(t, registry)
	assertPhysicalCompletionRetirement(t, pool, backend, registry)
	assertPrecommitCancellationRetirement(t, pool, backend, registry)
	assertNeverSentCommitRetirement(t, pool, backend, registry)
	assertDeadlineDuringPhysicalCommit(t, pool, backend, registry)
	coordinator, err := crossownertransaction.New(crossownertransaction.Options{
		Limits:           CrossOwnerLimits(registry),
		ValidationPolicy: crossOwnerTestValidationPolicy{CrossOwnerValidationPolicy{Coordinator: registry}},
		Backend:          backend, Catalog: []crossownertransaction.Descriptor{crossOwnerTestDescriptor()},
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
	oversized := crossownertransaction.ValidationResult{Status: "private", Findings: make([]crossownertransaction.Finding, 4097)}
	_, err = coordinator.Execute(ctx, crossownertransaction.Operation{OperationID: "overflow", NormalizedRequestSHA256: strings.Repeat("c", 64), Participants: []crossownertransaction.Participant{&crossOwnerParticipantFake{value: "must-not-write", validation: &oversized}}})
	var failure *crossownertransaction.ValidationError
	if !errors.As(err, &failure) || failure.ReasonCode != "extension_diagnostic_overflow" || len(failure.Findings) != 1 || countCrossOwnerRows(t, pool) != 1 {
		t.Fatalf("current overflow failure or rollback = %v", err)
	}
	var details map[string]any
	if err := json.Unmarshal(failure.Findings[0].Details, &details); err != nil || details["limit"] != float64(4096) || details["actual"] != float64(4097) || details["phase"] != "transaction" {
		t.Fatalf("current overflow details = %v", details)
	}

}

type retirementTestBackend struct {
	backend   *CrossOwnerBackend
	configure func(*crossOwnerTransaction)
}

func (b retirementTestBackend) Begin(ctx context.Context, descriptors []crossownertransaction.Descriptor) (crossownertransaction.Transaction, error) {
	transaction, err := b.backend.Begin(ctx, descriptors)
	if err == nil {
		b.configure(transaction.(*crossOwnerTransaction))
	}
	return transaction, err
}

func assertPhysicalCompletionRetirement(t *testing.T, pool *pgxpool.Pool, backend *CrossOwnerBackend, registry *extensions.Coordinator) {
	t.Helper()
	t.Run("completion panic before driver commit proves absence and retires the connection", func(t *testing.T) {
		var fatal atomic.Bool
		decorated := retirementTestBackend{backend: backend, configure: func(tx *crossOwnerTransaction) {
			registerPhysicalTransactionCleanup(t, tx.tx)
			tx.commitFinalized = func(context.Context) (extensionstore.CommitOutcome, error) { panic("private-completion-data") }
		}}
		coordinator, err := crossownertransaction.New(crossownertransaction.Options{
			Limits: CrossOwnerLimits(registry), ValidationPolicy: crossOwnerTestValidationPolicy{CrossOwnerValidationPolicy{Coordinator: registry}},
			Backend: decorated, Catalog: []crossownertransaction.Descriptor{crossOwnerTestDescriptor()},
			Timeout: time.Second, CancellationGrace: time.Second, FatalSink: func(error) { fatal.Store(true) },
		})
		if err != nil {
			t.Fatal(err)
		}
		_, err = coordinator.Execute(context.Background(), crossownertransaction.Operation{
			OperationID: "completion-panic", NormalizedRequestSHA256: strings.Repeat("d", 64),
			Participants: []crossownertransaction.Participant{&crossOwnerParticipantFake{value: "uncommitted-completion"}},
		})
		if err == nil || fatal.Load() || pool.Stat().AcquiredConns() != 0 {
			t.Fatalf("completion absence/retirement: err=%v fatal=%v acquired=%d", err, fatal.Load(), pool.Stat().AcquiredConns())
		}
		if strings.Contains(err.Error(), "private-completion-data") || countCrossOwnerRows(t, pool) != 0 {
			t.Fatal("completion failure leaked data or committed an effect")
		}
	})
	assertPhysicalCompletionOutcomes(t, pool, backend, registry)
}

type canceledWriteParticipant struct {
	*crossOwnerParticipantFake
	interrupt func(context.Context) error
}

func (p canceledWriteParticipant) Write(ctx context.Context, invocation crossownertransaction.Invocation) (crossownertransaction.WriteResult, error) {
	result, err := p.crossOwnerParticipantFake.Write(ctx, invocation)
	if err == nil {
		err = p.interrupt(ctx)
	}
	return result, err
}

func assertPrecommitCancellationRetirement(t *testing.T, pool *pgxpool.Pool, backend *CrossOwnerBackend, registry *extensions.Coordinator) {
	t.Helper()
	t.Run("canceled SQL before commit closes connection and proves absence", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		var physical *crossOwnerTransaction
		var fatal atomic.Bool
		decorated := retirementTestBackend{backend: backend, configure: func(tx *crossOwnerTransaction) {
			physical = tx
			registerPhysicalTransactionCleanup(t, tx.tx)
		}}
		coordinator, err := crossownertransaction.New(crossownertransaction.Options{
			Limits: CrossOwnerLimits(registry), ValidationPolicy: crossOwnerTestValidationPolicy{CrossOwnerValidationPolicy{Coordinator: registry}},
			Backend: decorated, Catalog: []crossownertransaction.Descriptor{crossOwnerTestDescriptor()},
			Timeout: 3 * time.Second, CancellationGrace: time.Second, FatalSink: func(error) { fatal.Store(true) },
		})
		if err != nil {
			t.Fatal(err)
		}
		participant := canceledWriteParticipant{crossOwnerParticipantFake: &crossOwnerParticipantFake{value: "canceled-before-commit"},
			interrupt: func(callCtx context.Context) error {
				timer := time.AfterFunc(100*time.Millisecond, cancel)
				defer timer.Stop()
				_, err := physical.tx.Exec(callCtx, `SELECT pg_sleep(10)`)
				return err
			}}
		_, err = coordinator.Execute(ctx, crossownertransaction.Operation{OperationID: "canceled-before-commit", NormalizedRequestSHA256: strings.Repeat("e", 64), Participants: []crossownertransaction.Participant{participant}})
		if !physical.tx.Conn().IsClosed() {
			t.Fatal("cancellation did not exercise driver connection closure")
		}
		// pgxpool destroys the retired resource asynchronously after the
		// driver's cleanup signal. Observe bounded pool-accounting completion.
		until := time.Now().Add(time.Second)
		for pool.Stat().AcquiredConns() != 0 && time.Now().Before(until) {
			time.Sleep(time.Millisecond)
		}
		if !errors.Is(err, crossownertransaction.ErrCanceled) || fatal.Load() || pool.Stat().AcquiredConns() != 0 || countCrossOwnerRows(t, pool) != 0 {
			t.Fatalf("precommit cancellation: err=%v fatal=%v acquired=%d", err, fatal.Load(), pool.Stat().AcquiredConns())
		}
	})
}

// A session that is already closed when COMMIT is due makes pgx reject the
// request before sending any of it. The never-sent request on a retired
// session is proven absent and classified, not fatal (EXT-REQ-215).
func assertNeverSentCommitRetirement(t *testing.T, pool *pgxpool.Pool, backend *CrossOwnerBackend, registry *extensions.Coordinator) {
	t.Helper()
	t.Run("commit never sent on a closed session is proven absent", func(t *testing.T) {
		ctx := context.Background()
		var physical *crossOwnerTransaction
		var fatal atomic.Bool
		decorated := retirementTestBackend{backend: backend, configure: func(tx *crossOwnerTransaction) {
			physical = tx
			registerPhysicalTransactionCleanup(t, tx.tx)
		}}
		coordinator, err := crossownertransaction.New(crossownertransaction.Options{
			Limits: CrossOwnerLimits(registry), ValidationPolicy: crossOwnerTestValidationPolicy{CrossOwnerValidationPolicy{Coordinator: registry}},
			Backend: decorated, Catalog: []crossownertransaction.Descriptor{crossOwnerTestDescriptor()},
			Timeout: time.Minute, CancellationGrace: 5 * time.Second, FatalSink: func(error) { fatal.Store(true) },
		})
		if err != nil {
			t.Fatal(err)
		}
		participant := canceledWriteParticipant{crossOwnerParticipantFake: &crossOwnerParticipantFake{value: "never-sent-commit"},
			interrupt: func(context.Context) error {
				// The participant reports success, but its session is gone.
				return physical.tx.Conn().Close(ctx)
			}}
		_, err = coordinator.Execute(ctx, crossownertransaction.Operation{OperationID: "never-sent-commit", NormalizedRequestSHA256: strings.Repeat("8", 64), Participants: []crossownertransaction.Participant{participant}})
		if err == nil || crossownertransaction.IsFatalIntegrity(err) || fatal.Load() {
			t.Fatalf("never-sent commit: err=%v fatal=%v", err, fatal.Load())
		}
		if !physical.tx.Conn().IsClosed() {
			t.Fatal("never-sent commit left its session open")
		}
		until := time.Now().Add(time.Second)
		for pool.Stat().AcquiredConns() != 0 && time.Now().Before(until) {
			time.Sleep(time.Millisecond)
		}
		if pool.Stat().AcquiredConns() != 0 || countCrossOwnerRows(t, pool) != 0 {
			t.Fatalf("never-sent commit: acquired=%d", pool.Stat().AcquiredConns())
		}
	})
}

type rollbackReplyLostTx struct{ pgx.Tx }

func (tx rollbackReplyLostTx) Rollback(ctx context.Context) error {
	if err := tx.Tx.Rollback(ctx); err != nil {
		return err
	}
	return errors.New("rollback reply unavailable")
}

func assertPhysicalCompletionOutcomes(t *testing.T, pool *pgxpool.Pool, backend *CrossOwnerBackend, registry *extensions.Coordinator) {
	t.Helper()
	cases := []struct {
		name, mode                       string
		proof, committed, fatal, success bool
	}{
		{name: "early_return", mode: "early", proof: true},
		{name: "panic_after_commit_without_proof", mode: "panic", committed: true, fatal: true},
		{name: "panic_after_commit_with_proof", mode: "panic", proof: true, committed: true, success: true},
		{name: "lost_reply_without_proof", mode: "reply", committed: true, fatal: true},
		{name: "lost_reply_with_proof", mode: "reply", proof: true, committed: true, success: true},
		{name: "lost_rollback_reply", mode: "rollback", fatal: true},
		{name: "committed_before_cancellation", mode: "cancel", committed: true, success: true},
	}
	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			value := "completion-" + test.name
			var fatal atomic.Bool
			var begins, proofs atomic.Int32
			decorated := retirementTestBackend{backend: backend, configure: func(tx *crossOwnerTransaction) {
				begins.Add(1)
				if test.mode == "rollback" {
					tx.tx = rollbackReplyLostTx{tx.tx}
				}
				registerPhysicalTransactionCleanup(t, tx.tx)
				if test.proof {
					tx.proveFinalizedCommit = func(proofCtx context.Context) bool {
						proofs.Add(1)
						var present bool
						err := pool.QueryRow(context.WithoutCancel(proofCtx), `SELECT EXISTS (SELECT 1 FROM cross_owner_protocol_test WHERE value=$1)`, value).Scan(&present)
						return err == nil && present
					}
				}
				tx.commitFinalized = func(callCtx context.Context) (extensionstore.CommitOutcome, error) {
					if test.mode == "early" || test.mode == "rollback" {
						return extensionstore.CommitUnknown, errors.New("completion stopped before driver commit")
					}
					if err := tx.tx.Commit(callCtx); err != nil {
						return extensionstore.CommitUnknown, err
					}
					switch test.mode {
					case "panic":
						panic("private-completion-data")
					case "cancel":
						cancel()
						return extensionstore.CommitProven, nil
					default:
						return extensionstore.CommitUnknown, errors.New("commit reply unavailable")
					}
				}
			}}
			t.Cleanup(func() {
				cleanupCtx, stop := context.WithTimeout(context.Background(), time.Second)
				defer stop()
				if _, err := pool.Exec(cleanupCtx, `DELETE FROM cross_owner_protocol_test WHERE value=$1`, value); err != nil {
					t.Error(err)
				}
			})
			coordinator, err := crossownertransaction.New(crossownertransaction.Options{
				Limits: CrossOwnerLimits(registry), ValidationPolicy: crossOwnerTestValidationPolicy{CrossOwnerValidationPolicy{Coordinator: registry}},
				Backend: decorated, Catalog: []crossownertransaction.Descriptor{crossOwnerTestDescriptor()},
				Timeout: time.Second, CancellationGrace: time.Second, FatalSink: func(error) { fatal.Store(true) },
			})
			if err != nil {
				t.Fatal(err)
			}
			_, err = coordinator.Execute(ctx, crossownertransaction.Operation{
				OperationID: value, NormalizedRequestSHA256: strings.Repeat("e", 64),
				Participants: []crossownertransaction.Participant{&crossOwnerParticipantFake{value: value}},
			})
			if (err == nil) != test.success || fatal.Load() != test.fatal || crossownertransaction.IsFatalIntegrity(err) != test.fatal || begins.Load() != 1 || pool.Stat().AcquiredConns() != 0 {
				t.Fatalf("completion: err=%v fatal=%v begins=%d acquired=%d", err, fatal.Load(), begins.Load(), pool.Stat().AcquiredConns())
			}
			if err != nil && strings.Contains(err.Error(), "private-completion-data") {
				t.Fatal("panic value escaped")
			}
			wantRows := 0
			if test.committed {
				wantRows = 1
			}
			if countCrossOwnerRows(t, pool) != wantRows {
				t.Fatal("physical effects disagree with closed outcome")
			}
			wantProofs := int32(0)
			if test.proof && test.committed {
				wantProofs = 1
			}
			if proofs.Load() != wantProofs {
				t.Fatalf("proof reads=%d want=%d", proofs.Load(), wantProofs)
			}
		})
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
	registerProtocolTransactionCleanup(t, first)
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
	registerProtocolTransactionCleanup(t, second)
	waitCtx, cancel := context.WithTimeout(ctx, 50*time.Millisecond)
	defer cancel()
	if err := second.AcquireSerializationLock(waitCtx, key); !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("same serialization key did not block: %v", err)
	}
	third, err := backend.Begin(ctx, descriptors)
	if err != nil {
		t.Fatal(err)
	}
	registerProtocolTransactionCleanup(t, third)
	distinct := key
	distinct.Key = "distinct"
	if err := third.AcquireSerializationLock(ctx, distinct); err != nil {
		t.Fatalf("distinct serialization key blocked: %v", err)
	}
}

// Register immediately after acquisition: failed product assertions must not
// leave the pool waiting for an owned handle. Closed physical handles are an
// expected result when the coordinator already completed retirement.
func registerPhysicalTransactionCleanup(t *testing.T, tx pgx.Tx) {
	t.Helper()
	t.Cleanup(func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := tx.Rollback(ctx); err != nil && !errors.Is(err, pgx.ErrTxClosed) {
			t.Errorf("physical transaction cleanup: %v", err)
		}
	})
}

func registerProtocolTransactionCleanup(t *testing.T, tx crossownertransaction.Transaction) {
	t.Helper()
	t.Cleanup(func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		outcome, err := tx.Rollback(ctx)
		if err != nil || outcome != crossownertransaction.CommitAbsent {
			t.Errorf("protocol transaction cleanup: outcome=%s err=%v", outcome, err)
		}
	})
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
	validation     *crossownertransaction.ValidationResult
}

func (*crossOwnerParticipantFake) ID() string { return "test.participant" }
func (p *crossOwnerParticipantFake) BuildInput(context.Context, crossownertransaction.OperationContext) (crossownertransaction.Input, error) {
	return crossownertransaction.Input{SchemaID: "test.input", CanonicalBytes: []byte(`{"value":"` + p.value + "\"}\n")}, nil
}
func (*crossOwnerParticipantFake) Prepare(context.Context, crossownertransaction.Invocation) (crossownertransaction.PrepareResult, error) {
	return crossownertransaction.PrepareResult{SerializationKeys: []crossownertransaction.SerializationKey{{KeyKind: "test.key", Key: "shared"}}}, nil
}
func (p *crossOwnerParticipantFake) Validate(context.Context, crossownertransaction.Invocation) (crossownertransaction.ValidationResult, error) {
	if p.validation != nil {
		return *p.validation, nil
	}
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

func assertAdmittedValidationPolicy(t *testing.T, registry *extensions.Coordinator) {
	t.Helper()
	policy := CrossOwnerValidationPolicy{Coordinator: registry}
	// Only descriptor binding is inspected here; owner input behavior is tested
	// with a configured Network Flow module and the real producer.
	bound := CrossOwnerValidationPolicy{Coordinator: registry, NetworkFlow: &networkflow.Module{}}
	for _, descriptor := range CrossOwnerDescriptors(registry.ParticipantContracts()) {
		if !bound.AcceptDescriptor(descriptor) {
			t.Fatalf("current descriptor not bound: %s", descriptor.ParticipantID)
		}
		for _, field := range []string{"digest", "schema", "algorithm", "families", "retired input", "retired algorithm"} {
			bad := descriptor
			switch field {
			case "digest":
				bad.ContractSHA256 = strings.Repeat("f", 64)
			case "schema":
				bad.InputSchemaID = "unresolved"
			case "algorithm":
				bad.WriteAlgorithmID = "unresolved"
			case "retired input":
				bad.InputSchemaID = strings.TrimSuffix(bad.InputSchemaID, ".v2") + ".v1"
				bad.ContractSHA256 = crossownertransaction.DescriptorDigest(bad)
			case "retired algorithm":
				bad.WriteAlgorithmID = strings.TrimSuffix(bad.WriteAlgorithmID, "_v2") + "_v1"
				bad.ContractSHA256 = crossownertransaction.DescriptorDigest(bad)
			case "families":
				bad.OwnedStateFamilyIDs = []string{"unowned"}
			}
			if bound.AcceptDescriptor(bad) {
				t.Fatalf("mismatching %s admitted", field)
			}
		}
		if !policy.AcceptDescriptor(descriptor) {
			t.Fatal("static typed owner contract required an operational module")
		}
		if descriptor.OwnerProfileID == networkflow.ProfileID && policy.AcceptInput(descriptor, nil) {
			t.Fatal("inactive owner input was admitted")
		}
	}
	var owner crossownertransaction.Descriptor
	for _, d := range CrossOwnerDescriptors(registry.ParticipantContracts()) {
		if d.OwnerProfileID == "network_flow_activity" {
			owner = d
			break
		}
	}
	if owner.ParticipantID == "" {
		t.Fatal("current Network Flow participant absent")
	}
	details := fmt.Sprintf(`{"algorithm_id":%q,"phase":"transaction","profile_id":%q,"timed_out":false,"timeout_seconds":60}`, owner.ValidationAlgorithmID, owner.OwnerProfileID)
	finding := crossownertransaction.Finding{Path: "$", ReasonCode: "extension_admission_validation_failed", Message: "Extension admission validation failed.", Details: []byte(details)}
	if !policy.AcceptFinding(owner, finding) {
		t.Fatal("registered participant finding rejected")
	}
	for _, mutation := range []string{"message", "extra", "identity", "duplicate", "reason", "path", "private path"} {
		bad := finding
		switch mutation {
		case "message":
			bad.Message = "private bytes"
		case "extra":
			bad.Details = []byte(strings.TrimSuffix(details, "}") + `,"private":"bytes"}`)
		case "identity":
			bad.Details = []byte(strings.Replace(details, owner.OwnerProfileID, "other_profile", 1))
		case "duplicate":
			bad.Details = []byte(strings.TrimSuffix(details, "}") + `,"phase":"transaction"}`)
		case "reason":
			bad.ReasonCode = "private_invented_reason"
		case "path":
			bad.Path = "/private/storage"
		case "private path":
			bad.Path = `$["/private/storage"]`
		}
		if policy.AcceptFinding(owner, bad) {
			t.Fatalf("admitted malformed %s finding", mutation)
		}
	}
	if finding, ok := policy.FailureFinding("extension_validation_result_overflow_transaction", owner, 60, false); !ok || finding.ReasonCode != "extension_diagnostic_overflow" {
		t.Fatal("transaction overflow condition absent")
	}
	if _, ok := policy.FailureFinding("invented", owner, 60, false); ok {
		t.Fatal("invented condition admitted")
	}
}

// The real PostgreSQL protocol test supplies its own closed synthetic input
// contract; production owner resolvers never admit this fixture participant.
type crossOwnerTestValidationPolicy struct{ CrossOwnerValidationPolicy }

func (crossOwnerTestValidationPolicy) AcceptDescriptor(d crossownertransaction.Descriptor) bool {
	return crossownertransaction.SameDescriptor(d, crossOwnerTestDescriptor())
}
func (crossOwnerTestValidationPolicy) AcceptInput(_ crossownertransaction.Descriptor, value any) bool {
	object, ok := value.(map[string]any)
	if !ok || len(object) != 1 {
		return false
	}
	_, ok = object["value"].(string)
	return ok
}

// commitWaitLockKey is the advisory lock a test-database deferred constraint
// trigger takes while PostgreSQL executes COMMIT. Holding it keeps COMMIT in
// pre-commit work, so the operation deadline expires with COMMIT in flight.
const commitWaitLockKey = 1106

func assertDeadlineDuringPhysicalCommit(t *testing.T, pool *pgxpool.Pool, backend *CrossOwnerBackend, registry *extensions.Coordinator) {
	t.Helper()
	ctx := context.Background()
	for _, statement := range []string{
		`CREATE FUNCTION cross_owner_commit_wait() RETURNS trigger LANGUAGE plpgsql AS $$
BEGIN
    PERFORM pg_advisory_xact_lock(1106);
    IF NEW.value LIKE '%-abort' THEN
        RAISE EXCEPTION 'injected commit-time abort' USING ERRCODE = '23514';
    ELSIF NEW.value LIKE '%-raise' THEN
        RAISE EXCEPTION 'injected deferred trigger failure' USING ERRCODE = 'P0001';
    ELSIF NEW.value LIKE '%-overflow' THEN
        RAISE EXCEPTION 'injected commit-time allocator exhaustion' USING ERRCODE = '22003';
    END IF;
    RETURN NULL;
END $$`,
		`CREATE CONSTRAINT TRIGGER cross_owner_commit_wait AFTER INSERT ON cross_owner_protocol_test
DEFERRABLE INITIALLY DEFERRED FOR EACH ROW WHEN (NEW.value LIKE 'commit-wait-%')
EXECUTE FUNCTION cross_owner_commit_wait()`,
	} {
		if _, err := pool.Exec(ctx, statement); err != nil {
			t.Fatal(err)
		}
	}
	t.Cleanup(func() {
		cleanupCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if _, err := pool.Exec(cleanupCtx, `DROP TRIGGER cross_owner_commit_wait ON cross_owner_protocol_test; DROP FUNCTION cross_owner_commit_wait()`); err != nil {
			t.Error(err)
		}
	})
	cases := []struct {
		name             string
		grace            time.Duration
		success, timeout bool
		fatalReason      string
		rows             int
	}{
		{name: "ok", grace: 5 * time.Second, success: true, rows: 1},
		{name: "abort", grace: 5 * time.Second, timeout: true},
		// The late commit lands after grace exhaustion: the operation stays
		// fatal, which is why a pending commit cannot report an ordinary outcome.
		{name: "grace", grace: 200 * time.Millisecond, fatalReason: "operation_cancellation_grace_exhausted", rows: 1},
	}
	for _, test := range cases {
		t.Run("deadline_during_physical_commit_"+test.name, func(t *testing.T) {
			value := "commit-wait-" + test.name
			t.Cleanup(func() {
				cleanupCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
				defer cancel()
				if _, err := pool.Exec(cleanupCtx, `DELETE FROM cross_owner_protocol_test WHERE value=$1`, value); err != nil {
					t.Error(err)
				}
			})
			holder, err := pool.Acquire(ctx)
			if err != nil {
				t.Fatal(err)
			}
			released := false
			release := func() {
				if released {
					return
				}
				released = true
				if _, err := holder.Exec(ctx, `SELECT pg_advisory_unlock($1)`, int64(commitWaitLockKey)); err != nil {
					t.Error(err)
				}
				holder.Release()
			}
			defer release()
			if _, err := holder.Exec(ctx, `SELECT pg_advisory_lock($1)`, int64(commitWaitLockKey)); err != nil {
				t.Fatal(err)
			}
			fatal := make(chan error, 2)
			decorated := retirementTestBackend{backend: backend, configure: func(tx *crossOwnerTransaction) {
				registerPhysicalTransactionCleanup(t, tx.tx)
			}}
			coordinator, err := crossownertransaction.New(crossownertransaction.Options{
				Limits: CrossOwnerLimits(registry), ValidationPolicy: crossOwnerTestValidationPolicy{CrossOwnerValidationPolicy{Coordinator: registry}},
				Backend: decorated, Catalog: []crossownertransaction.Descriptor{crossOwnerTestDescriptor()},
				Timeout: time.Minute, CancellationGrace: test.grace, FatalSink: func(err error) { fatal <- err },
			})
			if err != nil {
				t.Fatal(err)
			}
			parent, cancel := context.WithTimeout(ctx, 2*time.Second)
			defer cancel()
			done := make(chan error, 1)
			go func() {
				_, err := coordinator.Execute(parent, crossownertransaction.Operation{
					OperationID: value, NormalizedRequestSHA256: strings.Repeat("f", 64),
					Participants: []crossownertransaction.Participant{&crossOwnerParticipantFake{value: value}},
				})
				done <- err
			}()
			awaitCommitLockWait(t, pool, parent, done)
			<-parent.Done()
			if test.fatalReason != "" {
				select {
				case signal := <-fatal:
					fatal <- signal
				case <-time.After(5 * time.Second):
					t.Fatal("pending commit outlived the cancellation grace without a fatal signal")
				}
			}
			release()
			var result error
			select {
			case result = <-done:
			case <-time.After(10 * time.Second):
				t.Fatal("operation did not settle after the commit lock was released")
			}
			var signals []error
			for len(fatal) > 0 {
				signals = append(signals, <-fatal)
			}
			switch {
			case test.success:
				if result != nil || len(signals) != 0 {
					t.Fatalf("commit acknowledged after the deadline: err=%v fatal=%v", result, signals)
				}
			case test.timeout:
				if !errors.Is(result, crossownertransaction.ErrTimeout) || len(signals) != 0 {
					t.Fatalf("commit proven absent after the deadline: err=%v fatal=%v", result, signals)
				}
			default:
				var failure *crossownertransaction.FatalIntegrityError
				if !errors.As(result, &failure) || failure.FatalReasonCode() != test.fatalReason || len(signals) != 1 {
					t.Fatalf("pending commit past grace: err=%v fatal=%v", result, signals)
				}
				if !errors.As(signals[0], &failure) || failure.FatalReasonCode() != test.fatalReason {
					t.Fatalf("fatal signal = %v", signals[0])
				}
			}
			var rows int
			if err := pool.QueryRow(ctx, `SELECT count(*) FROM cross_owner_protocol_test WHERE value=$1`, value).Scan(&rows); err != nil {
				t.Fatal(err)
			}
			if rows != test.rows {
				t.Fatalf("committed rows=%d want=%d", rows, test.rows)
			}
			until := time.Now().Add(time.Second)
			for pool.Stat().AcquiredConns() != 0 && time.Now().Before(until) {
				time.Sleep(time.Millisecond)
			}
			if acquired := pool.Stat().AcquiredConns(); acquired != 0 {
				t.Fatalf("transaction connection not returned: acquired=%d", acquired)
			}
		})
	}
	// A definite ERROR acknowledgement to COMMIT proves the commit absent. It is
	// classified as the database failure, not as an indeterminate fatal.
	for _, test := range []struct{ name, code string }{
		{name: "raise", code: "P0001"},
		{name: "overflow", code: "22003"},
		{name: "cancel", code: "57014"},
	} {
		t.Run("definite_commit_error_"+test.code, func(t *testing.T) {
			value := "commit-wait-error-" + test.name
			t.Cleanup(func() {
				cleanupCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
				defer cancel()
				if _, err := pool.Exec(cleanupCtx, `DELETE FROM cross_owner_protocol_test WHERE value=$1`, value); err != nil {
					t.Error(err)
				}
			})
			var fatal atomic.Bool
			decorated := retirementTestBackend{backend: backend, configure: func(tx *crossOwnerTransaction) {
				registerPhysicalTransactionCleanup(t, tx.tx)
			}}
			coordinator, err := crossownertransaction.New(crossownertransaction.Options{
				Limits: CrossOwnerLimits(registry), ValidationPolicy: crossOwnerTestValidationPolicy{CrossOwnerValidationPolicy{Coordinator: registry}},
				Backend: decorated, Catalog: []crossownertransaction.Descriptor{crossOwnerTestDescriptor()},
				Timeout: time.Minute, CancellationGrace: 5 * time.Second, FatalSink: func(error) { fatal.Store(true) },
			})
			if err != nil {
				t.Fatal(err)
			}
			operation := crossownertransaction.Operation{
				OperationID: value, NormalizedRequestSHA256: strings.Repeat("9", 64),
				Participants: []crossownertransaction.Participant{&crossOwnerParticipantFake{value: value}},
			}
			if test.code == "57014" {
				// A query cancel reaches COMMIT while its deferred trigger waits.
				err = cancelCommitInLockWait(t, pool, func(parent context.Context) error {
					_, err := coordinator.Execute(parent, operation)
					return err
				})
			} else {
				_, err = coordinator.Execute(ctx, operation)
			}
			var response *pgconn.PgError
			if !errors.As(err, &response) || response.Code != test.code || crossownertransaction.IsFatalIntegrity(err) || fatal.Load() {
				t.Fatalf("definite commit error %s: err=%v fatal=%v", test.code, err, fatal.Load())
			}
			var rows int
			if err := pool.QueryRow(ctx, `SELECT count(*) FROM cross_owner_protocol_test WHERE value=$1`, value).Scan(&rows); err != nil || rows != 0 {
				t.Fatalf("absent commit left rows=%d (%v)", rows, err)
			}
			if acquired := pool.Stat().AcquiredConns(); acquired != 0 {
				t.Fatalf("transaction connection not returned: acquired=%d", acquired)
			}
		})
	}
}

// cancelCommitInLockWait holds the deferred trigger's lock, waits for COMMIT
// to block on it, and sends that backend a query cancel (pg_cancel_backend).
func cancelCommitInLockWait(t *testing.T, pool *pgxpool.Pool, execute func(context.Context) error) error {
	t.Helper()
	ctx := context.Background()
	holder, err := pool.Acquire(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer holder.Release()
	if _, err := holder.Exec(ctx, `SELECT pg_advisory_lock($1)`, int64(commitWaitLockKey)); err != nil {
		t.Fatal(err)
	}
	defer func() {
		if _, err := holder.Exec(ctx, `SELECT pg_advisory_unlock($1)`, int64(commitWaitLockKey)); err != nil {
			t.Error(err)
		}
	}()
	parent, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	done := make(chan error, 1)
	go func() { done <- execute(parent) }()
	awaitCommitLockWait(t, pool, parent, done)
	var canceled bool
	if err := pool.QueryRow(ctx, `
SELECT bool_and(pg_cancel_backend(pid))
  FROM pg_locks
 WHERE locktype = 'advisory' AND NOT granted AND classid = 0 AND objid = $1 AND objsubid = 1`, int64(commitWaitLockKey)).Scan(&canceled); err != nil || !canceled {
		t.Fatalf("cancel waiting COMMIT: canceled=%v err=%v", canceled, err)
	}
	select {
	case err := <-done:
		return err
	case <-time.After(10 * time.Second):
		t.Fatal("cancelled COMMIT did not settle")
		return nil
	}
}

// awaitCommitLockWait observes COMMIT blocked in the deferred trigger's lock
// wait. The deadline must not expire first, or the case would not exercise an
// issued commit.
func awaitCommitLockWait(t *testing.T, pool *pgxpool.Pool, parent context.Context, done <-chan error) {
	t.Helper()
	ticker := time.NewTicker(5 * time.Millisecond)
	defer ticker.Stop()
	for {
		var waiting bool
		if err := pool.QueryRow(context.Background(), `
SELECT EXISTS (
    SELECT 1
      FROM pg_locks l
      JOIN pg_stat_activity a ON a.pid = l.pid
     WHERE l.locktype = 'advisory' AND NOT l.granted
       AND l.classid = 0 AND l.objid = $1 AND l.objsubid = 1
       AND lower(a.query) = 'commit')`, int64(commitWaitLockKey)).Scan(&waiting); err != nil {
			t.Fatal(err)
		}
		if waiting {
			return
		}
		select {
		case err := <-done:
			t.Fatalf("operation settled before COMMIT reached the lock wait: %v", err)
		case <-parent.Done():
			t.Fatal("deadline expired before COMMIT reached the lock wait")
		case <-ticker.C:
		}
	}
}
