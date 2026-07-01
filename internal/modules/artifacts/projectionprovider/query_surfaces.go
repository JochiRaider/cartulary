package projectionprovider

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/JochiRaider/cartulary/internal/modules/projections/providercontract"
	"github.com/JochiRaider/cartulary/internal/platform/viewschema"
)

const (
	commLogViewSchemaID              = "cartulary.view.comm_log.v1"
	findingsViewSchemaID             = "cartulary.view.findings.v1"
	forensicKeywordsViewSchemaID     = "cartulary.view.forensic_keywords.v1"
	handoffViewSchemaID              = "cartulary.view.handoff.v1"
	investigativeQueriesViewSchemaID = "cartulary.view.investigative_queries.v1"
	lessonViewSchemaID               = "cartulary.view.lesson.v1"
	notesViewSchemaID                = "cartulary.view.notes.v1"
	statusReviewViewSchemaID         = "cartulary.view.status_review.v1"
)

func QuerySurfaces() []providercontract.QuerySurface {
	return []providercontract.QuerySurface{
		artifactSurface(findingsViewSchemaID, "finding", []providercontract.QueryField{
			{Key: "finding.statement", Expr: "p.finding_statement", Kind: providercontract.FieldKindText},
			{Key: "finding.kind", Expr: "p.finding_kind", SortExpr: enumSortExpr("p.finding_kind", "finding", "hypothesis"), Kind: providercontract.FieldKindText},
			{Key: "finding.state", Expr: "p.finding_state", SortExpr: enumSortExpr("p.finding_state", "open", "closed"), Kind: providercontract.FieldKindText},
			{Key: "finding.owner_user_id", Expr: "p.finding_owner_user_id", Kind: providercontract.FieldKindText},
			{Key: "finding.confidence_score", Expr: "p.finding_confidence_score", Kind: providercontract.FieldKindNumber},
			{Key: "finding.closed_at", Expr: "p.finding_closed_at", Kind: providercontract.FieldKindTimestamp},
			{Key: "finding.updated_at", Expr: "p.finding_updated_at", Kind: providercontract.FieldKindTimestamp},
			{Key: "finding.supporting_refs", Expr: recordRefCollectionExprFor("p", "finding.supporting_refs", "supported_by"), Kind: providercontract.FieldKindCollection},
			{Key: "finding.contradictory_refs", Expr: recordRefCollectionExprFor("p", "finding.contradictory_refs", "references_record"), Kind: providercontract.FieldKindCollection},
			{Key: "finding.confidence_band", Expr: "p.finding_confidence_band", SortExpr: enumSortExpr("p.finding_confidence_band", "unset", "low", "medium", "high"), Kind: providercontract.FieldKindText},
		}),
		artifactSurface(forensicKeywordsViewSchemaID, "forensic_keyword", []providercontract.QueryField{
			{Key: "forensic_keyword.pattern", Expr: "p.forensic_keyword_pattern", Kind: providercontract.FieldKindText},
			{Key: "forensic_keyword.reason", Expr: "p.forensic_keyword_reason", Kind: providercontract.FieldKindText},
			{Key: "forensic_keyword.match_mode", Expr: "p.forensic_keyword_match_mode", SortExpr: enumSortExpr("p.forensic_keyword_match_mode", "literal", "regex"), Kind: providercontract.FieldKindText},
			{Key: "forensic_keyword.case_sensitive", Expr: "p.forensic_keyword_case_sensitive", Kind: providercontract.FieldKindBool},
			{Key: "forensic_keyword.created_at", Expr: "p.forensic_keyword_created_at", Kind: providercontract.FieldKindTimestamp},
			{Key: "forensic_keyword.keyword_id", Expr: "p.forensic_keyword_keyword_id", Kind: providercontract.FieldKindText},
			{Key: "forensic_keyword.created_day", Expr: "p.forensic_keyword_created_day", Kind: providercontract.FieldKindDate},
		}),
		artifactSurface(investigativeQueriesViewSchemaID, "investigative_query", []providercontract.QueryField{
			{Key: "investigative_query.platform", Expr: "p.investigative_query_platform", Kind: providercontract.FieldKindText},
			{Key: "investigative_query.purpose", Expr: "p.investigative_query_purpose", Kind: providercontract.FieldKindText},
			{Key: "investigative_query.query_text", Expr: "p.investigative_query_query_text", Kind: providercontract.FieldKindText},
			{Key: "investigative_query.created_by_user_id", Expr: "p.investigative_query_created_by_user_id", Kind: providercontract.FieldKindText},
			{Key: "investigative_query.created_at", Expr: "p.investigative_query_created_at", Kind: providercontract.FieldKindTimestamp},
			{Key: "investigative_query.query_id", Expr: "p.investigative_query_query_id", Kind: providercontract.FieldKindText},
			{Key: "investigative_query.created_day", Expr: "p.investigative_query_created_day", Kind: providercontract.FieldKindDate},
		}),
		artifactSurface(notesViewSchemaID, "note", []providercontract.QueryField{
			{Key: "note.title", Expr: "p.title", Kind: providercontract.FieldKindText},
			{Key: "note.body", Expr: "p.body", Kind: providercontract.FieldKindText},
			{Key: "note.tags", Expr: tagCollectionExprFor("p"), Kind: providercontract.FieldKindCollection},
			{Key: "note.linked_record_count", Expr: "p.linked_record_count", Kind: providercontract.FieldKindNumber},
			{Key: "note.updated_at", Expr: "p.updated_at", Kind: providercontract.FieldKindTimestamp},
			{Key: "note.created_by_user_id", Expr: "p.created_by_user_id", Kind: providercontract.FieldKindText},
		}),
		artifactSurface(commLogViewSchemaID, "", []providercontract.QueryField{
			{Key: "comm_log.timestamp_utc", Expr: "p.timestamp_utc", Kind: providercontract.FieldKindTimestamp},
			{Key: "comm_log.comm_type", Expr: "p.comm_type", Kind: providercontract.FieldKindText},
			{Key: "comm_log.audience", Expr: "p.audience", Kind: providercontract.FieldKindText},
			{Key: "comm_log.channel_or_meeting", Expr: "p.channel_or_meeting", Kind: providercontract.FieldKindText},
			{Key: "comm_log.summary", Expr: "p.summary", Kind: providercontract.FieldKindText},
			{Key: "comm_log.next_report_at", Expr: "p.next_report_at", Kind: providercontract.FieldKindTimestamp},
			{Key: "comm_log.privilege_tag", Expr: "p.privilege_tag", Kind: providercontract.FieldKindText},
			{Key: "comm_log.decision_ids", Expr: recordRefCollectionExpr("comm_log.decision_ids"), Kind: providercontract.FieldKindCollection},
			{Key: "comm_log.action_task_ids", Expr: recordRefCollectionExpr("comm_log.action_task_ids"), Kind: providercontract.FieldKindCollection},
			{Key: "comm_log.audience_party_ids", Expr: partyRefCollectionExpr("comm_log.audience_party_ids"), Kind: providercontract.FieldKindCollection},
			{Key: "comm_log.attendee_party_ids", Expr: partyRefCollectionExpr("comm_log.attendee_party_ids"), Kind: providercontract.FieldKindCollection},
			{Key: "comm_log.comm_id", Expr: "p.comm_id", Kind: providercontract.FieldKindText},
			{Key: "comm_log.timestamp_day", Expr: "p.timestamp_day", Kind: providercontract.FieldKindDate},
			{Key: "comm_log.next_report_day", Expr: "p.next_report_day", Kind: providercontract.FieldKindDate},
			{Key: "comm_log.updated_at", Expr: "p.updated_at", Kind: providercontract.FieldKindTimestamp},
		}),
		artifactSurface(handoffViewSchemaID, "", []providercontract.QueryField{
			{Key: "handoff.timestamp_utc", Expr: "p.timestamp_utc", Kind: providercontract.FieldKindTimestamp},
			{Key: "handoff.outgoing_owner_user_id", Expr: "p.outgoing_owner_user_id", Kind: providercontract.FieldKindText},
			{Key: "handoff.incoming_owner_user_id", Expr: "p.incoming_owner_user_id", Kind: providercontract.FieldKindText},
			{Key: "handoff.current_state_summary", Expr: "p.current_state_summary", Kind: providercontract.FieldKindText},
			{Key: "handoff.open_task_ids", Expr: recordRefCollectionExpr("handoff.open_task_ids"), Kind: providercontract.FieldKindCollection},
			{Key: "handoff.open_decision_ids", Expr: recordRefCollectionExpr("handoff.open_decision_ids"), Kind: providercontract.FieldKindCollection},
			{Key: "handoff.open_risk_refs", Expr: riskRefCollectionExpr(), Kind: providercontract.FieldKindCollection},
			{Key: "handoff.next_checks", Expr: "p.next_checks", Kind: providercontract.FieldKindText},
			{Key: "handoff.acknowledged_at", Expr: "p.acknowledged_at", Kind: providercontract.FieldKindTimestamp},
			{Key: "handoff.handoff_id", Expr: "p.handoff_id", Kind: providercontract.FieldKindText},
			{Key: "handoff.timestamp_day", Expr: "p.timestamp_day", Kind: providercontract.FieldKindDate},
			{Key: "handoff.ack_state", Expr: "p.ack_state", SortExpr: enumSortExpr("p.ack_state", "pending", "acknowledged"), Kind: providercontract.FieldKindText},
			{Key: "handoff.updated_at", Expr: "p.updated_at", Kind: providercontract.FieldKindTimestamp},
		}),
		artifactSurface(statusReviewViewSchemaID, "", []providercontract.QueryField{
			{Key: "status_review.timestamp_utc", Expr: "p.timestamp_utc", Kind: providercontract.FieldKindTimestamp},
			{Key: "status_review.review_owner_user_id", Expr: "p.review_owner_user_id", Kind: providercontract.FieldKindText},
			{Key: "status_review.current_state_summary", Expr: "p.current_state_summary", Kind: providercontract.FieldKindText},
			{Key: "status_review.blocked_task_ids", Expr: recordRefCollectionExpr("status_review.blocked_task_ids"), Kind: providercontract.FieldKindCollection},
			{Key: "status_review.pending_evidence_ids", Expr: recordRefCollectionExpr("status_review.pending_evidence_ids"), Kind: providercontract.FieldKindCollection},
			{Key: "status_review.open_decision_ids", Expr: recordRefCollectionExpr("status_review.open_decision_ids"), Kind: providercontract.FieldKindCollection},
			{Key: "status_review.active_risks_summary", Expr: "p.active_risks_summary", Kind: providercontract.FieldKindText},
			{Key: "status_review.next_report_at", Expr: "p.next_report_at", Kind: providercontract.FieldKindTimestamp},
			{Key: "status_review.status_review_id", Expr: "p.status_review_id", Kind: providercontract.FieldKindText},
			{Key: "status_review.timestamp_day", Expr: "p.timestamp_day", Kind: providercontract.FieldKindDate},
			{Key: "status_review.next_report_day", Expr: "p.next_report_day", Kind: providercontract.FieldKindDate},
			{Key: "status_review.updated_at", Expr: "p.updated_at", Kind: providercontract.FieldKindTimestamp},
		}),
		artifactSurface(lessonViewSchemaID, "", []providercontract.QueryField{
			{Key: "lesson.timestamp_utc", Expr: "p.timestamp_utc", Kind: providercontract.FieldKindTimestamp},
			{Key: "lesson.summary", Expr: "p.summary", Kind: providercontract.FieldKindText},
			{Key: "lesson.owner_user_id", Expr: "p.owner_user_id", Kind: providercontract.FieldKindText},
			{Key: "lesson.closure_state", Expr: "p.closure_state", Kind: providercontract.FieldKindText},
			{Key: "lesson.follow_up_task_ids", Expr: recordRefCollectionExpr("lesson.follow_up_task_ids"), Kind: providercontract.FieldKindCollection},
			{Key: "lesson.evidence_refs", Expr: recordRefCollectionExpr("lesson.evidence_refs"), Kind: providercontract.FieldKindCollection},
			{Key: "lesson.lesson_id", Expr: "p.lesson_id", Kind: providercontract.FieldKindText},
			{Key: "lesson.timestamp_day", Expr: "p.timestamp_day", Kind: providercontract.FieldKindDate},
			{Key: "lesson.updated_at", Expr: "p.updated_at", Kind: providercontract.FieldKindTimestamp},
		}),
	}
}

