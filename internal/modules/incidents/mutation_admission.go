package incidents

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"io"
	"slices"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/google/uuid"
	"golang.org/x/text/unicode/norm"

	"github.com/JochiRaider/cartulary/internal/platform/authn"
	"github.com/JochiRaider/cartulary/internal/platform/strictjson"
)

const (
	incidentKeyMaxBytes      = 128
	singleLineTitleMaxRunes  = 512
	incidentMetadataMaxRunes = 128
	multilineBodyMaxRunes    = 16384
	reasonNoteMaxRunes       = 4096
)

var (
	membershipRoles    = []string{"viewer", "editor", "reviewer", "admin"}
	canonicalTLPTokens = map[string]struct{}{
		"TLP:CLEAR": {}, "TLP:GREEN": {}, "TLP:AMBER": {},
		"TLP:AMBER+STRICT": {}, "TLP:RED": {},
	}
)

// AdmissionError is the transport-neutral closed validation failure returned
// by the Incidents mutation boundary.
type AdmissionError struct {
	field      string
	reasonCode string
}

func (e *AdmissionError) Error() string { return "incidents: invalid mutation admission" }

func (e *AdmissionError) Field() (string, bool) {
	if e == nil || e.field == "" {
		return "", false
	}
	return e.field, true
}

func (e *AdmissionError) ReasonCode() string {
	if e == nil {
		return ""
	}
	return e.reasonCode
}

type LifecycleAction uint8

const (
	LifecycleActionClose LifecycleAction = iota + 1
	LifecycleActionReopen
)

func (a LifecycleAction) String() string {
	switch a {
	case LifecycleActionClose:
		return "close"
	case LifecycleActionReopen:
		return "reopen"
	default:
		return ""
	}
}

type IncidentCreateAdmission struct {
	admitted               bool
	clientTxnID            string
	incidentKey            string
	title                  string
	description            *string
	severity               *string
	tlp                    *string
	currentPhase           *string
	primaryExternalCaseRef *string
	requestHash            [sha256.Size]byte
}

func (a IncidentCreateAdmission) ClientTxnID() string { return a.clientTxnID }

type optionalNullableString struct {
	present bool
	value   *string
}

type IncidentPatchAdmission struct {
	admitted               bool
	baseIncidentVersion    int64
	description            optionalNullableString
	severity               optionalNullableString
	tlp                    optionalNullableString
	currentPhase           optionalNullableString
	primaryExternalCaseRef optionalNullableString
}

type IncidentLifecycleAdmission struct {
	admitted            bool
	action              LifecycleAction
	baseIncidentVersion int64
	clientTxnID         string
	reason              string
	requestHash         [sha256.Size]byte
}

func (a IncidentLifecycleAdmission) ClientTxnID() string     { return a.clientTxnID }
func (a IncidentLifecycleAdmission) Action() LifecycleAction { return a.action }

type MembershipCreateAdmission struct {
	admitted    bool
	clientTxnID string
	userID      *uuid.UUID
	email       *string
	role        string
	requestHash [sha256.Size]byte
}

func (a MembershipCreateAdmission) ClientTxnID() string { return a.clientTxnID }

func (a MembershipCreateAdmission) TargetUserID() (uuid.UUID, bool) {
	if a.userID == nil {
		return uuid.Nil, false
	}
	return *a.userID, true
}

func (a MembershipCreateAdmission) TargetEmail() (string, bool) {
	if a.email == nil {
		return "", false
	}
	return *a.email, true
}

type MembershipPatchAdmission struct {
	admitted              bool
	baseMembershipVersion int64
	role                  string
}

type MembershipDeleteAdmission struct {
	admitted              bool
	baseMembershipVersion int64
}

