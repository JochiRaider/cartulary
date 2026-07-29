package imports

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/JochiRaider/cartulary/internal/modules/imports/ownerfacade"
	"github.com/JochiRaider/cartulary/internal/modules/incidents"
	"github.com/JochiRaider/cartulary/internal/modules/revisions"
	"github.com/JochiRaider/cartulary/internal/platform/authn"
	"github.com/JochiRaider/cartulary/internal/platform/httpapi"
	"github.com/JochiRaider/cartulary/internal/platform/jobs"
	"github.com/JochiRaider/cartulary/internal/platform/pagination"
)

type Service struct {
	store                    *Store
	incidentAccess           incidents.Access
	authStore                *authn.Store
	jobManager               *jobs.Manager
	jobRunner                *jobs.Runner
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

type RouteOption func(*routeOptions)

type routeOptions struct {
	extensionProfileAdmitted func(string) bool
	jobSuccessFinalizer      JobSuccessFinalizer
	limits                   Limits
	archiveLimits            ArchiveLimits
	ownerCreateRegistry      *ownerfacade.ImportOwnerCreateRegistry
	revisionAppender         *revisions.Appender
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
		return httpapi.BindOwnerRoutes(mux, deps, "module.imports", map[string]http.HandlerFunc{
			"applyImportSession":                service.handleImportSessionsMember,
			"createImportUnitRegion":            service.handleImportSessionsMember,
			"createImportSession":               service.handleImportSessionsCollection,
			"getImportSession":                  service.handleImportSessionsMember,
			"getImportUnit":                     service.handleImportSessionsMember,
			"getImportUnitPreview":              service.handleImportSessionsMember,
			"listImportUnits":                   service.handleImportSessionsMember,
			"previewImportUnitExtensionMapping": service.handleImportSessionsMember,
			"putImportUnitMapping":              service.handleImportSessionsMember,
			"selectImportUnit":                  service.handleImportSessionsMember,
			"skipImportUnit":                    service.handleImportSessionsMember,
		})
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
	if deps.Jobs != nil && deps.JobTransactions == nil {
		return nil, fmt.Errorf("import admitted route requires the Jobs transaction service")
	}
	extensionImportFacades, err := extensionImportFacadesFromDependencies(deps)
	if err != nil {
		return nil, err
	}
	if deps.Jobs != nil && options.jobSuccessFinalizer == nil {
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
			deps.JobTransactions,
		),
		incidentAccess:           incidents.NewAccess(deps.PostgresHandle()),
		authStore:                authn.NewStore(deps.PostgresHandle()),
		jobManager:               deps.Jobs,
		jobRunner:                deps.JobRunner,
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
	if err := service.recoverImportJobs(context.Background()); err != nil {
		return nil, err
	}
	return service, nil
}
