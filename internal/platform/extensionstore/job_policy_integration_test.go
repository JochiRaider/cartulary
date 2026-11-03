package extensionstore

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/JochiRaider/cartulary/internal/platform/authn"
	"github.com/JochiRaider/cartulary/internal/platform/jobs"
	"github.com/JochiRaider/cartulary/internal/testutil/collaborationsupport"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

func checkJobPolicyLifecycle(t *testing.T, pool *pgxpool.Pool, now time.Time) {
	for _, proofRequired := range []bool{true, false} {
		for _, idempotent := range []bool{true, false} {
			for _, local := range []bool{false, true} {
				t.Run(fmt.Sprintf("proof=%t_idempotent=%t_local=%t", proofRequired, idempotent, local), func(t *testing.T) {
					ctx := context.Background()
					definitions := collaborationsupport.TestJobDefinitions()
					policy := definitions[1].Extension
					policy.ProofRequired = proofRequired
					policy.CancellationPolicy = "not_cancelable"
					if local && idempotent {
						policy.IdentitySchemaID = jobs.AttributedRouteIdentitySchema
					}
					if !idempotent {
						policy.IdempotencyPolicy = "none"
						policy.IdentitySchemaID = ""
					}
					catalog, err := jobs.NewCatalog(definitions)
					if err != nil {
						t.Fatal(err)
					}
					transactions := collaborationsupport.NewJobTransactionsForCatalog(catalog)
					manager, err := jobs.NewManager(jobs.ManagerOptions{Postgres: pool, Transactions: transactions, Catalog: catalog, Policy: jobs.ProductionRuntimePolicy(), Now: func() time.Time { return now }})
					if err != nil {
						t.Fatal(err)
					}
					store, err := New(pool, nil)
					if err != nil {
						t.Fatal(err)
					}
					finalizer, err := NewOwnerFinalizer(store, transactions, collaborationsupport.NewJobOwnerTransactionAdapters(), func() time.Time { return now }, func(err error) { t.Errorf("unexpected indeterminate commit: %v", err) })
					if err != nil {
						t.Fatal(err)
					}
					actor := uuid.New()
					clientTxn := uuid.NewString()
					if _, err := pool.Exec(ctx, `INSERT INTO users (id,email,display_name,password_hash,mfa_required,is_active,is_deployment_admin) VALUES ($1,$2,'Policy test','hash',false,true,true)`, actor, actor.String()+"@example.test"); err != nil {
						t.Fatal(err)
					}
					submitter, operation := actor, uuid.Nil
					if local {
						submitter = uuid.Nil
						operation = uuid.New()
						clientTxn = operation.String()
					}
					key := jobs.NewRouteIdempotencyKey("test.policy", submitter, "deployment", clientTxn)
					admission, err := jobs.NewNonIdempotentExtensionJobAdmission("test_profile", []byte(`{}`))
					if idempotent {
						if local {
							admission, err = jobs.NewAttributedExtensionJobAdmission("test_profile", key, jobs.Scope{Kind: jobs.ScopeKindDeployment}, operation, []byte(`{}`))
						} else {
							admission, err = jobs.NewExtensionJobAdmission("test_profile", key, jobs.Scope{Kind: jobs.ScopeKindDeployment}, []byte(`{}`))
						}
					}
					if err != nil {
						t.Fatal(err)
					}
					tx, err := pool.Begin(ctx)
					if err != nil {
						t.Fatal(err)
					}
					defer func() { _ = tx.Rollback(ctx) }()
					queued, err := transactions.CreateQueuedTx(ctx, tx, jobs.EnqueueParams{JobKind: definitions[1].JobKind, Scope: jobs.Scope{Kind: jobs.ScopeKindDeployment}, SubmittedByUserID: submitter, OperatorOperationID: operation, AuthPolicy: jobs.AuthPolicyDeploymentAdmin, Extension: admission, Cancelable: false, Progress: jobs.Progress{Total: intPointer(1)}}, now)
					if err != nil {
						t.Fatal(err)
					}
					if idempotent && !local {
						digest, _ := hex.DecodeString(admission.NormalizedRequestSHA256)
						if err := authn.InsertRouteIdempotencyPayload(ctx, tx, authn.RouteIdempotencyKey{RouteKey: key.RouteKey, ActorUserID: actor, ScopeKey: key.ScopeKey, ClientTxnID: clientTxn}, nil, digest, 202, queued); err != nil {
							t.Fatal(err)
						}
					}
					if err := tx.Commit(ctx); err != nil {
						t.Fatal(err)
					}
					id := uuid.MustParse(queued.JobID)
					canceled, err := manager.Cancel(ctx, jobs.CancelParams{JobID: id, ActorUserID: actor, ClientTxnID: uuid.NewString(), NormalizedRequest: []byte(`{}`)})
					if !errors.Is(err, jobs.ErrCancelRejected) || canceled.ReasonCode != jobs.CancelReasonNotCancelable {
						t.Fatalf("cancellation policy: %#v/%v", canceled, err)
					}
					var observations int
					if err := pool.QueryRow(ctx, `SELECT count(*) FROM extension_job_cancellation_observations WHERE job_id=$1`, id).Scan(&observations); err != nil || observations != 0 {
						t.Fatal("rejected cancellation wrote evidence", observations, err)
					}
					execution, claimed, err := manager.Claim(ctx, id)
					if err != nil || !claimed {
						t.Fatal("claim", claimed, err)
					}
					completion := jobs.SuccessCompletion{Progress: jobs.Progress{Completed: 1, Total: intPointer(1)}, ResultSummary: jobs.ResultSummary{Code: "done", Message: "Done."}}
					if _, err := manager.CompleteSucceeded(ctx, execution, completion); !errors.Is(err, jobs.ErrInvalidTransition) {
						t.Fatal("generic completion bypassed owner policy", err)
					}
					var result jobs.Resource
					if proofRequired {
						if _, err := finalizer.FinalizeJobOnlySuccess(ctx, execution, completion); err == nil {
							t.Fatal("proof-required job bypassed proof")
						}
						result, err = finalizer.FinalizeSuccess(ctx, JobFinalizationRequest{Execution: execution, Completion: completion, FinalCommitID: uuid.NewString()})
					} else {
						mutated := false
						if _, err := finalizer.FinalizeSuccess(ctx, JobFinalizationRequest{Execution: execution, Completion: completion, FinalCommitID: uuid.NewString(), Mutate: func(context.Context, pgx.Tx) error { mutated = true; return nil }}); err == nil || mutated {
							t.Fatal("proof-forbidden job reached mutation")
						}
						result, err = finalizer.FinalizeJobOnlySuccess(ctx, execution, completion)
					}
					if err != nil || result.Status != jobs.StatusSucceeded {
						t.Fatalf("policy finalization: %#v/%v", result, err)
					}
					var proofs int
					if err := pool.QueryRow(ctx, `SELECT count(*) FROM extension_job_commit_proofs WHERE job_id=$1`, id).Scan(&proofs); err != nil {
						t.Fatal(err)
					}
					if (proofs == 1) != proofRequired || proofs > 1 {
						t.Fatal("proof policy violated", proofs)
					}
					if proofRequired && !idempotent {
						proof, err := store.JobCommitProof(ctx, id)
						if err != nil || string(proof.IdempotencyIdentity) != "null" {
							t.Fatal("proof lost null identity", err)
						}
					}
					read, err := pool.Begin(ctx)
					if err != nil {
						t.Fatal(err)
					}
					defer func() { _ = read.Rollback(ctx) }()
					replay, present, err := collaborationsupport.NewJobOwnerTransactionAdapters().LookupRouteIdempotencyTx(ctx, read, key)
					if err != nil || present != (idempotent && !local) {
						t.Fatal("route replay policy", present, err)
					}
					if present {
						var retained jobs.Resource
						if json.Unmarshal(replay.ResponseJSON, &retained) != nil || retained.Status != jobs.StatusSucceeded || retained.JobID != queued.JobID {
							t.Fatal("terminal replay not atomic")
						}
					}
				})
			}
		}
	}
}