func AdmitIncidentCreateJSON(reader io.Reader) (IncidentCreateAdmission, *AdmissionError) {
	raw, admissionErr := decodeAdmissionObject(reader)
	if admissionErr != nil {
		return IncidentCreateAdmission{}, admissionErr
	}
	allowed := stringSet(
		"client_txn_id", "incident_key", "title", "description", "severity", "tlp",
		"current_phase", "primary_external_case_ref",
	)
	serverManaged := stringSet(
		"incident_id", "status", "incident_version", "closed_at", "created_at", "updated_at",
		"created_by_user_id", "updated_by_user_id",
	)
	for _, key := range sortedObjectKeys(raw) {
		switch {
		case key == "initial_memberships":
			return IncidentCreateAdmission{}, invalidAdmission(key, "collaborator_seeding_not_supported")
		case containsString(serverManaged, key):
			return IncidentCreateAdmission{}, invalidAdmission(key, "server_managed_field")
		case !containsString(allowed, key):
			return IncidentCreateAdmission{}, invalidAdmission(key, "unknown_field")
		}
	}

	admission := IncidentCreateAdmission{admitted: true}
	if value, ok := raw["client_txn_id"]; !ok {
		return IncidentCreateAdmission{}, invalidAdmission("client_txn_id", "missing_required_field")
	} else if json.Unmarshal(value, &admission.clientTxnID) != nil || strings.TrimSpace(admission.clientTxnID) == "" {
		return IncidentCreateAdmission{}, invalidAdmission("client_txn_id", "missing_required_field")
	}
	if value, ok := raw["incident_key"]; !ok {
		return IncidentCreateAdmission{}, invalidAdmission("incident_key", "missing_required_field")
	} else if admission.incidentKey, ok = normalizeIncidentKeyValue(value); !ok {
		return IncidentCreateAdmission{}, invalidAdmission("incident_key", "invalid_value")
	}
	if value, ok := raw["title"]; !ok {
		return IncidentCreateAdmission{}, invalidAdmission("title", "missing_required_field")
	} else if admission.title, ok = normalizeTitleValue(value); !ok {
		return IncidentCreateAdmission{}, invalidAdmission("title", "invalid_value")
	}
	var ok bool
	if admission.description, ok = normalizeNullableNoteField(raw, "description"); !ok {
		return IncidentCreateAdmission{}, invalidAdmission("description", "invalid_value")
	}
	if admission.severity, ok = normalizeNullableMetadataField(raw, "severity"); !ok {
		return IncidentCreateAdmission{}, invalidAdmission("severity", "invalid_value")
	}
	if admission.tlp, ok = normalizeNullableTLPField(raw, "tlp"); !ok {
		return IncidentCreateAdmission{}, invalidAdmission("tlp", "invalid_value")
	}
	if admission.currentPhase, ok = normalizeNullableMetadataField(raw, "current_phase"); !ok {
		return IncidentCreateAdmission{}, invalidAdmission("current_phase", "invalid_value")
	}
	if admission.primaryExternalCaseRef, ok = normalizeNullableMetadataField(raw, "primary_external_case_ref"); !ok {
		return IncidentCreateAdmission{}, invalidAdmission("primary_external_case_ref", "invalid_value")
	}
	admission.requestHash = hashRequestPayload(incidentCreateIdempotencyPayload{
		ClientTxnID: admission.clientTxnID, CurrentPhase: admission.currentPhase,
		Description: admission.description, IncidentKey: admission.incidentKey,
		PrimaryExternalCaseRef: admission.primaryExternalCaseRef, Severity: admission.severity,
		Title: admission.title, TLP: admission.tlp,
	})
	return admission, nil
}

