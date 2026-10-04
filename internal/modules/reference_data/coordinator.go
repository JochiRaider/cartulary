package reference_data

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"slices"
	"time"

	"github.com/JochiRaider/cartulary/internal/modules/reference_data/internal/packformat"
	"github.com/JochiRaider/cartulary/internal/platform/authn"
	"github.com/JochiRaider/cartulary/internal/platform/canonicaljson"
	"github.com/JochiRaider/cartulary/internal/platform/jobs"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Coordinator owns Reference Data application operations. Transport adapters
// supply admitted requests and actors; they do not choose trust inputs, perform
// verification, write owner rows or decide publication outcomes.
type Coordinator struct {
	observer        OperationObserver
	pool            *pgxpool.Pool
	storage         ArtifactStorage
	configuration   Configuration
	limits          Limits
	admission       referenceJobAdmission
	operations      referenceJobOperations
	finalizer       JobSuccessFinalizer
	actionFinalizer ActionFinalizer
	executionGuard  referenceJobExecutionGuard
	registryUsage   RegistryUsageReader
	now             func() time.Time
}

type referenceJobExecutionGuard interface {
	ExtensionCancellationContextTx(context.Context, pgx.Tx, jobs.Execution) (jobs.ExtensionFinalizationContext, error)
}

type CoordinatorOptions struct {
	Observer          OperationObserver
	Postgres          *pgxpool.Pool
	Storage           ArtifactStorage
	Configuration     Configuration
	Limits            Limits
	JobAdmission      referenceJobAdmission
	JobOperations     referenceJobOperations
	JobFinalizer      JobSuccessFinalizer
	JobExecutionGuard referenceJobExecutionGuard
	RegistryUsage     RegistryUsageReader
	Now               func() time.Time
}

type ImportAdmission interface {
	PrepareImport(context.Context, io.Reader) (*PendingImport, error)
}
type ImportAdmissionOptions struct {
	Observer      OperationObserver
	Postgres      *pgxpool.Pool
	Storage       ArtifactStorage
	Configuration Configuration
	Limits        Limits
	JobAdmission  referenceJobAdmission
	Now           func() time.Time
}

func NewImportAdmission(options ImportAdmissionOptions) (ImportAdmission, error) {
	if options.Postgres == nil || options.Storage == nil || options.JobAdmission == nil || options.Now == nil {
		return nil, errors.New("reference pack: incomplete import admission dependencies")
	}
	if err := validateCoordinatorLimits(options.Limits); err != nil {
		return nil, err
	}
	return &Coordinator{observer: options.Observer, pool: options.Postgres, storage: options.Storage, configuration: options.Configuration, limits: options.Limits, admission: options.JobAdmission, now: options.Now}, nil
}
func validateCoordinatorLimits(limits Limits) error {
	l := limits.verificationArchiveLimits()
	if l.ContainerBytes < 1 || l.ExtractedBytes < 1 || l.CompressionRatio < 1 || l.CompressionRatio > 1000 || l.Members < 1 || limits.ReferencePacks.MaxVerificationSeconds < 60 || limits.ReferencePacks.MaxVerificationSeconds > 86400 {
		return errors.New("reference pack: invalid coordinator limits")
	}
	return nil
}

func NewCoordinator(options CoordinatorOptions) (*Coordinator, error) {
	if options.Postgres == nil || options.Storage == nil || options.JobAdmission == nil || options.JobOperations == nil || options.JobFinalizer == nil || options.JobExecutionGuard == nil || options.Now == nil {
		return nil, errors.New("reference pack: incomplete coordinator dependencies")
	}
	if err := validateCoordinatorLimits(options.Limits); err != nil {
		return nil, err
	}
	return &Coordinator{observer: options.Observer, pool: options.Postgres, storage: options.Storage, configuration: options.Configuration, limits: options.Limits, admission: options.JobAdmission, operations: options.JobOperations, finalizer: options.JobFinalizer, actionFinalizer: options.JobFinalizer, executionGuard: options.JobExecutionGuard, registryUsage: options.RegistryUsage, now: options.Now}, nil
}

type VerificationRequest struct {
	OperatorOperationID uuid.UUID
	Kind                string
	ActorUserID         uuid.UUID
	ClientTxnID         string
	PackKeys            []string
	KeysProvided        bool
	PackVersion         string
	Reason              *string
}

