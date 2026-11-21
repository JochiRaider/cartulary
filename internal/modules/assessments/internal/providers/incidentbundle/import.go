package incidentbundle

import (
	"context"
	"encoding/json"
	"errors"
	"reflect"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"

	"github.com/JochiRaider/cartulary/internal/modules/assessments/internal/sourcevalue"
	"github.com/JochiRaider/cartulary/internal/modules/incidentbundles/sourceport"
	"github.com/JochiRaider/cartulary/internal/modules/incidentportability"
)

type preparationBinding struct{ contractMajor int }

// preparedImport carries bundle members only while preparing; the value the
// final transaction receives holds staged streams (Core 01 REQ-01-641).
type preparedImport struct {
	binding             *preparationBinding
	members             *incidentportability.BoundMembers
	incidentID, actorID uuid.UUID
	operationID         string
	version             int
	staged              bool
	rows                incidentportability.PreparedRowStream
	attribution         incidentportability.AttributionStream
}

func (p preparedImport) matches(input sourceport.ImportContext, binding *preparationBinding) bool {
	return p.binding == binding && binding.contractMajor == sourceport.ContractMajor && p.members == nil && p.staged && input.Preparation != nil && p.incidentID == input.IncidentID && p.actorID == input.ActorUserID && p.operationID == input.OperationID && p.version == input.BundleVersion && input.Attributions != nil
}
func prepare(ctx context.Context, bundle sourceport.Bundle, input sourceport.ImportContext, descriptor sourceport.Descriptor, binding *preparationBinding) (preparedImport, error) {
	if input.IncidentID == uuid.Nil || input.ActorUserID == uuid.Nil || input.BundleVersion != 8 {
		return preparedImport{}, sourceport.ErrPreparedBinding
	}
	members, err := incidentportability.BindMembers(bundle, []string{assessmentsBundlePath})
	if err != nil {
		return preparedImport{}, err
	}
	p := preparedImport{binding: binding, members: members, incidentID: input.IncidentID, actorID: input.ActorUserID, operationID: input.OperationID, version: input.BundleVersion}
	// Complete framing first, then stable identities, then typed and semantic
	// passes. Only one bounded logical row is retained at any time.
	if err := p.raw(ctx, descriptor, func(map[string]any) error { return nil }); err != nil {
		return preparedImport{}, err
	}
	previous := ""
	if err := p.raw(ctx, descriptor, func(raw map[string]any) error {
		id, valid := sourcevalue.CanonicalUUID(raw["record_id"])
		if !valid || id.String() <= previous {
			return descriptor.DeclaredFailure("assessments.source_identity_admitted")
		}
		previous = id.String()
		return nil
	}); err != nil {
		return preparedImport{}, err
	}
	for _, check := range []struct {
		invariant string
		valid     func(sourcevalue.Row) bool
	}{
		{"row_shape_exact", func(sourcevalue.Row) bool { return true }},
		{"subject_type_scope", func(r sourcevalue.Row) bool { return r.SubjectLegal(input.IncidentID) }},
		{"state_confidence_rationale_legal", sourcevalue.Row.JudgmentLegal},
		{"timestamps_lifecycle_legal", sourcevalue.Row.LifecycleLegal},
		{"attribution_exact", func(r sourcevalue.Row) bool {
			_, assessor := input.Actors.Lookup(r.AssessorUserID.String())
			if r.DeletedByUserID == nil {
				return assessor
			}
			_, deleter := input.Actors.Lookup(r.DeletedByUserID.String())
			return assessor && deleter
		}},
	} {
		if err := p.bundleWalk(ctx, descriptor, func(row sourcevalue.Row) error {
			if !check.valid(row) {
				return descriptor.DeclaredFailure("assessments." + check.invariant)
			}
			return nil
		}); err != nil {
			return preparedImport{}, err
		}
	}
	store := input.Preparation
	if store == nil {
		return preparedImport{}, errors.New("assessments: incident bundle preparation staging is required")
	}
	if p.rows, err = incidentportability.StagePreparedRows(ctx, store, "assessments.rows", assessmentRowCodec{}, func(emit func(sourcevalue.Row) error) error {
		return p.bundleWalk(ctx, descriptor, emit)
	}); err != nil {
		return preparedImport{}, err
	}
	admitted := func(actorID string) bool { _, ok := input.Actors.Lookup(actorID); return ok }
	if p.attribution, err = incidentportability.StageAttributions(ctx, store, "attribution.rows", admitted, func(emit func(incidentportability.AttributionRow) error) error {
		return incidentportability.ReadPreparedRows(ctx, store, p.rows, assessmentRowCodec{}, func(row sourcevalue.Row) error {
			for _, actor := range rowActors(row) {
				if err := emit(incidentportability.AttributionRow{SourceTable: "assessments", SourceRowID: row.RecordID.String(), SourceColumn: actor.column, SourceActorID: actor.id.String()}); err != nil {
					return err
				}
			}
			return nil
		})
	}); err != nil {
		return preparedImport{}, err
	}
	// The transaction receives no bundle reader.
	p.staged, p.members = true, nil
	return p, nil
}