func AdmitIncidentPatchJSON(reader io.Reader) (IncidentPatchAdmission, *AdmissionError) {
	raw, admissionErr := decodeAdmissionObject(reader)
	if admissionErr != nil {
		return IncidentPatchAdmission{}, admissionErr
	}
	allowed := stringSet(
		"base_incident_version", "description", "severity", "tlp", "current_phase",
		"primary_external_case_ref",
	)
	forbidden := stringSet(
		"incident_id", "incident_key", "title", "status", "created_at", "created_by_user_id",
		"updated_at", "updated_by_user_id", "closed_at", "incident_version", "memberships",
		"saved_views", "saved_view", "workbook_preferences", "default_workbook_preferences",
		"user_workbook_preferences",
	)
	for _, key := range sortedObjectKeys(raw) {
		switch {
		case containsString(forbidden, key):
			return IncidentPatchAdmission{}, invalidAdmission(key, "forbidden_field")
		case !containsString(allowed, key):
			return IncidentPatchAdmission{}, invalidAdmission(key, "unknown_field")
		}
	}
	admission := IncidentPatchAdmission{admitted: true}
	if value, ok := raw["base_incident_version"]; !ok {
		return IncidentPatchAdmission{}, invalidAdmission("base_incident_version", "missing_required_field")
	} else if json.Unmarshal(value, &admission.baseIncidentVersion) != nil || admission.baseIncidentVersion < 1 {
		return IncidentPatchAdmission{}, invalidAdmission("base_incident_version", "invalid_base_incident_version")
	}
	var ok bool
	if admission.description, ok = decodeOptionalNullableNoteField(raw, "description"); !ok {
		return IncidentPatchAdmission{}, invalidAdmission("description", "invalid_value")
	}
	if admission.severity, ok = decodeOptionalNullableMetadataField(raw, "severity"); !ok {
		return IncidentPatchAdmission{}, invalidAdmission("severity", "invalid_value")
	}
	if admission.tlp, ok = decodeOptionalNullableTLPField(raw, "tlp"); !ok {
		return IncidentPatchAdmission{}, invalidAdmission("tlp", "invalid_value")
	}
	if admission.currentPhase, ok = decodeOptionalNullableMetadataField(raw, "current_phase"); !ok {
		return IncidentPatchAdmission{}, invalidAdmission("current_phase", "invalid_value")
	}
	if admission.primaryExternalCaseRef, ok = decodeOptionalNullableMetadataField(raw, "primary_external_case_ref"); !ok {
		return IncidentPatchAdmission{}, invalidAdmission("primary_external_case_ref", "invalid_value")
	}
	return admission, nil
}

func AdmitIncidentLifecycleJSON(action LifecycleAction, reader io.Reader) (IncidentLifecycleAdmission, *AdmissionError) {
	if action.String() == "" {
		return IncidentLifecycleAdmission{}, invalidAdmission("action", "invalid_value")
	}
	raw, admissionErr := decodeAdmissionObject(reader)
	if admissionErr != nil {
		return IncidentLifecycleAdmission{}, admissionErr
	}
	allowed := stringSet("base_incident_version", "client_txn_id", "reason")
	if key, ok := firstUnknownObjectKey(raw, allowed); ok {
		return IncidentLifecycleAdmission{}, invalidAdmission(key, "unknown_field")
	}
	admission := IncidentLifecycleAdmission{admitted: true, action: action}
	if value, ok := raw["base_incident_version"]; !ok {
		return IncidentLifecycleAdmission{}, invalidAdmission("base_incident_version", "missing_required_field")
	} else if isJSONNull(value) {
		return IncidentLifecycleAdmission{}, invalidAdmission("base_incident_version", "field_not_nullable")
	} else if json.Unmarshal(value, &admission.baseIncidentVersion) != nil || admission.baseIncidentVersion < 1 {
		return IncidentLifecycleAdmission{}, invalidAdmission("base_incident_version", "invalid_base_incident_version")
	}
	if value, ok := raw["client_txn_id"]; !ok {
		return IncidentLifecycleAdmission{}, invalidAdmission("client_txn_id", "missing_required_field")
	} else if isJSONNull(value) {
		return IncidentLifecycleAdmission{}, invalidAdmission("client_txn_id", "field_not_nullable")
	} else if json.Unmarshal(value, &admission.clientTxnID) != nil || strings.TrimSpace(admission.clientTxnID) == "" {
		return IncidentLifecycleAdmission{}, invalidAdmission("client_txn_id", "invalid_client_txn_id")
	}
	if value, ok := raw["reason"]; !ok {
		return IncidentLifecycleAdmission{}, invalidAdmission("reason", "missing_required_field")
	} else if isJSONNull(value) {
		return IncidentLifecycleAdmission{}, invalidAdmission("reason", "field_not_nullable")
	} else {
		var reason string
		if json.Unmarshal(value, &reason) != nil {
			return IncidentLifecycleAdmission{}, invalidAdmission("reason", "invalid_reason")
		}
		normalized, reasonCode := normalizeLifecycleReason(reason)
		if reasonCode != "" {
			return IncidentLifecycleAdmission{}, invalidAdmission("reason", reasonCode)
		}
		admission.reason = normalized
	}
	admission.requestHash = hashRequestPayload(incidentLifecycleIdempotencyPayload{
		ActionRoute: action.String(), BaseIncidentVersion: admission.baseIncidentVersion,
		Reason: admission.reason,
	})
	return admission, nil
}