// Import completes an immutable input object before admission. Its retained
// operation reference is committed with the Job, so queued inputs participate
// in backup and recovery. Replay removes only this newly prepared copy.
func (c *Coordinator) Import(ctx context.Context, actor uuid.UUID, clientTxnID string, source io.Reader) (JobAcceptedResult, error) {
	pending, err := c.PrepareImport(ctx, source)
	if err != nil {
		return JobAcceptedResult{}, err
	}
	defer pending.Close()
	return pending.Accept(ctx, actor, clientTxnID)
}

// PendingImport owns a prepared immutable upload until the complete transport
// envelope has been admitted. Its opaque state prevents an adapter from
// changing the selected bytes between multipart parsing and Job admission.
type PendingImport struct {
	endObservation func(string)
	owner          *Coordinator
	identity       importIdentity
	input          preparedObject
	finished       bool
}

func (c *Coordinator) PrepareImport(ctx context.Context, source io.Reader) (*PendingImport, error) {
	ctx, end := observeReferenceOperation(ctx, c.observer, "reference_pack.import")
	identity, input, err := c.prepareImportInput(ctx, source)
	if err != nil {
		end(referenceOutcome(err))
		return nil, err
	}
	return &PendingImport{endObservation: end, owner: c, identity: identity, input: input}, nil
}
func (p *PendingImport) SHA256() string { return p.input.Digest }
func (p *PendingImport) Close() error {
	if p == nil || p.finished {
		return nil
	}
	p.finished = true
	err := errors.Join(p.owner.storage.RemovePublished(p.input.Reference), p.input.releasePublication())
	p.endObservation("rejected")
	return err
}
func (p *PendingImport) Identity() (string, string) { return p.identity.Key, p.identity.Version }
func (p *PendingImport) Accept(ctx context.Context, actor uuid.UUID, clientTxnID string) (JobAcceptedResult, error) {
	return p.accept(ctx, actor, uuid.Nil, clientTxnID)
}
func (p *PendingImport) AcceptLocalOperator(ctx context.Context, operationID uuid.UUID) (JobAcceptedResult, error) {
	if operationID == uuid.Nil {
		return JobAcceptedResult{}, errors.New("reference pack: local operation identity required")
	}
	return p.accept(ctx, uuid.Nil, operationID, operationID.String())
}
func (p *PendingImport) accept(ctx context.Context, actor, operatorOperation uuid.UUID, clientTxnID string) (JobAcceptedResult, error) {
	if p == nil || p.finished {
		return JobAcceptedResult{}, errors.New("reference pack: closed import preparation")
	}
	defer p.input.releasePublication()
	c := p.owner
	identity, input := p.identity, p.input
	request := VerificationRequest{Kind: "import", ActorUserID: actor, OperatorOperationID: operatorOperation, ClientTxnID: clientTxnID}
	a := operationAdmission{ID: uuid.New(), Kind: "import", ActorKind: "user", Actor: &actor, At: c.now().UTC(), Import: &identity, InputObject: &input, ContainerRef: &input.Reference, ContainerSHA: &input.Digest, ContainerBytes: &input.Size, ClockTrusted: c.configuration.ClockTrusted, TimeoutSeconds: c.limits.ReferencePacks.MaxVerificationSeconds}
	if operatorOperation != uuid.Nil {
		a.ID = operatorOperation
		a.ActorKind = "local_operator"
		a.Actor = nil
	}
	result, uncertain, err := c.accept(ctx, request, a)
	p.finished = true
	if result.Replayed {
		// A proven admission replay is authoritative. Failure to remove this
		// unused copy leaves cleanup work, not a failed or uncertain operation.
		_ = c.storage.RemovePublished(input.Reference)
	} else if err != nil && !uncertain {
		err = errors.Join(err, c.storage.RemovePublished(input.Reference))
	}
	p.endObservation(referenceOutcome(err))
	return result, err
}