func artifactSurface(viewSchemaID string, fallbackArtifactType string, fields []providercontract.QueryField) providercontract.QuerySurface {
	artifactType := artifactTypeForSurface(viewSchemaID, fallbackArtifactType)
	return providercontract.QuerySurface{
		ViewSchemaID: viewSchemaID,
		FromSQL:      "FROM artifact_grid_projection p JOIN records r ON r.record_id = p.record_id",
		RecordExpr:   "p.record_id",
		IncidentExpr: "p.incident_id",
		WhereSQL:     "p.artifact_type = '" + artifactType + "'",
		Fields:       fields,
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

func enumSortExpr(expr string, values ...string) string {
	var builder strings.Builder
	builder.WriteString("CASE ")
	builder.WriteString(expr)
	for index, value := range values {
		builder.WriteString(" WHEN '")
		builder.WriteString(value)
		builder.WriteString("' THEN ")
		builder.WriteString(strconv.Itoa(index))
	}
	builder.WriteString(" ELSE ")
	builder.WriteString(strconv.Itoa(len(values)))
	builder.WriteString(" END")
	return builder.String()
}

func recordRefCollectionExpr(fieldKey string) string {
	return recordRefCollectionExprFor("p", fieldKey, "references_record")
}

func recordRefCollectionExprFor(alias string, fieldKey string, linkType string) string {
	return `(SELECT COALESCE(jsonb_agg(jsonb_build_object(
        'item_ref', 'record_ref:' || dst.record_id::text,
        'item_kind', 'record_ref',
        'display_text', dst.record_type || ':' || dst.record_id::text,
        'linked_record_id', dst.record_id::text
    ) ORDER BY dst.record_type ASC, dst.record_id ASC), '[]'::jsonb)
      FROM record_links rl
      JOIN records dst
        ON dst.incident_id = rl.incident_id
       AND dst.record_id = rl.dst_record_id
       AND dst.deleted_at IS NULL
     WHERE rl.incident_id = ` + alias + `.incident_id
       AND rl.src_record_id = ` + alias + `.record_id
       AND rl.link_type = '` + linkType + `'
       AND rl.field_key = '` + fieldKey + `'
       AND rl.deleted_at IS NULL)::text`
}

func tagCollectionExprFor(alias string) string {
	return `(SELECT COALESCE(jsonb_agg(jsonb_build_object(
        'item_ref', 'record_tag:' || rt.record_id::text || ':' || rt.record_tag_id::text,
        'item_kind', 'tag',
        'display_text', rt.tag_name,
        'tag_id', rt.record_tag_id::text
    ) ORDER BY rt.normalized_tag_name ASC, rt.record_tag_id ASC), '[]'::jsonb)
      FROM record_tags rt
     WHERE rt.incident_id = ` + alias + `.incident_id
       AND rt.record_id = ` + alias + `.record_id
       AND rt.deleted_at IS NULL)::text`
}

func partyRefCollectionExpr(fieldKey string) string {
	return `(SELECT COALESCE(jsonb_agg(jsonb_build_object(
        'item_ref', 'party_ref:' || party.record_id::text,
        'item_kind', 'party_ref',
        'display_text', party.display_name,
        'party_id', party.record_id::text
    ) ORDER BY party.display_name ASC, party.record_id ASC), '[]'::jsonb)
      FROM record_links rl
      JOIN parties party
        ON party.incident_id = rl.incident_id
       AND party.record_id = rl.dst_record_id
      JOIN records dst
        ON dst.incident_id = rl.incident_id
       AND dst.record_id = rl.dst_record_id
       AND dst.deleted_at IS NULL
     WHERE rl.incident_id = p.incident_id
       AND rl.src_record_id = p.record_id
       AND rl.link_type = 'references_record'
       AND rl.field_key = '` + fieldKey + `'
       AND rl.deleted_at IS NULL)::text`
}

func riskRefCollectionExpr() string {
	return `(SELECT COALESCE(jsonb_agg(jsonb_build_object(
        'item_ref', 'risk_ref:' || risk_ref_id::text,
        'item_kind', 'risk_ref',
        'display_text', risk_ref_text,
        'risk_ref_id', risk_ref_id::text,
        'risk_ref_text', risk_ref_text
    ) ORDER BY risk_ref_text ASC, risk_ref_id ASC), '[]'::jsonb)
      FROM handoff_risk_refs hr
     WHERE hr.incident_id = p.incident_id
       AND hr.handoff_record_id = p.record_id
       AND hr.deleted_at IS NULL)::text`
}
