package entities

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/JochiRaider/cartulary/internal/modules/incidentbundles/sourceport"
)

const (
	entityMentionsBundlePath = "data/entity_mentions.ndjson"
	hostsBundlePath          = "data/hosts.ndjson"
	identitiesBundlePath     = "data/identities.ndjson"
	preservedIDsBundlePath   = "data/entity_preserved_identifiers.ndjson"
	entityAliasesBundlePath  = "data/entity_aliases.ndjson"
	entitySourceIdentity     = "entities.source_identity_admitted"
	entityMentionsObserved   = "entities.mentions_observational"
	entityEnvelopeTypeScope  = "entities.envelope_type_scope"
	entityResolutionMerge    = "entities.resolution_merge_coherent"
	entityNormalized         = "entities.alias_identifier_normalized"
	entityClassified         = "entities.alias_identifier_classified"
	entityUnique             = "entities.alias_identifier_unique"
	entitySameIncident       = "entities.alias_identifier_same_incident"
)

var (
	entityPortableIntegerPattern   = regexp.MustCompile(`^(?:0|[1-9][0-9]*)$`)
	entityPortableTimestampPattern = regexp.MustCompile(`^[0-9]{4}-[0-9]{2}-[0-9]{2}T[0-9]{2}:[0-9]{2}:[0-9]{2}(?:\.[0-9]{1,6})?\+00:00$`)
	hostPortableMembers            = []string{
		"record_id", "incident_id", "display_name", "hostname", "aad_device_id", "fqdn",
		"entity_origin", "seed_entity_mention_id", "host_state", "merged_into_record_id",
		"row_version", "created_at", "updated_at", "created_by_user_id", "updated_by_user_id",
		"location", "os_platform", "business_owner", "criticality", "containment_status",
	}
	identityPortableMembers = []string{
		"record_id", "incident_id", "display_name", "upn", "email", "sam_account_name",
		"aad_object_id", "sid", "entity_origin", "seed_entity_mention_id", "identity_state",
		"merged_into_record_id", "row_version", "created_at", "updated_at", "created_by_user_id",
		"updated_by_user_id", "privilege_level", "mfa_state", "reset_status",
	}
	mentionPortableMembers = []string{
		"entity_mention_id", "source_record_id", "entity_type", "source_field_key", "origin_kind",
		"origin_locator", "raw_text", "normalized_text", "resolution_status", "row_version", "ordinal",
		"created_by_user_id", "created_at", "resolved_record_id", "resolved_by_user_id", "resolved_at",
		"resolution_method",
	}
	preservedPortableMembers = []string{
		"entity_preserved_identifier_id", "incident_id", "record_id", "entity_type", "identifier_type",
		"raw_value", "normalized_value", "classification", "created_by_user_id", "created_at", "deleted_at",
	}
	aliasPortableMembers = []string{
		"entity_alias_id", "incident_id", "record_id", "entity_type", "raw_text", "normalized_text",
		"classification", "created_by_user_id", "created_at", "deleted_at",
	}
)

type entityPortableBinding struct {
	operationID   string
	incidentID    uuid.UUID
	bundleVersion int
	contractMajor int
}

type preparedEntityImport struct {
	binding    entityPortableBinding
	hosts      []portableHostRow
	identities []portableIdentityRow
	mentions   []portableMentionRow
	preserved  []portablePreservedIdentifierRow
	aliases    []portableAliasRow
}

type portableHostRow struct {
	RecordID, IncidentID                                                uuid.UUID
	DisplayName                                                         string
	Hostname, AADDeviceID, FQDN                                         *string
	EntityOrigin                                                        string
	SeedMentionID, MergedIntoRecordID                                   *uuid.UUID
	State                                                               string
	RowVersion                                                          int64
	CreatedAt, UpdatedAt                                                time.Time
	PortableCreatedByID, PortableUpdatedByID                            uuid.UUID
	RuntimeCreatedByID, RuntimeUpdatedByID                              uuid.UUID
	Location, OSPlatform, BusinessOwner, Criticality, ContainmentStatus *string
}

type portableIdentityRow struct {
	RecordID, IncidentID                         uuid.UUID
	DisplayName                                  string
	UPN, Email, SAMAccountName, AADObjectID, SID *string
	EntityOrigin                                 string
	SeedMentionID, MergedIntoRecordID            *uuid.UUID
	State                                        string
	RowVersion                                   int64
	CreatedAt, UpdatedAt                         time.Time
	PortableCreatedByID, PortableUpdatedByID     uuid.UUID
	RuntimeCreatedByID, RuntimeUpdatedByID       uuid.UUID
	PrivilegeLevel, MFAState, ResetStatus        *string
}

type portableMentionRow struct {
	MentionID, SourceRecordID                 uuid.UUID
	EntityType, SourceFieldKey, OriginKind    string
	OriginLocator, RawText, NormalizedText    string
	ResolutionStatus                          string
	RowVersion                                int64
	Ordinal                                   int32
	PortableCreatedByID, RuntimeCreatedByID   uuid.UUID
	CreatedAt                                 time.Time
	ResolvedRecordID                          *uuid.UUID
	PortableResolvedByID, RuntimeResolvedByID *uuid.UUID
	ResolvedAt                                *time.Time
	ResolutionMethod                          *string
}

type portablePreservedIdentifierRow struct {
	PreservedID, IncidentID, RecordID         uuid.UUID
	EntityType, IdentifierType                string
	RawValue, NormalizedValue, Classification string
	PortableCreatedByID, RuntimeCreatedByID   uuid.UUID
	CreatedAt                                 time.Time
	DeletedAt                                 *time.Time
}

