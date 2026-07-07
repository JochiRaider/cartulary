package reportcomposition

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"unicode/utf8"
)

var (
	compositionVersionPattern = regexp.MustCompile(`^v([1-9][0-9]*)$`)
	generatedTargetPattern    = regexp.MustCompile(`^(vx|ed|slide|chunk|split|svg|dom|rf|mermaid)[_-]`)
)

func canonicalJSON(value any) ([]byte, error) {
	return json.Marshal(value)
}

func hashHex(data []byte) string {
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:])
}

func formatCompositionVersion(version int64) string {
	return "v" + strconv.FormatInt(version, 10)
}

func parseCompositionVersion(value string) (int64, bool) {
	matches := compositionVersionPattern.FindStringSubmatch(value)
	if len(matches) != 2 {
		return 0, false
	}
	parsed, err := strconv.ParseInt(matches[1], 10, 64)
	if err != nil || parsed <= 0 {
		return 0, false
	}
	return parsed, true
}

func canonicalComposition(record ResourceRecord, compositionVersion int64) (json.RawMessage, string, error) {
	base := map[string]any{
		"schema_id":                    CompositionSchemaID,
		"composition_id":               record.CompositionID.String(),
		"composition_version":          formatCompositionVersion(compositionVersion),
		"incident_id":                  record.IncidentID.String(),
		"template_id":                  record.TemplateID,
		"template_version":             record.TemplateVersion,
		"authored_against_snapshot_id": optionalStringForJSON(record.AuthoredAgainstSnapshotID),
		"deck_ops":                     rawJSONValue(record.DeckOps),
		"diagram_decls":                rawJSONValue(record.DiagramDecls),
		"authored_texts":               rawJSONValue(record.AuthoredTexts),
	}
	digestBytes, err := canonicalJSON(base)
	if err != nil {
		return nil, "", err
	}
	digest := hashHex(digestBytes)
	base["composition_sha256"] = digest
	canonicalBytes, err := canonicalJSON(base)
	if err != nil {
		return nil, "", err
	}
	if recomputed, err := digestFromCompositionBytes(canonicalBytes); err != nil || recomputed != digest {
		if err != nil {
			return nil, "", err
		}
		return nil, "", fmt.Errorf("composition digest mismatch")
	}
	return json.RawMessage(canonicalBytes), digest, nil
}

func digestFromCompositionBytes(data []byte) (string, error) {
	var decoded map[string]any
	if err := json.Unmarshal(data, &decoded); err != nil {
		return "", err
	}
	delete(decoded, "composition_sha256")
	digestBytes, err := canonicalJSON(decoded)
	if err != nil {
		return "", err
	}
	return hashHex(digestBytes), nil
}

