package artifacts

import (
	"context"
	"maps"
	"slices"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"

	"github.com/JochiRaider/cartulary/internal/modules/artifacts/internal/portablevalue"
	"github.com/JochiRaider/cartulary/internal/modules/incidentbundles/sourceport"
	"github.com/JochiRaider/cartulary/internal/platform/sqlshape"
)

func (p preparedArtifactImport) apply(ctx context.Context, tx pgx.Tx, input sourceport.ImportContext, descriptor sourceport.Descriptor) error {
	// All envelope preconditions finish before the first owner write.
	if err := p.typed(ctx, descriptor, func(_ sourceStateRelation, row any) error {
		if _, risk := row.(*portablevalue.Risk); risk {
			return nil
		}
		id, incident := portablevalue.IDs(row)
		var valid bool
		if err := tx.QueryRow(ctx, `SELECT EXISTS (SELECT 1 FROM records WHERE record_id=$1 AND incident_id=$2 AND record_type='artifact')`, id, incident).Scan(&valid); err != nil {
			return err
		}
		if !valid {
			return descriptor.DeclaredFailure("artifacts.envelope_type_scope")
		}
		return nil
	}); err != nil {
		return err
	}
	if err := p.typed(ctx, descriptor, func(relation sourceStateRelation, row any) error {
		query, valid := artifactIdentityExistsQuery(relation.tableName)
		if !valid {
			return sourceport.ErrInvalidCatalog
		}
		id, _ := portablevalue.IDs(row)
		var exists bool
		if err := tx.QueryRow(ctx, sqlshape.Text(query), id).Scan(&exists); err != nil {
			return err
		}
		if exists {
			return descriptor.DeclaredFailure("artifacts.source_identity_admitted")
		}
		return nil
	}); err != nil {
		return err
	}
	return p.typed(ctx, descriptor, func(relation sourceStateRelation, row any) error {
		if err := validateArtifactInsertReferences(ctx, tx, row, descriptor); err != nil {
			return err
		}
		tag, err := insertArtifactPortableRow(ctx, tx, portablevalue.WithLocalActor(row, p.actorID))
		if err != nil {
			return err
		}
		if tag.RowsAffected() != 1 {
			return descriptor.DeclaredFailure("artifacts.source_identity_admitted")
		}
		id, _ := portablevalue.IDs(row)
		actors := portablevalue.Actors(row)
		for _, column := range slices.Sorted(maps.Keys(actors)) {
			if err := input.Attributions.RecordImportedAttribution(relation.tableName, id.String(), column, actors[column].String()); err != nil {
				return err
			}
		}
		return nil
	})
}

func artifactIdentityExistsQuery(table string) (sqlshape.Fragment, bool) {
	switch table {
	case "artifacts":
		return sqlshape.Literal(`SELECT EXISTS (SELECT 1 FROM artifacts WHERE record_id=$1)`), true
	case "artifact_findings":
		return sqlshape.Literal(`SELECT EXISTS (SELECT 1 FROM artifact_findings WHERE record_id=$1)`), true
	case "artifact_investigative_queries":
		return sqlshape.Literal(`SELECT EXISTS (SELECT 1 FROM artifact_investigative_queries WHERE record_id=$1)`), true
	case "artifact_forensic_keywords":
		return sqlshape.Literal(`SELECT EXISTS (SELECT 1 FROM artifact_forensic_keywords WHERE record_id=$1)`), true
	case "handoff_risk_refs":
		return sqlshape.Literal(`SELECT EXISTS (SELECT 1 FROM handoff_risk_refs WHERE risk_ref_id=$1)`), true
	default:
		return sqlshape.Fragment{}, false
	}
}

