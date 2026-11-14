package evidence

import (
	"context"
	"errors"
	"github.com/JochiRaider/cartulary/internal/modules/evidence/internal/sourcevalue"
	"github.com/JochiRaider/cartulary/internal/modules/incidentbundles/sourceport"
	"github.com/JochiRaider/cartulary/internal/modules/incidentportability"
	"github.com/google/uuid"
)

const evidenceRecordPath = "data/evidence_records.ndjson"

type evidencePreparationBinding struct{ contractMajor int }
type preparedEvidenceImport struct {
	binding             *evidencePreparationBinding
	members             *incidentportability.BoundMembers
	incidentID, actorID uuid.UUID
	operationID         string
	version             int
	blobRewrite         *portableBlobRewrite
}

func (p preparedEvidenceImport) matches(input sourceport.ImportContext, binding *evidencePreparationBinding) bool {
	return binding != nil && p.binding == binding && binding.contractMajor == sourceport.ContractMajor && p.members != nil && p.incidentID == input.IncidentID && p.actorID == input.ActorUserID && p.operationID == input.OperationID && p.version == input.BundleVersion && input.Attributions != nil
}
func prepareEvidenceImport(ctx context.Context, bundle sourceport.Bundle, input sourceport.ImportContext, d sourceport.Descriptor, binding *evidencePreparationBinding) (preparedEvidenceImport, error) {
	if input.IncidentID == uuid.Nil || input.ActorUserID == uuid.Nil || input.BundleVersion != 8 || input.OperationID == "" {
		return preparedEvidenceImport{}, sourceport.ErrPreparedBinding
	}
	members, err := incidentportability.BindMembers(bundle, []string{evidenceRecordPath, "data/object_blobs.ndjson", "data/evidence_custody_events.ndjson"})
	if err != nil {
		return preparedEvidenceImport{}, err
	}
	p := preparedEvidenceImport{binding: binding, members: members, incidentID: input.IncidentID, actorID: input.ActorUserID, operationID: input.OperationID, version: input.BundleVersion, blobRewrite: &portableBlobRewrite{}}
	if err := p.raw(ctx, d, func(map[string]any) error { return nil }); err != nil {
		return preparedEvidenceImport{}, err
	}
	previous := ""
	if err := p.raw(ctx, d, func(raw map[string]any) error {
		text, ok := raw["record_id"].(string)
		id, err := uuid.Parse(text)
		if !ok || err != nil || id == uuid.Nil || text != id.String() || text <= previous {
			return d.DeclaredFailure("evidence.source_identity_admitted")
		}
		previous = text
		return nil
	}); err != nil {
		return preparedEvidenceImport{}, err
	}
	for _, phase := range []struct {
		invariant string
		valid     func(sourcevalue.Row) bool
	}{
		{"row_shape_exact", func(sourcevalue.Row) bool { return true }},
		{"envelope_type_scope", func(row sourcevalue.Row) bool { return row.IncidentID == p.incidentID }},
		{"lifecycle_legal", func(row sourcevalue.Row) bool { return row.LifecycleLegal() }},
		{"storage_reference_legal", func(row sourcevalue.Row) bool { return row.StorageReferenceLegal() }},
		{"attribution_exact", func(row sourcevalue.Row) bool {
			if row.OwnerUserID == nil {
				return true
			}
			_, ok := input.Actors.Lookup(row.OwnerUserID.String())
			return ok
		}},
	} {
		if err := p.typed(ctx, d, func(row sourcevalue.Row) error {
			if !phase.valid(row) {
				return d.DeclaredFailure("evidence." + phase.invariant)
			}
			return nil
		}); err != nil {
			return preparedEvidenceImport{}, err
		}
	}
	if err := p.prepareCustody(ctx, input, d); err != nil {
		return preparedEvidenceImport{}, err
	}
	if err := p.prepareBlobs(ctx, bundle, input, d); err != nil {
		return preparedEvidenceImport{}, err
	}
	return p, nil
}
func (p preparedEvidenceImport) raw(ctx context.Context, d sourceport.Descriptor, visit func(map[string]any) error) error {
	err := incidentportability.WalkPath(ctx, p.members, evidenceRecordPath, visit)
	var malformed *incidentportability.StrictNDJSONDecodeError
	if errors.As(err, &malformed) {
		return d.DeclaredFailure("evidence.row_shape_exact")
	}
	return err
}
func (p preparedEvidenceImport) typed(ctx context.Context, d sourceport.Descriptor, visit func(sourcevalue.Row) error) error {
	return p.raw(ctx, d, func(raw map[string]any) error {
		row, ok := sourcevalue.Parse(raw)
		if !ok {
			return d.DeclaredFailure("evidence.row_shape_exact")
		}
		return visit(row)
	})
}
