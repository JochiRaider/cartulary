package graphprojection

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"sort"
	"strings"
	"time"

	"github.com/google/uuid"

	contractrecovery "github.com/JochiRaider/cartulary/internal/gen/contractrecovery"
)

const (
	RestoreAlgorithmID                         = "graphprojection.restore_rebuild.v1"
	RestoreSourceRegistrySchemaID              = "cartulary.graph_projection_restore_source_registry.v1"
	RestoreImplementationBindingSchemaID       = "cartulary.graph_projection_restore_implementation_binding.v1"
	RestoreRebuildResultSchemaID               = "cartulary.graph_projection_restore_rebuild_result.v1"
	RestoreMaximumSourceRegistrations          = 128
	RestoreMaximumCandidates                   = 1024
	RestoreMaximumNormalizedInputBytes   int64 = 268435456
	RestoreMaximumVertices                     = 500000
	RestoreMaximumEdges                        = 1000000
)

var restoreGraphTableIDs = []string{
	"graph_projection_edges",
	"graph_projection_idempotency",
	"graph_projection_runs",
	"graph_projection_vertices",
	"graph_projection_views",
}

func RestoreGraphTableIDs() []string {
	return append([]string(nil), restoreGraphTableIDs...)
}

type RestoreErrorCode string

const (
	RestoreErrorInvalidRequest          RestoreErrorCode = "invalid_restore_request"
	RestoreErrorRecoveryCatalogMismatch RestoreErrorCode = "recovery_catalog_mismatch"
	RestoreErrorSourceRegistryMismatch  RestoreErrorCode = "source_registry_mismatch"
	RestoreErrorBindingUnavailable      RestoreErrorCode = "implementation_binding_unavailable"
	RestoreErrorSourceEnumeration       RestoreErrorCode = "source_enumeration_failed"
	RestoreErrorInvalidCandidate        RestoreErrorCode = "invalid_restore_candidate"
	RestoreErrorResourceOverflow        RestoreErrorCode = "restore_resource_overflow"
	RestoreErrorPublicationFailed       RestoreErrorCode = "restore_publication_failed"
	RestoreErrorPostconditionFailed     RestoreErrorCode = "restore_postcondition_failed"
	RestoreErrorOutcomeIndeterminate    RestoreErrorCode = "restore_outcome_indeterminate"
)

type RestoreError struct {
	Code RestoreErrorCode
}

func (err *RestoreError) Error() string {
	if err == nil {
		return "graphprojection restore failed"
	}
	return "graphprojection restore failed: " + string(err.Code)
}

func NewRestoreError(code RestoreErrorCode) error {
	if !validRestoreErrorCode(code) {
		code = RestoreErrorInvalidRequest
	}
	return &RestoreError{Code: code}
}

func RestoreErrorCodeOf(err error) (RestoreErrorCode, bool) {
	var classified *RestoreError
	if !errors.As(err, &classified) || classified == nil || !validRestoreErrorCode(classified.Code) {
		return "", false
	}
	return classified.Code, true
}

func validRestoreErrorCode(code RestoreErrorCode) bool {
	switch code {
	case RestoreErrorInvalidRequest,
		RestoreErrorRecoveryCatalogMismatch,
		RestoreErrorSourceRegistryMismatch,
		RestoreErrorBindingUnavailable,
		RestoreErrorSourceEnumeration,
		RestoreErrorInvalidCandidate,
		RestoreErrorResourceOverflow,
		RestoreErrorPublicationFailed,
		RestoreErrorPostconditionFailed,
		RestoreErrorOutcomeIndeterminate:
		return true
	default:
		return false
	}
}

type RestoreSourceRegistryEntry struct {
	SourceRegistrationID      string     `json:"source_registration_id"`
	SourceOwnerID             string     `json:"source_owner_id"`
	EnumeratorBindingID       string     `json:"enumerator_binding_id"`
	ValidityBindingID         string     `json:"validity_binding_id"`
	ProjectionInputContractID string     `json:"projection_input_contract_id"`
	ImplementationBindingID   string     `json:"implementation_binding_id"`
	Status                    string     `json:"status"`
	IntroducedAt              time.Time  `json:"introduced_at"`
	RetiredAt                 *time.Time `json:"retired_at"`
}

type RestoreSourceRegistryDocument struct {
	SchemaID string                       `json:"schema_id"`
	Entries  []RestoreSourceRegistryEntry `json:"entries"`
}

type RestoreSourceState interface {
	GraphProjectionRestoreSourceState()
}

type RestoreCandidate struct {
	CandidateID          string
	GraphViewID          string
	NormalizedInput      []byte
	DeclarationExpiresAt *time.Time
}

