package workbook

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/JochiRaider/cartulary/internal/modules/incidents/admission"
	workbookstartup "github.com/JochiRaider/cartulary/internal/modules/workbook/startup"
	"github.com/JochiRaider/cartulary/internal/platform/authn"
	"github.com/JochiRaider/cartulary/internal/platform/httpapi"
	"github.com/JochiRaider/cartulary/internal/platform/pagination"
)

type service struct {
	noteAssociations NoteAssociationProvider
	contributions    *WorkbookContributionCatalog
	recordTargets    RecordTargetResolver
	conflictTokens   ConflictTokenDecoder
	incidentAccess   *admission.Checker
	startupStore     *workbookstartup.Store
	authStore        *authn.Store
	cursorCodec      *pagination.Codec
	now              func() time.Time
	serviceVersion   string
}

type StartupStoreFactory func(httpapi.DependencySet) (*workbookstartup.Store, error)

type RouteDependencies struct {
	NoteAssociations    NoteAssociationProvider
	Catalog             *WorkbookContributionCatalog
	RecordTargets       RecordTargetResolver
	ConflictTokens      ConflictTokenDecoder
	StartupStoreFactory StartupStoreFactory
}

// RegisterRoutes binds the workbook routes for central admission (Core 01
// REQ-01-707): the registry authenticates by each operation's declared
// security and admits the incident, record, view-schema and conflict-token
// path parameters before a handler runs.
func RegisterRoutes(routeDependencies RouteDependencies) httpapi.RouteRegistrar {
	return func(mux *http.ServeMux, deps httpapi.DependencySet) error {
		if routeDependencies.StartupStoreFactory == nil {
			return errors.New("workbook route composition requires a startup store factory")
		}
		startupStore, err := routeDependencies.StartupStoreFactory(deps)
		if err != nil {
			return fmt.Errorf("compose workbook startup store: %w", err)
		}
		if startupStore == nil {
			return errors.New("workbook startup store factory returned nil")
		}
		if isNilContributionProvider(routeDependencies.NoteAssociations) {
			return errors.New("workbook route composition requires Note associations")
		}
		service, err := newService(
			deps,
			routeDependencies,
			startupStore,
		)
		if err != nil {
			return err
		}
		return httpapi.BindAdmittedOwnerRoutes(mux, deps, "module.workbook", map[string]http.HandlerFunc{
			"applyWorkbookBulkMutation":             service.handleBulkMutations,
			"createRecordLinkedNote":                service.handleLinkedNoteCreate,
			"createViewRow":                         service.handleCreate,
			"getCurrentUserWorkbookPreferences":     service.handleGetUserWorkbookPreferences,
			"getIncidentDefaultWorkbookPreferences": service.handleGetDefaultWorkbookPreferences,
			"getIncidentWorkbookStartup":            service.handleWorkbookStartup,
			"listNoteAssociations":                  service.handleListNoteAssociations,
			"locateWorkbookViewRecord":              service.handleLocate,
			"mutateNoteAssociations":                service.handleMutateNoteAssociations,
			"patchRecord":                           service.handlePatch,
			"pasteWorkbookClipboard":                service.handleClipboardPaste,
			"putCurrentUserWorkbookPreferences":     service.handlePutUserWorkbookPreferences,
			"putIncidentDefaultWorkbookPreferences": service.handlePutDefaultWorkbookPreferences,
			"queryWorkbookView":                     service.handleQuery,
			"resolveRecordSameFieldConflict":        service.handleConflictResolve,
			"supersedeRecord":                       service.handleSupersede,
			// Every resolved startup response, failures included, carries
			// Cache-Control: no-store (Core 01 REQ-01-151.1).
		}, httpapi.NoStore("getIncidentWorkbookStartup"))
	}
}

func newService(
	deps httpapi.DependencySet,
	routeDependencies RouteDependencies,
	startupStore *workbookstartup.Store,
) (*service, error) {
	now := deps.Now
	if now == nil {
		now = func() time.Time { return time.Now().UTC() }
	}
	cursorCodec := deps.CursorCodec
	if cursorCodec == nil {
		keys, err := authn.LoadMasterKeys(deps.Env)
		if err != nil {
			return nil, err
		}
		cursorKey, err := authn.DerivePurposeKey(keys, "pagination-cursor-v2")
		if err != nil {
			return nil, err
		}
		cursorCodec = pagination.NewCodec(cursorKey)
	}
	if routeDependencies.Catalog == nil {
		return nil, errors.New("workbook route composition requires a contribution catalog")
	}
	if isNilContributionProvider(routeDependencies.RecordTargets) {
		return nil, errors.New("workbook route composition requires a record target owner")
	}
	if isNilContributionProvider(routeDependencies.ConflictTokens) {
		return nil, errors.New("workbook route composition requires a conflict token decoder")
	}
	if startupStore == nil {
		return nil, errors.New("workbook route composition requires a startup store")
	}
	return &service{
		noteAssociations: routeDependencies.NoteAssociations,
		contributions:    routeDependencies.Catalog,
		recordTargets:    routeDependencies.RecordTargets,
		conflictTokens:   routeDependencies.ConflictTokens,
		incidentAccess:   admission.NewChecker(deps.PostgresHandle()),
		startupStore:     startupStore,
		authStore:        authn.NewStore(deps.PostgresHandle()),
		cursorCodec:      cursorCodec,
		now:              now,
		serviceVersion:   deps.Telemetry.ServiceVersion,
	}, nil
}
