package incidentbundle

import (
	"context"
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

type preparedImport struct {
	binding             *preparationBinding
	members             *incidentportability.BoundMembers
	incidentID, actorID uuid.UUID
	operationID         string
	version             int
}

func (p preparedImport) matches(input sourceport.ImportContext, binding *preparationBinding) bool {
	return p.binding == binding && binding.contractMajor == sourceport.ContractMajor && p.members != nil && p.incidentID == input.IncidentID && p.actorID == input.ActorUserID && p.operationID == input.OperationID && p.version == input.BundleVersion && input.Attributions != nil
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
		if err := p.walk(ctx, descriptor, func(row sourcevalue.Row) error {
			if !check.valid(row) {
				return descriptor.DeclaredFailure("assessments." + check.invariant)
			}
			return nil
		}); err != nil {
			return preparedImport{}, err
		}
	}
	return p, nil
}
func (p preparedImport) raw(ctx context.Context, descriptor sourceport.Descriptor, visit func(map[string]any) error) error {
	err := incidentportability.WalkPath(ctx, p.members, assessmentsBundlePath, visit)
	var malformed *incidentportability.StrictNDJSONDecodeError
	if errors.As(err, &malformed) {
		return descriptor.DeclaredFailure("assessments.row_shape_exact")
	}
	return err
}
func (p preparedImport) walk(ctx context.Context, descriptor sourceport.Descriptor, visit func(sourcevalue.Row) error) error {
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
	if err := p.walk(ctx, descriptor, func(row sourcevalue.Row) error {
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
	return p.walk(ctx, descriptor, func(row sourcevalue.Row) error {
		var deletedBy *uuid.UUID
		if row.DeletedByUserID != nil {
			deletedBy = &p.actorID
		}
		tag, err := tx.Exec(ctx, `INSERT INTO assessments (
 record_id,incident_id,subject_record_id,subject_type,assessment_state,confidence_score,
 assessor_user_id,assessed_at,created_at,updated_at,deleted_at,deleted_by_user_id,rationale
 ) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13)`,
			row.RecordID, row.IncidentID, row.SubjectRecordID, row.SubjectType, row.AssessmentState, row.ConfidenceScore,
			p.actorID, row.AssessedAt, row.CreatedAt, row.UpdatedAt, row.DeletedAt, deletedBy, row.Rationale)
		if err != nil {
			var pgErr *pgconn.PgError
			if errors.As(err, &pgErr) && pgErr.Code == "23505" {
				return descriptor.DeclaredFailure("assessments.source_identity_admitted")
			}
			return err
		}
		if tag.RowsAffected() != 1 {
			return descriptor.DeclaredFailure("assessments.source_identity_admitted")
		}
		for _, actor := range rowActors(row) {
			if err := input.Attributions.RecordImportedAttribution("assessments", row.RecordID.String(), actor.column, actor.id.String()); err != nil {
				return err
			}
		}
		return nil
	})
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
	err = p.walk(ctx, descriptor, func(expected sourcevalue.Row) error {
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
	return p.validateAttributions(ctx, input.Attributions, descriptor)
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
func (p preparedImport) validateAttributions(ctx context.Context, recorder incidentportability.AttributionRecorder, descriptor sourceport.Descriptor) error {
	var expectedCount, actualCount uint64
	err := p.walk(ctx, descriptor, func(row sourcevalue.Row) error {
		expected := map[string]string{}
		for _, actor := range rowActors(row) {
			expected[actor.column] = actor.id.String()
			expectedCount++
		}
		err := recorder.VisitImportedAttributions("assessments", row.RecordID.String(), func(value incidentportability.ImportedAttribution) error {
			if value.LocalUserID != p.actorID || value.SourceTable != "assessments" || value.SourceRowID != row.RecordID.String() || expected[value.SourceColumn] != value.SourceActorID || value.SourceActorID == "" {
				return descriptor.DeclaredFailure("assessments.attribution_exact")
			}
			delete(expected, value.SourceColumn)
			return nil
		})
		if err != nil {
			return err
		}
		if len(expected) != 0 {
			return descriptor.DeclaredFailure("assessments.attribution_exact")
		}
		return nil
	})
	if err != nil {
		return err
	}
	if err := recorder.VisitImportedAttributions("assessments", "", func(incidentportability.ImportedAttribution) error { actualCount++; return ctx.Err() }); err != nil {
		return err
	}
	if actualCount != expectedCount {
		return descriptor.DeclaredFailure("assessments.attribution_exact")
	}
	return nil
}
