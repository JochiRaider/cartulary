package projections

import (
	"fmt"

	"github.com/JochiRaider/cartulary/internal/platform/viewschema"
)

func artifactQuerySurfaces() []genericSurface {
	return []genericSurface{
		artifactSurface(findingsViewSchemaID, "finding", []genericField{
			{key: "finding.statement", expr: "p.finding_statement", kind: fieldKindText},
			{key: "finding.kind", expr: "p.finding_kind", sortExpr: enumSortExpr("p.finding_kind", "finding", "hypothesis"), kind: fieldKindText},
			{key: "finding.state", expr: "p.finding_state", sortExpr: enumSortExpr("p.finding_state", "open", "closed"), kind: fieldKindText},
			{key: "finding.owner_user_id", expr: "p.finding_owner_user_id", kind: fieldKindText},
			{key: "finding.confidence_score", expr: "p.finding_confidence_score", kind: fieldKindNumber},
			{key: "finding.closed_at", expr: "p.finding_closed_at", kind: fieldKindTimestamp},
			{key: "finding.updated_at", expr: "p.finding_updated_at", kind: fieldKindTimestamp},
			{key: "finding.supporting_refs", expr: recordRefCollectionExprFor("p", "finding.supporting_refs", "supported_by"), kind: fieldKindCollection},
			{key: "finding.contradictory_refs", expr: recordRefCollectionExprFor("p", "finding.contradictory_refs", "references_record"), kind: fieldKindCollection},
			{key: "finding.confidence_band", expr: "p.finding_confidence_band", sortExpr: enumSortExpr("p.finding_confidence_band", "unset", "low", "medium", "high"), kind: fieldKindText},
		}),
		artifactSurface(forensicKeywordsViewSchemaID, "forensic_keyword", []genericField{
			{key: "forensic_keyword.pattern", expr: "p.forensic_keyword_pattern", kind: fieldKindText},
			{key: "forensic_keyword.reason", expr: "p.forensic_keyword_reason", kind: fieldKindText},
			{key: "forensic_keyword.match_mode", expr: "p.forensic_keyword_match_mode", sortExpr: enumSortExpr("p.forensic_keyword_match_mode", "literal", "regex"), kind: fieldKindText},
			{key: "forensic_keyword.case_sensitive", expr: "p.forensic_keyword_case_sensitive", kind: fieldKindBool},
			{key: "forensic_keyword.created_at", expr: "p.forensic_keyword_created_at", kind: fieldKindTimestamp},
			{key: "forensic_keyword.keyword_id", expr: "p.forensic_keyword_keyword_id", kind: fieldKindText},
			{key: "forensic_keyword.created_day", expr: "p.forensic_keyword_created_day", kind: fieldKindDate},
		}),
		artifactSurface(investigativeQueriesViewSchemaID, "investigative_query", []genericField{
			{key: "investigative_query.platform", expr: "p.investigative_query_platform", kind: fieldKindText},
			{key: "investigative_query.purpose", expr: "p.investigative_query_purpose", kind: fieldKindText},
			{key: "investigative_query.query_text", expr: "p.investigative_query_query_text", kind: fieldKindText},
			{key: "investigative_query.created_by_user_id", expr: "p.investigative_query_created_by_user_id", kind: fieldKindText},
			{key: "investigative_query.created_at", expr: "p.investigative_query_created_at", kind: fieldKindTimestamp},
			{key: "investigative_query.query_id", expr: "p.investigative_query_query_id", kind: fieldKindText},
			{key: "investigative_query.created_day", expr: "p.investigative_query_created_day", kind: fieldKindDate},
		}),
		artifactSurface(notesViewSchemaID, "note", []genericField{
			{key: "note.title", expr: "p.title", kind: fieldKindText},
			{key: "note.body", expr: "p.body", kind: fieldKindText},
			{key: "note.tags", expr: tagCollectionExprFor("p"), kind: fieldKindCollection},
			{key: "note.linked_record_count", expr: "p.linked_record_count", kind: fieldKindNumber},
			{key: "note.updated_at", expr: "p.updated_at", kind: fieldKindTimestamp},
			{key: "note.created_by_user_id", expr: "p.created_by_user_id", kind: fieldKindText},
		}),
		artifactSurface(commLogViewSchemaID, "", []genericField{
			{key: "comm_log.timestamp_utc", expr: "p.timestamp_utc", kind: fieldKindTimestamp},
			{key: "comm_log.comm_type", expr: "p.comm_type", kind: fieldKindText},
			{key: "comm_log.audience", expr: "p.audience", kind: fieldKindText},
			{key: "comm_log.channel_or_meeting", expr: "p.channel_or_meeting", kind: fieldKindText},
			{key: "comm_log.summary", expr: "p.summary", kind: fieldKindText},
			{key: "comm_log.next_report_at", expr: "p.next_report_at", kind: fieldKindTimestamp},
			{key: "comm_log.privilege_tag", expr: "p.privilege_tag", kind: fieldKindText},
			{key: "comm_log.decision_ids", expr: recordRefCollectionExpr("comm_log.decision_ids"), kind: fieldKindCollection},
			{key: "comm_log.action_task_ids", expr: recordRefCollectionExpr("comm_log.action_task_ids"), kind: fieldKindCollection},
			{key: "comm_log.audience_party_ids", expr: partyRefCollectionExpr("comm_log.audience_party_ids"), kind: fieldKindCollection},
			{key: "comm_log.attendee_party_ids", expr: partyRefCollectionExpr("comm_log.attendee_party_ids"), kind: fieldKindCollection},
			{key: "comm_log.comm_id", expr: "p.comm_id", kind: fieldKindText},
			{key: "comm_log.timestamp_day", expr: "p.timestamp_day", kind: fieldKindDate},
			{key: "comm_log.next_report_day", expr: "p.next_report_day", kind: fieldKindDate},
			{key: "comm_log.updated_at", expr: "p.updated_at", kind: fieldKindTimestamp},
		}),
		artifactSurface(handoffViewSchemaID, "", []genericField{
			{key: "handoff.timestamp_utc", expr: "p.timestamp_utc", kind: fieldKindTimestamp},
			{key: "handoff.outgoing_owner_user_id", expr: "p.outgoing_owner_user_id", kind: fieldKindText},
			{key: "handoff.incoming_owner_user_id", expr: "p.incoming_owner_user_id", kind: fieldKindText},
			{key: "handoff.current_state_summary", expr: "p.current_state_summary", kind: fieldKindText},
			{key: "handoff.open_task_ids", expr: recordRefCollectionExpr("handoff.open_task_ids"), kind: fieldKindCollection},
			{key: "handoff.open_decision_ids", expr: recordRefCollectionExpr("handoff.open_decision_ids"), kind: fieldKindCollection},
			{key: "handoff.open_risk_refs", expr: riskRefCollectionExpr(), kind: fieldKindCollection},
			{key: "handoff.next_checks", expr: "p.next_checks", kind: fieldKindText},
			{key: "handoff.acknowledged_at", expr: "p.acknowledged_at", kind: fieldKindTimestamp},
			{key: "handoff.handoff_id", expr: "p.handoff_id", kind: fieldKindText},
			{key: "handoff.timestamp_day", expr: "p.timestamp_day", kind: fieldKindDate},
			{key: "handoff.ack_state", expr: "p.ack_state", sortExpr: enumSortExpr("p.ack_state", "pending", "acknowledged"), kind: fieldKindText},
			{key: "handoff.updated_at", expr: "p.updated_at", kind: fieldKindTimestamp},
		}),
		artifactSurface(statusReviewViewSchemaID, "", []genericField{
			{key: "status_review.timestamp_utc", expr: "p.timestamp_utc", kind: fieldKindTimestamp},
			{key: "status_review.review_owner_user_id", expr: "p.review_owner_user_id", kind: fieldKindText},
			{key: "status_review.current_state_summary", expr: "p.current_state_summary", kind: fieldKindText},
			{key: "status_review.blocked_task_ids", expr: recordRefCollectionExpr("status_review.blocked_task_ids"), kind: fieldKindCollection},
			{key: "status_review.pending_evidence_ids", expr: recordRefCollectionExpr("status_review.pending_evidence_ids"), kind: fieldKindCollection},
			{key: "status_review.open_decision_ids", expr: recordRefCollectionExpr("status_review.open_decision_ids"), kind: fieldKindCollection},
			{key: "status_review.active_risks_summary", expr: "p.active_risks_summary", kind: fieldKindText},
			{key: "status_review.next_report_at", expr: "p.next_report_at", kind: fieldKindTimestamp},
			{key: "status_review.status_review_id", expr: "p.status_review_id", kind: fieldKindText},
			{key: "status_review.timestamp_day", expr: "p.timestamp_day", kind: fieldKindDate},
			{key: "status_review.next_report_day", expr: "p.next_report_day", kind: fieldKindDate},
			{key: "status_review.updated_at", expr: "p.updated_at", kind: fieldKindTimestamp},
		}),
		artifactSurface(lessonViewSchemaID, "", []genericField{
			{key: "lesson.timestamp_utc", expr: "p.timestamp_utc", kind: fieldKindTimestamp},
			{key: "lesson.summary", expr: "p.summary", kind: fieldKindText},
			{key: "lesson.owner_user_id", expr: "p.owner_user_id", kind: fieldKindText},
			{key: "lesson.closure_state", expr: "p.closure_state", kind: fieldKindText},
			{key: "lesson.follow_up_task_ids", expr: recordRefCollectionExpr("lesson.follow_up_task_ids"), kind: fieldKindCollection},
			{key: "lesson.evidence_refs", expr: recordRefCollectionExpr("lesson.evidence_refs"), kind: fieldKindCollection},
			{key: "lesson.lesson_id", expr: "p.lesson_id", kind: fieldKindText},
			{key: "lesson.timestamp_day", expr: "p.timestamp_day", kind: fieldKindDate},
			{key: "lesson.updated_at", expr: "p.updated_at", kind: fieldKindTimestamp},
		}),
	}
}