var assessmentColumns = []string{
	"record_id", "incident_id", "subject_record_id", "subject_type", "assessment_state", "confidence_score",
	"assessor_user_id", "assessed_at", "created_at", "updated_at", "deleted_at", "deleted_by_user_id", "rationale",
}

// assessmentRowCodec stages one admitted row; COPY writes the importing actor
// as assessor and, for a deleted row, as deleter.
type assessmentRowCodec struct{ actor uuid.UUID }

func (assessmentRowCodec) Encode(row sourcevalue.Row) ([]byte, error) { return json.Marshal(row) }
func (assessmentRowCodec) Decode(line []byte) (sourcevalue.Row, error) {
	var row sourcevalue.Row
	if err := json.Unmarshal(line, &row); err != nil {
		return sourcevalue.Row{}, incidentportability.ErrPreparedStream
	}
	return row, nil
}
func (assessmentRowCodec) Columns() []string { return assessmentColumns }
func (c assessmentRowCodec) Values(row sourcevalue.Row) ([]any, error) {
	var deletedBy *uuid.UUID
	if row.DeletedByUserID != nil {
		local := c.actor
		deletedBy = &local
	}
	return []any{row.RecordID, row.IncidentID, row.SubjectRecordID, row.SubjectType, row.AssessmentState, row.ConfidenceScore,
		c.actor, row.AssessedAt, row.CreatedAt, row.UpdatedAt, row.DeletedAt, deletedBy, row.Rationale}, nil
}
func (p preparedImport) raw(ctx context.Context, descriptor sourceport.Descriptor, visit func(map[string]any) error) error {
	err := incidentportability.WalkPath(ctx, p.members, assessmentsBundlePath, visit)
	var malformed *incidentportability.StrictNDJSONDecodeError
	if errors.As(err, &malformed) {
		return descriptor.DeclaredFailure("assessments.row_shape_exact")
	}
	return err
}

// walk visits the staged rows; the final transaction reads no bundle member.
func (p preparedImport) walk(ctx context.Context, input sourceport.ImportContext, visit func(sourcevalue.Row) error) error {
	return incidentportability.ReadPreparedRows(ctx, input.Preparation, p.rows, assessmentRowCodec{}, visit)
}