type RestoreCandidateValidity struct {
	Eligible   bool
	SkipReason RestoreSkipReason
}

type RestoreCandidateEnumerator func(context.Context, RestoreSourceState, time.Time) ([]RestoreCandidate, error)
type RestoreCandidateValidityEvaluator func(context.Context, RestoreCandidate, time.Time) (RestoreCandidateValidity, error)

type RestoreSourceRegistration struct {
	Entry            RestoreSourceRegistryEntry
	Enumerate        RestoreCandidateEnumerator
	EvaluateValidity RestoreCandidateValidityEvaluator
}

type RestoreSourceRegistry struct {
	document      RestoreSourceRegistryDocument
	registrations []RestoreSourceRegistration
	digestSHA256  string
}

func NewRestoreSourceRegistry(registrations ...RestoreSourceRegistration) (*RestoreSourceRegistry, error) {
	if len(registrations) > RestoreMaximumSourceRegistrations {
		return nil, NewRestoreError(RestoreErrorResourceOverflow)
	}
	document := RestoreSourceRegistryDocument{
		SchemaID: RestoreSourceRegistrySchemaID,
		Entries:  make([]RestoreSourceRegistryEntry, len(registrations)),
	}
	copyRegistrations := append([]RestoreSourceRegistration(nil), registrations...)
	for index, registration := range copyRegistrations {
		entry := registration.Entry
		if strings.TrimSpace(entry.SourceRegistrationID) == "" ||
			strings.TrimSpace(entry.SourceOwnerID) == "" ||
			strings.TrimSpace(entry.EnumeratorBindingID) == "" ||
			strings.TrimSpace(entry.ValidityBindingID) == "" ||
			strings.TrimSpace(entry.ProjectionInputContractID) == "" ||
			strings.TrimSpace(entry.ImplementationBindingID) == "" ||
			(entry.Status != "active" && entry.Status != "retired") ||
			entry.IntroducedAt.IsZero() || registration.Enumerate == nil || registration.EvaluateValidity == nil {
			return nil, NewRestoreError(RestoreErrorInvalidRequest)
		}
		entry.IntroducedAt = entry.IntroducedAt.UTC()
		if entry.RetiredAt != nil {
			retiredAt := entry.RetiredAt.UTC()
			entry.RetiredAt = &retiredAt
		}
		if index > 0 && copyRegistrations[index-1].Entry.SourceRegistrationID >= entry.SourceRegistrationID {
			return nil, NewRestoreError(RestoreErrorInvalidRequest)
		}
		copyRegistrations[index].Entry = entry
		document.Entries[index] = entry
	}
	body, err := json.Marshal(map[string]any{
		"entries":   document.Entries,
		"schema_id": document.SchemaID,
	})
	if err != nil {
		return nil, NewRestoreError(RestoreErrorInvalidRequest)
	}
	digest := sha256.Sum256(body)
	return &RestoreSourceRegistry{
		document: document, registrations: copyRegistrations, digestSHA256: hex.EncodeToString(digest[:]),
	}, nil
}

func CurrentRestoreSourceRegistry() *RestoreSourceRegistry {
	return &RestoreSourceRegistry{
		document:      RestoreSourceRegistryDocument{SchemaID: RestoreSourceRegistrySchemaID, Entries: []RestoreSourceRegistryEntry{}},
		registrations: []RestoreSourceRegistration{},
		digestSHA256:  contractrecovery.CurrentGraphProjectionRestoreSourceRegistrySHA256,
	}
}

func (registry *RestoreSourceRegistry) Document() RestoreSourceRegistryDocument {
	if registry == nil {
		return RestoreSourceRegistryDocument{}
	}
	document := registry.document
	document.Entries = append([]RestoreSourceRegistryEntry(nil), document.Entries...)
	return document
}

func (registry *RestoreSourceRegistry) Registrations() []RestoreSourceRegistration {
	if registry == nil {
		return nil
	}
	return append([]RestoreSourceRegistration(nil), registry.registrations...)
}

func (registry *RestoreSourceRegistry) DigestSHA256() string {
	if registry == nil {
		return ""
	}
	return registry.digestSHA256
}