func AdmitMembershipCreateJSON(reader io.Reader) (MembershipCreateAdmission, *AdmissionError) {
	raw, admissionErr := decodeAdmissionObject(reader)
	if admissionErr != nil {
		return MembershipCreateAdmission{}, admissionErr
	}
	allowed := stringSet("client_txn_id", "user_id", "email", "role")
	if key, ok := firstUnknownObjectKey(raw, allowed); ok {
		return MembershipCreateAdmission{}, invalidAdmission(key, "unknown_field")
	}
	admission := MembershipCreateAdmission{admitted: true}
	if value, ok := raw["client_txn_id"]; !ok {
		return MembershipCreateAdmission{}, invalidAdmission("client_txn_id", "missing_required_field")
	} else if json.Unmarshal(value, &admission.clientTxnID) != nil || strings.TrimSpace(admission.clientTxnID) == "" {
		return MembershipCreateAdmission{}, invalidAdmission("client_txn_id", "missing_required_field")
	}
	if value, ok := raw["user_id"]; ok {
		var parsed string
		if json.Unmarshal(value, &parsed) != nil {
			return MembershipCreateAdmission{}, invalidAdmission("user_id", "invalid_user_id")
		}
		userID, err := uuid.Parse(parsed)
		if err != nil {
			return MembershipCreateAdmission{}, invalidAdmission("user_id", "invalid_user_id")
		}
		admission.userID = &userID
	}
	if value, ok := raw["email"]; ok {
		var rawEmail string
		if json.Unmarshal(value, &rawEmail) != nil {
			return MembershipCreateAdmission{}, invalidAdmission("email", "invalid_email")
		}
		normalized, _, ok := authn.NormalizeEmailAddress(rawEmail)
		if !ok {
			return MembershipCreateAdmission{}, invalidAdmission("email", "invalid_email")
		}
		admission.email = &normalized
	}
	if (admission.userID == nil) == (admission.email == nil) {
		return MembershipCreateAdmission{}, invalidAdmission("user_id", "exactly_one_target_selector")
	}
	if value, ok := raw["role"]; !ok {
		return MembershipCreateAdmission{}, invalidAdmission("role", "missing_required_field")
	} else if json.Unmarshal(value, &admission.role) != nil || !slices.Contains(membershipRoles, admission.role) {
		return MembershipCreateAdmission{}, invalidAdmission("role", "invalid_role")
	}
	admission.requestHash = hashRequestPayload(membershipCreateIdempotencyPayload{
		ClientTxnID: admission.clientTxnID, Email: admission.email, Role: admission.role,
		UserID: admission.userID,
	})
	return admission, nil
}

func AdmitMembershipPatchJSON(reader io.Reader) (MembershipPatchAdmission, *AdmissionError) {
	raw, admissionErr := decodeAdmissionObject(reader)
	if admissionErr != nil {
		return MembershipPatchAdmission{}, admissionErr
	}
	allowed := stringSet("base_membership_version", "role")
	forbidden := stringSet(
		"incident_id", "user_id", "joined_at", "added_by_user_id", "updated_at",
		"updated_by_user_id", "membership_version",
	)
	for _, key := range sortedObjectKeys(raw) {
		switch {
		case containsString(forbidden, key):
			return MembershipPatchAdmission{}, invalidAdmission(key, "forbidden_field")
		case !containsString(allowed, key):
			return MembershipPatchAdmission{}, invalidAdmission(key, "unknown_field")
		}
	}
	admission := MembershipPatchAdmission{admitted: true}
	if value, ok := raw["base_membership_version"]; !ok {
		return MembershipPatchAdmission{}, invalidAdmission("base_membership_version", "missing_required_field")
	} else if json.Unmarshal(value, &admission.baseMembershipVersion) != nil || admission.baseMembershipVersion < 1 {
		return MembershipPatchAdmission{}, invalidAdmission("base_membership_version", "invalid_base_membership_version")
	}
	if value, ok := raw["role"]; !ok {
		return MembershipPatchAdmission{}, invalidAdmission("role", "missing_required_field")
	} else if json.Unmarshal(value, &admission.role) != nil || !slices.Contains(membershipRoles, admission.role) {
		return MembershipPatchAdmission{}, invalidAdmission("role", "invalid_role")
	}
	return admission, nil
}