func (p preparedImport) bundleWalk(ctx context.Context, descriptor sourceport.Descriptor, visit func(sourcevalue.Row) error) error {
	return p.raw(ctx, descriptor, func(raw map[string]any) error {
		row, valid := sourcevalue.Parse(raw)
		if !valid {
			return descriptor.DeclaredFailure("assessments.row_shape_exact")
		}
		return visit(row)
	})
}
func (p preparedImport) apply(ctx context.Context, tx pgx.Tx, input sourceport.ImportContext, descriptor sourceport.Descriptor) error {
	// Scope validation finishes for all inputs before the first write; database
	// constraint ordering cannot select the source owner's semantic error.
	if err := p.walk(ctx, input, func(row sourcevalue.Row) error {
		var valid bool
		if err := tx.QueryRow(ctx, `SELECT
 EXISTS (SELECT 1 FROM records WHERE record_id=$1 AND incident_id=$2 AND record_type='assessment')
 AND EXISTS (SELECT 1 FROM records WHERE record_id=$3 AND incident_id=$2 AND record_type=$4)`, row.RecordID, row.IncidentID, row.SubjectRecordID, row.SubjectType).Scan(&valid); err != nil {
			return err
		}
		if !valid {
			return descriptor.DeclaredFailure("assessments.subject_type_scope")
		}
		return nil
	}); err != nil {
		return err
	}
	// One COPY of the staged rows; attribution is flushed by the core in step 9.
	source, err := incidentportability.OpenPreparedCopy(ctx, input.Preparation, p.rows, assessmentRowCodec{actor: p.actorID})
	if err != nil {
		return err
	}
	written, err := tx.CopyFrom(ctx, pgx.Identifier{"assessments"}, assessmentColumns, source)
	if err := source.Finish(written, err); err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return descriptor.DeclaredFailure("assessments.source_identity_admitted")
		}
		return err
	}
	return nil
}
func (p preparedImport) validate(ctx context.Context, tx pgx.Tx, input sourceport.ImportContext, descriptor sourceport.Descriptor) error {
	var invalid bool
	if err := tx.QueryRow(ctx, `SELECT EXISTS (
 SELECT 1 FROM assessments a
 LEFT JOIN records r ON r.record_id=a.record_id
 LEFT JOIN records s ON s.record_id=a.subject_record_id
 WHERE a.incident_id=$1 AND
 (r.record_id IS NULL OR r.incident_id<>$1 OR r.record_type<>'assessment' OR
 s.record_id IS NULL OR s.incident_id<>$1 OR s.record_type<>a.subject_type))`, p.incidentID).Scan(&invalid); err != nil {
		return err
	}
	if invalid {
		return descriptor.DeclaredFailure("assessments.subject_type_scope")
	}
	rows, err := tx.Query(ctx, `SELECT record_id,incident_id,subject_record_id,subject_type,assessment_state,confidence_score,
 assessor_user_id,assessed_at,created_at,updated_at,deleted_at,deleted_by_user_id,rationale
 FROM assessments WHERE incident_id=$1 ORDER BY record_id`, p.incidentID)
	if err != nil {
		return err
	}
	defer rows.Close()
	err = p.walk(ctx, input, func(expected sourcevalue.Row) error {
		if !rows.Next() {
			if err := rows.Err(); err != nil {
				return err
			}
			return descriptor.DeclaredFailure("assessments.source_state_exact")
		}
		var actual sourcevalue.Row
		if err := rows.Scan(&actual.RecordID, &actual.IncidentID, &actual.SubjectRecordID, &actual.SubjectType, &actual.AssessmentState, &actual.ConfidenceScore,
			&actual.AssessorUserID, &actual.AssessedAt, &actual.CreatedAt, &actual.UpdatedAt, &actual.DeletedAt, &actual.DeletedByUserID, &actual.Rationale); err != nil {
			return err
		}
		actual.AssessedAt, actual.CreatedAt, actual.UpdatedAt = actual.AssessedAt.UTC(), actual.CreatedAt.UTC(), actual.UpdatedAt.UTC()
		if actual.DeletedAt != nil {
			value := actual.DeletedAt.UTC()
			actual.DeletedAt = &value
		}
		expected.AssessorUserID = p.actorID
		if expected.DeletedByUserID != nil {
			expected.DeletedByUserID = &p.actorID
		}
		if !reflect.DeepEqual(actual, expected) {
			return descriptor.DeclaredFailure("assessments.source_state_exact")
		}
		return nil
	})
	if err != nil {
		return err
	}
	if rows.Next() {
		return descriptor.DeclaredFailure("assessments.source_state_exact")
	}
	if err := rows.Err(); err != nil {
		return err
	}
	rows.Close()
	// Attribution is flushed by the core in step 9 and checked there by counts.
	return nil
}

type rowActor struct {
	column string
	id     uuid.UUID
}

func rowActors(row sourcevalue.Row) []rowActor {
	actors := []rowActor{{"assessor_user_id", row.AssessorUserID}}
	if row.DeletedByUserID != nil {
		actors = append(actors, rowActor{"deleted_by_user_id", *row.DeletedByUserID})
	}
	return actors
}
