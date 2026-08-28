package artifacts

import (
	"encoding/json"
	"strings"

	"github.com/google/uuid"
	"golang.org/x/text/cases"

	"github.com/JochiRaider/cartulary/internal/modules/links"
	"github.com/JochiRaider/cartulary/internal/platform/fieldnorm"
)

var riskReferenceCaseFolder = cases.Fold()

func decodeArtifactCollectionActionPayload(fieldKey string, raw json.RawMessage) (collectionActionPayload, *AdmissionError) {
	var object map[string]json.RawMessage
	if json.Unmarshal(raw, &object) != nil || !artifactObjectHasOnlyFields(object, "kind", "actions") {
		return collectionActionPayload{}, newAdmissionError(fieldKey, admissionInvalidValue)
	}
	var kind string
	if json.Unmarshal(object["kind"], &kind) != nil || kind != "collection_actions_v1" {
		return collectionActionPayload{}, newAdmissionError(fieldKey, admissionInvalidValue)
	}
	var rawActions []json.RawMessage
	if json.Unmarshal(object["actions"], &rawActions) != nil {
		return collectionActionPayload{}, newAdmissionError(fieldKey, admissionInvalidValue)
	}
	if len(rawActions) == 0 {
		return collectionActionPayload{}, &AdmissionError{
			field: fieldKey + ".actions", collectionField: fieldKey, reason: admissionEmptyCollectionActions,
		}
	}
	if len(rawActions) > maxMutationCollectionActions {
		return collectionActionPayload{}, newAdmissionLimitError(
			fieldKey+".actions", fieldKey, admissionCollectionActionCountExceeded,
			len(rawActions), maxMutationCollectionActions,
		)
	}
	payload := collectionActionPayload{Actions: make([]collectionAction, 0, len(rawActions))}
	for _, rawAction := range rawActions {
		action, admissionErr := decodeArtifactCollectionAction(fieldKey, rawAction)
		if admissionErr != nil {
			return collectionActionPayload{}, admissionErr
		}
		payload.Actions = append(payload.Actions, action)
	}
	return payload, nil
}