func previewSource(resource ResourceRecord, version *VersionRecord, request PreviewRequest) (json.RawMessage, string, *int64, *string, error) {
	var draftVersion *int64
	var compositionVersion any
	var compositionSHA any
	authoredAgainst := resource.AuthoredAgainstSnapshotID
	deckOps := resource.DeckOps
	diagramDecls := resource.DiagramDecls
	authoredTexts := resource.AuthoredTexts
	var compositionSHAPtr *string
	if version == nil {
		draft := resource.DraftVersion
		draftVersion = &draft
		compositionVersion = nil
		compositionSHA = nil
	} else {
		var decoded map[string]any
		if err := json.Unmarshal(version.CanonicalComposition, &decoded); err != nil {
			return nil, "", nil, nil, err
		}
		authoredAgainst = stringPtrFromAny(decoded["authored_against_snapshot_id"])
		deckOps = mustRawFromAny(decoded["deck_ops"])
		diagramDecls = mustRawFromAny(decoded["diagram_decls"])
		authoredTexts = mustRawFromAny(decoded["authored_texts"])
		compositionVersion = formatCompositionVersion(version.CompositionVersion)
		compositionSHA = version.CompositionSHA256
		sha := version.CompositionSHA256
		compositionSHAPtr = &sha
	}
	base := map[string]any{
		"schema_id":                    PreviewSourceSchemaID,
		"source_kind":                  request.SourceKind,
		"incident_id":                  resource.IncidentID.String(),
		"composition_id":               resource.CompositionID.String(),
		"draft_version":                draftVersion,
		"composition_version":          compositionVersion,
		"composition_sha256":           compositionSHA,
		"template_id":                  resource.TemplateID,
		"template_version":             resource.TemplateVersion,
		"authored_against_snapshot_id": optionalStringForJSON(authoredAgainst),
		"deck_ops":                     rawJSONValue(deckOps),
		"diagram_decls":                rawJSONValue(diagramDecls),
		"authored_texts":               rawJSONValue(authoredTexts),
	}
	digestBytes, err := canonicalJSON(base)
	if err != nil {
		return nil, "", nil, nil, err
	}
	digest := hashHex(digestBytes)
	base["preview_source_sha256"] = digest
	canonicalBytes, err := canonicalJSON(base)
	if err != nil {
		return nil, "", nil, nil, err
	}
	return json.RawMessage(canonicalBytes), digest, draftVersion, compositionSHAPtr, nil
}

func validateDraft(deckOps json.RawMessage, diagramDecls json.RawMessage, authoredTexts json.RawMessage, version *VersionRecord) map[string]any {
	issues := []map[string]any{}
	authoredRoles := map[string]string{}
	diagramIDs := map[string]struct{}{}
	issues = append(issues, validateAuthoredTexts(authoredTexts, authoredRoles)...)
	issues = append(issues, validateDiagramDecls(diagramDecls, diagramIDs)...)
	issues = append(issues, validateCompositionOps(deckOps, authoredRoles, diagramIDs)...)
	if version != nil {
		if digest, err := digestFromCompositionBytes(version.CanonicalComposition); err != nil || digest != version.CompositionSHA256 {
			issues = append(issues, issue("composition_digest_mismatch", map[string]any{
				"composition_id":      version.CompositionID.String(),
				"composition_version": formatCompositionVersion(version.CompositionVersion),
			}))
		}
	}
	stage := any(nil)
	if len(issues) > 0 {
		stage = "schema_validation"
	}
	if version != nil && len(issues) > 0 && issues[0]["code"] == "composition_digest_mismatch" {
		stage = "canonical_digest_validation"
	}
	return map[string]any{
		"valid":               len(issues) == 0,
		"stage":               stage,
		"issues":              issues,
		"composition_id":      nil,
		"composition_version": nil,
		"composition_sha256":  nil,
	}
}