func (c *Coordinator) prepareImportInput(ctx context.Context, source io.Reader) (identity importIdentity, input preparedObject, resultErr error) {
	ref, digest, size, err := c.storage.StageStream(ctx, source, c.limits.ReferencePacks.MaxContainerBytes)
	if err != nil {
		return identity, input, importStagingError(ctx, err)
	}
	defer func() {
		cleanupErr := c.storage.RemoveStaged(ref)
		if (resultErr != nil || cleanupErr != nil) && input.ID != uuid.Nil {
			cleanupErr = errors.Join(cleanupErr, c.storage.RemovePublished(input.Reference), input.releasePublication())
		}
		// An operational cleanup failure cannot carry a publishable transport
		// rejection through errors.Is/As from a joined lower-level failure.
		if cleanupErr != nil {
			resultErr = cleanupErr
		}
	}()
	identity, err = probeContainerIdentity(ctx, c.storage, ref, c.limits.verificationArchiveLimits())
	if err != nil {
		return identity, input, err
	}
	reader, actual, err := c.storage.OpenStaged(ctx, ref)
	if err != nil {
		return identity, input, err
	}
	if actual != size {
		_ = reader.Close()
		return identity, input, errors.New("reference pack: staged container changed")
	}
	retained, lease, writeErr := c.storage.PublishStream(ctx, digest, size, io.NewSectionReader(reader, 0, size))
	closeErr := reader.Close()
	if writeErr != nil {
		return identity, input, errors.Join(writeErr, closeErr)
	}
	if closeErr != nil {
		return identity, input, errors.Join(closeErr, c.storage.RemovePublished(retained), lease.Close())
	}
	input = preparedObject{ID: uuid.New(), Path: "input/container", Digest: digest, Size: size, Reference: retained, publication: lease}
	return identity, input, nil
}

func (c *Coordinator) VerifyRetained(ctx context.Context, request VerificationRequest) (JobAcceptedResult, error) {
	if request.OperatorOperationID != uuid.Nil || (request.Kind != "reverify" && request.Kind != "refresh") {
		return JobAcceptedResult{}, errors.New("reference pack: invalid verification operation")
	}
	a := operationAdmission{ID: uuid.New(), Kind: request.Kind, ActorKind: "user", Actor: &request.ActorUserID, At: c.now().UTC(), Keys: request.PackKeys, Version: request.PackVersion, ClockTrusted: c.configuration.ClockTrusted, TimeoutSeconds: c.limits.ReferencePacks.MaxVerificationSeconds}
	result, _, err := c.accept(ctx, request, a)
	return result, err
}

