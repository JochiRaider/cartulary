package savedviews

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
)

type savedViewRepository interface {
	create(context.Context, uuid.UUID, uuid.UUID, createRequest, time.Time) (savedViewRecord, error)
	createSystemFixture(context.Context, uuid.UUID, createRequest, time.Time) (savedViewRecord, error)
	listVisible(context.Context, uuid.UUID, uuid.UUID, listPageRequest) ([]savedViewRecord, error)
	getVisibleForUpdate(context.Context, uuid.UUID, uuid.UUID, uuid.UUID) (savedViewRecord, error)
	patchVisible(
		context.Context,
		uuid.UUID,
		uuid.UUID,
		uuid.UUID,
		func(savedViewRecord) (savedViewRecord, bool, error),
	) (savedViewRecord, error)
	deleteVisible(context.Context, uuid.UUID, uuid.UUID, uuid.UUID, func(savedViewRecord) error) error
}

type savedViewApplication struct {
	repository savedViewRepository
}

func newSavedViewApplication(repository savedViewRepository) *savedViewApplication {
	return &savedViewApplication{repository: repository}
}

func (s *savedViewApplication) create(
	ctx context.Context,
	incidentID uuid.UUID,
	actorUserID uuid.UUID,
	request createRequest,
	now time.Time,
) (savedViewRecord, error) {
	return s.repository.create(ctx, actorUserID, incidentID, request, now)
}

func (s *savedViewApplication) createSystemFixture(
	ctx context.Context,
	incidentID uuid.UUID,
	request createRequest,
	now time.Time,
) (savedViewRecord, error) {
	return s.repository.createSystemFixture(ctx, incidentID, request, now)
}

func (s *savedViewApplication) listVisible(
	ctx context.Context,
	incidentID uuid.UUID,
	actorUserID uuid.UUID,
	page listPageRequest,
) ([]savedViewRecord, error) {
	return s.repository.listVisible(ctx, incidentID, actorUserID, page)
}

func (s *savedViewApplication) visibleForPatch(
	ctx context.Context,
	incidentID uuid.UUID,
	savedViewID uuid.UUID,
	actorUserID uuid.UUID,
) (savedViewRecord, error) {
	return s.repository.getVisibleForUpdate(ctx, incidentID, savedViewID, actorUserID)
}

func (s *savedViewApplication) patch(
	ctx context.Context,
	incidentID uuid.UUID,
	savedViewID uuid.UUID,
	actorUserID uuid.UUID,
	membershipRole string,
	request patchRequest,
	now time.Time,
) (savedViewRecord, error) {
	return s.repository.patchVisible(ctx, incidentID, savedViewID, actorUserID, func(current savedViewRecord) (savedViewRecord, bool, error) {
		if !canMutate(current, actorUserID, membershipRole) {
			return savedViewRecord{}, false, errSavedViewMutationDenied
		}
		if current.SavedViewVersion != request.BaseSavedViewVersion {
			return savedViewRecord{}, false, &savedViewVersionConflictError{
				SavedViewID:             current.SavedViewID,
				BaseSavedViewVersion:    request.BaseSavedViewVersion,
				CurrentSavedViewVersion: current.SavedViewVersion,
			}
		}
		next, changed, err := applyPatch(current, request, now)
		if err != nil {
			return savedViewRecord{}, false, fmt.Errorf("apply saved view patch: %w", err)
		}
		return next, changed, nil
	})
}

func (s *savedViewApplication) delete(
	ctx context.Context,
	incidentID uuid.UUID,
	savedViewID uuid.UUID,
	actorUserID uuid.UUID,
	membershipRole string,
) error {
	return s.repository.deleteVisible(ctx, incidentID, savedViewID, actorUserID, func(current savedViewRecord) error {
		if !canMutate(current, actorUserID, membershipRole) {
			return errSavedViewMutationDenied
		}
		return nil
	})
}