func validateCompositionOps(raw json.RawMessage, authoredRoles map[string]string, diagramIDs map[string]struct{}) []map[string]any {
	items, issues := decodeObjectArray(raw, "deck_ops")
	if len(issues) > 0 {
		return issues
	}
	issues = append(issues, duplicateIDIssuesFromItems(items, "op_id", "composition_op_id")...)
	payloadRules := map[string][]string{
		"exclude_section":         {"section_anchor"},
		"reorder_sections":        {"section_anchors"},
		"override_slide_layout":   {"section_anchor", "layout_id"},
		"override_title":          {"section_anchor", "authored_text_ref"},
		"set_speaker_notes":       {"section_anchor", "authored_text_ref"},
		"insert_authored_block":   {"block_anchor", "position", "authored_text_ref"},
		"exclude_block":           {"block_anchor"},
		"override_click_profile":  {"section_anchor", "click_profile"},
		"insert_diagram_slide":    {"diagram_anchor", "section_anchor", "position"},
		"exclude_diagram":         {"diagram_anchor"},
		"override_diagram_labels": {"diagram_anchor", "label_overrides"},
	}
	textRoles := map[string]string{
		"override_title":        "title_override",
		"set_speaker_notes":     "speaker_notes",
		"insert_authored_block": "authored_text",
	}
	for _, item := range items {
		if !objectHasOnly(item, "op_id", "op_kind", "on_unresolved", "payload") {
			issues = append(issues, issue("composition_schema_invalid", map[string]any{"field": "deck_ops"}))
			continue
		}
		opID, _ := item["op_id"].(string)
		opKind, ok := item["op_kind"].(string)
		if opID == "" || !ok || opKind == "" {
			issues = append(issues, issue("composition_schema_invalid", map[string]any{"composition_op_id": opID}))
			continue
		}
		if onUnresolved, ok := item["on_unresolved"]; ok && onUnresolved != "fail" && onUnresolved != "drop" {
			issues = append(issues, issue("composition_schema_invalid", map[string]any{"composition_op_id": opID, "field": "on_unresolved"}))
		}
		required, ok := payloadRules[opKind]
		if !ok {
			issues = append(issues, issue("composition_schema_invalid", map[string]any{"composition_op_id": opID, "field": "op_kind"}))
			continue
		}
		payload, ok := item["payload"].(map[string]any)
		if !ok || !objectHasOnly(payload, required...) || !objectHasRequired(payload, required...) {
			issues = append(issues, issue("composition_schema_invalid", map[string]any{"composition_op_id": opID, "field": "payload"}))
			continue
		}
		for key, value := range payload {
			switch key {
			case "section_anchor", "block_anchor", "diagram_anchor":
				if code := validateAnchorObject(value); code != "" {
					issues = append(issues, issue(code, map[string]any{"composition_op_id": opID, "field": key}))
				}
			case "section_anchors":
				anchors, ok := value.([]any)
				if !ok || len(anchors) == 0 {
					issues = append(issues, issue("composition_schema_invalid", map[string]any{"composition_op_id": opID, "field": key}))
					continue
				}
				seen := map[string]struct{}{}
				for _, anchor := range anchors {
					if code := validateAnchorObject(anchor); code != "" {
						issues = append(issues, issue(code, map[string]any{"composition_op_id": opID, "field": key}))
					}
					canonical, _ := canonicalJSON(anchor)
					if _, ok := seen[string(canonical)]; ok {
						issues = append(issues, issue("composition_schema_invalid", map[string]any{"composition_op_id": opID, "field": key}))
					}
					seen[string(canonical)] = struct{}{}
				}
			case "authored_text_ref":
				ref, ok := value.(string)
				wantRole := textRoles[opKind]
				if !ok || ref == "" || authoredRoles[ref] != wantRole {
					issues = append(issues, issue("composition_schema_invalid", map[string]any{"composition_op_id": opID, "authored_text_id": ref}))
				}
			case "position":
				if value != "before" && value != "after" {
					issues = append(issues, issue("composition_schema_invalid", map[string]any{"composition_op_id": opID, "field": key}))
				}
			case "click_profile":
				if value != "none" && value != "reveal_blocks" && value != "reveal_list_items" {
					issues = append(issues, issue("composition_schema_invalid", map[string]any{"composition_op_id": opID, "field": key}))
				}
			case "label_overrides":
				if labelIssues := validateLabelOverrides(value, opID); len(labelIssues) > 0 {
					issues = append(issues, labelIssues...)
				}
			}
		}
		if opKind == "insert_diagram_slide" {
			if anchor, ok := payload["diagram_anchor"].(map[string]any); ok {
				if declID, _ := anchor["decl_id"].(string); declID != "" {
					if _, ok := diagramIDs[declID]; !ok {
						issues = append(issues, issue("composition_schema_invalid", map[string]any{"composition_op_id": opID, "diagram_id": declID}))
					}
				}
			}
		}
	}
	return issues
}

