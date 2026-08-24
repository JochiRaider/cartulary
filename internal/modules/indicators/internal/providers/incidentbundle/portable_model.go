package incidentbundle

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
	indicatororigin "github.com/JochiRaider/cartulary/internal/modules/indicators/internal/origin"
)

const (
	indicatorsBundlePath   = "data/indicators.ndjson"
	observationsBundlePath = "data/indicator_observations.ndjson"
	intervalsBundlePath    = "data/indicator_state_intervals.ndjson"

	representationInvariant      = "indicators.representation_legal"
	normalizationInvariant       = "indicators.normalization_exact"
	identityUniqueInvariant      = "indicators.identity_unique"
	observationIncidentInvariant = "indicators.observation_same_incident"
	observationOrderedInvariant  = "indicators.observation_ordered"
	observationCoherentInvariant = "indicators.observation_coherent"
	intervalIncidentInvariant    = "indicators.interval_same_incident"
	intervalOrderedInvariant     = "indicators.interval_ordered"
	intervalCoherentInvariant    = "indicators.interval_coherent"
	repeatedObservationInvariant = "indicators.repeated_observations_preserved"
)

var (
	portableIntegerPattern   = regexp.MustCompile(`^(?:0|-?[1-9][0-9]*)$`)
	portableTimestampPattern = regexp.MustCompile(`^[0-9]{4}-[0-9]{2}-[0-9]{2}T[0-9]{2}:[0-9]{2}:[0-9]{2}(?:\.[0-9]{1,6})?\+00:00$`)
	portableDedupePattern    = regexp.MustCompile(`^[0-9a-f]{64}$`)
)

type portableBinding struct {
	operationID   string
	incidentID    uuid.UUID
	bundleVersion int
	contractMajor int
}

type preparedIndicatorImport struct {
	binding      portableBinding
	indicators   []portableIndicatorRow
	observations []portableObservationRow
	intervals    []portableIntervalRow
}

type portableIndicatorRow struct {
	RecordID            uuid.UUID
	IncidentID          uuid.UUID
	IndicatorType       string
	ValueKind           string
	DisplayValue        string
	NormalizedValue     *string
	DedupeKey           string
	DefangedValue       *string
	HashAlgorithm       *string
	HashValue           *string
	STIXPattern         *string
	RowVersion          int64
	CreatedAt           time.Time
	UpdatedAt           time.Time
	PortableCreatedByID uuid.UUID
	PortableUpdatedByID uuid.UUID
	RuntimeCreatedByID  uuid.UUID
	RuntimeUpdatedByID  uuid.UUID
	DeletedAt           *time.Time
	PortableDeletedByID *uuid.UUID
	RuntimeDeletedByID  *uuid.UUID
}

type portableObservationRow struct {
	ObservationID        uuid.UUID
	IncidentID           uuid.UUID
	SourceRecordID       uuid.UUID
	SourceFieldKey       string
	OriginKind           indicatororigin.ObservationOrigin
	OriginLocator        string
	ObservedText         string
	ParsedIndicatorType  *string
	NormalizedCandidate  *string
	ResolutionStatus     string
	ResolvedIndicatorID  *uuid.UUID
	RowVersion           int64
	PortableCreatedByID  uuid.UUID
	RuntimeCreatedByID   uuid.UUID
	CreatedAt            time.Time
	PortableResolvedByID *uuid.UUID
	RuntimeResolvedByID  *uuid.UUID
	ResolvedAt           *time.Time
	ResolutionMethod     *string
	DeletedAt            *time.Time
	PortableDeletedByID  *uuid.UUID
	RuntimeDeletedByID   *uuid.UUID
}

type portableIntervalRow struct {
	IntervalID          uuid.UUID
	IncidentID          uuid.UUID
	IndicatorRecordID   uuid.UUID
	LifecycleState      string
	ValidFrom           time.Time
	ValidTo             *time.Time
	Confidence          *int
	Rationale           *string
	SupportRefs         []uuid.UUID
	Assessor            *string
	AssessedAt          time.Time
	RowVersion          int64
	PortableCreatedByID uuid.UUID
	RuntimeCreatedByID  uuid.UUID
	CreatedAt           time.Time
	DeletedAt           *time.Time
	PortableDeletedByID *uuid.UUID
	RuntimeDeletedByID  *uuid.UUID
}

type indicatorFailureCandidate struct {
	invariant string
	path      string
	identity  string
	digest    string
}

func (binding portableBinding) matches(importContext sourceport.ImportContext) bool {
	return binding.operationID != "" &&
		binding.operationID == importContext.OperationID &&
		binding.incidentID != uuid.Nil &&
		binding.incidentID == importContext.IncidentID &&
		binding.bundleVersion == importContext.BundleVersion &&
		binding.bundleVersion == 3 &&
		binding.contractMajor == sourceport.ContractMajor
}

