package artifacts

import (
	"context"
	"errors"

	"github.com/google/uuid"

	"github.com/JochiRaider/cartulary/internal/modules/artifacts/internal/portablevalue"
	"github.com/JochiRaider/cartulary/internal/modules/incidentbundles/sourceport"
	"github.com/JochiRaider/cartulary/internal/modules/incidentportability"
)

type artifactPreparationBinding struct{ contractMajor int }

type preparedArtifactImport struct {
	binding             *artifactPreparationBinding
	manifest            *sourceStateManifest
	members             *incidentportability.BoundMembers
	incidentID, actorID uuid.UUID
	operationID         string
	version             int
}

func (p preparedArtifactImport) matches(input sourceport.ImportContext, binding *artifactPreparationBinding) bool {
	return binding != nil && p.binding == binding && binding.contractMajor == sourceport.ContractMajor && p.manifest != nil && p.members != nil && p.incidentID == input.IncidentID && p.actorID == input.ActorUserID && p.operationID == input.OperationID && p.version == input.BundleVersion && input.Attributions != nil
}

func prepareArtifactImport(ctx context.Context, bundle sourceport.Bundle, input sourceport.ImportContext, descriptor sourceport.Descriptor, manifest *sourceStateManifest, binding *artifactPreparationBinding) (preparedArtifactImport, error) {
	if input.IncidentID == uuid.Nil || input.ActorUserID == uuid.Nil || input.BundleVersion != 8 || input.OperationID == "" {
		return preparedArtifactImport{}, sourceport.ErrPreparedBinding
	}
	paths := make([]string, 0, len(manifest.relations))
	for _, relation := range manifest.relations {
		paths = append(paths, relation.logicalBundlePath)
	}
	members, err := incidentportability.BindMembers(bundle, paths)
	if err != nil {
		return preparedArtifactImport{}, err
	}
	p := preparedArtifactImport{binding: binding, manifest: manifest, members: members, incidentID: input.IncidentID, actorID: input.ActorUserID, operationID: input.OperationID, version: input.BundleVersion}
	// Finish each phase across every member before starting the next phase.
	// Only the current logical row and one previous identity per path survive.
	if err := p.raw(ctx, descriptor, func(sourceStateRelation, map[string]any) error { return nil }); err != nil {
		return preparedArtifactImport{}, err
	}
	previous := map[string]string{}
	if err := p.raw(ctx, descriptor, func(relation sourceStateRelation, raw map[string]any) error {
		id, valid := portablevalue.CanonicalUUID(raw[relation.stableIdentity[0]])
		if !valid || id.String() <= previous[relation.logicalBundlePath] {
			return descriptor.DeclaredFailure("artifacts.source_identity_admitted")
		}
		previous[relation.logicalBundlePath] = id.String()
		return nil
	}); err != nil {
		return preparedArtifactImport{}, err
	}
	for _, phase := range []struct {
		invariant string
		valid     func(any) bool
	}{
		{"row_shape_exact", func(any) bool { return true }},
		{"references_same_incident", func(row any) bool { _, incident := portablevalue.IDs(row); return incident == p.incidentID }},
		{"subtype_exact", portablevalue.SubtypeLegal},
		{"lifecycle_fields_legal", portablevalue.LifecycleLegal},
		{"attribution_exact", func(row any) bool {
			for _, actor := range portablevalue.Actors(row) {
				if _, exists := input.Actors.Lookup(actor.String()); !exists {
					return false
				}
			}
			return true
		}},
	} {
		if err := p.typed(ctx, descriptor, func(_ sourceStateRelation, row any) error {
			if !phase.valid(row) {
				return descriptor.DeclaredFailure("artifacts." + phase.invariant)
			}
			return nil
		}); err != nil {
			return preparedArtifactImport{}, err
		}
	}
	return p, nil
}

func (p preparedArtifactImport) raw(ctx context.Context, descriptor sourceport.Descriptor, visit func(sourceStateRelation, map[string]any) error) error {
	for _, relation := range p.manifest.relations {
		if err := p.rawPath(ctx, descriptor, relation, func(raw map[string]any) error { return visit(relation, raw) }); err != nil {
			return err
		}
	}
	return nil
}

func (p preparedArtifactImport) rawPath(ctx context.Context, descriptor sourceport.Descriptor, relation sourceStateRelation, visit func(map[string]any) error) error {
	err := incidentportability.WalkPath(ctx, p.members, relation.logicalBundlePath, visit)
	var malformed *incidentportability.StrictNDJSONDecodeError
	if errors.As(err, &malformed) {
		return descriptor.DeclaredFailure("artifacts.row_shape_exact")
	}
	return err
}

func (p preparedArtifactImport) typed(ctx context.Context, descriptor sourceport.Descriptor, visit func(sourceStateRelation, any) error) error {
	return p.raw(ctx, descriptor, func(relation sourceStateRelation, raw map[string]any) error {
		row, valid := portablevalue.Parse(relation.tableName, raw)
		if !valid {
			return descriptor.DeclaredFailure("artifacts.row_shape_exact")
		}
		return visit(relation, row)
	})
}
