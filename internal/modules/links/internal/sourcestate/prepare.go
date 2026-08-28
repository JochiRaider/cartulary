package sourcestate

import (
	"bytes"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/JochiRaider/cartulary/internal/modules/incidentbundles/sourceport"
	"github.com/JochiRaider/cartulary/internal/modules/incidentportability"
	"github.com/JochiRaider/cartulary/internal/modules/links/internal/valuecodec"
	"github.com/JochiRaider/cartulary/internal/platform/fieldnorm"
)

type preparedLinkRow struct {
	raw       map[string]any
	incident  uuid.UUID
	source    uuid.UUID
	target    uuid.UUID
	linkType  string
	fieldKey  string
	deleted   bool
	targetKey string
}

type preparedTagCatalogRow struct {
	raw        map[string]any
	tagName    string
	normalized string
}

type preparedRecordTagRow struct {
	raw        map[string]any
	incident   uuid.UUID
	recordID   uuid.UUID
	tagName    string
	normalized string
	deleted    bool
	targetKey  string
}

type preparedImport struct {
	incidentID    uuid.UUID
	bundleVersion int
	operationID   string
	links         []preparedLinkRow
	tagCatalog    []preparedTagCatalogRow
	recordTags    []preparedRecordTagRow
}

type decodedFiles struct {
	links      []map[string]any
	tagCatalog []map[string]any
	recordTags []map[string]any
}

type sourceStateError struct {
	reason string
}

func (err *sourceStateError) Error() string {
	return "links: invalid prepared source state"
}

func prepareImport(descriptor sourceport.Descriptor, stateManifest manifest, bundle sourceport.Bundle, importContext sourceport.ImportContext) (preparedImport, error) {
	files, err := readStrictFiles(stateManifest, bundle, importContext.BundleVersion)
	if err != nil {
		return preparedImport{}, err
	}
	prepared, invariantID := validateAndPrepare(files, stateManifest, importContext.IncidentID)
	if invariantID != "" {
		return preparedImport{}, descriptor.DeclaredFailure(invariantID)
	}
	prepared.incidentID = importContext.IncidentID
	prepared.bundleVersion = importContext.BundleVersion
	prepared.operationID = importContext.OperationID
	return prepared, nil
}

func readStrictFiles(stateManifest manifest, bundle sourceport.Bundle, version int) (decodedFiles, error) {
	if bundle == nil {
		return decodedFiles{}, &incidentportability.VerificationFailure{ReasonCode: "missing_required_file"}
	}
	var result decodedFiles
	for _, path := range stateManifest.pathSpecs() {
		if !containsVersion(path.versions, version) {
			continue
		}
		payload, ok := bundle.File(path.logicalPath)
		if !ok {
			return decodedFiles{}, &incidentportability.VerificationFailure{ReasonCode: "missing_required_file"}
		}
		rows, err := decodeStrictNDJSON(payload)
		if err != nil {
			return decodedFiles{}, err
		}
		switch path.kind {
		case pathRecordLinks:
			result.links = rows
		case pathTagCatalog:
			result.tagCatalog = rows
		case pathRecordTags:
			result.recordTags = rows
		default:
			return decodedFiles{}, manifestError("unknown path kind")
		}
	}
	return result, nil
}

func decodeStrictNDJSON(payload []byte) ([]map[string]any, error) {
	if len(payload) == 0 {
		return []map[string]any{}, nil
	}
	lines := bytes.Split(payload, []byte{'\n'})
	if len(lines[len(lines)-1]) == 0 {
		lines = lines[:len(lines)-1]
	}
	rows := make([]map[string]any, 0, len(lines))
	for _, line := range lines {
		if len(bytes.TrimSpace(line)) == 0 {
			return nil, &incidentportability.MalformedPayloadError{Err: fmt.Errorf("blank NDJSON line")}
		}
		var row map[string]any
		decoder := json.NewDecoder(bytes.NewReader(line))
		decoder.UseNumber()
		if err := decoder.Decode(&row); err != nil || row == nil {
			if err == nil {
				err = fmt.Errorf("row is not an object")
			}
			return nil, &incidentportability.MalformedPayloadError{Err: err}
		}
		if decoder.More() {
			return nil, &incidentportability.MalformedPayloadError{Err: fmt.Errorf("multiple JSON values")}
		}
		var trailing any
		if err := decoder.Decode(&trailing); err == nil {
			return nil, &incidentportability.MalformedPayloadError{Err: fmt.Errorf("multiple JSON values")}
		}
		rows = append(rows, row)
	}
	return rows, nil
}