func decodeArtifactCollectionAction(fieldKey string, raw json.RawMessage) (collectionAction, *AdmissionError) {
	var object map[string]json.RawMessage
	if json.Unmarshal(raw, &object) != nil {
		return collectionAction{}, newAdmissionError(fieldKey, admissionInvalidValue)
	}
	var op string
	if json.Unmarshal(object["op"], &op) != nil {
		return collectionAction{}, newAdmissionError(fieldKey, admissionInvalidValue)
	}
	action := collectionAction{Op: op}
	switch op {
	case "add_token":
		if !artifactObjectHasOnlyFields(object, "op", "raw_text") {
			return collectionAction{}, newAdmissionError(fieldKey, admissionInvalidValue)
		}
		text, ok := artifactStringActionField(object, "raw_text")
		normalized, valid := fieldnorm.NormalizeLine(text)
		if !ok || !valid {
			return collectionAction{}, newAdmissionError(fieldKey, admissionInvalidValue)
		}
		action.RawText, action.NormalizedText = text, normalized
	case "add_tag":
		if !artifactObjectHasOnlyFields(object, "op", "tag_name") {
			return collectionAction{}, newAdmissionError(fieldKey, admissionInvalidValue)
		}
		text, ok := artifactStringActionField(object, "tag_name")
		label, normalized, valid := fieldnorm.NormalizeTagLabel(text)
		if !ok || !valid {
			return collectionAction{}, newAdmissionError(fieldKey, admissionInvalidValue)
		}
		action.RawText, action.NormalizedText = label, normalized
	case "remove_tag":
		if !artifactObjectHasOnlyFields(object, "op", "item_ref") {
			return collectionAction{}, newAdmissionError(fieldKey, admissionInvalidValue)
		}
		itemRef, ok := artifactStringActionField(object, "item_ref")
		if !ok {
			return collectionAction{}, newAdmissionError(fieldKey, admissionInvalidValue)
		}
		if _, _, err := links.ParseRecordTagItemRef(itemRef); err != nil {
			return collectionAction{}, newAdmissionError(fieldKey, admissionInvalidValue)
		}
		action.ItemRef = itemRef
	case "add_record_ref":
		if !artifactObjectHasOnlyFields(object, "op", "linked_record_id") {
			return collectionAction{}, newAdmissionError(fieldKey, admissionInvalidValue)
		}
		id, ok := artifactUUIDActionField(object, "linked_record_id")
		if !ok {
			return collectionAction{}, newAdmissionError(fieldKey, admissionInvalidValue)
		}
		action.LinkedRecordID = &id
	case "remove_record_ref":
		if !artifactObjectHasOnlyFields(object, "op", "item_ref") {
			return collectionAction{}, newAdmissionError(fieldKey, admissionInvalidValue)
		}
		itemRef, ok := artifactStringActionField(object, "item_ref")
		if !ok {
			return collectionAction{}, newAdmissionError(fieldKey, admissionInvalidValue)
		}
		if _, err := links.ParseRecordRefItemRef(itemRef); err != nil {
			return collectionAction{}, newAdmissionError(fieldKey, admissionInvalidValue)
		}
		action.ItemRef = itemRef
	case "add_party_ref":
		if !artifactObjectHasOnlyFields(object, "op", "party_id") {
			return collectionAction{}, newAdmissionError(fieldKey, admissionInvalidValue)
		}
		id, ok := artifactUUIDActionField(object, "party_id")
		if !ok {
			return collectionAction{}, newAdmissionError(fieldKey, admissionInvalidValue)
		}
		action.PartyID = &id
	case "remove_party_ref":
		if !artifactObjectHasOnlyFields(object, "op", "item_ref") {
			return collectionAction{}, newAdmissionError(fieldKey, admissionInvalidValue)
		}
		itemRef, ok := artifactStringActionField(object, "item_ref")
		if !ok {
			return collectionAction{}, newAdmissionError(fieldKey, admissionInvalidValue)
		}
		if _, err := links.ParsePartyRefItemRef(itemRef); err != nil {
			return collectionAction{}, newAdmissionError(fieldKey, admissionInvalidValue)
		}
		action.ItemRef = itemRef
	case "add_risk_ref":
		if !artifactObjectHasOnlyFields(object, "op", "risk_ref_text") {
			return collectionAction{}, newAdmissionError(fieldKey, admissionInvalidValue)
		}
		text, ok := artifactStringActionField(object, "risk_ref_text")
		normalized, valid := fieldnorm.NormalizeLine(text)
		if !ok || !valid {
			return collectionAction{}, newAdmissionError(fieldKey, admissionInvalidValue)
		}
		action.RiskRefText = normalized
		action.NormalizedText = riskReferenceCaseFolder.String(normalized)
	case "remove_risk_ref":
		if !artifactObjectHasOnlyFields(object, "op", "item_ref") {
			return collectionAction{}, newAdmissionError(fieldKey, admissionInvalidValue)
		}
		itemRef, ok := artifactStringActionField(object, "item_ref")
		if !ok {
			return collectionAction{}, newAdmissionError(fieldKey, admissionInvalidValue)
		}
		if _, err := parseRiskRefItemRef(itemRef); err != nil {
			return collectionAction{}, newAdmissionError(fieldKey, admissionInvalidValue)
		}
		action.ItemRef = itemRef
	default:
		return collectionAction{}, newAdmissionError(fieldKey, admissionInvalidValue)
	}
	return action, nil
}

func artifactObjectHasOnlyFields(object map[string]json.RawMessage, fields ...string) bool {
	allowed := make(map[string]struct{}, len(fields))
	for _, field := range fields {
		allowed[field] = struct{}{}
		if _, present := object[field]; !present {
			return false
		}
	}
	for key := range object {
		if _, admitted := allowed[key]; !admitted {
			return false
		}
	}
	return true
}

func artifactStringActionField(object map[string]json.RawMessage, field string) (string, bool) {
	value, present := object[field]
	if !present {
		return "", false
	}
	var text string
	if json.Unmarshal(value, &text) != nil || strings.TrimSpace(text) == "" {
		return "", false
	}
	return text, true
}

func artifactUUIDActionField(object map[string]json.RawMessage, field string) (uuid.UUID, bool) {
	text, ok := artifactStringActionField(object, field)
	if !ok {
		return uuid.Nil, false
	}
	parsed, err := uuid.Parse(text)
	return parsed, err == nil && parsed.String() == text
}
