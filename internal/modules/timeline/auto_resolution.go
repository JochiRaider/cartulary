package timeline

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/JochiRaider/cartulary/internal/platform/fieldnorm"
)

const autoResolutionMethod = "auto_match"

var autoResolutionSuppressorTokens = map[string]struct{}{
	"maybe":         {},
	"prob":          {},
	"probably":      {},
	"approx":        {},
	"approximately": {},
}

type autoResolutionMatch struct {
	RecordID  uuid.UUID
	AliasText string
}

type autoResolutionAliasCandidate struct {
	RecordID uuid.UUID
	RawText  string
}

type collectionLinkMetadata struct {
	Provenance string
	Confidence *int
}

func lookupInteractiveAutoResolutionMatchTx(ctx context.Context, tx pgx.Tx, incidentID uuid.UUID, fieldKey string, rawText string) (*autoResolutionMatch, error) {
	entityType, ok := timelineRelationshipEntityType(fieldKey)
	if !ok {
		return nil, nil
	}

	candidateText, ok := fieldnorm.AutoResolutionCandidateText(rawText)
	if !ok || autoResolutionSuppressed(candidateText) {
		return nil, nil
	}

	aliases, err := loadActiveAutoResolutionAliasesTx(ctx, tx, incidentID, entityType)
	if err != nil {
		return nil, err
	}

	matches := make(map[uuid.UUID]string)
	for _, alias := range aliases {
		aliasCandidateText, ok := fieldnorm.AutoResolutionCandidateText(alias.RawText)
		if !ok || aliasCandidateText != candidateText {
			continue
		}
		if _, exists := matches[alias.RecordID]; !exists {
			matches[alias.RecordID] = alias.RawText
		}
	}
	if len(matches) != 1 {
		return nil, nil
	}

	for recordID, aliasText := range matches {
		return &autoResolutionMatch{RecordID: recordID, AliasText: aliasText}, nil
	}
	return nil, nil
}

func autoResolutionSuppressed(candidateText string) bool {
	if strings.ContainsAny(candidateText, "?~") {
		return true
	}
	for _, token := range strings.Fields(candidateText) {
		if _, ok := autoResolutionSuppressorTokens[token]; ok {
			return true
		}
	}
	return false
}

func loadActiveAutoResolutionAliasesTx(ctx context.Context, tx pgx.Tx, incidentID uuid.UUID, entityType string) ([]autoResolutionAliasCandidate, error) {
	var (
		query string
		args  = []any{incidentID, entityType}
	)

	switch entityType {
	case "host":
		query = `
SELECT ea.record_id, ea.raw_text
  FROM entity_aliases ea
  JOIN hosts h
    ON h.record_id = ea.record_id
   AND h.incident_id = ea.incident_id
 WHERE ea.incident_id = $1
   AND ea.entity_type = $2
   AND ea.deleted_at IS NULL
   AND h.host_state IN ('stub', 'canonical')
 ORDER BY ea.record_id ASC, ea.created_at ASC, ea.entity_alias_id ASC
`
	case "identity":
		query = `
SELECT ea.record_id, ea.raw_text
  FROM entity_aliases ea
  JOIN identities i
    ON i.record_id = ea.record_id
   AND i.incident_id = ea.incident_id
 WHERE ea.incident_id = $1
   AND ea.entity_type = $2
   AND ea.deleted_at IS NULL
   AND i.identity_state IN ('stub', 'canonical')
 ORDER BY ea.record_id ASC, ea.created_at ASC, ea.entity_alias_id ASC
`
	default:
		return nil, nil
	}

	rows, err := tx.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("load active %s aliases: %w", entityType, err)
	}
	defer rows.Close()

	aliases := make([]autoResolutionAliasCandidate, 0)
	for rows.Next() {
		var alias autoResolutionAliasCandidate
		if err := rows.Scan(&alias.RecordID, &alias.RawText); err != nil {
			return nil, fmt.Errorf("scan active %s alias: %w", entityType, err)
		}
		aliases = append(aliases, alias)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate active %s aliases: %w", entityType, err)
	}
	return aliases, nil
}

func lookupMatchedAliasText(ctx context.Context, querier mentionQueryer, recordID uuid.UUID, entityType string, rawText string) (*string, error) {
	candidateText, ok := fieldnorm.AutoResolutionCandidateText(rawText)
	if !ok {
		return nil, nil
	}

	rows, err := querier.Query(ctx, `
SELECT raw_text
  FROM entity_aliases
 WHERE record_id = $1
   AND entity_type = $2
   AND deleted_at IS NULL
 ORDER BY created_at ASC, entity_alias_id ASC
`, recordID, entityType)
	if err != nil {
		return nil, fmt.Errorf("query matched alias text: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var aliasText string
		if err := rows.Scan(&aliasText); err != nil {
			return nil, fmt.Errorf("scan matched alias text: %w", err)
		}
		aliasCandidateText, ok := fieldnorm.AutoResolutionCandidateText(aliasText)
		if ok && aliasCandidateText == candidateText {
			value := aliasText
			return &value, nil
		}
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate matched alias texts: %w", err)
	}
	return nil, nil
}

func loadActiveCollectionLinkMetadata(ctx context.Context, querier mentionQueryer, incidentID uuid.UUID, sourceRecordID uuid.UUID, targetRecordID uuid.UUID, linkType string) (*collectionLinkMetadata, error) {
	rows, err := querier.Query(ctx, `
SELECT provenance, confidence
  FROM record_links
 WHERE incident_id = $1
   AND src_record_id = $2
   AND dst_record_id = $3
   AND link_type = $4
   AND deleted_at IS NULL
 ORDER BY created_at DESC, record_link_id DESC
 LIMIT 1
`, incidentID, sourceRecordID, targetRecordID, linkType)
	if err != nil {
		return nil, fmt.Errorf("query active link metadata: %w", err)
	}
	defer rows.Close()

	if !rows.Next() {
		if err := rows.Err(); err != nil {
			return nil, fmt.Errorf("iterate active link metadata: %w", err)
		}
		return nil, nil
	}

	var (
		metadata   collectionLinkMetadata
		confidence pgtype.Int4
	)
	if err := rows.Scan(&metadata.Provenance, &confidence); err != nil {
		return nil, fmt.Errorf("scan active link metadata: %w", err)
	}
	if confidence.Valid {
		value := int(confidence.Int32)
		metadata.Confidence = &value
	}
	return &metadata, nil
}

func timelineRelationshipEntityType(fieldKey string) (string, bool) {
	switch fieldKey {
	case "timeline.host_refs":
		return "host", true
	case "timeline.identity_refs":
		return "identity", true
	default:
		return "", false
	}
}
