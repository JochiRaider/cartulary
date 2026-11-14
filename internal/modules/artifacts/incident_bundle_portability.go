package artifacts

import (
	"context"

	"github.com/JochiRaider/cartulary/internal/modules/incidentbundles/sourceport"
	"github.com/JochiRaider/cartulary/internal/platform/sqlshape"
)

func exportIncidentBundleFiles(ctx context.Context, input sourceport.ExportContext) error {
	manifest, err := loadSourceStateManifest()
	if err != nil {
		return err
	}
	actors := map[string][]string{
		"artifacts":         {"outgoing_owner_user_id", "incoming_owner_user_id", "review_owner_user_id", "owner_user_id", "created_by_user_id"},
		"artifact_findings": {"owner_user_id"}, "artifact_investigative_queries": {"created_by_user_id"}, "handoff_risk_refs": {"created_by_user_id", "deleted_by_user_id"},
	}
	for _, relation := range manifest.relations {
		if len(relation.stableIdentity) != 1 {
			return errInvalidSourceStateManifest
		}
		identity := relation.stableIdentity[0]
		query, err := sourceExportQuery(relation.tableName)
		if err != nil {
			return err
		}
		if err := sourceport.ExportAttributedRows(ctx, input, sourceport.AttributedRows{Path: relation.logicalBundlePath, Table: relation.tableName, Identity: identity, ActorColumns: actors[relation.tableName], Query: query}); err != nil {
			return err
		}
	}
	return nil
}

func sourceExportQuery(table string) (sqlshape.Fragment, error) {
	switch table {
	case "artifacts":
		return sqlshape.Literal(`SELECT to_jsonb(t) FROM (
 SELECT record_id,incident_id,artifact_type,title,body,timestamp_utc,updated_at,created_at,
 comm_id,comm_type,audience,channel_or_meeting,summary,next_report_at,privilege_tag,
 handoff_id,outgoing_owner_user_id,incoming_owner_user_id,current_state_summary,next_checks,acknowledged_at,
 status_review_id,review_owner_user_id,active_risks_summary,lesson_id,owner_user_id,closure_state,created_by_user_id
 FROM artifacts WHERE incident_id=$1 AND ($2::uuid IS NULL OR record_id>$2) ORDER BY record_id LIMIT 128) t`), nil
	case "artifact_findings":
		return sqlshape.Literal(`SELECT to_jsonb(t) FROM (
 SELECT record_id,incident_id,kind,statement,state,confidence_score,owner_user_id,closed_at,created_at,updated_at
 FROM artifact_findings WHERE incident_id=$1 AND ($2::uuid IS NULL OR record_id>$2) ORDER BY record_id LIMIT 128) t`), nil
	case "artifact_investigative_queries":
		return sqlshape.Literal(`SELECT to_jsonb(t) FROM (
 SELECT record_id,incident_id,query_id,platform,purpose,query_text,created_by_user_id,created_at,updated_at
 FROM artifact_investigative_queries WHERE incident_id=$1 AND ($2::uuid IS NULL OR record_id>$2) ORDER BY record_id LIMIT 128) t`), nil
	case "artifact_forensic_keywords":
		return sqlshape.Literal(`SELECT to_jsonb(t) FROM (
 SELECT record_id,incident_id,keyword_id,pattern,reason,match_mode,case_sensitive,created_at,updated_at
 FROM artifact_forensic_keywords WHERE incident_id=$1 AND ($2::uuid IS NULL OR record_id>$2) ORDER BY record_id LIMIT 128) t`), nil
	case "handoff_risk_refs":
		return sqlshape.Literal(`SELECT to_jsonb(t) FROM (
 SELECT risk_ref_id,incident_id,handoff_record_id,risk_ref_text,created_by_user_id,created_at,deleted_at,deleted_by_user_id
 FROM handoff_risk_refs WHERE incident_id=$1 AND ($2::uuid IS NULL OR risk_ref_id>$2) ORDER BY risk_ref_id LIMIT 128) t`), nil
	default:
		return sqlshape.Fragment{}, errInvalidSourceStateManifest
	}
}
