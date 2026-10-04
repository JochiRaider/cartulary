package operator

import (
	"context"
	"errors"
	"fmt"
	"time"

	dbmigrations "github.com/JochiRaider/cartulary/db/migrations"
	"github.com/JochiRaider/cartulary/internal/app/configassembly"
	"github.com/JochiRaider/cartulary/internal/app/referenceassembly"
	"github.com/JochiRaider/cartulary/internal/modules/database_migrations"
	"github.com/JochiRaider/cartulary/internal/modules/reference_data"
	"github.com/JochiRaider/cartulary/internal/platform/canonicaljson"
	"github.com/JochiRaider/cartulary/internal/platform/jobs"
	"github.com/JochiRaider/cartulary/internal/platform/postgres"
	"github.com/JochiRaider/cartulary/internal/platform/processlease"
	"github.com/google/uuid"
)

const referencePackImportUsage = "operator reference-pack import <bundle_name>"

type referencePackOperatorResult struct {
	SchemaID        string  `json:"schema_id"`
	OperationID     string  `json:"operation_id"`
	Result          string  `json:"result"`
	ContainerSHA256 *string `json:"container_sha256"`
	PackKey         *string `json:"pack_key"`
	PackVersion     *string `json:"pack_version"`
	JobID           *string `json:"job_id"`
	ErrorCode       *string `json:"error_code"`
	ReasonCode      *string `json:"reason_code"`
}

type referencePackLocalImport interface {
	importAndObserve(context.Context, uuid.UUID, string) referencePackOperatorResult
}
type referencePackExecutor struct {
	transport      operatorTransport
	loadConfig     func(string) (configassembly.Loaded, error)
	newOperationID func() uuid.UUID
	open           func(context.Context, configassembly.Loaded) (referencePackLocalImport, func(), error)
}

func (e referencePackExecutor) runCommand(ctx context.Context, args []string) int {
	id := e.newOperationID()
	result := referencePackOperatorResult{SchemaID: "cartulary.reference_pack_operator_result.v1", OperationID: id.String(), Result: "failed"}
	exit := 3
	fail := func(code string) { result.ErrorCode = &code }
	switch {
	case len(args) != 3 || args[0] != "reference-pack" || args[1] != "import" || !reference_data.ValidOperatorBundleName(args[2]):
		fail("invalid_operator_request")
		exit = 2
		_, _ = fmt.Fprintln(normalizeOperatorWriter(e.transport.stderr), "usage: "+referencePackImportUsage)
	default:
		loaded, err := e.loadConfig("")
		if err != nil {
			fail("internal_error")
			break
		}
		port, close, err := e.open(ctx, loaded)
		if err != nil {
			fail("internal_error")
			break
		}
		result = port.importAndObserve(ctx, id, args[2])
		close()
		if result.Result == "succeeded" {
			exit = 0
		}
	}
	data, err := canonicaljson.Marshal(result)
	if err != nil {
		return 3
	}
	if n, err := normalizeOperatorWriter(e.transport.stdout).Write(append(data, '\n')); err != nil || n != len(data)+1 {
		return 3
	}
	return exit
}

type referencePackLocalRuntime struct {
	ctx       context.Context
	storage   *referenceassembly.RootStorage
	admission reference_data.ImportAdmission
	jobs      interface {
		Get(context.Context, uuid.UUID) (jobs.Resource, error)
	}
}

