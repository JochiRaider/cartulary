package timeline

import (
	"context"
	"errors"
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
	RecordID uuid.UUID
}

type autoResolutionKey struct{ Field, RawText string }
type preparedAutoResolution map[autoResolutionKey]*autoResolutionMatch

func (s *store) lookupInteractiveAutoResolutionMatchTx(ctx context.Context, tx pgx.Tx, incidentID uuid.UUID, fieldKey string, rawText string, prepared preparedAutoResolution) (*autoResolutionMatch, error) {
	match, present := prepared[autoResolutionKey{fieldKey, rawText}]
	if !present {
		return nil, errors.New("timeline: automatic reference was not prepared")
	}
	if match == nil {
		return nil, nil
	}
	entityType, _ := timelineRelationshipEntityType(fieldKey)
	if err := s.entityStore.AdmitResolvedTargetTx(ctx, tx, incidentID, entityType, match.RecordID); err != nil {
		if errors.Is(err, ErrResolvedRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	current, err := s.initialAutoResolutionMatchTx(ctx, tx, incidentID, fieldKey, rawText)
	if err != nil {
		return nil, err
	}
	if current == nil || current.RecordID != match.RecordID {
		return nil, nil
	}
	return current, nil
}

func (s *store) initialAutoResolutionMatchTx(ctx context.Context, tx pgx.Tx, incidentID uuid.UUID, fieldKey, rawText string) (*autoResolutionMatch, error) {
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
	return uniqueAutoResolutionMatch(aliases, candidateText), nil
}

func uniqueAutoResolutionMatch(aliases []EntityAlias, candidateText string) *autoResolutionMatch {
	var match *autoResolutionMatch
	for _, alias := range aliases {
		aliasCandidateText, ok := fieldnorm.AutoResolutionCandidateText(alias.RawText)
		if !ok || aliasCandidateText != candidateText {
			continue
		}
		if match != nil && match.RecordID != alias.RecordID {
			return nil
		}
		if match == nil {
			match = &autoResolutionMatch{RecordID: alias.RecordID}
		}
	}
	return match
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
