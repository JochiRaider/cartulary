package artifacts

import (
	"context"
	"reflect"

	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/modules/artifacts/internal/portablevalue"
	"github.com/JochiRaider/cartulary/internal/modules/incidentbundles/sourceport"
	"github.com/JochiRaider/cartulary/internal/modules/incidentportability"
	"github.com/JochiRaider/cartulary/internal/platform/sqlshape"
)

func (p preparedArtifactImport) validate(ctx context.Context, tx pgx.Tx, input sourceport.ImportContext, descriptor sourceport.Descriptor) error {
	for _, check := range []struct{ invariant, query string }{
		{"envelope_type_scope", `SELECT EXISTS (
 SELECT 1 FROM artifacts a LEFT JOIN records r ON r.record_id=a.record_id
 WHERE a.incident_id=$1 AND (r.record_id IS NULL OR r.incident_id<>$1 OR r.record_type<>'artifact')
 UNION ALL SELECT 1 FROM artifact_findings a LEFT JOIN records r ON r.record_id=a.record_id
 WHERE a.incident_id=$1 AND (r.record_id IS NULL OR r.incident_id<>$1 OR r.record_type<>'artifact')
 UNION ALL SELECT 1 FROM artifact_investigative_queries a LEFT JOIN records r ON r.record_id=a.record_id
 WHERE a.incident_id=$1 AND (r.record_id IS NULL OR r.incident_id<>$1 OR r.record_type<>'artifact')
 UNION ALL SELECT 1 FROM artifact_forensic_keywords a LEFT JOIN records r ON r.record_id=a.record_id
 WHERE a.incident_id=$1 AND (r.record_id IS NULL OR r.incident_id<>$1 OR r.record_type<>'artifact'))`},
		{"subtype_exact", `SELECT EXISTS (
 SELECT 1 FROM artifacts a
 LEFT JOIN artifact_findings f USING (incident_id,record_id)
 LEFT JOIN artifact_investigative_queries q USING (incident_id,record_id)
 LEFT JOIN artifact_forensic_keywords k USING (incident_id,record_id)
 WHERE a.incident_id=$1 AND
 (a.artifact_type NOT IN ('note','comm_log','handoff','status_review','lesson','finding','investigative_query','forensic_keyword')
 OR ((a.artifact_type='finding')<>(f.record_id IS NOT NULL))
 OR ((a.artifact_type='investigative_query')<>(q.record_id IS NOT NULL))
 OR ((a.artifact_type='forensic_keyword')<>(k.record_id IS NOT NULL)))
 UNION ALL SELECT 1 FROM artifact_findings f LEFT JOIN artifacts a USING (incident_id,record_id)
 WHERE f.incident_id=$1 AND a.record_id IS NULL
 UNION ALL SELECT 1 FROM artifact_investigative_queries q LEFT JOIN artifacts a USING (incident_id,record_id)
 WHERE q.incident_id=$1 AND a.record_id IS NULL
 UNION ALL SELECT 1 FROM artifact_forensic_keywords k LEFT JOIN artifacts a USING (incident_id,record_id)
 WHERE k.incident_id=$1 AND a.record_id IS NULL)`},
		{"handoff_risk_target", `SELECT EXISTS (
 SELECT 1 FROM handoff_risk_refs h LEFT JOIN artifacts a ON a.record_id=h.handoff_record_id
 LEFT JOIN records r ON r.record_id=h.handoff_record_id
 WHERE h.incident_id=$1 AND
 (a.record_id IS NULL OR a.incident_id<>$1 OR a.artifact_type<>'handoff'
 OR r.record_id IS NULL OR r.incident_id<>$1 OR r.record_type<>'artifact'))`},
		{"handoff_risk_legal", `SELECT EXISTS (
 SELECT 1 FROM handoff_risk_refs WHERE incident_id=$1 AND deleted_at IS NULL
 GROUP BY handoff_record_id,risk_ref_text HAVING count(*)<>1)`},
	} {
		var invalid bool
		if err := tx.QueryRow(ctx, check.query, p.incidentID).Scan(&invalid); err != nil {
			return err
		}
		if invalid {
			return descriptor.DeclaredFailure("artifacts." + check.invariant)
		}
	}
	for _, relation := range p.manifest.relations {
		if err := p.compareRelation(ctx, tx, descriptor, relation); err != nil {
			return err
		}
	}
	return p.validateAttributions(ctx, input.Attributions, descriptor)
}