func artifactSurface(viewSchemaID string, fallbackArtifactType string, fields []genericField) genericSurface {
	artifactType := artifactTypeForSurface(viewSchemaID, fallbackArtifactType)
	return genericSurface{
		viewSchemaID: viewSchemaID,
		fromSQL:      "FROM artifact_grid_projection p JOIN records r ON r.record_id = p.record_id",
		recordExpr:   "p.record_id",
		incidentExpr: "p.incident_id",
		whereSQL:     "p.artifact_type = '" + artifactType + "'",
		fields:       fields,
	}
}

func artifactTypeForSurface(viewSchemaID string, fallbackArtifactType string) string {
	schema, ok := viewschema.Lookup(viewSchemaID)
	if ok {
		if filter, hasFilter := schema.CanonicalSourceFilter(); hasFilter {
			if schema.BaseProjection != "artifact_grid_projection" {
				panic(fmt.Sprintf("artifact surface %s declares base_projection=%q", viewSchemaID, schema.BaseProjection))
			}
			if filter.Kind != "artifact_type" || filter.Field != "artifact_type" || filter.Value == "" {
				panic(fmt.Sprintf("artifact surface %s declares invalid canonical source filter %#v", viewSchemaID, filter))
			}
			if fallbackArtifactType != "" && fallbackArtifactType != filter.Value {
				panic(fmt.Sprintf("artifact surface %s fallback artifact_type=%q contradicts contract value %q", viewSchemaID, fallbackArtifactType, filter.Value))
			}
			return filter.Value
		}
	}
	if fallbackArtifactType == "" {
		panic(fmt.Sprintf("artifact surface %s missing canonical artifact_type filter", viewSchemaID))
	}
	return fallbackArtifactType
}