func indicatorFailure(invariant string, path string, identity string, value any) indicatorFailureCandidate {
	encoded, _ := json.Marshal(value)
	sum := sha256.Sum256(encoded)
	return indicatorFailureCandidate{
		invariant: invariant,
		path:      path,
		identity:  identity,
		digest:    hex.EncodeToString(sum[:]),
	}
}

func selectedIndicatorFailure(candidates []indicatorFailureCandidate) error {
	if len(candidates) == 0 {
		return nil
	}
	sort.Slice(candidates, func(i, j int) bool {
		left, right := candidates[i], candidates[j]
		if indicatorInvariantRank(left.invariant) != indicatorInvariantRank(right.invariant) {
			return indicatorInvariantRank(left.invariant) < indicatorInvariantRank(right.invariant)
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
	return indicatorSourceFailure(candidates[0].invariant)
}

func indicatorInvariantRank(invariant string) int {
	switch invariant {
	case representationInvariant:
		return 1
	case normalizationInvariant:
		return 2
	case identityUniqueInvariant:
		return 3
	case observationIncidentInvariant:
		return 4
	case observationOrderedInvariant:
		return 5
	case observationCoherentInvariant:
		return 6
	case intervalIncidentInvariant:
		return 7
	case intervalOrderedInvariant:
		return 8
	case intervalCoherentInvariant:
		return 9
	case repeatedObservationInvariant:
		return 10
	default:
		return 0
	}
}

func exactPortableMembers(row map[string]any, members []string) bool {
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

func portableStableIdentity(value any) string {
	parsed, ok := canonicalPortableUUID(value)
	if !ok {
		return ""
	}
	return parsed.String()
}

func canonicalPortableUUID(value any) (uuid.UUID, bool) {
	text, ok := value.(string)
	if !ok {
		return uuid.Nil, false
	}
	parsed, err := uuid.Parse(text)
	return parsed, err == nil && parsed.String() == text
}

func nullablePortableUUID(value any) (*uuid.UUID, bool) {
	if value == nil {
		return nil, true
	}
	parsed, ok := canonicalPortableUUID(value)
	if !ok {
		return nil, false
	}
	return &parsed, true
}

func admittedPortableActor(value any, importContext sourceport.ImportContext) (uuid.UUID, bool) {
	actorID, ok := canonicalPortableUUID(value)
	if !ok {
		return uuid.Nil, false
	}
	_, admitted := importContext.Actors.Lookup(actorID.String())
	return actorID, admitted
}

func nullableAdmittedPortableActor(value any, importContext sourceport.ImportContext) (*uuid.UUID, bool) {
	if value == nil {
		return nil, true
	}
	actorID, ok := admittedPortableActor(value, importContext)
	if !ok {
		return nil, false
	}
	return &actorID, true
}

func portableText(value any, allowEmpty bool) (string, bool) {
	text, ok := value.(string)
	if !ok || strings.ContainsRune(text, '\x00') || (!allowEmpty && text == "") {
		return "", false
	}
	return text, true
}

func nullablePortableText(value any, allowEmpty bool) (*string, bool) {
	if value == nil {
		return nil, true
	}
	text, ok := portableText(value, allowEmpty)
	if !ok {
		return nil, false
	}
	return &text, true
}

func canonicalPortableInteger(value any) (int64, bool) {
	number, ok := value.(json.Number)
	if !ok || !portableIntegerPattern.MatchString(number.String()) {
		return 0, false
	}
	parsed, err := strconv.ParseInt(number.String(), 10, 64)
	return parsed, err == nil
}

func canonicalPortableTimestamp(value any) (time.Time, bool) {
	text, ok := value.(string)
	if !ok || !portableTimestampPattern.MatchString(text) {
		return time.Time{}, false
	}
	parsed, err := time.Parse(time.RFC3339Nano, text)
	if err != nil || formatPortableTimestamp(parsed) != text {
		return time.Time{}, false
	}
	return parsed.UTC(), true
}

func nullablePortableTimestamp(value any) (*time.Time, bool) {
	if value == nil {
		return nil, true
	}
	parsed, ok := canonicalPortableTimestamp(value)
	if !ok {
		return nil, false
	}
	return &parsed, true
}

func formatPortableTimestamp(value time.Time) string {
	return value.UTC().Format("2006-01-02T15:04:05.999999+00:00")
}

func portableStringPointersEqual(left *string, right *string) bool {
	if left == nil || right == nil {
		return left == right
	}
	return *left == *right
}

func portableUUIDPointersEqual(left *uuid.UUID, right *uuid.UUID) bool {
	if left == nil || right == nil {
		return left == right
	}
	return *left == *right
}

func portableTimePointersEqual(left *time.Time, right *time.Time) bool {
	if left == nil || right == nil {
		return left == right
	}
	return left.Equal(*right)
}

func runtimeActorPointer(portable *uuid.UUID, runtimeActor uuid.UUID) *uuid.UUID {
	if portable == nil {
		return nil
	}
	value := runtimeActor
	return &value
}

func uuidStrings(values []uuid.UUID) []string {
	result := make([]string, 0, len(values))
	for _, value := range values {
		result = append(result, value.String())
	}
	return result
}
