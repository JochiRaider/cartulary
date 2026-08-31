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
	"github.com/JochiRaider/cartulary/internal/platform/postgres"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type service struct {
	store                    *store
	incidentAccess           *admission.Checker
	authStore                *authn.Store
	jobManager               importJobOperations
	jobRunner                importJobRunner
	keys                     authn.MasterKeys
	cursorCodec              *pagination.Codec
	limits                   Limits
	archiveLimits            ArchiveLimits
	ownerCreateRegistry      *ownerfacade.ImportOwnerCreateRegistry
	extensionImportFacades   map[analyticalImportTargetKey]ExtensionImportFacade
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

type ModuleDependencies struct {
	Postgres                  postgres.DB
	JobTransactions           importJobTransactions
	JobOperations             importJobOperations
	JobRunner                 importJobRunner
	Limits                    Limits
	ArchiveLimits             ArchiveLimits
	OwnerCreateRegistry       *ownerfacade.ImportOwnerCreateRegistry
	RevisionAppender          *revisions.Appender
	ExtensionProfileAdmission func(string) bool
	JobSuccessFinalizer       JobSuccessFinalizer
	ExtensionImportFacades    []ExtensionImportFacade
	Env                       map[string]string
	CursorCodec               *pagination.Codec
	Now                       func() time.Time
}

type Module struct {
	service *service
}

func NewModule(dependencies ModuleDependencies) (*Module, error) {
	if nilInterface(dependencies.Postgres) {
		return nil, fmt.Errorf("imports module requires PostgreSQL")
	}
	if nilInterface(dependencies.JobTransactions) {
		return nil, fmt.Errorf("imports module requires Jobs transactions")
	}
	if nilInterface(dependencies.JobOperations) {
		return nil, fmt.Errorf("imports module requires Jobs operations")
	}
	if nilInterface(dependencies.JobRunner) {
		return nil, fmt.Errorf("imports module requires a Jobs runner")
	}
	if dependencies.OwnerCreateRegistry == nil {
		return nil, fmt.Errorf("imports module requires an owner-create registry")
	}
	if dependencies.RevisionAppender == nil {
		return nil, fmt.Errorf("imports module requires a Revisions appender")
	}
	if dependencies.ExtensionProfileAdmission == nil {
		return nil, fmt.Errorf("imports module requires extension profile admission")
	}
	if nilInterface(dependencies.JobSuccessFinalizer) {
		return nil, fmt.Errorf("imports module requires a job finalizer")
	}
	if err := validateModuleLimits(dependencies.Limits, dependencies.ArchiveLimits); err != nil {
		return nil, err
	}
	keys, err := authn.LoadMasterKeys(dependencies.Env)
	if err != nil {
		return nil, fmt.Errorf("load auth master key: %w", err)
	}
	now := dependencies.Now
	if now == nil {
		now = func() time.Time { return time.Now().UTC() }
	}
	cursorCodec := dependencies.CursorCodec
	if cursorCodec == nil {
		cursorKey := authn.DerivePurposeKey(keys, "pagination-cursor-v1")
		cursorCodec = pagination.NewCodec(cursorKey[:])
	}
	extensionImportFacades, err := validateExtensionImportFacades(
		dependencies.ExtensionImportFacades,
		dependencies.ExtensionProfileAdmission,
	)
	if err != nil {
		return nil, err
	}
	service := &service{
		store: newStore(
			dependencies.Postgres,
			dependencies.RevisionAppender,
			dependencies.JobTransactions,
		),
		incidentAccess:           admission.NewChecker(dependencies.Postgres),
		authStore:                authn.NewStore(dependencies.Postgres),
		jobManager:               dependencies.JobOperations,
		jobRunner:                dependencies.JobRunner,
		keys:                     keys,
		cursorCodec:              cursorCodec,
		limits:                   dependencies.Limits,
		archiveLimits:            dependencies.ArchiveLimits,
		ownerCreateRegistry:      dependencies.OwnerCreateRegistry,
		extensionImportFacades:   extensionImportFacades,
		extensionProfileAdmitted: dependencies.ExtensionProfileAdmission,
		jobSuccessFinalizer:      dependencies.JobSuccessFinalizer,
		now:                      now,
	}
	return &Module{service: service}, nil
}

func (m *Module) RegisterRoutes() httpapi.RouteRegistrar {
	return func(mux *http.ServeMux, deps httpapi.DependencySet) error {
		if m == nil || m.service == nil {
			return fmt.Errorf("imports module unavailable")
		}
		return bindOwnerRoutes(mux, deps, m.service)
	}
}

func (m *Module) RegisterWorkers() error {
	if m == nil || m.service == nil {
		return fmt.Errorf("imports module unavailable")
	}
	return m.service.registerJobHandlers()
}

func validateModuleLimits(limits Limits, archiveLimits ArchiveLimits) error {
	checks := []struct {
		name  string
		value int64
	}{
		{name: "MaxCSVSourceBytes", value: limits.MaxCSVSourceBytes},
		{name: "MaxXLSXSourceBytes", value: limits.MaxXLSXSourceBytes},
		{name: "MaxRows", value: limits.MaxRows},
		{name: "MaxColumns", value: limits.MaxColumns},
		{name: "MaxCells", value: limits.MaxCells},
		{name: "DefaultMaxExtractedBytes", value: archiveLimits.DefaultMaxExtractedBytes},
		{name: "MaxCompressionRatio", value: archiveLimits.MaxCompressionRatio},
		{name: "MaxMembers", value: archiveLimits.MaxMembers},
	}
	for _, check := range checks {
		if check.value <= 0 {
			return fmt.Errorf("imports module limit %s must be positive", check.name)
		}
	}
	return nil
}
