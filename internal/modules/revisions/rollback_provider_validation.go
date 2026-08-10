package revisions

import (
	"fmt"
	"sort"

	"github.com/google/uuid"

	"github.com/JochiRaider/cartulary/internal/modules/revisions/rollbackcontract"
)

func validateNonRowApplyResult(descriptor rollbackcontract.TargetDescriptor, result rollbackcontract.ApplyInverseResult) error {
	want := canonicalRecordIDs(descriptor.AffectedRecordIDs)
	got := canonicalRecordIDs(result.AffectedRecordIDs)
	if len(want) != len(got) {
		return fmt.Errorf("%w: affected record set changed", ErrRollbackPreconditionFailed)
	}
	for index := range want {
		if want[index] != got[index] {
			return fmt.Errorf("%w: affected record set changed", ErrRollbackPreconditionFailed)
		}
	}
	return nil
}

func canonicalRecordIDs(values []uuid.UUID) []uuid.UUID {
	seen := make(map[uuid.UUID]struct{}, len(values))
	result := make([]uuid.UUID, 0, len(values))
	for _, value := range values {
		if value == uuid.Nil {
			continue
		}
		if _, duplicate := seen[value]; duplicate {
			continue
		}
		seen[value] = struct{}{}
		result = append(result, value)
	}
	sort.Slice(result, func(left, right int) bool { return result[left].String() < result[right].String() })
	return result
}
