package links

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/modules/links/valuecodec"
)

func (s *Store) LoadRecordLinkValueTx(ctx context.Context, tx pgx.Tx, recordLinkID uuid.UUID) (map[string]any, error) {
	value, err := valuecodec.LoadRecordLinkValueTx(ctx, tx, recordLinkID)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrRecordLinkNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("load record link value: %w", err)
	}
	return value, nil
}

func (s *Store) LoadRecordTagValueTx(ctx context.Context, tx pgx.Tx, recordTagID uuid.UUID) (map[string]any, error) {
	value, err := valuecodec.LoadRecordTagValueTx(ctx, tx, recordTagID)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrTagNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("load record tag value: %w", err)
	}
	return value, nil
}

func compactRecordLinkMutationValue(record RecordLink) map[string]any {
	return map[string]any{
		"record_link_id": record.RecordLinkID.String(),
		"incident_id":    record.IncidentID.String(),
		"src_record_id":  record.SrcRecordID.String(),
		"dst_record_id":  record.DstRecordID.String(),
		"link_type":      record.LinkType,
		"provenance":     record.Provenance,
		"confidence":     record.Confidence,
		"deleted_at":     formatMutationTimestampPointer(record.DeletedAt),
	}
}

func compactRecordTagMutationValue(recordTagID uuid.UUID, incidentID uuid.UUID, recordID uuid.UUID, tagName string, normalizedTagName string, deletedAt *time.Time, deletedByUserID *uuid.UUID) map[string]any {
	return map[string]any{
		"record_tag_id":       recordTagID.String(),
		"incident_id":         incidentID.String(),
		"record_id":           recordID.String(),
		"tag_name":            tagName,
		"normalized_tag_name": normalizedTagName,
		"deleted_at":          formatMutationTimestampPointer(deletedAt),
		"deleted_by_user_id":  formatMutationUUIDPointer(deletedByUserID),
	}
}

func formatMutationTimestampPointer(value *time.Time) any {
	if value == nil {
		return nil
	}
	return value.UTC().Format(time.RFC3339Nano)
}

func formatMutationUUIDPointer(value *uuid.UUID) any {
	if value == nil {
		return nil
	}
	return value.String()
}