func validateAuthoredTexts(raw json.RawMessage, roles map[string]string) []map[string]any {
	items, issues := decodeObjectArray(raw, "authored_texts")
	if len(issues) > 0 {
		return issues
	}
	issues = append(issues, duplicateIDIssuesFromItems(items, "authored_text_id", "authored_text_id")...)
	for _, item := range items {
		if !objectHasOnly(item, "authored_text_id", "text_role", "body", "disclosure_partition_ref") {
			issues = append(issues, issue("composition_schema_invalid", map[string]any{"field": "authored_texts"}))
			continue
		}
		id, _ := item["authored_text_id"].(string)
		role, _ := item["text_role"].(string)
		body, bodyOK := item["body"].(string)
		partition, _ := item["disclosure_partition_ref"].(string)
		if id == "" || !bodyOK || partition == "" || partition == "blocked" || !validTextRole(role) {
			issues = append(issues, issue("composition_schema_invalid", map[string]any{"authored_text_id": id}))
			continue
		}
		roles[id] = role
		if code := validateAuthoredBody(role, body); code != "" {
			issues = append(issues, issue(code, map[string]any{"authored_text_id": id}))
		}
	}
	return issues
}

func validateDiagramDecls(raw json.RawMessage, ids map[string]struct{}) []map[string]any {
	items, issues := decodeObjectArray(raw, "diagram_decls")
	if len(issues) > 0 {
		return issues
	}
	issues = append(issues, duplicateIDIssuesFromItems(items, "decl_id", "diagram_id")...)
	for _, item := range items {
		if !objectHasOnly(item, "decl_id", "replaces_decl_id", "diagram_kind", "diagram_source_kind", "source_graph_view_id", "selection_rule", "label_overrides", "layout_mode", "layout") {
			issues = append(issues, issue("composition_schema_invalid", map[string]any{"field": "diagram_decls"}))
			continue
		}
		id, _ := item["decl_id"].(string)
		if id == "" {
			issues = append(issues, issue("composition_schema_invalid", map[string]any{"diagram_id": id}))
			continue
		}
		ids[id] = struct{}{}
		sourceKind, _ := item["diagram_source_kind"].(string)
		sourceGraph, graphIsString := item["source_graph_view_id"].(string)
		if sourceKind == "graph" {
			if !graphIsString || sourceGraph == "" {
				issues = append(issues, issue("composition_schema_invalid", map[string]any{"diagram_id": id, "field": "source_graph_view_id"}))
			}
		} else if sourceKind == "timeline" {
			if value, ok := item["source_graph_view_id"]; ok && value != nil {
				issues = append(issues, issue("composition_schema_invalid", map[string]any{"diagram_id": id, "field": "source_graph_view_id"}))
			}
		} else {
			issues = append(issues, issue("composition_schema_invalid", map[string]any{"diagram_id": id, "field": "diagram_source_kind"}))
		}
		if hasRawGeneratedSource(item) {
			issues = append(issues, issue("raw_generated_source_invalid", map[string]any{"diagram_id": id}))
		}
		if selectionRule, ok := item["selection_rule"].(map[string]any); !ok || hasRawGeneratedSource(selectionRule) {
			issues = append(issues, issue("raw_generated_source_invalid", map[string]any{"diagram_id": id, "field": "selection_rule"}))
		}
		if labelOverrides, ok := item["label_overrides"]; ok {
			issues = append(issues, validateLabelOverrides(labelOverrides, id)...)
		}
		layoutMode := "auto"
		if value, ok := item["layout_mode"].(string); ok && value != "" {
			layoutMode = value
		}
		switch layoutMode {
		case "auto":
			if layout, ok := item["layout"]; ok && layout != nil {
				issues = append(issues, issue("diagram_layout_invalid", map[string]any{"diagram_id": id, "layout_mode": layoutMode}))
			}
		case "manual":
			if item["diagram_kind"] != "flowchart" {
				issues = append(issues, issue("diagram_layout_invalid", map[string]any{"diagram_id": id, "layout_mode": layoutMode}))
			}
			layout, ok := item["layout"].(map[string]any)
			if !ok {
				issues = append(issues, issue("diagram_layout_invalid", map[string]any{"diagram_id": id, "layout_mode": layoutMode}))
				continue
			}
			issues = append(issues, validateDiagramLayout(id, layout)...)
		default:
			issues = append(issues, issue("composition_schema_invalid", map[string]any{"diagram_id": id, "field": "layout_mode"}))
		}
	}
	return issues
}

