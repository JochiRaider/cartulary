package evidence

import (
	"context"
	"errors"

	"github.com/JochiRaider/cartulary/internal/modules/incidentbundles/sourceport"
	"github.com/JochiRaider/cartulary/internal/modules/incidentportability"
	"github.com/google/uuid"
)

func (p preparedEvidenceImport) custodyRaw(ctx context.Context, d sourceport.Descriptor, visit func(map[string]any) error) error {
	err := incidentportability.WalkPath(ctx, p.members, portableCustodyPath, visit)
	var malformed *incidentportability.StrictNDJSONDecodeError
	if errors.As(err, &malformed) {
		return d.DeclaredFailure("evidence.row_shape_exact")
	}
	return err
}
func (p preparedEvidenceImport) custody(ctx context.Context, d sourceport.Descriptor, visit func(portableCustodyRow) error) error {
	return p.custodyRaw(ctx, d, func(raw map[string]any) error {
		row, ok := parsePortableCustodyShape(raw)
		if !ok {
			return d.DeclaredFailure("evidence.row_shape_exact")
		}
		return visit(row)
	})
}
func (p preparedEvidenceImport) prepareCustody(ctx context.Context, input sourceport.ImportContext, d sourceport.Descriptor) error {
	if err := p.custodyRaw(ctx, d, func(map[string]any) error { return nil }); err != nil {
		return err
	}
	previous := ""
	if err := p.custodyRaw(ctx, d, func(raw map[string]any) error {
		text, ok := raw["custody_event_id"].(string)
		id, err := uuid.Parse(text)
		if !ok || err != nil || id == uuid.Nil || id.String() != text || text <= previous {
			return d.DeclaredFailure("evidence.source_identity_admitted")
		}
		previous = text
		return nil
	}); err != nil {
		return err
	}
	for _, phase := range []struct {
		invariant string
		valid     func(portableCustodyRow) bool
	}{
		{"row_shape_exact", func(portableCustodyRow) bool { return true }},
		{"custody_ordered", func(row portableCustodyRow) bool { _, ok := portableSourceClock(row.OccurredAt); return ok }},
		{"custody_same_incident", func(row portableCustodyRow) bool { return row.IncidentID == p.incidentID }},
		{"attribution_exact", func(row portableCustodyRow) bool {
			if row.ActorUserID == nil {
				return true
			}
			_, ok := input.Actors.Lookup(row.ActorUserID.String())
			return ok
		}},
	} {
		if err := p.custody(ctx, d, func(row portableCustodyRow) error {
			if !phase.valid(row) {
				return d.DeclaredFailure("evidence." + phase.invariant)
			}
			return nil
		}); err != nil {
			return err
		}
	}
	return nil
}
