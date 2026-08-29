package evidence

import (
	"context"
	"errors"
	"reflect"

	"github.com/google/uuid"

	incidentbundleprovider "github.com/JochiRaider/cartulary/internal/modules/evidence/internal/providers/incidentbundle"
	"github.com/JochiRaider/cartulary/internal/modules/incidentportability"
	"github.com/JochiRaider/cartulary/internal/platform/objectstore"
)

// IncidentBundleBlobPortability is Evidence's source-owned portability
// capability. The concrete storage adapter remains private to Evidence.
type IncidentBundleBlobPortability interface {
	ExportBlobFiles(context.Context, incidentportability.Queryer, uuid.UUID, map[string][]byte) error
	RewriteAndStageObjectBlobs(context.Context, map[string][]byte, uuid.UUID, uuid.UUID, incidentportability.AttributionRecorder) ([]byte, []string, error)
	CleanupStagedObjects(context.Context, []string)
}

func NewIncidentBundleBlobPortability(store objectstore.TypedStore) (IncidentBundleBlobPortability, error) {
	if nilIncidentBundleStore(store) {
		return nil, errors.New("compose Evidence Incident Bundle blob portability: typed object store is required")
	}
	return incidentbundleprovider.NewBlobPortability(store), nil
}

func nilIncidentBundleStore(store objectstore.TypedStore) bool {
	if store == nil {
		return true
	}
	value := reflect.ValueOf(store)
	switch value.Kind() {
	case reflect.Chan, reflect.Func, reflect.Interface, reflect.Map, reflect.Pointer, reflect.Slice:
		return value.IsNil()
	default:
		return false
	}
}
