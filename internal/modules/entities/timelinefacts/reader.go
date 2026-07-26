package timelinefacts

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/JochiRaider/cartulary/internal/modules/timeline/workbookprojection"
	"github.com/JochiRaider/cartulary/internal/platform/fieldnorm"
)

type Reader struct{}

func (Reader) LoadMentionsTx(ctx context.Context, tx pgx.Tx, recordID uuid.UUID) ([]workbookprojection.MentionFact, error) {
	rows, err := tx.Query(ctx, `
SELECT entity_mention_id, entity_type, source_field_key, raw_text, resolution_status,
       row_version, resolved_record_id, resolution_method
  FROM entity_mentions
 WHERE source_record_id = $1
   AND resolution_status IN ('unresolved', 'resolved')
 ORDER BY ordinal ASC, entity_mention_id ASC
`, recordID)
	if err != nil {
		return nil, fmt.Errorf("load timeline mention facts: %w", err)
	}
	defer rows.Close()

	facts := make([]workbookprojection.MentionFact, 0)
	autoResolvedIDs := make([]uuid.UUID, 0)
	for rows.Next() {
		var (
			fact             workbookprojection.MentionFact
			resolvedRecordID pgtype.UUID
			resolutionMethod pgtype.Text
		)
		if err := rows.Scan(
			&fact.MentionID,
			&fact.EntityType,
			&fact.SourceFieldKey,
			&fact.RawText,
			&fact.ResolutionStatus,
			&fact.RowVersion,
			&resolvedRecordID,
			&resolutionMethod,
		); err != nil {
			return nil, fmt.Errorf("scan timeline mention fact: %w", err)
		}
		if resolvedRecordID.Valid {
			value := uuid.Must(uuid.FromBytes(resolvedRecordID.Bytes[:]))
			fact.ResolvedRecordID = &value
		}
		if resolutionMethod.Valid {
			value := resolutionMethod.String
			fact.ResolutionMethod = &value
		}
		if fact.ResolvedRecordID != nil && fact.ResolutionMethod != nil && *fact.ResolutionMethod == "auto_match" {
			autoResolvedIDs = append(autoResolvedIDs, *fact.ResolvedRecordID)
		}
		facts = append(facts, fact)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate timeline mention facts: %w", err)
	}
	if len(autoResolvedIDs) == 0 {
		return facts, nil
	}

	aliases, err := loadAliasesTx(ctx, tx, autoResolvedIDs)
	if err != nil {
		return nil, err
	}
	for index := range facts {
		fact := &facts[index]
		if fact.ResolvedRecordID == nil || fact.ResolutionMethod == nil || *fact.ResolutionMethod != "auto_match" {
			continue
		}
		candidate, ok := fieldnorm.AutoResolutionCandidateText(fact.RawText)
		if !ok {
			continue
		}
		for _, alias := range aliases[*fact.ResolvedRecordID] {
			normalized, ok := fieldnorm.AutoResolutionCandidateText(alias)
			if ok && normalized == candidate {
				value := alias
				fact.MatchedAliasText = &value
				break
			}
		}
	}
	return facts, nil
}

func loadAliasesTx(ctx context.Context, tx pgx.Tx, recordIDs []uuid.UUID) (map[uuid.UUID][]string, error) {
	rows, err := tx.Query(ctx, `
SELECT record_id, raw_text
  FROM entity_aliases
 WHERE record_id = ANY($1)
   AND deleted_at IS NULL
 ORDER BY record_id ASC, created_at ASC, entity_alias_id ASC
`, recordIDs)
	if err != nil {
		return nil, fmt.Errorf("load timeline matched alias facts: %w", err)
	}
	defer rows.Close()
	result := make(map[uuid.UUID][]string)
	for rows.Next() {
		var recordID uuid.UUID
		var rawText string
		if err := rows.Scan(&recordID, &rawText); err != nil {
			return nil, fmt.Errorf("scan timeline matched alias fact: %w", err)
		}
		result[recordID] = append(result[recordID], rawText)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate timeline matched alias facts: %w", err)
	}
	return result, nil
}
