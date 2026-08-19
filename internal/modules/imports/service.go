package imports

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/JochiRaider/cartulary/internal/modules/imports/ownerfacade"
	"github.com/JochiRaider/cartulary/internal/modules/incidents/admission"
	"github.com/JochiRaider/cartulary/internal/modules/revisions"
	"github.com/JochiRaider/cartulary/internal/platform/authn"
	"github.com/JochiRaider/cartulary/internal/platform/httpapi"
	"github.com/JochiRaider/cartulary/internal/platform/jobs"
	"github.com/JochiRaider/cartulary/internal/platform/pagination"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type Service struct {
	store                    *Store
	incidentAccess           *admission.Checker
	authStore                *authn.Store
	jobManager               importJobOperations
	jobRunner                importJobRunner
	keys                     authn.MasterKeys
	cursorCodec              *pagination.Codec
	limits                   Limits
	archiveLimits            ArchiveLimits
	ownerCreateRegistry      *ownerfacade.ImportOwnerCreateRegistry
	extensionImportFacades   map[string]ExtensionImportFacade
	extensionProfileAdmitted func(string) bool
	jobSuccessFinalizer      JobSuccessFinalizer
	now                      func() time.Time
}

type importJobTransactions interface {
	CreateQueuedTx(context.Context, pgx.Tx, jobs.EnqueueParams, time.Time) (jobs.Resource, error)
	ValidateExecutionTx(context.Context, pgx.Tx, jobs.Execution) error
	ValidateCancellationExecutionTx(context.Context, pgx.Tx, jobs.Execution) error
}

type importJobOperations interface {
	HandlerPayload(context.Context, jobs.Execution) (json.RawMessage, error)
	ObserveExecution(context.Context, jobs.Execution) (jobs.Resource, error)
	UpdateProgress(context.Context, jobs.Execution, jobs.Progress, *string) (jobs.Resource, error)
	CompleteCanceled(context.Context, jobs.Execution, jobs.CancellationCompletion) (jobs.Resource, error)
}

type importJobRunner interface {
	RegisterHandler(string, jobs.HandlerFunc) error
	Notify(uuid.UUID)
}

type RouteOption func(*routeOptions)

type routeOptions struct {
	extensionProfileAdmitted func(string) bool
	jobSuccessFinalizer      JobSuccessFinalizer
	limits                   Limits
	archiveLimits            ArchiveLimits
	ownerCreateRegistry      *ownerfacade.ImportOwnerCreateRegistry
	revisionAppender         *revisions.Appender
	jobTransactions          importJobTransactions
	jobOperations            importJobOperations
	jobRunner                importJobRunner
}

func WithJobs(transactions importJobTransactions, operations importJobOperations, runner importJobRunner) RouteOption {
	return func(options *routeOptions) {
		options.jobTransactions = transactions
		options.jobOperations = operations
		options.jobRunner = runner
	}
}

func WithOwnerCreateRegistry(
	registry *ownerfacade.ImportOwnerCreateRegistry,
) RouteOption {
	return func(options *routeOptions) {
		options.ownerCreateRegistry = registry
	}
}

func WithRevisionAppender(appender *revisions.Appender) RouteOption {
	return func(options *routeOptions) {
		options.revisionAppender = appender
	}
}

func WithLimits(limits Limits, archiveLimits ArchiveLimits) RouteOption {
	return func(options *routeOptions) {
		options.limits = limits
		options.archiveLimits = archiveLimits
	}
}

func WithExtensionProfileAdmission(admitted func(string) bool) RouteOption {
	return func(options *routeOptions) {
		options.extensionProfileAdmitted = admitted
	}
}

func WithJobSuccessFinalizer(finalizer JobSuccessFinalizer) RouteOption {
	return func(options *routeOptions) {
		options.jobSuccessFinalizer = finalizer
	}
}

func RegisterRoutes(options ...RouteOption) httpapi.RouteRegistrar {
	return func(mux *http.ServeMux, deps httpapi.DependencySet) error {
		settings := routeOptions{}
		for _, option := range options {
			if option != nil {
				option(&settings)
			}
		}
		service, err := newService(deps, settings)
		if err != nil {
			return err
		}
		return bindOwnerRoutes(mux, deps, service)
	}
}

func newService(deps httpapi.DependencySet, options routeOptions) (*Service, error) {
	keys, err := authn.LoadMasterKeys(deps.Env)
	if err != nil {
		return nil, fmt.Errorf("load auth master key: %w", err)
	}
	now := deps.Now
	if now == nil {
		now = func() time.Time { return time.Now().UTC() }
	}
	cursorCodec := deps.CursorCodec
	if cursorCodec == nil {
		cursorKey := authn.DerivePurposeKey(keys, "pagination-cursor-v1")
		cursorCodec = pagination.NewCodec(cursorKey[:])
	}
	if options.ownerCreateRegistry == nil {
		return nil, fmt.Errorf("import route composition requires an owner-create registry")
	}
	if options.revisionAppender == nil {
		return nil, fmt.Errorf("import route composition requires a Revisions appender")
	}
	if options.jobOperations != nil && options.jobTransactions == nil {
		return nil, fmt.Errorf("import admitted route requires the Jobs transaction service")
	}
	extensionImportFacades, err := extensionImportFacadesFromDependencies(deps)
	if err != nil {
		return nil, err
	}
	if options.jobOperations != nil && options.jobSuccessFinalizer == nil {
		return nil, fmt.Errorf("import admitted route requires a job success finalizer")
	}
	extensionProfileAdmitted := options.extensionProfileAdmitted
	if extensionProfileAdmitted == nil {
		extensionProfileAdmitted = func(string) bool { return false }
	}
	service := &Service{
		store: NewStore(
			deps.Postgres,
			options.revisionAppender,
			options.jobTransactions,
		),
		incidentAccess:           admission.NewChecker(deps.PostgresHandle()),
		authStore:                authn.NewStore(deps.PostgresHandle()),
		jobManager:               options.jobOperations,
		jobRunner:                options.jobRunner,
		keys:                     keys,
		cursorCodec:              cursorCodec,
		limits:                   options.limits,
		archiveLimits:            options.archiveLimits,
		ownerCreateRegistry:      options.ownerCreateRegistry,
		extensionImportFacades:   extensionImportFacades,
		extensionProfileAdmitted: extensionProfileAdmitted,
		jobSuccessFinalizer:      options.jobSuccessFinalizer,
		now:                      now,
	}
	if err := service.registerJobHandlers(); err != nil {
		return nil, err
	}
	return service, nil
}
