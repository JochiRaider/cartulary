package evidence

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"

	"github.com/JochiRaider/cartulary/internal/platform/authn"
	"github.com/JochiRaider/cartulary/internal/platform/httpapi"
	"github.com/JochiRaider/cartulary/internal/platform/objectstore"
)

type service struct {
	operations     routeService
	admission      routeAdmission
	objects        routeObjectStoreAdapter
	uploads        uploadCapabilityService
	keys           authn.MasterKeys
	now            func() time.Time
	maxBlobBytes   int64
	previewMax     int64
	textPreviewMax int64
}

type Settings struct {
	MaxBlobBytes   int64
	PreviewMax     int64
	TextPreviewMax int64
}

type routeDependencies struct {
	objectStore objectstore.TypedStore
	now         func() time.Time
	operations  routeService
	admission   routeAdmission
}

type routeService interface {
	CreateBlobSlot(context.Context, blobSlotParams) (blobSlotResult, error)
	GetBlob(context.Context, uuid.UUID) (blobRecord, error)
	GetUploadLease(context.Context, uuid.UUID) (uploadLeaseRecord, error)
	ClaimUploadLease(context.Context, uuid.UUID, []byte, time.Time) error
	CompleteUploadLease(context.Context, uuid.UUID, time.Time) error
	PreflightAttachBlob(context.Context, authn.UserRecord, uuid.UUID, attachBlobRequest, []byte, time.Time) (attachBlobPreflightResult, error)
	AttachBlob(context.Context, authn.UserRecord, uuid.UUID, attachBlobRequest, []byte, *observedObject, string, time.Time) (attachBlobResult, error)
	LoadEvidenceAccess(context.Context, uuid.UUID) (evidenceAccessRecord, error)
	InsertHandle(context.Context, handleRecord, uuid.UUID) error
	LoadHandle(context.Context, string) (handleRecord, error)
	CheckHandleAccess(context.Context, handleRecord) (string, error)
	ConsumeDownloadHandle(context.Context, string, time.Time) error
}

var _ routeService = (*routeOperations)(nil)

func registerRoutes(settings Settings, captured routeDependencies) httpapi.RouteRegistrar {
	return func(mux *http.ServeMux, deps httpapi.DependencySet) error {
		service, err := newService(deps.Env, settings, captured)
		if err != nil {
			return err
		}
		return httpapi.BindOwnerRoutes(mux, deps, "module.evidence", map[string]http.HandlerFunc{
			"attachBlobToEvidenceRecord":  service.handleAttachBlob,
			"createObjectBlobSlot":        service.handleCreateBlob,
			"issueEvidenceDownloadHandle": service.handleDownloadHandle,
			"issueEvidencePreviewHandle":  service.handlePreviewHandle,
			"redeemEvidenceHandle":        service.handleRedeemHandle,
			"uploadObjectBlobContent":     service.handleUploadTarget,
		})
	}
}

func newService(env map[string]string, settings Settings, dependencies routeDependencies) (*service, error) {
	keys, err := authn.LoadMasterKeys(env)
	if err != nil {
		return nil, err
	}
	if dependencies.objectStore == nil {
		return nil, fmt.Errorf("compose Evidence routes: object store is required")
	}
	if dependencies.now == nil {
		return nil, fmt.Errorf("compose Evidence routes: clock is required")
	}
	if dependencies.operations == nil {
		return nil, fmt.Errorf("compose Evidence routes: operations are required")
	}
	if dependencies.admission.incidents == nil || dependencies.admission.auth == nil {
		return nil, fmt.Errorf("compose Evidence routes: admission is required")
	}
	admission := dependencies.admission
	admission.keys = keys
	return &service{
		operations:     dependencies.operations,
		admission:      admission,
		objects:        routeObjectStoreAdapter{store: dependencies.objectStore},
		uploads:        uploadCapabilityService{keys: keys},
		keys:           keys,
		now:            dependencies.now,
		maxBlobBytes:   settings.MaxBlobBytes,
		previewMax:     settings.PreviewMax,
		textPreviewMax: settings.TextPreviewMax,
	}, nil
}