func (p preparedArtifactImport) compareRelation(ctx context.Context, tx pgx.Tx, descriptor sourceport.Descriptor, relation sourceStateRelation) error {
	query, valid := artifactSourceReadQuery(relation.tableName)
	if !valid {
		return sourceport.ErrInvalidCatalog
	}
	rows, err := tx.Query(ctx, sqlshape.Text(query), p.incidentID)
	if err != nil {
		return err
	}
	defer rows.Close()
	err = p.rawPath(ctx, descriptor, relation, func(raw map[string]any) error {
		expected, valid := portablevalue.Parse(relation.tableName, raw)
		if !valid {
			return descriptor.DeclaredFailure("artifacts.row_shape_exact")
		}
		if !rows.Next() {
			if err := rows.Err(); err != nil {
				return err
			}
			return descriptor.DeclaredFailure("artifacts.source_state_exact")
		}
		actual, err := scanArtifactPortableRow(rows, relation.tableName)
		if err != nil {
			return err
		}
		portablevalue.UTCClocks(actual)
		if !reflect.DeepEqual(actual, portablevalue.WithLocalActor(expected, p.actorID)) {
			return descriptor.DeclaredFailure("artifacts.source_state_exact")
		}
		return nil
	})
	if err != nil {
		return err
	}
	if rows.Next() {
		return descriptor.DeclaredFailure("artifacts.source_state_exact")
	}
	return rows.Err()
}

func (p preparedArtifactImport) validateAttributions(ctx context.Context, recorder incidentportability.AttributionRecorder, descriptor sourceport.Descriptor) error {
	expectedCounts := map[string]uint64{}
	if err := p.typed(ctx, descriptor, func(relation sourceStateRelation, row any) error {
		id, _ := portablevalue.IDs(row)
		expected := portablevalue.Actors(row)
		expectedCounts[relation.tableName] += uint64(len(expected))
		if err := recorder.VisitImportedAttributions(relation.tableName, id.String(), func(value incidentportability.ImportedAttribution) error {
			actor, exists := expected[value.SourceColumn]
			if !exists || value.LocalUserID != p.actorID || value.SourceTable != relation.tableName || value.SourceRowID != id.String() || value.SourceActorID != actor.String() {
				return descriptor.DeclaredFailure("artifacts.attribution_exact")
			}
			delete(expected, value.SourceColumn)
			return ctx.Err()
		}); err != nil {
			return err
		}
		if len(expected) != 0 {
			return descriptor.DeclaredFailure("artifacts.attribution_exact")
		}
		return nil
	}); err != nil {
		return err
	}
	for _, relation := range p.manifest.relations {
		var count uint64
		if err := recorder.VisitImportedAttributions(relation.tableName, "", func(incidentportability.ImportedAttribution) error { count++; return ctx.Err() }); err != nil {
			return err
		}
		if count != expectedCounts[relation.tableName] {
			return descriptor.DeclaredFailure("artifacts.attribution_exact")
		}
	}
	return nil
}