func validateDiagramLayout(diagramID string, layout map[string]any) []map[string]any {
	issues := []map[string]any{}
	if !objectHasOnly(layout, "schema_id", "coordinate_space", "node_positions", "edge_routes") || layout["schema_id"] != "composition_diagram_layout.v1" {
		return []map[string]any{issue("diagram_layout_invalid", map[string]any{"diagram_id": diagramID})}
	}
	space, ok := layout["coordinate_space"].(map[string]any)
	if !ok || !objectHasOnly(space, "unit", "origin", "width", "height") || space["unit"] != "css_px" || space["origin"] != "top_left" {
		issues = append(issues, issue("diagram_layout_invalid", map[string]any{"diagram_id": diagramID, "field": "coordinate_space"}))
	}
	width, widthOK := int64Value(space["width"])
	height, heightOK := int64Value(space["height"])
	if !widthOK || !heightOK || width < 1 || height < 1 || width > 10000 || height > 10000 {
		issues = append(issues, issue("diagram_layout_invalid", map[string]any{"diagram_id": diagramID, "field": "coordinate_space"}))
	}
	issues = append(issues, validateNodePositions(diagramID, layout["node_positions"], width, height)...)
	issues = append(issues, validateEdgeRoutes(diagramID, layout["edge_routes"], width, height)...)
	return issues
}

func validateNodePositions(diagramID string, raw any, maxWidth int64, maxHeight int64) []map[string]any {
	items, ok := raw.([]any)
	if !ok {
		return []map[string]any{issue("diagram_layout_invalid", map[string]any{"diagram_id": diagramID, "field": "node_positions"})}
	}
	issues := []map[string]any{}
	seen := map[string]struct{}{}
	var previous string
	for _, rawItem := range items {
		item, ok := rawItem.(map[string]any)
		if !ok || !objectHasOnly(item, "target_ref", "x", "y", "width", "height") {
			issues = append(issues, issue("diagram_layout_invalid", map[string]any{"diagram_id": diagramID, "field": "node_positions"}))
			continue
		}
		target, _ := item["target_ref"].(string)
		if target == "" || isGeneratedTarget(target) {
			issues = append(issues, issue("diagram_layout_unknown_target", map[string]any{"diagram_id": diagramID, "target_ref": target}))
		}
		if previous != "" && target < previous {
			issues = append(issues, issue("diagram_layout_invalid", map[string]any{"diagram_id": diagramID, "target_ref": target}))
		}
		previous = target
		if _, ok := seen[target]; ok {
			issues = append(issues, issue("diagram_layout_duplicate_target", map[string]any{"diagram_id": diagramID, "target_ref": target}))
		}
		seen[target] = struct{}{}
		x, xOK := int64Value(item["x"])
		y, yOK := int64Value(item["y"])
		w, wOK := int64Value(item["width"])
		h, hOK := int64Value(item["height"])
		if !xOK || !yOK || !wOK || !hOK || x < 0 || y < 0 || w < 1 || h < 1 || x+w > maxWidth || y+h > maxHeight {
			issues = append(issues, issue("diagram_layout_invalid", map[string]any{"diagram_id": diagramID, "target_ref": target}))
		}
	}
	return issues
}