type RestoreImplementationBinding struct {
	SchemaID                    string   `json:"schema_id"`
	AlgorithmID                 string   `json:"algorithm_id"`
	BindingID                   string   `json:"binding_id"`
	GraphProjectionContractID   string   `json:"graph_projection_contract_id"`
	RecoveryStateCatalogSHA256  string   `json:"recovery_state_catalog_sha256"`
	SourceRegistrySHA256        string   `json:"source_registry_sha256"`
	GraphTableIDs               []string `json:"graph_table_ids"`
	GraphEngineAlgorithmIDs     []string `json:"graph_engine_algorithm_ids"`
	GraphEngineAlgorithmDigests []string `json:"graph_engine_algorithm_digests"`
	DatabaseSchemaLineage       string   `json:"database_schema_lineage"`
	DatabaseSchemaHead          int64    `json:"database_schema_head"`
	PackagedSubjectSHA256       string   `json:"packaged_subject_sha256"`
	BuildProvenanceSHA256       string   `json:"build_provenance_sha256"`
}

type RestoreImplementationBindingRef struct {
	Binding RestoreImplementationBinding
	SHA256  string
	Legacy  bool
}

func CurrentRestoreImplementationBinding() RestoreImplementationBindingRef {
	return decodePackagedRestoreBinding(
		contractrecovery.CurrentGraphProjectionRestoreImplementationBindingJSON,
		contractrecovery.CurrentGraphProjectionRestoreImplementationBindingSHA256,
		false,
	)
}

func LegacyEmptyRegistryRestoreImplementationBinding() RestoreImplementationBindingRef {
	return decodePackagedRestoreBinding(
		contractrecovery.LegacyGraphProjectionRestoreImplementationBindingJSON,
		contractrecovery.LegacyGraphProjectionRestoreImplementationBindingSHA256,
		true,
	)
}

func decodePackagedRestoreBinding(body string, digest string, legacy bool) RestoreImplementationBindingRef {
	var binding RestoreImplementationBinding
	if err := strictRestoreJSON([]byte(body), &binding); err != nil {
		return RestoreImplementationBindingRef{}
	}
	return RestoreImplementationBindingRef{Binding: binding, SHA256: digest, Legacy: legacy}
}

type RestoreRecoveryCatalogRef struct {
	DigestSHA256  string
	AlgorithmID   string
	GraphTableIDs []string
}

type RestoreSourceRegistryRef struct {
	Registry *RestoreSourceRegistry
	SHA256   string
}

type RestoreRebuildRequest struct {
	Context               context.Context
	RestoreOperationID    uuid.UUID
	RestoredSourceState   RestoreSourceState
	BackupSetID           uuid.UUID
	ConsistencyPointAt    time.Time
	TargetGenerationID    uuid.UUID
	RecoveryStateCatalog  RestoreRecoveryCatalogRef
	SourceRegistry        RestoreSourceRegistryRef
	ImplementationBinding RestoreImplementationBindingRef
}

type RestoreStatus string
type RestoreReadinessOutcome string

const (
	RestoreStatusSucceeded     RestoreStatus           = "succeeded"
	RestoreStatusFailed        RestoreStatus           = "failed"
	RestoreReadinessReady      RestoreReadinessOutcome = "ready"
	RestoreReadinessIncomplete RestoreReadinessOutcome = "incomplete"
)

type RestoreRebuiltView struct {
	SourceRegistrationID          string `json:"source_registration_id"`
	CandidateID                   string `json:"candidate_id"`
	GraphViewID                   string `json:"graph_view_id"`
	ProjectionRunID               string `json:"projection_run_id"`
	SourceSnapshotID              string `json:"source_snapshot_id"`
	ProjectionVersion             string `json:"projection_version"`
	NormalizedConfigurationSHA256 string `json:"normalized_configuration_sha256"`
	NormalizedSourceSHA256        string `json:"normalized_source_sha256"`
	VertexCount                   int    `json:"vertex_count"`
	EdgeCount                     int    `json:"edge_count"`
	CanonicalOutputSHA256         string `json:"canonical_output_sha256"`
}

type RestoreSkipReason string

const (
	RestoreSkipDeclarationExpired       RestoreSkipReason = "declaration_expired_at_consistency_point"
	RestoreSkipRegistrationNotYetActive RestoreSkipReason = "registration_not_yet_active"
	RestoreSkipRegistrationRetired      RestoreSkipReason = "registration_retired_before_consistency_point"
)

type RestoreSkippedCandidate struct {
	SourceRegistrationID string            `json:"source_registration_id"`
	CandidateID          string            `json:"candidate_id"`
	Reason               RestoreSkipReason `json:"reason"`
}

type RestoreSafeMessage struct {
	Code RestoreErrorCode `json:"code"`
}