func validateAndPrepare(files decodedFiles, stateManifest manifest, incidentID uuid.UUID) (preparedImport, string) {
	if incidentID == uuid.Nil || !exactRows(files, stateManifest) || !sourceIdentitiesValid(files, stateManifest) {
		return preparedImport{}, "links_tags.source_identity_admitted"
	}
	prepared, ok := parseSourceRows(files)
	if !ok {
		return preparedImport{}, "links_tags.source_identity_admitted"
	}
	if !linkTuplesLegal(files.links) {
		return preparedImport{}, "links_tags.link_tuple_legal"
	}
	if !deletionTuplesLegal(files.links, files.recordTags) {
		return preparedImport{}, "links_tags.deletion_tuple_legal"
	}
	if !tagsCanonical(prepared) {
		return preparedImport{}, "links_tags.tag_normalized"
	}
	if !tagCatalogExact(prepared) {
		return preparedImport{}, "links_tags.tag_catalog_exact"
	}
	if !activeRowsUnique(prepared) {
		return preparedImport{}, "links_tags.link_unique"
	}
	if !sourceIncidentScopeValid(prepared, incidentID) {
		return preparedImport{}, "links_tags.endpoints_same_incident"
	}
	return prepared, ""
}

func exactRows(files decodedFiles, stateManifest manifest) bool {
	for _, path := range stateManifest.pathSpecs() {
		var rows []map[string]any
		switch path.kind {
		case pathRecordLinks:
			rows = files.links
		case pathTagCatalog:
			rows = files.tagCatalog
		case pathRecordTags:
			rows = files.recordTags
		default:
			return false
		}
		for _, row := range rows {
			if len(row) != len(path.allowedColumns) {
				return false
			}
			for _, column := range path.allowedColumns {
				if _, ok := row[column]; !ok {
					return false
				}
			}
		}
	}
	return true
}

func sourceIdentitiesValid(files decodedFiles, stateManifest manifest) bool {
	for _, path := range stateManifest.pathSpecs() {
		var rows []map[string]any
		switch path.kind {
		case pathRecordLinks:
			rows = files.links
		case pathTagCatalog:
			rows = files.tagCatalog
		case pathRecordTags:
			rows = files.recordTags
		}
		seen := map[string]struct{}{}
		for _, row := range rows {
			parts := make([]string, 0, len(path.stableIdentity))
			for _, column := range path.stableIdentity {
				value, ok := row[column].(string)
				if !ok || value == "" {
					return false
				}
				parts = append(parts, value)
			}
			key := strings.Join(parts, "\x00")
			if _, duplicate := seen[key]; duplicate {
				return false
			}
			seen[key] = struct{}{}
		}
	}
	return true
}

