package timeline

import (
	"context"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

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

func (s *store) lookupInteractiveAutoResolutionMatchTx(ctx context.Context, tx pgx.Tx, incidentID uuid.UUID, fieldKey string, rawText string) (*autoResolutionMatch, error) {
	entityType, ok := timelineRelationshipEntityType(fieldKey)
	if !ok {
		return nil, nil
	}

	candidateText, ok := fieldnorm.AutoResolutionCandidateText(rawText)
	if !ok || autoResolutionSuppressed(candidateText) {
		return nil, nil
	}

	aliases, err := s.entityStore.ListEligibleAliasesTx(ctx, tx, incidentID, entityType)
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