func (c *Coordinator) accept(ctx context.Context, request VerificationRequest, a operationAdmission) (JobAcceptedResult, bool, error) {
	local := request.OperatorOperationID != uuid.Nil
	if request.ClientTxnID == "" || (local && (request.Kind != "import" || request.ActorUserID != uuid.Nil || request.ClientTxnID != request.OperatorOperationID.String() || a.ID != request.OperatorOperationID || a.ActorKind != "local_operator" || a.Actor != nil)) || (!local && request.ActorUserID == uuid.Nil) {
		return JobAcceptedResult{}, false, errors.New("reference pack: missing admitted actor or idempotency identity")
	}
	keys := slices.Clone(request.PackKeys)
	slices.Sort(keys)
	keys = slices.Compact(keys)
	if request.Kind == "refresh" && request.KeysProvided && len(keys) == 0 {
		return JobAcceptedResult{}, false, wrapAPIError(invalidReferencePackRequest("pack_keys", "empty_pack_keys"))
	}
	if request.Kind == "reverify" && (len(keys) != 1 || request.PackVersion == "") {
		return JobAcceptedResult{}, false, consumerError("invalid_pack_request")
	}
	for _, key := range keys {
		if !validConsumerKey(key) {
			return JobAcceptedResult{}, false, consumerError("invalid_pack_request")
		}
	}
	var selector any
	if request.KeysProvided || request.Kind == "reverify" {
		selector = keys
	}
	normalized, err := canonicaljson.Marshal(map[string]any{"schema_id": "cartulary.reference_pack_verification_request.v2", "kind": request.Kind, "pack_keys": selector, "pack_version": request.PackVersion, "container_sha256": a.ContainerSHA, "reason": request.Reason})
	if err != nil {
		return JobAcceptedResult{}, false, err
	}
	key := authn.RouteIdempotencyKey{RouteKey: "reference_packs." + request.Kind, ActorUserID: request.ActorUserID, ScopeKey: "deployment", ClientTxnID: request.ClientTxnID}
	if request.Kind == "reverify" {
		key.ScopeKey = keys[0] + ":" + request.PackVersion
	}
	tx, err := c.pool.Begin(ctx)
	if err != nil {
		return JobAcceptedResult{}, false, err
	}
	defer func() { _ = tx.Rollback(ctx) }()
	if !local {
		if err := authn.LockRouteIdempotencyTx(ctx, tx, key); err != nil {
			return JobAcceptedResult{}, false, err
		}
		if prior, err := authn.GetRouteIdempotencyTx(ctx, tx, key); err == nil {
			if !bytes.Equal(prior.RequestHash, hashBytes(normalized)) {
				return JobAcceptedResult{}, false, authn.ErrClientTxnConflict
			}
			var job jobs.Resource
			if err := json.Unmarshal(prior.ResponseJSON, &job); err != nil {
				return JobAcceptedResult{}, false, err
			}
			return JobAcceptedResult{Job: job, Replayed: true}, false, nil
		} else if !errors.Is(err, authn.ErrNotFound) {
			return JobAcceptedResult{}, false, err
		}
	}
	for _, key := range keys {
		known := false
		for _, profile := range packformat.Profiles() {
			if profile.Key == key {
				known = true
				break
			}
		}
		if !known {
			return JobAcceptedResult{}, false, wrapAPIError(invalidReferencePackRequest("pack_keys", "invalid_pack_keys"))
		}
	}
	a.Keys = keys
	if request.Kind == "refresh" && !request.KeysProvided {
		rows, err := tx.Query(ctx, `SELECT DISTINCT pack_key COLLATE "C" FROM reference_pack_candidates WHERE distribution_kind='operator_imported' AND current_envelope_id IS NOT NULL AND NOT removed ORDER BY 1`)
		if err != nil {
			return JobAcceptedResult{}, false, err
		}
		a.Keys = []string{}
		for rows.Next() {
			var k string
			if err := rows.Scan(&k); err != nil {
				rows.Close()
				return JobAcceptedResult{}, false, err
			}
			a.Keys = append(a.Keys, k)
		}
		err = rows.Err()
		rows.Close()
		if err != nil {
			return JobAcceptedResult{}, false, err
		}
	}
	scope := jobs.Scope{Kind: jobs.ScopeKindDeployment}
	var admission *jobs.ExtensionJobAdmission
	if request.Kind == "import" {
		admission, err = jobs.NewAttributedExtensionJobAdmission(ProfileID, jobs.NewRouteIdempotencyKey(key.RouteKey, key.ActorUserID, key.ScopeKey, key.ClientTxnID), scope, request.OperatorOperationID, normalized)
	} else {
		admission, err = jobs.NewExtensionJobAdmission(ProfileID, jobs.NewRouteIdempotencyKey(key.RouteKey, key.ActorUserID, key.ScopeKey, key.ClientTxnID), scope, normalized)
	}
	if err != nil {
		return JobAcceptedResult{}, false, err
	}
	job, err := c.admission.CreateQueuedTx(ctx, tx, jobs.EnqueueParams{JobKind: referenceCatalogJobKind(request.Kind), Scope: scope, SubmittedByUserID: request.ActorUserID, OperatorOperationID: request.OperatorOperationID, AuthPolicy: jobs.AuthPolicyDeploymentAdmin, Cancelable: true, Progress: jobs.Progress{Completed: 0, Total: intPtr(1)}, Extension: admission}, a.At)
	if err != nil {
		return JobAcceptedResult{}, false, err
	}
	id, err := uuid.Parse(job.JobID)
	if err != nil {
		return JobAcceptedResult{}, false, err
	}
	a.JobID = &id
	if _, err := admitOperationTx(ctx, tx, a); err != nil {
		return JobAcceptedResult{}, false, err
	}
	if err := appendPackAuditTx(ctx, tx, a.ID, a.Kind+"_admitted", "admitted", a.At, "", "", nil); err != nil {
		return JobAcceptedResult{}, false, err
	}
	if !local {
		if err := authn.InsertRouteIdempotencyPayload(ctx, tx, key, nil, hashBytes(normalized), http.StatusAccepted, job); err != nil {
			return JobAcceptedResult{}, false, err
		}
	}
	if err := tx.Commit(ctx); err != nil {
		absent := tx.Rollback(ctx) == nil || errors.Is(err, pgx.ErrTxCommitRollback)
		return JobAcceptedResult{}, !absent, err
	}
	return JobAcceptedResult{Job: job}, false, nil
}