func AdmitMembershipDeleteJSON(reader io.Reader) (MembershipDeleteAdmission, *AdmissionError) {
	raw, admissionErr := decodeAdmissionObject(reader)
	if admissionErr != nil {
		return MembershipDeleteAdmission{}, admissionErr
	}
	allowed := stringSet("base_membership_version")
	if key, ok := firstUnknownObjectKey(raw, allowed); ok {
		return MembershipDeleteAdmission{}, invalidAdmission(key, "unknown_field")
	}
	admission := MembershipDeleteAdmission{admitted: true}
	if value, ok := raw["base_membership_version"]; !ok {
		return MembershipDeleteAdmission{}, invalidAdmission("base_membership_version", "missing_required_field")
	} else if json.Unmarshal(value, &admission.baseMembershipVersion) != nil || admission.baseMembershipVersion < 1 {
		return MembershipDeleteAdmission{}, invalidAdmission("base_membership_version", "invalid_base_membership_version")
	}
	return admission, nil
}

func decodeAdmissionObject(reader io.Reader) (map[string]json.RawMessage, *AdmissionError) {
	raw, err := strictjson.DecodeObject(reader)
	if err != nil {
		return nil, invalidAdmission("", "request_not_object")
	}
	return raw, nil
}

func invalidAdmission(field string, reasonCode string) *AdmissionError {
	return &AdmissionError{field: field, reasonCode: reasonCode}
}

func stringSet(values ...string) map[string]struct{} {
	set := make(map[string]struct{}, len(values))
	for _, value := range values {
		set[value] = struct{}{}
	}
	return set
}

func containsString(set map[string]struct{}, value string) bool {
	_, ok := set[value]
	return ok
}

func sortedObjectKeys(raw map[string]json.RawMessage) []string {
	keys := make([]string, 0, len(raw))
	for key := range raw {
		keys = append(keys, key)
	}
	slices.Sort(keys)
	return keys
}

func firstUnknownObjectKey(raw map[string]json.RawMessage, allowed map[string]struct{}) (string, bool) {
	for _, key := range sortedObjectKeys(raw) {
		if !containsString(allowed, key) {
			return key, true
		}
	}
	return "", false
}

type singleLineConstraint struct {
	maxBytes int
	maxRunes int
}

func normalizeIncidentKeyValue(value json.RawMessage) (string, bool) {
	var raw string
	if json.Unmarshal(value, &raw) != nil {
		return "", false
	}
	normalized, ok := normalizeSingleLine(raw, singleLineConstraint{maxBytes: incidentKeyMaxBytes})
	return normalized, ok && normalized != ""
}

func normalizeTitleValue(value json.RawMessage) (string, bool) {
	var raw string
	if json.Unmarshal(value, &raw) != nil {
		return "", false
	}
	normalized, ok := normalizeSingleLine(raw, singleLineConstraint{maxRunes: singleLineTitleMaxRunes})
	return normalized, ok && normalized != ""
}

func normalizeNullableMetadataField(raw map[string]json.RawMessage, field string) (*string, bool) {
	value, ok := raw[field]
	if !ok || isJSONNull(value) {
		return nil, true
	}
	var line string
	if json.Unmarshal(value, &line) != nil {
		return nil, false
	}
	normalized, ok := normalizeSingleLine(line, singleLineConstraint{maxRunes: incidentMetadataMaxRunes})
	if !ok || normalized == "" {
		return nil, ok
	}
	return &normalized, true
}

func normalizeNullableTLPField(raw map[string]json.RawMessage, field string) (*string, bool) {
	value, ok := raw[field]
	if !ok || isJSONNull(value) {
		return nil, true
	}
	normalized, ok := normalizeTLPValue(value)
	if !ok {
		return nil, false
	}
	return &normalized, true
}

func normalizeNullableNoteField(raw map[string]json.RawMessage, field string) (*string, bool) {
	value, ok := raw[field]
	if !ok || isJSONNull(value) {
		return nil, true
	}
	var note string
	if json.Unmarshal(value, &note) != nil {
		return nil, false
	}
	normalized, ok := normalizeNote(note, multilineBodyMaxRunes)
	if !ok || normalized == "" {
		return nil, ok
	}
	return &normalized, true
}