func parseSourceRows(files decodedFiles) (preparedImport, bool) {
	result := preparedImport{
		links: make([]preparedLinkRow, 0, len(files.links)), tagCatalog: make([]preparedTagCatalogRow, 0, len(files.tagCatalog)),
		recordTags: make([]preparedRecordTagRow, 0, len(files.recordTags)),
	}
	for _, raw := range files.links {
		linkID, ok := canonicalUUIDValue(raw["record_link_id"])
		incident, incidentOK := canonicalUUIDValue(raw["incident_id"])
		source, sourceOK := canonicalUUIDValue(raw["src_record_id"])
		target, targetOK := canonicalUUIDValue(raw["dst_record_id"])
		if !ok || !incidentOK || !sourceOK || !targetOK || !canonicalRequiredUUID(raw["owner_user_id"]) ||
			!canonicalRequiredUUID(raw["created_by_user_id"]) || !canonicalNullableUUID(raw["deleted_by_user_id"]) {
			return preparedImport{}, false
		}
		fieldKey := ""
		if raw["field_key"] != nil {
			var fieldOK bool
			fieldKey, fieldOK = raw["field_key"].(string)
			if !fieldOK || strings.TrimSpace(fieldKey) == "" || strings.TrimSpace(fieldKey) != fieldKey {
				return preparedImport{}, false
			}
		}
		linkType, typeOK := raw["link_type"].(string)
		result.links = append(result.links, preparedLinkRow{
			raw: cloneRow(raw), incident: incident, source: source, target: target, linkType: linkType,
			fieldKey: fieldKey, deleted: raw["deleted_at"] != nil, targetKey: linkID.String(),
		})
		if !typeOK {
			return preparedImport{}, false
		}
	}
	for _, raw := range files.tagCatalog {
		tagName, tagOK := raw["tag_name"].(string)
		normalized, normalizedOK := raw["normalized_tag_name"].(string)
		if !tagOK || !normalizedOK {
			return preparedImport{}, false
		}
		result.tagCatalog = append(result.tagCatalog, preparedTagCatalogRow{raw: cloneRow(raw), tagName: tagName, normalized: normalized})
	}
	for _, raw := range files.recordTags {
		tagID, ok := canonicalUUIDValue(raw["record_tag_id"])
		incident, incidentOK := canonicalUUIDValue(raw["incident_id"])
		recordID, recordOK := canonicalUUIDValue(raw["record_id"])
		tagName, tagOK := raw["tag_name"].(string)
		normalized, normalizedOK := raw["normalized_tag_name"].(string)
		if !ok || !incidentOK || !recordOK || !tagOK || !normalizedOK ||
			!canonicalRequiredUUID(raw["created_by_user_id"]) || !canonicalNullableUUID(raw["deleted_by_user_id"]) {
			return preparedImport{}, false
		}
		result.recordTags = append(result.recordTags, preparedRecordTagRow{
			raw: cloneRow(raw), incident: incident, recordID: recordID, tagName: tagName, normalized: normalized,
			deleted: raw["deleted_at"] != nil, targetKey: tagID.String(),
		})
	}
	return result, true
}

func linkTuplesLegal(rows []map[string]any) bool {
	for _, raw := range rows {
		withoutDeletion, ok := canonicalCodecRow(raw, "decided_at", "created_at", "deleted_at")
		if !ok {
			return false
		}
		withoutDeletion["deleted_at"] = nil
		withoutDeletion["deleted_by_user_id"] = nil
		if _, err := valuecodec.DecodeRecordLinkMutationValue(withoutDeletion); err != nil {
			return false
		}
	}
	return true
}

func deletionTuplesLegal(linkRows []map[string]any, tagRows []map[string]any) bool {
	for _, raw := range linkRows {
		canonical, ok := canonicalCodecRow(raw, "decided_at", "created_at", "deleted_at")
		if !ok {
			return false
		}
		if _, err := valuecodec.DecodeRecordLinkMutationValue(canonical); err != nil {
			return false
		}
	}
	for _, raw := range tagRows {
		canonical, ok := canonicalCodecRow(raw, "created_at", "updated_at", "deleted_at")
		if !ok {
			return false
		}
		if _, err := valuecodec.DecodeRecordTagMutationValue(canonical); err != nil {
			return false
		}
	}
	return true
}

var portableTimestampPattern = regexp.MustCompile(`^[0-9]{4}-[0-9]{2}-[0-9]{2}T[0-9]{2}:[0-9]{2}:[0-9]{2}(?:\.[0-9]{1,6})?\+00:00$`)