func openReferencePackLocalRuntime(ctx context.Context, loaded configassembly.Loaded) (port referencePackLocalImport, close func(), err error) {
	if err := loaded.ValidateForStartup(); err != nil {
		return nil, nil, err
	}
	cfg := loaded.Deployment()
	settings, err := postgres.ResolveSettings(configassembly.PostgresBinding(cfg), postgres.PurposeRuntime, nil)
	if err != nil {
		return nil, nil, err
	}
	pool, err := postgres.Setup(ctx, settings)
	if err != nil {
		return nil, nil, err
	}
	cleanups := []func(){pool.Close}
	cleanup := func() {
		for i := len(cleanups) - 1; i >= 0; i-- {
			cleanups[i]()
		}
	}
	defer func() {
		if err != nil {
			cleanup()
		}
	}()
	operationCtx, cancel := context.WithCancel(ctx)
	cleanups = append(cleanups, cancel)
	lease, err := processlease.AcquireApplicationRecoveryServing(operationCtx, pool.Pool(), time.Duration(cfg.Timeouts.Extensions.ProcessLeaseAcquireSeconds)*time.Second, time.Duration(cfg.Timeouts.Extensions.ProcessLeaseLossDetectionSeconds)*time.Second)
	if err != nil {
		return nil, nil, err
	}
	cleanups = append(cleanups, lease.Close)
	lease.StartMonitor(operationCtx)
	go func() {
		for {
			select {
			case <-operationCtx.Done():
				return
			case event := <-lease.Events():
				if event.State == processlease.StateLost || event.State == processlease.StateUncertain {
					cancel()
					return
				}
			}
		}
	}()
	source, err := dbmigrations.Source()
	if err != nil {
		return nil, nil, err
	}
	if err := database_migrations.EnsureSchemaReady(operationCtx, pool.Pool(), source); err != nil {
		return nil, nil, err
	}
	storage, err := referenceassembly.NewRootStorage(cfg.Roots.TemporaryWork.Path, cfg.Roots.ReferencePackStorage.Path)
	if err != nil {
		return nil, nil, err
	}
	cleanups = append(cleanups, storage.Close)
	if err := storage.ReconcileWorkspaces(operationCtx); err != nil {
		return nil, nil, err
	}
	if err := reference_data.CollectUnreferencedObjects(operationCtx, pool.Pool(), storage); err != nil {
		return nil, nil, err
	}
	limits := reference_data.Limits{Archives: reference_data.ArchiveLimits{DefaultMaxExtractedBytes: cfg.Limits.Archives.DefaultMaxExtractedBytes, MaxCompressionRatio: cfg.Limits.Archives.MaxCompressionRatio, MaxMembers: cfg.Limits.Archives.MaxMembers}, ReferencePacks: reference_data.ReferenceLimits{MaxExtractedBytes: cfg.Limits.ReferencePacks.MaxExtractedBytes, MaxContainerBytes: cfg.Limits.ReferencePacks.MaxContainerBytes, MaxVerificationSeconds: cfg.Limits.ReferencePacks.MaxVerificationSeconds}}
	admission, manager, err := referenceassembly.NewOperatorAdmission(operationCtx, pool.Pool(), storage, cfg.ReferencePacks, limits, loaded.RequestedClaims().ProfileIDs(), func() time.Time { return time.Now().UTC() })
	if err != nil {
		return nil, nil, err
	}
	return &referencePackLocalRuntime{ctx: operationCtx, storage: storage, admission: admission, jobs: manager}, cleanup, nil
}

func (r *referencePackLocalRuntime) importAndObserve(_ context.Context, id uuid.UUID, name string) referencePackOperatorResult {
	result := referencePackOperatorResult{SchemaID: "cartulary.reference_pack_operator_result.v1", OperationID: id.String(), Result: "failed"}
	internal := "internal_error"
	result.ErrorCode = &internal
	reader, err := r.storage.OpenIncoming(r.ctx, name)
	if err != nil {
		return result
	}
	pending, err := r.admission.PrepareImport(r.ctx, reader)
	closeErr := reader.Close()
	if pending != nil {
		defer pending.Close()
	}
	if closeErr != nil {
		return result
	}
	if err != nil {
		if rejected, ok := reference_data.ImportContentRejection(err); ok {
			code := "reference_pack_verification_failed"
			result.ErrorCode = &code
			result.ReasonCode = &rejected.Code
		}
		return result
	}
	digest := pending.SHA256()
	result.ContainerSHA256 = &digest
	key, version := pending.Identity()
	if key != "" {
		result.PackKey = &key
		result.PackVersion = &version
	}
	accepted, err := pending.AcceptLocalOperator(r.ctx, id)
	if err != nil {
		var rejection *reference_data.OperationRejection
		if errors.As(err, &rejection) {
			code := "reference_pack_operation_rejected"
			result.ErrorCode = &code
			result.ReasonCode = &rejection.Reason
		}
		return result
	}
	result.JobID = &accepted.Job.JobID
	jobID, err := uuid.Parse(accepted.Job.JobID)
	if err != nil {
		return result
	}
	tick := time.NewTicker(time.Second)
	defer tick.Stop()
	for {
		job, err := r.jobs.Get(r.ctx, jobID)
		if err != nil {
			return result
		}
		switch job.Status {
		case jobs.StatusSucceeded:
			result.Result = "succeeded"
			result.ErrorCode = nil
			return result
		case jobs.StatusFailed:
			if job.ErrorSummary != nil {
				result.ErrorCode = &job.ErrorSummary.Code
				if reason, ok := job.ErrorSummary.Details["reason_code"].(string); ok {
					result.ReasonCode = &reason
				}
			}
			return result
		case jobs.StatusCanceled:
			code := "job_canceled"
			result.ErrorCode = &code
			return result
		}
		select {
		case <-r.ctx.Done():
			return result
		case <-tick.C:
		}
	}
}