func decodeOptionalNullableMetadataField(raw map[string]json.RawMessage, field string) (optionalNullableString, bool) {
	value, ok := raw[field]
	if !ok {
		return optionalNullableString{}, true
	}
	if isJSONNull(value) {
		return optionalNullableString{present: true}, true
	}
	var line string
	if json.Unmarshal(value, &line) != nil {
		return optionalNullableString{}, false
	}
	normalized, ok := normalizeSingleLine(line, singleLineConstraint{maxRunes: incidentMetadataMaxRunes})
	if !ok || normalized == "" {
		return optionalNullableString{present: true}, ok
	}
	return optionalNullableString{present: true, value: &normalized}, true
}

func decodeOptionalNullableTLPField(raw map[string]json.RawMessage, field string) (optionalNullableString, bool) {
	value, ok := raw[field]
	if !ok {
		return optionalNullableString{}, true
	}
	if isJSONNull(value) {
		return optionalNullableString{present: true}, true
	}
	normalized, ok := normalizeTLPValue(value)
	if !ok {
		return optionalNullableString{}, false
	}
	return optionalNullableString{present: true, value: &normalized}, true
}

func decodeOptionalNullableNoteField(raw map[string]json.RawMessage, field string) (optionalNullableString, bool) {
	value, ok := raw[field]
	if !ok {
		return optionalNullableString{}, true
	}
	if isJSONNull(value) {
		return optionalNullableString{present: true}, true
	}
	var note string
	if json.Unmarshal(value, &note) != nil {
		return optionalNullableString{}, false
	}
	normalized, ok := normalizeNote(note, multilineBodyMaxRunes)
	if !ok || normalized == "" {
		return optionalNullableString{present: true}, ok
	}
	return optionalNullableString{present: true, value: &normalized}, true
}

func normalizeSingleLine(raw string, constraint singleLineConstraint) (string, bool) {
	normalized := norm.NFC.String(strings.TrimFunc(raw, unicode.IsSpace))
	for _, r := range normalized {
		if unicode.Is(unicode.Cc, r) {
			return "", false
		}
	}
	if constraint.maxBytes > 0 && len([]byte(normalized)) > constraint.maxBytes {
		return "", false
	}
	if constraint.maxRunes > 0 && len([]rune(normalized)) > constraint.maxRunes {
		return "", false
	}
	return normalized, true
}

func normalizeNote(raw string, maxRunes int) (string, bool) {
	normalized := norm.NFC.String(raw)
	normalized = strings.ReplaceAll(normalized, "\r\n", "\n")
	normalized = strings.ReplaceAll(normalized, "\r", "\n")
	normalized = strings.TrimFunc(normalized, unicode.IsSpace)
	for _, r := range normalized {
		switch {
		case r == '\n' || r == '\t':
		case unicode.Is(unicode.Cc, r):
			return "", false
		}
	}
	if maxRunes > 0 && len([]rune(normalized)) > maxRunes {
		return "", false
	}
	return normalized, true
}

func normalizeTLPValue(value json.RawMessage) (string, bool) {
	var raw string
	if json.Unmarshal(value, &raw) != nil {
		return "", false
	}
	_, ok := canonicalTLPTokens[raw]
	return raw, ok
}

func normalizeLifecycleReason(raw string) (string, string) {
	normalized := norm.NFC.String(raw)
	normalized = strings.ReplaceAll(normalized, "\r\n", "\n")
	normalized = strings.ReplaceAll(normalized, "\r", "\n")
	normalized = strings.TrimFunc(normalized, unicode.IsSpace)
	if normalized == "" {
		return "", "reason_empty_after_normalization"
	}
	if utf8.RuneCountInString(normalized) > reasonNoteMaxRunes {
		return "", "reason_too_long"
	}
	for _, r := range normalized {
		switch {
		case r == '\n' || r == '\t':
		case unicode.Is(unicode.Cc, r) || unicode.Is(unicode.Cf, r):
			return "", "control_character_not_allowed"
		}
	}
	return normalized, ""
}

func isJSONNull(value json.RawMessage) bool {
	return bytes.Equal(bytes.TrimSpace(value), []byte("null"))
}
