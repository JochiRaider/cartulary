package savedviews

import (
	"context"
	"io"
	"time"

	"github.com/google/uuid"

	"github.com/JochiRaider/cartulary/internal/platform/authn"
	"github.com/JochiRaider/cartulary/internal/platform/httpapi"
	"github.com/JochiRaider/cartulary/internal/platform/postgres"
)

type CreateRequest = createRequest
type PatchRequest = patchRequest
type SavedViewVersionConflictError = savedViewVersionConflictError

const ScopeShared = scopeShared

func ParseScope(value string) (scope, bool) {
	return parseScope(value)
}

func DecodeCreateRequest(reader io.Reader) (createRequest, *httpapi.APIError) {
	return decodeCreateRequest(reader)
}

func DecodePatchRequest(reader io.Reader, viewSchemaID string) (patchRequest, *httpapi.APIError) {
	return decodePatchRequest(reader, viewSchemaID)
}

type Store struct {
	application *savedViewApplication
}

func NewStore(pool postgres.DB) *Store {
	return &Store{application: newSavedViewApplication(newPostgresSavedViewRepository(pool))}
}

func (s *Store) Create(
	ctx context.Context,
	actor authn.UserRecord,
	incidentID uuid.UUID,
	request createRequest,
	now time.Time,
) (savedViewRecord, error) {
	return s.application.create(ctx, incidentID, actor.ID, request, now)
}

func (s *Store) Patch(
	ctx context.Context,
	actor authn.UserRecord,
	membershipRole string,
	incidentID uuid.UUID,
	savedViewID uuid.UUID,
	request patchRequest,
	now time.Time,
) (savedViewRecord, error) {
	return s.application.patch(ctx, incidentID, savedViewID, actor.ID, membershipRole, request, now)
}
