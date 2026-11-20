package evidence

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"

	"github.com/JochiRaider/cartulary/internal/modules/evidence/blobref"
	"github.com/JochiRaider/cartulary/internal/platform/authn"
	"github.com/JochiRaider/cartulary/internal/platform/httpapi"
	"github.com/JochiRaider/cartulary/internal/platform/objectstore"
)

type service struct {
	operations     routeService
	verification   *blobVerificationService
	admission      routeAdmission
	objects        routeObjectStoreAdapter
	uploads        uploadCapabilityService
	keys           authn.MasterKeys
	now            func() time.Time
	maxBlobBytes   int64
	previewMax     int64
	textPreviewMax int64
	// storageKeys generates a new slot's physical key. Nil selects the
	// production object_blob_storage_key_v2 generator; tests substitute a
	// defective generator to exercise the server-generated-key contract.
	storageKeys func(incidentID, objectBlobID uuid.UUID) (string, error)
}

// objectBlobStorageKey generates the physical key for a new blob slot.
func (s *service) objectBlobStorageKey(incidentID, objectBlobID uuid.UUID) (string, error) {
	if s.storageKeys != nil {
		return s.storageKeys(incidentID, objectBlobID)
	}
	return blobref.ObjectBlobStorageKey(incidentID, objectBlobID, uuid.New())
}

type Settings struct {
	MaxBlobBytes   int64
	PreviewMax     int64
	TextPreviewMax int64
}

type routeDependencies struct {
	objectStore  objectstore.TypedStore
	verification *blobVerificationService
	now          func() time.Time
	operations   routeService
	admission    routeAdmission
}

type routeService interface {
	CreateBlobSlot(context.Context, blobSlotParams) (blobSlotResult, error)
	GetBlob(context.Context, uuid.UUID) (blobRecord, error)
	GetUploadLease(context.Context, uuid.UUID) (uploadLeaseRecord, error)
	ClaimUploadLease(context.Context, uuid.UUID, []byte, time.Time) error
	CompleteUploadLease(context.Context, uuid.UUID, time.Time) error
	LoadAttachTarget(context.Context, uuid.UUID) (evidenceAttachTarget, error)
	PreflightAttachBlob(context.Context, authn.UserRecord, uuid.UUID, attachBlobRequest, []byte, time.Time) (attachBlobPreflightResult, error)
	AttachBlob(context.Context, authn.UserRecord, uuid.UUID, attachBlobRequest, []byte, string, time.Time) (attachBlobResult, error)
	LoadEvidenceAccess(context.Context, uuid.UUID) (evidenceAccessRecord, error)
	LoadEvidenceMetadata(context.Context, uuid.UUID, []uuid.UUID) ([]evidenceAccessRecord, error)
	InsertHandle(context.Context, handleRecord, uuid.UUID) error
	LoadHandle(context.Context, string) (handleRecord, error)
	CheckHandleAccess(context.Context, handleRecord) (string, error)
	ConsumeDownloadHandle(context.Context, string, time.Time) error
}

var _ routeService = (*routeOperations)(nil)

// registerRoutes binds the Evidence routes for central admission (Core 01
// REQ-01-707): the registry authenticates by each operation's declared
// security and admits its path parameters before a handler runs. Handle and
// upload tokens are admitted by syntax only; their capability checks stay in
// route handling.
func registerRoutes(settings Settings, captured routeDependencies) httpapi.RouteRegistrar {
	return func(mux *http.ServeMux, deps httpapi.DependencySet) error {
		service, err := newService(deps.Env, settings, captured)
		if err != nil {
			return err
		}
		return httpapi.BindAdmittedOwnerRoutes(mux, deps, "module.evidence", service.operationHandlers())
	}
}

// operationHandlers maps each Evidence operation to its one handler.
func (s *service) operationHandlers() map[string]http.HandlerFunc {
	return map[string]http.HandlerFunc{
		"attachBlobToEvidenceRecord":  s.handleAttachBlob,
		"createObjectBlobSlot":        s.handleCreateBlob,
		"verifyObjectBlob":            s.handleVerifyBlob,
		"issueEvidenceDownloadHandle": s.handleDownloadHandle,
		"issueEvidencePreviewHandle":  s.handlePreviewHandle,
		"queryEvidenceMetadata":       s.handleMetadataQuery,
		"redeemEvidenceHandle":        s.handleRedeemHandle,
		"uploadObjectBlobContent":     s.handleUploadTarget,
	}
}

func newService(env map[string]string, settings Settings, dependencies routeDependencies) (*service, error) {
	keys, err := authn.LoadMasterKeys(env)
	if err != nil {
		return nil, err
	}
	if dependencies.verification == nil {
		return nil, fmt.Errorf("compose Evidence routes: background verification is required")
	}
	if isNilCapability(dependencies.objectStore) {
		return nil, fmt.Errorf("compose Evidence routes: object store is required")
	}
	if dependencies.now == nil {
		return nil, fmt.Errorf("compose Evidence routes: clock is required")
	}
	if isNilCapability(dependencies.operations) {
		return nil, fmt.Errorf("compose Evidence routes: operations are required")
	}
	if dependencies.admission.incidents == nil || isNilCapability(dependencies.admission.auth) {
		return nil, fmt.Errorf("compose Evidence routes: admission is required")
	}
	return &service{
		operations:     dependencies.operations,
		verification:   dependencies.verification,
		admission:      dependencies.admission,
		objects:        routeObjectStoreAdapter{store: dependencies.objectStore},
		uploads:        uploadCapabilityService{keys: keys},
		keys:           keys,
		now:            dependencies.now,
		maxBlobBytes:   settings.MaxBlobBytes,
		previewMax:     settings.PreviewMax,
		textPreviewMax: settings.TextPreviewMax,
	}, nil
}