func canonicalCodecRow(raw map[string]any, timestampFields ...string) (map[string]any, bool) {
	result := cloneRow(raw)
	for _, field := range timestampFields {
		if result[field] == nil {
			continue
		}
		text, ok := result[field].(string)
		if !ok || !portableTimestampPattern.MatchString(text) {
			return nil, false
		}
		parsed, err := time.Parse(time.RFC3339Nano, text)
		if err != nil || parsed.IsZero() {
			return nil, false
		}
		result[field] = parsed.UTC().Format(time.RFC3339Nano)
	}
	return result, true
}

func tagsCanonical(value preparedImport) bool {
	for _, row := range value.tagCatalog {
		if !canonicalTag(row.tagName, row.normalized) {
			return false
		}
	}
	for _, row := range value.recordTags {
		if !canonicalTag(row.tagName, row.normalized) {
			return false
		}
	}
	return true
}

func canonicalTag(tagName string, normalized string) bool {
	label, expectedNormalized, ok := fieldnorm.NormalizeTagLabel(tagName)
	return ok && label == tagName && expectedNormalized == normalized
}

func tagCatalogExact(value preparedImport) bool {
	catalog := make(map[string]struct{}, len(value.tagCatalog))
	for _, row := range value.tagCatalog {
		catalog[row.normalized+"\x00"+row.tagName] = struct{}{}
	}
	derived := make(map[string]struct{}, len(value.recordTags))
	for _, row := range value.recordTags {
		derived[row.normalized+"\x00"+row.tagName] = struct{}{}
	}
	if len(catalog) != len(derived) {
		return false
	}
	for key := range catalog {
		if _, ok := derived[key]; !ok {
			return false
		}
	}
	return true
}

func activeRowsUnique(value preparedImport) bool {
	links := map[string]struct{}{}
	supersedes := map[string]struct{}{}
	for _, row := range value.links {
		if row.deleted {
			continue
		}
		key := row.incident.String() + "\x00" + row.source.String() + "\x00" + row.target.String() + "\x00" + row.linkType + "\x00" + row.fieldKey
		if _, duplicate := links[key]; duplicate {
			return false
		}
		links[key] = struct{}{}
		if row.linkType == "supersedes" {
			key = row.incident.String() + "\x00" + row.target.String()
			if _, duplicate := supersedes[key]; duplicate {
				return false
			}
			supersedes[key] = struct{}{}
		}
	}
	tags := map[string]struct{}{}
	for _, row := range value.recordTags {
		if row.deleted {
			continue
		}
		key := row.incident.String() + "\x00" + row.recordID.String() + "\x00" + row.normalized
		if _, duplicate := tags[key]; duplicate {
			return false
		}
		tags[key] = struct{}{}
	}
	return true
}

func sourceIncidentScopeValid(value preparedImport, incidentID uuid.UUID) bool {
	for _, row := range value.links {
		if row.incident != incidentID {
			return false
		}
	}
	for _, row := range value.recordTags {
		if row.incident != incidentID {
			return false
		}
	}
	return true
}

func canonicalUUIDValue(raw any) (uuid.UUID, bool) {
	text, ok := raw.(string)
	if !ok {
		return uuid.Nil, false
	}
	value, err := uuid.Parse(text)
	return value, err == nil && value != uuid.Nil && value.String() == text
}

func canonicalRequiredUUID(raw any) bool {
	_, ok := canonicalUUIDValue(raw)
	return ok
}

func canonicalNullableUUID(raw any) bool {
	if raw == nil {
		return true
	}
	return canonicalRequiredUUID(raw)
}

func containsVersion(versions []int, version int) bool {
	for _, candidate := range versions {
		if candidate == version {
			return true
		}
	}
	return false
}

func cloneRow(source map[string]any) map[string]any {
	result := make(map[string]any, len(source))
	for key, value := range source {
		result[key] = value
	}
	return result
}