func validateEdgeRoutes(diagramID string, raw any, maxWidth int64, maxHeight int64) []map[string]any {
	items, ok := raw.([]any)
	if !ok {
		return []map[string]any{issue("diagram_layout_invalid", map[string]any{"diagram_id": diagramID, "field": "edge_routes"})}
	}
	issues := []map[string]any{}
	seen := map[string]struct{}{}
	var previous string
	for _, rawItem := range items {
		item, ok := rawItem.(map[string]any)
		if !ok || !objectHasOnly(item, "target_ref", "route_kind", "waypoints") || item["route_kind"] != "polyline" {
			issues = append(issues, issue("diagram_layout_invalid", map[string]any{"diagram_id": diagramID, "field": "edge_routes"}))
			continue
		}
		target, _ := item["target_ref"].(string)
		if target == "" || isGeneratedTarget(target) {
			issues = append(issues, issue("diagram_layout_unknown_target", map[string]any{"diagram_id": diagramID, "target_ref": target}))
		}
		if previous != "" && target < previous {
			issues = append(issues, issue("diagram_layout_invalid", map[string]any{"diagram_id": diagramID, "target_ref": target}))
		}
		previous = target
		if _, ok := seen[target]; ok {
			issues = append(issues, issue("diagram_layout_duplicate_target", map[string]any{"diagram_id": diagramID, "target_ref": target}))
		}
		seen[target] = struct{}{}
		waypoints, ok := item["waypoints"].([]any)
		if !ok || len(waypoints) > 32 {
			issues = append(issues, issue("diagram_layout_invalid", map[string]any{"diagram_id": diagramID, "target_ref": target}))
			continue
		}
		for _, waypoint := range waypoints {
			point, ok := waypoint.(map[string]any)
			x, xOK := int64Value(point["x"])
			y, yOK := int64Value(point["y"])
			if !ok || !objectHasOnly(point, "x", "y") || !xOK || !yOK || x < 0 || y < 0 || x > maxWidth || y > maxHeight {
				issues = append(issues, issue("diagram_layout_invalid", map[string]any{"diagram_id": diagramID, "target_ref": target}))
			}
		}
	}
	return issues
}

func validationSummaryForResource(record ResourceRecord) map[string]any {
	summary := validateDraft(record.DeckOps, record.DiagramDecls, record.AuthoredTexts, nil)
	summary["composition_id"] = record.CompositionID.String()
	return summary
}

func validationSummaryForVersion(version VersionRecord) map[string]any {
	summary := validateDraft(nil, nil, nil, &version)
	summary["composition_id"] = version.CompositionID.String()
	summary["composition_version"] = formatCompositionVersion(version.CompositionVersion)
	summary["composition_sha256"] = version.CompositionSHA256
	return summary
}

func summaryValid(summary map[string]any) bool {
	valid, _ := summary["valid"].(bool)
	return valid
}

func duplicateIDIssuesFromItems(items []map[string]any, idField string, detailKey string) []map[string]any {
	seen := map[string]struct{}{}
	for _, item := range items {
		id, ok := item[idField].(string)
		if !ok || id == "" {
			return []map[string]any{issue("composition_schema_invalid", map[string]any{"field": idField})}
		}
		if _, ok := seen[id]; ok {
			return []map[string]any{issue("composition_duplicate_id", map[string]any{detailKey: id})}
		}
		seen[id] = struct{}{}
	}
	return nil
}

func decodeObjectArray(raw json.RawMessage, field string) ([]map[string]any, []map[string]any) {
	if len(raw) == 0 {
		return nil, nil
	}
	var items []map[string]any
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.UseNumber()
	if err := decoder.Decode(&items); err != nil {
		return nil, []map[string]any{issue("composition_schema_invalid", map[string]any{"field": field})}
	}
	return items, nil
}

func objectHasOnly(value map[string]any, keys ...string) bool {
	allowed := map[string]struct{}{}
	for _, key := range keys {
		allowed[key] = struct{}{}
	}
	for key := range value {
		if _, ok := allowed[key]; !ok {
			return false
		}
	}
	return true
}

