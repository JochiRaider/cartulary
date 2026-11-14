package evidence

import (
	"context"
	"errors"
	"reflect"
	"strings"

	"github.com/google/uuid"

	"github.com/JochiRaider/cartulary/internal/modules/evidence/blobref"
	incidentbundleprovider "github.com/JochiRaider/cartulary/internal/modules/evidence/internal/providers/incidentbundle"
	"github.com/JochiRaider/cartulary/internal/modules/incidentportability"
	"github.com/JochiRaider/cartulary/internal/modules/stagedobjects"
	"github.com/JochiRaider/cartulary/internal/platform/objectstore"
)

// IncidentBundleBlobPortability is Evidence's source-owned portability
// capability. The concrete storage adapter remains private to Evidence.
type IncidentBundleBlobPortability interface {
	ExportBlobFiles(context.Context, incidentportability.Queryer, uuid.UUID, incidentportability.MemberWriter) error
	RewriteAndStageObjectBlobs(context.Context, incidentportability.MemberReader, incidentportability.MemberWriter, uuid.UUID, string) (*stagedobjects.PreparedResources, error)
}

func NewIncidentBundleBlobPortability(store objectstore.TypedStore, staging stagedobjects.ResourceAllocator) (IncidentBundleBlobPortability, error) {
	if nilIncidentBundleStore(store) || staging == nil {
		return nil, errors.New("compose Evidence Incident Bundle blob portability: typed object store is required")
	}
	return &incidentBundleBlobPortability{
		IncidentBundleBlobPortability: incidentbundleprovider.NewBlobPortability(store, staging),
		metadataFailure:               incidentBundleSourceDescriptor().DeclaredFailure("evidence.byte_size_digest_agree"),
	}, nil
}

type incidentBundleBlobPortability struct {
	IncidentBundleBlobPortability
	metadataFailure error
}

func (p *incidentBundleBlobPortability) ExportBlobFiles(ctx context.Context, queryer incidentportability.Queryer, incidentID uuid.UUID, members incidentportability.MemberWriter) error {
	err := p.IncidentBundleBlobPortability.ExportBlobFiles(ctx, queryer, incidentID, members)
	if errors.Is(err, incidentbundleprovider.ErrBlobMetadata) {
		return p.metadataFailure
	}
	return err
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

// IncidentBundleResourceIdentity admits Evidence's logical reservation and derives
// its physical identity. The archive never supplies a destination storage key.
func IncidentBundleResourceIdentity(profileID string, resource stagedobjects.ResourceBinding, stagingID string) (string, error) {
	if profileID != "evidence" || resource.Kind != "object_blob" {
		return "", stagedobjects.ErrScopeDenied
	}
	parts := strings.Split(resource.ID, ":")
	if len(parts) != 2 {
		return "", stagedobjects.ErrScopeDenied
	}
	incident, err := uuid.Parse(parts[0])
	if err != nil || incident.String() != parts[0] {
		return "", stagedobjects.ErrScopeDenied
	}
	blob, err := uuid.Parse(parts[1])
	if err != nil || blob.String() != parts[1] {
		return "", stagedobjects.ErrScopeDenied
	}
	generation, err := uuid.Parse(stagingID)
	if err != nil {
		return "", stagedobjects.ErrScopeDenied
	}
	return blobref.ObjectBlobStorageKey(incident, blob, generation)
}
