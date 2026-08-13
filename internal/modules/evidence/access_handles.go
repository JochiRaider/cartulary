package evidence

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"

	"github.com/JochiRaider/cartulary/internal/platform/postgres"
)

// accessHandleService owns Evidence access snapshots and opaque handle state.
// It has no blob mutation, projection, revision, or object-store dependency.
type accessHandleService struct {
	accessHandles accessHandleRepository
}

type evidenceAccessRecord struct {
	IncidentID             uuid.UUID
	RecordID               uuid.UUID
	RecordRowVersion       int64
	ObjectBlobID           *uuid.UUID
	BlobMetadataVisible    bool
	StorageKey             *string
	EvidenceLifecycleState string
	UploadState            string
	FilenameSource         string
	ContentType            string
	SizeBytes              int64
	SHA256                 *string
	MediaClass             string
	PreviewKind            *string
}

type handleRecord struct {
	Token                  string
	IncidentID             uuid.UUID
	RecordID               uuid.UUID
	RecordRowVersion       int64
	ObjectBlobID           uuid.UUID
	StorageKey             string
	SessionID              uuid.UUID
	HandleKind             string
	MediaClass             string
	PreviewKind            *string
	Disposition            string
	Filename               string
	ContentType            string
	SizeBytes              int64
	SHA256                 *string
	EvidenceLifecycleState string
	UploadState            string
	ExpiresAt              time.Time
	ConsumedAt             *time.Time
}

func newAccessHandleService(pool postgres.DB) (*accessHandleService, error) {
	if pool == nil {
		return nil, errors.New("compose Evidence access handles: Postgres is required")
	}
	return &accessHandleService{accessHandles: accessHandleRepository{db: pool}}, nil
}

func (s *accessHandleService) LoadEvidenceAccess(ctx context.Context, recordID uuid.UUID) (evidenceAccessRecord, error) {
	return s.accessHandles.loadEvidence(ctx, recordID)
}

func classifyEvidenceAccess(access evidenceAccessRecord, boundObjectBlobID *uuid.UUID) string {
	if access.ObjectBlobID == nil {
		return "no_visible_blob"
	}
	if boundObjectBlobID != nil && *access.ObjectBlobID != *boundObjectBlobID {
		return "evidence_inconsistent"
	}
	if !access.BlobMetadataVisible {
		return "blob_missing"
	}
	if access.EvidenceLifecycleState == "quarantined" || access.UploadState == "quarantined" {
		return "evidence_quarantined"
	}
	switch access.UploadState {
	case "pending":
		return "blob_pending"
	case "failed":
		return "blob_failed"
	}
	if (access.EvidenceLifecycleState != "available" && access.EvidenceLifecycleState != "released") || access.UploadState != "available" {
		return "evidence_inconsistent"
	}
	return ""
}

func (s *accessHandleService) InsertHandle(ctx context.Context, handle handleRecord, issuedByUserID uuid.UUID) error {
	return s.accessHandles.insert(ctx, handle, issuedByUserID)
}

func (s *accessHandleService) LoadHandle(ctx context.Context, token string) (handleRecord, error) {
	return s.accessHandles.load(ctx, token)
}

func (s *accessHandleService) ConsumeDownloadHandle(ctx context.Context, token string, now time.Time) error {
	return s.accessHandles.consumeDownload(ctx, token, now)
}

func (s *accessHandleService) CheckHandleAccess(ctx context.Context, handle handleRecord) (string, error) {
	return s.accessHandles.checkCurrent(ctx, handle)
}