func validateArtifactInsertReferences(ctx context.Context, tx pgx.Tx, row any, descriptor sourceport.Descriptor) error {
	parentType := ""
	switch value := row.(type) {
	case *portablevalue.Finding:
		parentType = "finding"
	case *portablevalue.Query:
		parentType = "investigative_query"
		var duplicate bool
		if err := tx.QueryRow(ctx, `SELECT EXISTS (SELECT 1 FROM artifact_investigative_queries WHERE incident_id=$1 AND query_id=$2)`, value.IncidentID, value.QueryID).Scan(&duplicate); err != nil {
			return err
		}
		if duplicate {
			return descriptor.DeclaredFailure("artifacts.subtype_exact")
		}
	case *portablevalue.Keyword:
		parentType = "forensic_keyword"
		var duplicate bool
		if err := tx.QueryRow(ctx, `SELECT EXISTS (SELECT 1 FROM artifact_forensic_keywords WHERE incident_id=$1 AND keyword_id=$2)`, value.IncidentID, value.KeywordID).Scan(&duplicate); err != nil {
			return err
		}
		if duplicate {
			return descriptor.DeclaredFailure("artifacts.subtype_exact")
		}
	case *portablevalue.Risk:
		var valid bool
		if err := tx.QueryRow(ctx, `SELECT EXISTS (
 SELECT 1 FROM artifacts a JOIN records r ON r.record_id=a.record_id AND r.incident_id=a.incident_id
 WHERE a.record_id=$1 AND a.incident_id=$2 AND a.artifact_type='handoff' AND r.record_type='artifact')`, value.HandoffRecordID, value.IncidentID).Scan(&valid); err != nil {
			return err
		}
		if !valid {
			return descriptor.DeclaredFailure("artifacts.handoff_risk_target")
		}
		if value.DeletedAt == nil {
			var duplicate bool
			if err := tx.QueryRow(ctx, `SELECT EXISTS (SELECT 1 FROM handoff_risk_refs WHERE handoff_record_id=$1 AND risk_ref_text=$2 AND deleted_at IS NULL)`, value.HandoffRecordID, value.RiskRefText).Scan(&duplicate); err != nil {
				return err
			}
			if duplicate {
				return descriptor.DeclaredFailure("artifacts.handoff_risk_legal")
			}
		}
	}
	if parentType != "" {
		id, incident := portablevalue.IDs(row)
		var valid bool
		if err := tx.QueryRow(ctx, `SELECT EXISTS (SELECT 1 FROM artifacts WHERE record_id=$1 AND incident_id=$2 AND artifact_type=$3)`, id, incident, parentType).Scan(&valid); err != nil {
			return err
		}
		if !valid {
			return descriptor.DeclaredFailure("artifacts.subtype_exact")
		}
	}
	return nil
}

func insertArtifactPortableRow(ctx context.Context, tx pgx.Tx, row any) (pgconn.CommandTag, error) {
	switch value := row.(type) {
	case *portablevalue.Parent:
		return tx.Exec(ctx, `INSERT INTO artifacts (
 record_id,incident_id,artifact_type,title,body,timestamp_utc,updated_at,created_at,
 comm_id,comm_type,audience,channel_or_meeting,summary,next_report_at,privilege_tag,
 handoff_id,outgoing_owner_user_id,incoming_owner_user_id,current_state_summary,next_checks,acknowledged_at,
 status_review_id,review_owner_user_id,active_risks_summary,lesson_id,owner_user_id,closure_state,created_by_user_id
 ) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18,$19,$20,$21,$22,$23,$24,$25,$26,$27,$28)`,
			value.RecordID, value.IncidentID, value.ArtifactType, value.Title, value.Body, value.TimestampUTC, value.UpdatedAt, value.CreatedAt,
			value.CommID, value.CommType, value.Audience, value.ChannelOrMeeting, value.Summary, value.NextReportAt, value.PrivilegeTag,
			value.HandoffID, value.OutgoingOwnerUserID, value.IncomingOwnerUserID, value.CurrentStateSummary, value.NextChecks, value.AcknowledgedAt,
			value.StatusReviewID, value.ReviewOwnerUserID, value.ActiveRisksSummary, value.LessonID, value.OwnerUserID, value.ClosureState, value.CreatedByUserID)
	case *portablevalue.Finding:
		return tx.Exec(ctx, `INSERT INTO artifact_findings (record_id,incident_id,kind,statement,state,confidence_score,owner_user_id,closed_at,created_at,updated_at)
 VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)`, value.RecordID, value.IncidentID, value.Kind, value.Statement, value.State, value.ConfidenceScore, value.OwnerUserID, value.ClosedAt, value.CreatedAt, value.UpdatedAt)
	case *portablevalue.Query:
		return tx.Exec(ctx, `INSERT INTO artifact_investigative_queries (record_id,incident_id,query_id,platform,purpose,query_text,created_by_user_id,created_at,updated_at)
 VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)`, value.RecordID, value.IncidentID, value.QueryID, value.Platform, value.Purpose, value.QueryText, value.CreatedByUserID, value.CreatedAt, value.UpdatedAt)
	case *portablevalue.Keyword:
		return tx.Exec(ctx, `INSERT INTO artifact_forensic_keywords (record_id,incident_id,keyword_id,pattern,reason,match_mode,case_sensitive,created_at,updated_at)
 VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)`, value.RecordID, value.IncidentID, value.KeywordID, value.Pattern, value.Reason, value.MatchMode, value.CaseSensitive, value.CreatedAt, value.UpdatedAt)
	case *portablevalue.Risk:
		return tx.Exec(ctx, `INSERT INTO handoff_risk_refs (risk_ref_id,incident_id,handoff_record_id,risk_ref_text,created_by_user_id,created_at,deleted_at,deleted_by_user_id)
 VALUES ($1,$2,$3,$4,$5,$6,$7,$8)`, value.RiskRefID, value.IncidentID, value.HandoffRecordID, value.RiskRefText, value.CreatedByUserID, value.CreatedAt, value.DeletedAt, value.DeletedByUserID)
	default:
		return pgconn.CommandTag{}, sourceport.ErrPreparedBinding
	}
}