func objectHasRequired(value map[string]any, keys ...string) bool {
	for _, key := range keys {
		if _, ok := value[key]; !ok || value[key] == nil {
			return false
		}
	}
	return true
}

func validTextRole(role string) bool {
	switch role {
	case "title_override", "speaker_notes", "authored_text":
		return true
	default:
		return false
	}
}

func validateAuthoredBody(role string, body string) string {
	if !utf8.ValidString(body) || strings.TrimSpace(body) == "" {
		if role == "title_override" {
			return "authored_title_limit_exceeded"
		}
		return "authored_text_limit_exceeded"
	}
	for _, r := range body {
		if r == '\r' || r == '\t' || r == 0 || (r < 0x20 && r != '\n') || (r >= 0x7f && r <= 0x9f) {
			if role == "title_override" {
				return "authored_title_limit_exceeded"
			}
			return "authored_text_limit_exceeded"
		}
		if role == "title_override" && r == '\n' {
			return "authored_title_limit_exceeded"
		}
	}
	if role == "title_override" && len([]rune(body)) > 120 {
		return "authored_title_limit_exceeded"
	}
	if role == "authored_text" && len([]rune(body)) > 5000 {
		return "authored_text_limit_exceeded"
	}
	if role == "speaker_notes" && len([]rune(body)) > 5000 {
		return "authored_text_limit_exceeded"
	}
	if malformedSubjectPlaceholder(body) {
		return "authored_subject_ref_unresolved"
	}
	return ""
}

func malformedSubjectPlaceholder(body string) bool {
	offset := 0
	for {
		relativeStart := strings.Index(body[offset:], "{{subject:")
		if relativeStart < 0 {
			return false
		}
		start := offset + relativeStart
		end := strings.Index(body[start:], "}}")
		if end < 0 {
			return true
		}
		inner := body[start+len("{{subject:") : start+end]
		if strings.TrimSpace(inner) == "" {
			return true
		}
		offset = start + end + len("}}")
	}
}

func validateAnchorObject(value any) string {
	anchor, ok := value.(map[string]any)
	if !ok {
		return "composition_schema_invalid"
	}
	kind, _ := anchor["anchor_kind"].(string)
	switch kind {
	case "section_anchor":
		if !objectHasOnly(anchor, "anchor_kind", "template_section_decl_id", "expansion_key") || !objectHasRequired(anchor, "anchor_kind", "template_section_decl_id") {
			return "composition_schema_invalid"
		}
	case "record_anchor":
		if !objectHasOnly(anchor, "anchor_kind", "record_id") || !objectHasRequired(anchor, "anchor_kind", "record_id") {
			return "composition_schema_invalid"
		}
	case "block_anchor":
		if !objectHasOnly(anchor, "anchor_kind", "section_anchor", "block_kind", "record_anchor") || !objectHasRequired(anchor, "anchor_kind", "section_anchor", "block_kind") {
			return "composition_schema_invalid"
		}
		if code := validateAnchorObject(anchor["section_anchor"]); code != "" {
			return code
		}
		if recordAnchor, ok := anchor["record_anchor"]; ok && recordAnchor != nil {
			if code := validateAnchorObject(recordAnchor); code != "" {
				return code
			}
		}
	case "diagram_anchor":
		if !objectHasOnly(anchor, "anchor_kind", "decl_id") || !objectHasRequired(anchor, "anchor_kind", "decl_id") {
			return "composition_schema_invalid"
		}
	default:
		return "composition_schema_invalid"
	}
	if containsGeneratedTarget(anchor) {
		return "composition_anchor_invalid"
	}
	return ""
}

