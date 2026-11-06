package indicators

import (
	"context"
	"fmt"

	"github.com/JochiRaider/cartulary/internal/modules/incidentportability"
	"github.com/google/uuid"
)

// PortableReferences grants dependent owners only the identity and value of an
// already prepared Indicator source. It includes historical/deleted targets;
// active-target admission for new links is a different operation.
type PortableReferences struct {
	incidentID uuid.UUID
	members    *incidentportability.BoundMembers
}

func NewPortableReferences(members incidentportability.MemberReader, incidentID uuid.UUID) (*PortableReferences, error) {
	if incidentID == uuid.Nil {
		return nil, fmt.Errorf("portable indicator scope is invalid")
	}
	bound, err := incidentportability.BindMembers(members, []string{"data/indicators.ndjson"})
	if err != nil {
		return nil, err
	}
	return &PortableReferences{incidentID: incidentID, members: bound}, nil
}
func (r *PortableReferences) FindIndicator(ctx context.Context, id uuid.UUID) (IndicatorReference, error) {
	var result IndicatorReference
	if r == nil || id == uuid.Nil {
		return result, fmt.Errorf("portable indicator reference is invalid")
	}
	found := false
	err := incidentportability.WalkPath(ctx, r.members, "data/indicators.ndjson", func(row map[string]any) error {
		if row["record_id"] != id.String() {
			return nil
		}
		if found || row["incident_id"] != r.incidentID.String() {
			return fmt.Errorf("portable indicator reference is ambiguous")
		}
		kind, ok := row["indicator_type"].(string)
		if !ok {
			return fmt.Errorf("portable indicator type is invalid")
		}
		valueKind, ok := row["value_kind"].(string)
		if !ok {
			return fmt.Errorf("portable indicator value kind is invalid")
		}
		var normalized *string
		if row["normalized_value"] != nil {
			value, ok := row["normalized_value"].(string)
			if !ok {
				return fmt.Errorf("portable indicator value is invalid")
			}
			normalized = &value
		}
		result = IndicatorReference{RecordID: id, IncidentID: r.incidentID, IndicatorType: kind, ValueKind: valueKind, NormalizedValue: normalized}
		found = true
		return nil
	})
	if err == nil && !found {
		err = fmt.Errorf("portable indicator target is missing")
	}
	return result, err
}