func artifactSourceReadQuery(table string) (sqlshape.Fragment, bool) {
	switch table {
	case "artifacts":
		return sqlshape.Literal(`SELECT record_id,incident_id,artifact_type,title,body,timestamp_utc,updated_at,created_at,
 comm_id,comm_type,audience,channel_or_meeting,summary,next_report_at,privilege_tag,
 handoff_id,outgoing_owner_user_id,incoming_owner_user_id,current_state_summary,next_checks,acknowledged_at,
 status_review_id,review_owner_user_id,active_risks_summary,lesson_id,owner_user_id,closure_state,created_by_user_id
 FROM artifacts WHERE incident_id=$1 ORDER BY record_id`), true
	case "artifact_findings":
		return sqlshape.Literal(`SELECT record_id,incident_id,kind,statement,state,confidence_score,owner_user_id,closed_at,created_at,updated_at
 FROM artifact_findings WHERE incident_id=$1 ORDER BY record_id`), true
	case "artifact_investigative_queries":
		return sqlshape.Literal(`SELECT record_id,incident_id,query_id,platform,purpose,query_text,created_by_user_id,created_at,updated_at
 FROM artifact_investigative_queries WHERE incident_id=$1 ORDER BY record_id`), true
	case "artifact_forensic_keywords":
		return sqlshape.Literal(`SELECT record_id,incident_id,keyword_id,pattern,reason,match_mode,case_sensitive,created_at,updated_at
 FROM artifact_forensic_keywords WHERE incident_id=$1 ORDER BY record_id`), true
	case "handoff_risk_refs":
		return sqlshape.Literal(`SELECT risk_ref_id,incident_id,handoff_record_id,risk_ref_text,created_by_user_id,created_at,deleted_at,deleted_by_user_id
 FROM handoff_risk_refs WHERE incident_id=$1 ORDER BY risk_ref_id`), true
	default:
		return sqlshape.Fragment{}, false
	}
}

func scanArtifactPortableRow(row pgx.Row, table string) (any, error) {
	switch table {
	case "artifacts":
		value := &portablevalue.Parent{}
		err := row.Scan(&value.RecordID, &value.IncidentID, &value.ArtifactType, &value.Title, &value.Body, &value.TimestampUTC, &value.UpdatedAt, &value.CreatedAt,
			&value.CommID, &value.CommType, &value.Audience, &value.ChannelOrMeeting, &value.Summary, &value.NextReportAt, &value.PrivilegeTag,
			&value.HandoffID, &value.OutgoingOwnerUserID, &value.IncomingOwnerUserID, &value.CurrentStateSummary, &value.NextChecks, &value.AcknowledgedAt,
			&value.StatusReviewID, &value.ReviewOwnerUserID, &value.ActiveRisksSummary, &value.LessonID, &value.OwnerUserID, &value.ClosureState, &value.CreatedByUserID)
		return value, err
	case "artifact_findings":
		value := &portablevalue.Finding{}
		err := row.Scan(&value.RecordID, &value.IncidentID, &value.Kind, &value.Statement, &value.State, &value.ConfidenceScore, &value.OwnerUserID, &value.ClosedAt, &value.CreatedAt, &value.UpdatedAt)
		return value, err
	case "artifact_investigative_queries":
		value := &portablevalue.Query{}
		err := row.Scan(&value.RecordID, &value.IncidentID, &value.QueryID, &value.Platform, &value.Purpose, &value.QueryText, &value.CreatedByUserID, &value.CreatedAt, &value.UpdatedAt)
		return value, err
	case "artifact_forensic_keywords":
		value := &portablevalue.Keyword{}
		err := row.Scan(&value.RecordID, &value.IncidentID, &value.KeywordID, &value.Pattern, &value.Reason, &value.MatchMode, &value.CaseSensitive, &value.CreatedAt, &value.UpdatedAt)
		return value, err
	case "handoff_risk_refs":
		value := &portablevalue.Risk{}
		err := row.Scan(&value.RiskRefID, &value.IncidentID, &value.HandoffRecordID, &value.RiskRefText, &value.CreatedByUserID, &value.CreatedAt, &value.DeletedAt, &value.DeletedByUserID)
		return value, err
	default:
		return nil, sourceport.ErrPreparedBinding
	}
}
