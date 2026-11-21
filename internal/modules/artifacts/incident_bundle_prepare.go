package artifacts

import (
	"context"
	"encoding/json"
	"errors"
	"maps"
	"slices"

	"github.com/google/uuid"

	"github.com/JochiRaider/cartulary/internal/modules/artifacts/internal/portablevalue"
	"github.com/JochiRaider/cartulary/internal/modules/incidentbundles/sourceport"
	"github.com/JochiRaider/cartulary/internal/modules/incidentportability"
)

type artifactPreparationBinding struct{ contractMajor int }

// preparedArtifactImport carries bundle members only while preparing; the
// value the final transaction receives holds staged streams (Core 01
// REQ-01-641).
type preparedArtifactImport struct {
	binding             *artifactPreparationBinding
	manifest            *sourceStateManifest
	members             *incidentportability.BoundMembers
	incidentID, actorID uuid.UUID
	operationID         string
	version             int
	staged              map[string]incidentportability.PreparedRowStream
	attribution         incidentportability.AttributionStream
}

func (p preparedArtifactImport) matches(input sourceport.ImportContext, binding *artifactPreparationBinding) bool {
	return binding != nil && p.binding == binding && binding.contractMajor == sourceport.ContractMajor && p.manifest != nil && p.members == nil && p.staged != nil && input.Preparation != nil && p.incidentID == input.IncidentID && p.actorID == input.ActorUserID && p.operationID == input.OperationID && p.version == input.BundleVersion && input.Attributions != nil
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
		if err := p.bundleTyped(ctx, descriptor, func(_ sourceStateRelation, row any) error {
			if !phase.valid(row) {
				return descriptor.DeclaredFailure("artifacts." + phase.invariant)
			}
			return nil
		}); err != nil {
			return preparedArtifactImport{}, err
		}
	}
	if err := p.stage(ctx, input, descriptor); err != nil {
		return preparedArtifactImport{}, err
	}
	// The transaction receives no bundle reader.
	p.members = nil
	return p, nil
}

// stage writes each relation's admitted typed rows and the historical
// attribution of their actor columns.
func (p *preparedArtifactImport) stage(ctx context.Context, input sourceport.ImportContext, descriptor sourceport.Descriptor) error {
	store := input.Preparation
	if store == nil {
		return errors.New("artifacts: incident bundle preparation staging is required")
	}
	p.staged = map[string]incidentportability.PreparedRowStream{}
	for _, relation := range p.manifest.relations {
		stream, err := incidentportability.StagePreparedRows(ctx, store, relation.tableName+".rows", artifactRowCodec{table: relation.tableName}, func(emit func(any) error) error {
			return p.rawPath(ctx, descriptor, relation, func(raw map[string]any) error {
				row, valid := portablevalue.Parse(relation.tableName, raw)
				if !valid {
					return descriptor.DeclaredFailure("artifacts.row_shape_exact")
				}
				return emit(row)
			})
		})
		if err != nil {
			return err
		}
		p.staged[relation.tableName] = stream
	}
	admitted := func(actorID string) bool { _, ok := input.Actors.Lookup(actorID); return ok }
	var err error
	p.attribution, err = incidentportability.StageAttributions(ctx, store, "attribution.rows", admitted, func(emit func(incidentportability.AttributionRow) error) error {
		return p.typed(ctx, store, func(relation sourceStateRelation, row any) error {
			id, _ := portablevalue.IDs(row)
			actors := portablevalue.Actors(row)
			for _, column := range slices.Sorted(maps.Keys(actors)) {
				if err := emit(incidentportability.AttributionRow{SourceTable: relation.tableName, SourceRowID: id.String(), SourceColumn: column, SourceActorID: actors[column].String()}); err != nil {
					return err
				}
			}
			return nil
		})
	})
	return err
}

// artifactRowCodec stages one relation's admitted typed row as JSON.
type artifactRowCodec struct{ table string }

func (artifactRowCodec) Encode(row any) ([]byte, error) { return json.Marshal(row) }
func (c artifactRowCodec) Decode(line []byte) (any, error) {
	var row any
	switch c.table {
	case "artifacts":
		row = &portablevalue.Parent{}
	case "artifact_findings":
		row = &portablevalue.Finding{}
	case "artifact_investigative_queries":
		row = &portablevalue.Query{}
	case "artifact_forensic_keywords":
		row = &portablevalue.Keyword{}
	case "handoff_risk_refs":
		row = &portablevalue.Risk{}
	default:
		return nil, incidentportability.ErrPreparedStream
	}
	if err := json.Unmarshal(line, row); err != nil {
		return nil, incidentportability.ErrPreparedStream
	}
	return row, nil
}
func (artifactRowCodec) Columns() []string { return nil }
func (artifactRowCodec) Values(any) ([]any, error) {
	return nil, incidentportability.ErrPreparedStream
}

// typed visits each relation's staged rows in relation order; the final
// transaction reads no bundle member.
func (p preparedArtifactImport) typed(ctx context.Context, store incidentportability.MemberReader, visit func(sourceStateRelation, any) error) error {
	for _, relation := range p.manifest.relations {
		if err := p.typedRelation(ctx, store, relation, func(row any) error { return visit(relation, row) }); err != nil {
			return err
		}
	}
	return nil
}

func (p preparedArtifactImport) typedRelation(ctx context.Context, store incidentportability.MemberReader, relation sourceStateRelation, visit func(any) error) error {
	stream, ok := p.staged[relation.tableName]
	if !ok || store == nil {
		return sourceport.ErrPreparedBinding
	}
	return incidentportability.ReadPreparedRows(ctx, store, stream, artifactRowCodec{table: relation.tableName}, visit)
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

func (p preparedArtifactImport) bundleTyped(ctx context.Context, descriptor sourceport.Descriptor, visit func(sourceStateRelation, any) error) error {
	return p.raw(ctx, descriptor, func(relation sourceStateRelation, raw map[string]any) error {
		row, valid := portablevalue.Parse(relation.tableName, raw)
		if !valid {
			return descriptor.DeclaredFailure("artifacts.row_shape_exact")
		}
		return visit(relation, row)
	})
}