func validateLabelOverrides(raw any, ownerID string) []map[string]any {
	items, ok := raw.([]any)
	if !ok {
		return []map[string]any{issue("composition_schema_invalid", map[string]any{"diagram_id": ownerID, "field": "label_overrides"})}
	}
	issues := []map[string]any{}
	seen := map[string]struct{}{}
	for _, rawItem := range items {
		item, ok := rawItem.(map[string]any)
		if !ok || !objectHasOnly(item, "target", "label") || !objectHasRequired(item, "target", "label") {
			issues = append(issues, issue("composition_schema_invalid", map[string]any{"diagram_id": ownerID, "field": "label_overrides"}))
			continue
		}
		target, ok := item["target"].(map[string]any)
		label, labelOK := item["label"].(string)
		if !ok || !objectHasOnly(target, "target_kind", "ref") || !objectHasRequired(target, "target_kind", "ref") || !labelOK || strings.Contains(label, "\n") {
			issues = append(issues, issue("diagram_label_override_invalid", map[string]any{"diagram_id": ownerID}))
			continue
		}
		if target["target_kind"] != "vertex" && target["target_kind"] != "edge" {
			issues = append(issues, issue("diagram_label_override_invalid", map[string]any{"diagram_id": ownerID}))
		}
		ref, _ := target["ref"].(string)
		if ref == "" || isGeneratedTarget(ref) {
			issues = append(issues, issue("diagram_selection_missing_ref", map[string]any{"diagram_id": ownerID, "target_ref": ref}))
		}
		canonical, _ := canonicalJSON(target)
		if _, ok := seen[string(canonical)]; ok {
			issues = append(issues, issue("diagram_label_override_invalid", map[string]any{"diagram_id": ownerID, "target_ref": ref}))
		}
		seen[string(canonical)] = struct{}{}
	}
	return issues
}

func hasRawGeneratedSource(value any) bool {
	switch typed := value.(type) {
	case map[string]any:
		for key, child := range typed {
			switch key {
			case "raw_mermaid", "mermaid", "mermaid_source", "generated_source", "raw_generated_source", "raw_graph_query", "renderer_syntax", "mmd", "raw_markdown", "raw_html", "nodes", "edges", "vertices":
				return true
			}
			if hasRawGeneratedSource(child) {
				return true
			}
		}
	case []any:
		for _, child := range typed {
			if hasRawGeneratedSource(child) {
				return true
			}
		}
	case string:
		lower := strings.ToLower(typed)
		return strings.Contains(lower, "mermaid") || strings.Contains(lower, "<script") || strings.Contains(lower, "graph td") || strings.Contains(lower, "flowchart ")
	}
	return false
}

func containsGeneratedTarget(value any) bool {
	switch typed := value.(type) {
	case map[string]any:
		for _, child := range typed {
			if containsGeneratedTarget(child) {
				return true
			}
		}
	case []any:
		for _, child := range typed {
			if containsGeneratedTarget(child) {
				return true
			}
		}
	case string:
		return isGeneratedTarget(typed)
	}
	return false
}

func isGeneratedTarget(value string) bool {
	return generatedTargetPattern.MatchString(value) || strings.Contains(value, "reactflow") || strings.Contains(value, "mermaid")
}

func int64Value(value any) (int64, bool) {
	switch typed := value.(type) {
	case float64:
		if typed != float64(int64(typed)) {
			return 0, false
		}
		return int64(typed), true
	case int64:
		return typed, true
	case int:
		return int64(typed), true
	case json.Number:
		parsed, err := strconv.ParseInt(typed.String(), 10, 64)
		return parsed, err == nil
	default:
		return 0, false
	}
}

func issue(code string, details map[string]any) map[string]any {
	return map[string]any{
		"severity":     "error",
		"code":         code,
		"message_key":  "report_composition." + code,
		"safe_details": details,
	}
}

func mustRawFromAny(value any) json.RawMessage {
	encoded, _ := canonicalJSON(value)
	return json.RawMessage(encoded)
}

func stringPtrFromAny(value any) *string {
	parsed, ok := value.(string)
	if !ok {
		return nil
	}
	return &parsed
}
