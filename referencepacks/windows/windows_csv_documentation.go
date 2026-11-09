package producer

import (
	"fmt"
	"strings"

	"github.com/JochiRaider/cartulary/referencepacks/format/canonicaljson"
)

func csvDocumentation(r windowsCSVRow) []WindowsDocumentation {
	out := []WindowsDocumentation{}
	add := func(kind, scope, column string, index int, content map[string]any) {
		loc := fmt.Sprintf("csv/record/%d/field/%s", r.Number, column)
		if index >= 0 {
			loc += fmt.Sprintf("/%d", index)
		}
		out = append(out, WindowsDocumentation{kind, scope, loc, content})
	}
	text := func(kind, scope, column string) {
		if v := r.Fields[column]; v != "" {
			add(kind, scope, column, -1, map[string]any{"text": v})
		}
	}
	for _, c := range []string{"xlsx_message_summaries", "docx_event_titles"} {
		text("title", "event", c)
	}
	text("message_template", "event_version", "xlsx_message_template")
	text("message_template", "event", "docx_message_template")
	text("explanation", "event", "docx_event_description")
	text("explanation", "event", "docx_other_event_context")
	text("field_reference", "event", "docx_event_field_reference")
	if r.Fields["docx_version_history"] != "" {
		versions := []any{}
		for _, a := range r.JSON["docx_section_contexts_json"] {
			if v, ok := a["documented_versions"].([]any); ok {
				versions = append(versions, v...)
			}
		}
		add("version_history", "event", "docx_version_history", -1, map[string]any{"text": r.Fields["docx_version_history"], "documented_versions": uniqueCSVValues(versions)})
	}
	if r.Fields["docx_minimum_os_as_documented"] != "" || r.Fields["docx_required_server_roles"] != "" {
		add("applicability", "event", "docx_minimum_os_as_documented", -1, map[string]any{"minimum_os": csvNullable(r.Fields["docx_minimum_os_as_documented"]), "required_roles": csvNullable(r.Fields["docx_required_server_roles"])})
	}
	if r.Fields["xlsx_minimum_os_for_message_version"] != "" {
		add("applicability", "event_version", "xlsx_minimum_os_for_message_version", -1, map[string]any{"minimum_os": r.Fields["xlsx_minimum_os_for_message_version"], "required_roles": nil})
	}
	for i, a := range r.JSON["xlsx_catalog_rows_json"] {
		add("classification", "event", "xlsx_catalog_rows_json", i, map[string]any{"category": a["category"], "subcategory": a["subcategory"], "outcomes": []any{}, "minimum_os": a["minimum_os_requirement"]})
	}
	for i, a := range r.JSON["docx_section_contexts_json"] {
		add("classification", "event", "docx_section_contexts_json", i, map[string]any{"category": a["category"], "subcategory": a["subcategory"], "outcomes": a["outcome_codes"], "minimum_os": nil})
	}
	for _, column := range []string{"docx_security_monitoring_recommendations", "docx_shared_monitoring_guidance"} {
		if v := r.Fields[column]; v != "" {
			scope := "event"
			if strings.Contains(column, "shared") {
				scope = "shared_guidance"
			}
			add("monitoring_guidance", scope, column, -1, map[string]any{"text": v, "category": nil, "subcategory": nil, "association_basis": nil})
		}
	}
	for i, a := range r.JSON["docx_subcategory_guidance_json"] {
		add("monitoring_guidance", "audit_subcategory", "docx_subcategory_guidance_json", i, selectCSV(a, "text", "category", "subcategory", "association_basis"))
	}
	for i, a := range r.JSON["docx_field_definitions_json"] {
		add("field_definition", "event", "docx_field_definitions_json", i, selectCSV(a, "label", "type_as_documented", "definition", "continuation_text", "group_heading", "versions_as_documented", "annotations", "table_indices", "section_id"))
	}
	for i, a := range r.JSON["docx_reference_tables_json"] {
		scope := "event"
		if a["role"] == "shared_monitoring_guidance" {
			scope = "shared_guidance"
		}
		if a["role"] == "audit_subcategory_context" {
			scope = "audit_subcategory"
		}
		add("reference_table", scope, "docx_reference_tables_json", i, selectCSV(a, "table_index", "owner", "role", "section", "rows"))
	}
	for _, column := range []string{"docx_sample_metadata_for_row_version_json", "docx_other_sample_metadata_json"} {
		for i, a := range r.JSON[column] {
			c := selectCSV(a, "event_id", "event_version", "provider_name", "provider_guid", "level", "task", "opcode", "channel", "keywords", "payload_container", "payload_fields", "parse_status", "parse_error", "matches_section_event_id")
			fields := []any{}
			if values, ok := a["payload_fields"].([]any); ok {
				for _, v := range values {
					m, _ := v.(map[string]any)
					fields = append(fields, selectCSV(m, "position", "element", "name", "path", "namespace", "extraction_method"))
				}
			}
			c["payload_fields"] = fields
			scope := "source_context"
			if column == "docx_sample_metadata_for_row_version_json" && a["matches_section_event_id"] == true {
				scope = "event_version"
			}
			add("example", scope, column, i, c)
		}
	}
	for i, a := range r.JSON["docx_links_json"] {
		add("source_link", "event", "docx_links_json", i, selectCSV(a, "text", "target", "internal_anchor"))
	}
	for i, a := range r.JSON["docx_image_references_json"] {
		add("untranscribed_image", "source_context", "docx_image_references_json", i, selectCSV(a, "name", "title", "description", "media_parts"))
	}
	for i, a := range r.JSON["quality_notes_json"] {
		b, _ := canonicaljson.Marshal(a)
		add("quality_issue", "source_context", "quality_notes_json", i, map[string]any{"code": a["code"], "text": string(b)})
	}
	for _, flag := range strings.Split(r.Fields["quality_flags"], "; ") {
		if flag != "" {
			add("quality_issue", "source_context", "quality_flags", -1, map[string]any{"code": flag, "text": "Source extraction quality flag; associated claims remain qualified."})
		}
	}
	for i, a := range r.JSON["source_artifact_metadata_json"] {
		add("source_metadata", "source_context", "source_artifact_metadata_json", i, selectCSV(a, "filename", "sha256", "publisher", "download_page", "download_listing_version", "download_listing_publication_date", "embedded_created_utc", "embedded_modified_utc", "source_os_build", "document_scope"))
	}
	add("extraction_context", "source_context", "version_evidence", -1, map[string]any{"version_evidence": r.Fields["version_evidence"], "source_coverage": r.Fields["source_coverage"], "context_scope": r.Fields["docx_context_scope"], "source_filenames": r.Fields["source_filenames"], "source_download_urls": r.Fields["source_download_urls"], "provider_names_as_extracted": r.Fields["provider_names_from_docx_examples"], "provider_guids_as_extracted": r.Fields["provider_guids_from_docx_examples"]})
	return out
}
func selectCSV(m map[string]any, keys ...string) map[string]any {
	out := map[string]any{}
	for _, k := range keys {
		out[k] = m[k]
	}
	return out
}
func csvNullable(s string) any {
	if s == "" {
		return nil
	}
	return s
}
func uniqueCSVValues(values []any) []any {
	out := []any{}
	seen := map[string]bool{}
	for _, v := range values {
		b, _ := canonicaljson.Marshal(v)
		if !seen[string(b)] {
			seen[string(b)] = true
			out = append(out, v)
		}
	}
	return out
}
