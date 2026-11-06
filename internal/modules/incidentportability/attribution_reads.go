package incidentportability

import (
	"context"
	"errors"
	"strings"
)

// ValidateTopLevelAttributions compares exact owner-selected sidecars without
// retaining a collection-wide actor map. Each row has a schema-bounded number
// of actor columns; the final count rejects unrelated extra sidecars.
func ValidateTopLevelAttributions(ctx context.Context, members MemberReader, path, table, identity string, recorder AttributionRecorder) error {
	failure := errors.New("portable attribution does not match source")
	if recorder == nil {
		return failure
	}
	var expectedCount, actualCount uint64
	err := WalkPath(ctx, members, path, func(row map[string]any) error {
		expected := map[string]string{}
		for column, value := range row {
			if strings.HasSuffix(column, "_user_id") && value != nil {
				expected[column] = StringFromAny(value)
				expectedCount++
			}
		}
		rowID := StringFromAny(row[identity])
		err := recorder.VisitImportedAttributions(table, rowID, func(value ImportedAttribution) error {
			actor, ok := expected[value.SourceColumn]
			if !ok || actor != value.SourceActorID || value.SourceTable != table || value.SourceRowID != rowID {
				return failure
			}
			delete(expected, value.SourceColumn)
			return nil
		})
		if err != nil {
			return err
		}
		if len(expected) != 0 {
			return failure
		}
		return nil
	})
	if err != nil {
		return err
	}
	if err := recorder.VisitImportedAttributions(table, "", func(ImportedAttribution) error { actualCount++; return nil }); err != nil {
		return err
	}
	if expectedCount != actualCount {
		return failure
	}
	return nil
}
