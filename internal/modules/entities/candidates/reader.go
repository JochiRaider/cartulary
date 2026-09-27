// Package candidates owns non-mutating, incident-scoped entity discovery.
package candidates

import (
	"context"
	"errors"
	"strings"

	"github.com/JochiRaider/cartulary/internal/platform/listquery"
	"github.com/JochiRaider/cartulary/internal/platform/pagination"
	"github.com/JochiRaider/cartulary/internal/platform/postgres"
	"github.com/google/uuid"
)

type Candidate struct {
	RecordID    uuid.UUID `json:"record_id"`
	RowVersion  int64     `json:"row_version"`
	EntityType  string    `json:"entity_type"`
	DisplayName string    `json:"display_name"`
}
type Position struct {
	DisplayName string
	RecordID    uuid.UUID
}
type Request struct {
	IncidentID uuid.UUID
	EntityType string
	Search     string // canonical list_search_v1 tokens, validated by the route
	Limit      int
	After      *Position
}
type Page struct {
	Candidates []Candidate
	Next       *Position
}
type PageReader interface {
	Page(context.Context, Request) (Page, error)
}
type Reader struct{ db postgres.DB }

func NewReader(db postgres.DB) *Reader { return &Reader{db: db} }

// Page retains at most limit+1 matches and one bounded source batch. A sparse
// predicate scans until exhaustion; cancellation never becomes an empty result.
func (r *Reader) Page(ctx context.Context, request Request) (Page, error) {
	if request.Limit < 1 || request.Limit > pagination.MaxLimit || (request.EntityType != "host" && request.EntityType != "identity") {
		return Page{}, errors.New("invalid candidate request")
	}
	query := hostBatch
	if request.EntityType == "identity" {
		query = identityBatch
	}
	batchSize := max(request.Limit, pagination.DefaultLimit)
	after := request.After
	matches := make([]Candidate, 0, request.Limit+1)
	tokens := strings.Fields(request.Search)
	for {
		if err := ctx.Err(); err != nil {
			return Page{}, err
		}
		var name *string
		var id *uuid.UUID
		if after != nil {
			name = &after.DisplayName
			id = &after.RecordID
		}
		rows, err := r.db.Query(ctx, query, request.IncidentID, name, id, batchSize)
		if err != nil {
			return Page{}, err
		}
		count := 0
		for rows.Next() {
			count++
			var candidate Candidate
			var values []string
			if err := rows.Scan(&candidate.RecordID, &candidate.RowVersion, &candidate.DisplayName, &values); err != nil {
				rows.Close()
				return Page{}, err
			}
			candidate.EntityType = request.EntityType
			after = &Position{DisplayName: candidate.DisplayName, RecordID: candidate.RecordID}
			if listquery.MatchSearchTokens(tokens, values...) {
				matches = append(matches, candidate)
			}
			if len(matches) > request.Limit {
				break
			}
		}
		err = rows.Err()
		rows.Close()
		if err != nil {
			return Page{}, err
		}
		if len(matches) > request.Limit {
			last := matches[request.Limit-1]
			return Page{Candidates: matches[:request.Limit], Next: &Position{DisplayName: last.DisplayName, RecordID: last.RecordID}}, nil
		}
		if count < batchSize {
			return Page{Candidates: matches}, nil
		}
	}
}

// Both ORDER BY and the keyset tuple use the same native text comparison as
// entity workbook views. Values are hydrated only for the bounded source page.
const hostBatch = `
WITH batch AS (
 SELECT e.record_id, r.row_version, e.display_name,
        ARRAY[e.display_name,e.hostname,e.fqdn,e.aad_device_id] AS values
 FROM hosts e JOIN records r ON r.record_id=e.record_id AND r.incident_id=e.incident_id
 WHERE e.incident_id=$1 AND r.deleted_at IS NULL
   AND e.host_state IN ('stub','canonical') AND e.merged_into_record_id IS NULL
   AND ($2::text IS NULL OR (e.display_name,e.record_id)>($2::text,$3::uuid))
 ORDER BY e.display_name,e.record_id LIMIT $4
)
SELECT b.record_id,b.row_version,b.display_name,
 array_remove(b.values,NULL) || ARRAY(
  SELECT raw_text FROM entity_aliases WHERE incident_id=$1 AND record_id=b.record_id AND entity_type='host' AND deleted_at IS NULL
  UNION ALL
  SELECT raw_value FROM entity_preserved_identifiers WHERE incident_id=$1 AND record_id=b.record_id AND entity_type='host' AND deleted_at IS NULL AND classification IN ('exact_match_reuse','suggestion_only')
 ) FROM batch b ORDER BY b.display_name,b.record_id`
const identityBatch = `
WITH batch AS (
 SELECT e.record_id, r.row_version, e.display_name,
        ARRAY[e.display_name,e.upn,e.email::text,e.sam_account_name,e.sid,e.aad_object_id] AS values
 FROM identities e JOIN records r ON r.record_id=e.record_id AND r.incident_id=e.incident_id
 WHERE e.incident_id=$1 AND r.deleted_at IS NULL
   AND e.identity_state IN ('stub','canonical') AND e.merged_into_record_id IS NULL
   AND ($2::text IS NULL OR (e.display_name,e.record_id)>($2::text,$3::uuid))
 ORDER BY e.display_name,e.record_id LIMIT $4
)
SELECT b.record_id,b.row_version,b.display_name,
 array_remove(b.values,NULL) || ARRAY(
  SELECT raw_text FROM entity_aliases WHERE incident_id=$1 AND record_id=b.record_id AND entity_type='identity' AND deleted_at IS NULL
  UNION ALL
  SELECT raw_value FROM entity_preserved_identifiers WHERE incident_id=$1 AND record_id=b.record_id AND entity_type='identity' AND deleted_at IS NULL AND classification IN ('exact_match_reuse','suggestion_only')
 ) FROM batch b ORDER BY b.display_name,b.record_id`
