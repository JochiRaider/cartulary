package links

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type Store struct{}

type SupersedesLink struct {
	RecordLinkID uuid.UUID
	IncidentID   uuid.UUID
	SrcRecordID  uuid.UUID
	DstRecordID  uuid.UUID
}

func NewStore() *Store {
	return &Store{}
}

func (s *Store) InsertSupersedesTx(ctx context.Context, tx pgx.Tx, incidentID uuid.UUID, replacementRecordID uuid.UUID, supersededRecordID uuid.UUID, ownerUserID uuid.UUID, now time.Time) (SupersedesLink, error) {
	var link SupersedesLink
	if err := tx.QueryRow(ctx, `
INSERT INTO record_links (
    incident_id,
    src_record_id,
    dst_record_id,
    link_type,
    provenance,
    confidence,
    owner_user_id,
    decided_at,
    created_at
)
VALUES ($1, $2, $3, 'supersedes', 'manual', NULL, $4, $5, $5)
RETURNING record_link_id, incident_id, src_record_id, dst_record_id
`, incidentID, replacementRecordID, supersededRecordID, ownerUserID, now.UTC()).Scan(&link.RecordLinkID, &link.IncidentID, &link.SrcRecordID, &link.DstRecordID); err != nil {
		return SupersedesLink{}, fmt.Errorf("insert supersedes link: %w", err)
	}
	return link, nil
}
