package links

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

// AssociationLookup is a bounded keyset read of explicit-route (null-field)
// links. The source owner decides which relationship semantics to expose.
type AssociationLookup struct {
	IncidentID uuid.UUID
	RecordID   uuid.UUID
	LinkType   LinkType
	Incoming   bool
	Outgoing   bool
	Limit      int
	BeforeTime *time.Time
	BeforeID   uuid.UUID
}

type AssociationLink struct {
	ID            uuid.UUID
	SourceID      uuid.UUID
	TargetID      uuid.UUID
	CounterpartID uuid.UUID
	CreatedAt     time.Time
}

func (s *Store) ListAssociationsTx(ctx context.Context, tx pgx.Tx, q AssociationLookup) ([]AssociationLink, error) {
	if q.Limit < 1 || q.Limit > 501 || (!q.Incoming && !q.Outgoing) || q.LinkType == LinkTypeInvalid {
		return nil, fmt.Errorf("links: invalid association window")
	}
	// UNION ALL allows each direction to use its own bounded lookup index.
	rows, err := tx.Query(ctx, `
SELECT l.record_link_id, l.src_record_id, l.dst_record_id, CASE WHEN l.src_record_id=$2 THEN l.dst_record_id ELSE l.src_record_id END, l.created_at
FROM (
 SELECT * FROM record_links WHERE $4 AND incident_id=$1 AND src_record_id=$2 AND link_type=$3
   AND field_key IS NULL AND deleted_at IS NULL
 UNION ALL
 SELECT * FROM record_links WHERE $5 AND incident_id=$1 AND dst_record_id=$2 AND link_type=$3
   AND field_key IS NULL AND deleted_at IS NULL
) l
WHERE ($6::timestamptz IS NULL OR (l.created_at,l.record_link_id)<($6,$7::uuid))
ORDER BY l.created_at DESC,l.record_link_id DESC LIMIT $8`, q.IncidentID, q.RecordID, q.LinkType.String(), q.Outgoing, q.Incoming, q.BeforeTime, q.BeforeID, q.Limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	result := make([]AssociationLink, 0)
	for rows.Next() {
		var link AssociationLink
		if err := rows.Scan(&link.ID, &link.SourceID, &link.TargetID, &link.CounterpartID, &link.CreatedAt); err != nil {
			return nil, err
		}
		result = append(result, link)
	}
	return result, rows.Err()
}

func (s *Store) LoadAssociationTx(ctx context.Context, tx pgx.Tx, incidentID, linkID uuid.UUID) (AssociationLink, LinkType, error) {
	var link AssociationLink
	var kind string
	err := tx.QueryRow(ctx, `SELECT record_link_id,src_record_id,dst_record_id,created_at,link_type FROM record_links
WHERE incident_id=$1 AND record_link_id=$2 AND field_key IS NULL AND deleted_at IS NULL`, incidentID, linkID).Scan(&link.ID, &link.SourceID, &link.TargetID, &link.CreatedAt, &kind)
	if err != nil {
		return AssociationLink{}, LinkTypeInvalid, err
	}
	linkType, err := ParseLinkType(kind)
	return link, linkType, err
}

// EnsureAssociationTx deliberately preserves an existing link's provenance.
// Callers hold endpoint locks in deterministic order for the entire operation.
func (s *Store) EnsureAssociationTx(ctx context.Context, tx pgx.Tx, command UpsertLinkCommand) (RecordLinkCommandResult, error) {
	var id uuid.UUID
	err := tx.QueryRow(ctx, `SELECT record_link_id FROM record_links WHERE incident_id=$1 AND src_record_id=$2
AND dst_record_id=$3 AND link_type=$4 AND field_key IS NULL AND deleted_at IS NULL`, command.IncidentID, command.SrcRecordID, command.DstRecordID, command.LinkType.String()).Scan(&id)
	if err == nil {
		return RecordLinkCommandResult{RecordLinkID: id, SrcRecordID: command.SrcRecordID, DstRecordID: command.DstRecordID, LinkType: command.LinkType}, nil
	}
	if err != pgx.ErrNoRows {
		return RecordLinkCommandResult{}, err
	}
	return s.UpsertLinkCommandTx(ctx, tx, command)
}