type RestoreRebuildResult struct {
	SchemaID                    string                    `json:"schema_id"`
	RestoreOperationID          string                    `json:"restore_operation_id"`
	TargetGenerationID          string                    `json:"target_generation_id"`
	Status                      RestoreStatus             `json:"status"`
	ReadinessOutcome            RestoreReadinessOutcome   `json:"readiness_outcome"`
	AlgorithmID                 string                    `json:"algorithm_id"`
	ImplementationBindingSHA256 string                    `json:"implementation_binding_sha256"`
	SourceRegistrySHA256        string                    `json:"source_registry_sha256"`
	ClearedTableIDs             []string                  `json:"cleared_table_ids"`
	RebuiltViews                []RestoreRebuiltView      `json:"rebuilt_views"`
	SkippedCandidates           []RestoreSkippedCandidate `json:"skipped_candidates"`
	PostconditionSHA256         *string                   `json:"postcondition_sha256"`
	Warnings                    []RestoreSafeMessage      `json:"warnings"`
	Errors                      []RestoreSafeMessage      `json:"errors"`
}

func (result RestoreRebuildResult) ReadinessSatisfied() bool {
	return result.Validate() == nil && result.Status == RestoreStatusSucceeded
}

func (result RestoreRebuildResult) Validate() error {
	if result.SchemaID != RestoreRebuildResultSchemaID || result.AlgorithmID != RestoreAlgorithmID ||
		strings.TrimSpace(result.RestoreOperationID) == "" || uuid.Validate(result.TargetGenerationID) != nil ||
		!validSHA256String(result.ImplementationBindingSHA256) || !validSHA256String(result.SourceRegistrySHA256) ||
		result.ClearedTableIDs == nil || result.RebuiltViews == nil || result.SkippedCandidates == nil ||
		result.Warnings == nil || result.Errors == nil {
		return NewRestoreError(RestoreErrorPostconditionFailed)
	}
	switch {
	case result.Status == RestoreStatusSucceeded && result.ReadinessOutcome == RestoreReadinessReady:
		if !equalStrings(result.ClearedTableIDs, restoreGraphTableIDs) || result.PostconditionSHA256 == nil ||
			!validSHA256String(*result.PostconditionSHA256) || len(result.Errors) != 0 {
			return NewRestoreError(RestoreErrorPostconditionFailed)
		}
	case result.Status == RestoreStatusFailed && result.ReadinessOutcome == RestoreReadinessIncomplete:
		if len(result.ClearedTableIDs) != 0 || len(result.RebuiltViews) != 0 || len(result.SkippedCandidates) != 0 ||
			result.PostconditionSHA256 != nil || len(result.Errors) == 0 {
			return NewRestoreError(RestoreErrorPostconditionFailed)
		}
	default:
		return NewRestoreError(RestoreErrorPostconditionFailed)
	}
	for _, message := range append(append([]RestoreSafeMessage{}, result.Warnings...), result.Errors...) {
		if !validRestoreErrorCode(message.Code) {
			return NewRestoreError(RestoreErrorPostconditionFailed)
		}
	}
	if !sort.SliceIsSorted(result.RebuiltViews, func(left, right int) bool {
		if result.RebuiltViews[left].SourceRegistrationID != result.RebuiltViews[right].SourceRegistrationID {
			return result.RebuiltViews[left].SourceRegistrationID < result.RebuiltViews[right].SourceRegistrationID
		}
		return result.RebuiltViews[left].CandidateID < result.RebuiltViews[right].CandidateID
	}) || !sort.SliceIsSorted(result.SkippedCandidates, func(left, right int) bool {
		if result.SkippedCandidates[left].SourceRegistrationID != result.SkippedCandidates[right].SourceRegistrationID {
			return result.SkippedCandidates[left].SourceRegistrationID < result.SkippedCandidates[right].SourceRegistrationID
		}
		return result.SkippedCandidates[left].CandidateID < result.SkippedCandidates[right].CandidateID
	}) {
		return NewRestoreError(RestoreErrorPostconditionFailed)
	}
	return nil
}

type RestoreParticipant interface {
	Rebuild(context.Context, RestoreRebuildRequest) (RestoreRebuildResult, error)
}

func strictRestoreJSON(data []byte, destination any) error {
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(destination); err != nil {
		return err
	}
	if err := decoder.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		return fmt.Errorf("trailing JSON value")
	}
	return nil
}

func validSHA256String(value string) bool {
	if len(value) != sha256.Size*2 {
		return false
	}
	decoded, err := hex.DecodeString(value)
	return err == nil && hex.EncodeToString(decoded) == value
}

func equalStrings(left, right []string) bool {
	if len(left) != len(right) {
		return false
	}
	for index := range left {
		if left[index] != right[index] {
			return false
		}
	}
	return true
}
