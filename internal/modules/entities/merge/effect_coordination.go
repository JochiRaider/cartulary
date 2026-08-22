package merge

import (
	"errors"
	"slices"
	"strings"

	"github.com/google/uuid"

	"github.com/JochiRaider/cartulary/internal/modules/timeline/mentioneffects"
)

func classifyAssessmentRepointError(err error) error {
	var protectedSetChanged *AssessmentProtectedSetChangedError
	if errors.As(err, &protectedSetChanged) {
		return &MergePreconditionError{
			ReasonCode: "protected_set_changed",
			Details: map[string]any{
				"record_id": protectedSetChanged.RecordID.String(),
			},
		}
	}
	return err
}

func mergeTimelineInvalidations(groups ...[]mentioneffects.TimelineInvalidation) []mentioneffects.TimelineInvalidation {
	byRecord := map[uuid.UUID]mentioneffects.TimelineInvalidation{}
	for _, group := range groups {
		for _, invalidation := range group {
			current := byRecord[invalidation.RecordID]
			if current.RecordID == uuid.Nil {
				current.RecordID = invalidation.RecordID
				current.RowVersion = invalidation.RowVersion
			}
			current.ChangedFieldKeys = append(current.ChangedFieldKeys, invalidation.ChangedFieldKeys...)
			byRecord[invalidation.RecordID] = current
		}
	}
	recordIDs := make([]uuid.UUID, 0, len(byRecord))
	for recordID := range byRecord {
		recordIDs = append(recordIDs, recordID)
	}
	slices.SortFunc(recordIDs, func(left uuid.UUID, right uuid.UUID) int {
		return strings.Compare(left.String(), right.String())
	})
	result := make([]mentioneffects.TimelineInvalidation, 0, len(recordIDs))
	for _, recordID := range recordIDs {
		invalidation := byRecord[recordID]
		slices.Sort(invalidation.ChangedFieldKeys)
		invalidation.ChangedFieldKeys = slices.Compact(invalidation.ChangedFieldKeys)
		result = append(result, invalidation)
	}
	return result
}