type portableAliasRow struct {
	AliasID, IncidentID, RecordID           uuid.UUID
	EntityType, RawText, NormalizedText     string
	Classification                          string
	PortableCreatedByID, RuntimeCreatedByID uuid.UUID
	CreatedAt                               time.Time
	DeletedAt                               *time.Time
}

type entityFailureCandidate struct {
	invariant, path, identity, digest string
}

func (binding entityPortableBinding) matches(importContext sourceport.ImportContext) bool {
	return binding.operationID != "" && binding.operationID == importContext.OperationID &&
		binding.incidentID != uuid.Nil && binding.incidentID == importContext.IncidentID &&
		binding.bundleVersion == 2 && binding.bundleVersion == importContext.BundleVersion &&
		binding.contractMajor == sourceport.ContractMajor
}

func entityFailure(invariant, path, identity string, value any) entityFailureCandidate {
	encoded, _ := json.Marshal(value)
	digest := sha256.Sum256(encoded)
	return entityFailureCandidate{invariant: invariant, path: path, identity: identity, digest: hex.EncodeToString(digest[:])}
}

func selectedEntityFailure(candidates []entityFailureCandidate) error {
	if len(candidates) == 0 {
		return nil
	}
	sort.Slice(candidates, func(i, j int) bool {
		left, right := candidates[i], candidates[j]
		if entityInvariantRank(left.invariant) != entityInvariantRank(right.invariant) {
			return entityInvariantRank(left.invariant) < entityInvariantRank(right.invariant)
		}
		if left.path != right.path {
			return left.path < right.path
		}
		if (left.identity != "") != (right.identity != "") {
			return left.identity != ""
		}
		if left.identity != right.identity {
			return left.identity < right.identity
		}
		return left.digest < right.digest
	})
	return entitySourceFailure(candidates[0].invariant)
}

func entityInvariantRank(invariant string) int {
	switch invariant {
	case entitySourceIdentity:
		return 1
	case entityMentionsObserved:
		return 2
	case entityEnvelopeTypeScope:
		return 3
	case entityResolutionMerge:
		return 4
	case entityNormalized:
		return 5
	case entityClassified:
		return 6
	case entityUnique:
		return 7
	case entitySameIncident:
		return 8
	default:
		return 0
	}
}

func exactEntityPortableMembers(row map[string]any, members []string) bool {
	if len(row) != len(members) {
		return false
	}
	for _, member := range members {
		if _, present := row[member]; !present {
			return false
		}
	}
	return true
}

func entityCanonicalUUID(value any) (uuid.UUID, bool) {
	text, ok := value.(string)
	if !ok {
		return uuid.Nil, false
	}
	parsed, err := uuid.Parse(text)
	return parsed, err == nil && parsed.String() == text
}

func entityNullableUUID(value any) (*uuid.UUID, bool) {
	if value == nil {
		return nil, true
	}
	parsed, ok := entityCanonicalUUID(value)
	if !ok {
		return nil, false
	}
	return &parsed, true
}

func entityAdmittedActor(value any, importContext sourceport.ImportContext) (uuid.UUID, bool) {
	actorID, ok := entityCanonicalUUID(value)
	if !ok {
		return uuid.Nil, false
	}
	_, admitted := importContext.Actors.Lookup(actorID.String())
	return actorID, admitted
}

func entityNullableAdmittedActor(value any, importContext sourceport.ImportContext) (*uuid.UUID, bool) {
	if value == nil {
		return nil, true
	}
	actor, ok := entityAdmittedActor(value, importContext)
	if !ok {
		return nil, false
	}
	return &actor, true
}

func entityPortableText(value any, allowEmpty bool) (string, bool) {
	text, ok := value.(string)
	if !ok || strings.ContainsRune(text, '\x00') || (!allowEmpty && text == "") {
		return "", false
	}
	return text, true
}

func entityNullableText(value any, allowEmpty bool) (*string, bool) {
	if value == nil {
		return nil, true
	}
	text, ok := entityPortableText(value, allowEmpty)
	if !ok {
		return nil, false
	}
	return &text, true
}

func entityCanonicalInteger(value any) (int64, bool) {
	number, ok := value.(json.Number)
	if !ok || !entityPortableIntegerPattern.MatchString(number.String()) {
		return 0, false
	}
	parsed, err := strconv.ParseInt(number.String(), 10, 64)
	return parsed, err == nil
}

func entityCanonicalTimestamp(value any) (time.Time, bool) {
	text, ok := value.(string)
	if !ok || !entityPortableTimestampPattern.MatchString(text) {
		return time.Time{}, false
	}
	parsed, err := time.Parse(time.RFC3339Nano, text)
	if err != nil || entityFormatTimestamp(parsed) != text {
		return time.Time{}, false
	}
	return parsed.UTC(), true
}

func entityNullableTimestamp(value any) (*time.Time, bool) {
	if value == nil {
		return nil, true
	}
	parsed, ok := entityCanonicalTimestamp(value)
	if !ok {
		return nil, false
	}
	return &parsed, true
}

func entityFormatTimestamp(value time.Time) string {
	return value.UTC().Format("2006-01-02T15:04:05.999999+00:00")
}

func entityUUIDPtrEqual(left, right *uuid.UUID) bool {
	return (left == nil && right == nil) || (left != nil && right != nil && *left == *right)
}

func entityStringPtrEqual(left, right *string) bool {
	return (left == nil && right == nil) || (left != nil && right != nil && *left == *right)
}

func entityTimePtrEqual(left, right *time.Time) bool {
	return (left == nil && right == nil) || (left != nil && right != nil && left.Equal(*right))
}

func entityRuntimeActorPointer(portable *uuid.UUID, runtime uuid.UUID) *uuid.UUID {
	if portable == nil {
		return nil
	}
	value := runtime
	return &value
}
