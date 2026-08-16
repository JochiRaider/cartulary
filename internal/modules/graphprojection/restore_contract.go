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
	RestoreAlgorithmID                         = "graphprojection.restore_rebuild.v2"
	RestoreSourceRegistrySchemaID              = "cartulary.graph_projection_restore_source_registry.v2"
	RestoreImplementationBindingSchemaID       = "cartulary.graph_projection_restore_implementation_binding.v2"
	RestoreRebuildResultSchemaID               = "cartulary.graph_projection_restore_rebuild_result.v2"
	RestoreMaximumSourceRegistrations          = 128
	RestoreMaximumCandidates                   = 128
	RestoreMaximumNormalizedInputBytes   int64 = 268435456
	RestoreMaximumVertices                     = MaximumResultVerticesV2
	RestoreMaximumEdges                        = MaximumResultEdgesV2
)

var restoreGraphTableIDs = []string{
	"graph_projection_result_edges",
	"graph_projection_result_leases",
	"graph_projection_result_vertices",
	"graph_projection_results",
}

func RestoreGraphTableIDs() []string { return append([]string(nil), restoreGraphTableIDs...) }

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

type RestoreError struct{ Code RestoreErrorCode }

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
	case RestoreErrorInvalidRequest, RestoreErrorRecoveryCatalogMismatch, RestoreErrorSourceRegistryMismatch,
		RestoreErrorBindingUnavailable, RestoreErrorSourceEnumeration, RestoreErrorInvalidCandidate,
		RestoreErrorResourceOverflow, RestoreErrorPublicationFailed, RestoreErrorPostconditionFailed,
		RestoreErrorOutcomeIndeterminate:
		return true
	default:
		return false
	}
}

// RestoreSourceRegistryEntry is the adopted v2 source-owner registration.
type RestoreSourceRegistryEntry struct {
	SourceRegistrationID       string `json:"source_registration_id"`
	SourceOwnerID              string `json:"source_owner_id"`
	AuthoritativeFamilyID      string `json:"authoritative_family_id"`
	EnumeratorBindingID        string `json:"enumerator_binding_id"`
	ValidityBindingID          string `json:"validity_binding_id"`
	ProjectionInputContractID  string `json:"projection_input_contract_id"`
	ProjectionResultContractID string `json:"projection_result_contract_id"`
	Status                     string `json:"status"`
}

type RestoreSourceRegistryDocument struct {
	SchemaID string                       `json:"schema_id"`
	Entries  []RestoreSourceRegistryEntry `json:"entries"`
}

type RestoreSourceState interface{ GraphProjectionRestoreSourceState() }

// RestoreCandidate is a complete immutable rebuild request from a source
// owner. ExpectedBinding is the selected result recorded by authoritative
// state and is proved byte-for-byte after deterministic recomputation.
type RestoreCandidate struct {
	CandidateID     string
	GraphViewID     string
	SemanticInput   []byte
	ExpectedBinding ResultBindingV2
}

type RestoreCandidateValidity struct {
	Eligible bool
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
	if len(registrations) == 0 || len(registrations) > RestoreMaximumSourceRegistrations {
		return nil, NewRestoreError(RestoreErrorInvalidRequest)
	}
	document := RestoreSourceRegistryDocument{SchemaID: RestoreSourceRegistrySchemaID, Entries: make([]RestoreSourceRegistryEntry, len(registrations))}
	copyRegistrations := append([]RestoreSourceRegistration(nil), registrations...)
	for index := range copyRegistrations {
		registration := copyRegistrations[index]
		entry := registration.Entry
		if !validRestoreRegistryEntry(entry) || registration.Enumerate == nil ||
			(index > 0 && copyRegistrations[index-1].Entry.SourceRegistrationID >= entry.SourceRegistrationID) {
			return nil, NewRestoreError(RestoreErrorInvalidRequest)
		}
		document.Entries[index] = entry
	}
	body, err := restoreCanonicalJSON(document)
	if err != nil {
		return nil, NewRestoreError(RestoreErrorInvalidRequest)
	}
	digest := sha256.Sum256(body)
	return &RestoreSourceRegistry{document: document, registrations: copyRegistrations, digestSHA256: hex.EncodeToString(digest[:])}, nil
}

func NewCurrentRestoreSourceRegistry(registrations ...RestoreSourceRegistration) (*RestoreSourceRegistry, error) {
	registry, err := NewRestoreSourceRegistry(registrations...)
	if err != nil || registry.DigestSHA256() != contractrecovery.CurrentGraphProjectionRestoreSourceRegistrySHA256 ||
		string(mustRestoreCanonicalJSON(registry.document)) != contractrecovery.CurrentGraphProjectionRestoreSourceRegistryJSON {
		return nil, NewRestoreError(RestoreErrorSourceRegistryMismatch)
	}
	return registry, nil
}

func CurrentRestoreSourceRegistry() *RestoreSourceRegistry {
	return decodePackagedRestoreRegistry(contractrecovery.CurrentGraphProjectionRestoreSourceRegistryJSON,
		contractrecovery.CurrentGraphProjectionRestoreSourceRegistrySHA256)
}

func decodePackagedRestoreRegistry(body, digest string) *RestoreSourceRegistry {
	var document RestoreSourceRegistryDocument
	if err := strictRestoreJSON([]byte(body), &document); err != nil {
		return nil
	}
	return &RestoreSourceRegistry{document: document, registrations: []RestoreSourceRegistration{}, digestSHA256: digest}
}

func validRestoreRegistryEntry(entry RestoreSourceRegistryEntry) bool {
	return strings.TrimSpace(entry.SourceRegistrationID) != "" && strings.TrimSpace(entry.SourceOwnerID) != "" &&
		strings.TrimSpace(entry.AuthoritativeFamilyID) != "" && strings.TrimSpace(entry.EnumeratorBindingID) != "" &&
		strings.TrimSpace(entry.ValidityBindingID) != "" && entry.ProjectionInputContractID == ProjectionSchemaIDV2 &&
		entry.ProjectionResultContractID == "graph_projection_result.v2" && entry.Status == "active"
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
}

func CurrentRestoreImplementationBinding() RestoreImplementationBindingRef {
	return decodePackagedRestoreBinding(contractrecovery.CurrentGraphProjectionRestoreImplementationBindingJSON,
		contractrecovery.CurrentGraphProjectionRestoreImplementationBindingSHA256)
}

func decodePackagedRestoreBinding(body, digest string) RestoreImplementationBindingRef {
	var binding RestoreImplementationBinding
	if err := strictRestoreJSON([]byte(body), &binding); err != nil {
		return RestoreImplementationBindingRef{}
	}
	return RestoreImplementationBindingRef{Binding: binding, SHA256: digest}
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
	ProjectionResultID            string `json:"projection_result_id"`
	SourceSnapshotID              string `json:"source_snapshot_id"`
	ProjectionVersion             string `json:"projection_version"`
	NormalizedConfigurationSHA256 string `json:"normalized_configuration_sha256"`
	NormalizedSourceSHA256        string `json:"normalized_source_sha256"`
	VertexCount                   int    `json:"vertex_count"`
	EdgeCount                     int    `json:"edge_count"`
	CanonicalOutputSHA256         string `json:"canonical_output_sha256"`
}

type RestoreSafeMessage struct {
	Code RestoreErrorCode `json:"code"`
}

type RestoreRebuildResult struct {
	SchemaID                      string                  `json:"schema_id"`
	RestoreOperationID            string                  `json:"restore_operation_id"`
	TargetGenerationID            string                  `json:"target_generation_id"`
	Status                        RestoreStatus           `json:"status"`
	ReadinessOutcome              RestoreReadinessOutcome `json:"readiness_outcome"`
	AlgorithmID                   string                  `json:"algorithm_id"`
	ImplementationBindingSHA256   string                  `json:"implementation_binding_sha256"`
	SourceRegistrySHA256          string                  `json:"source_registry_sha256"`
	ClearedTableIDs               []string                `json:"cleared_table_ids"`
	RebuiltViews                  []RestoreRebuiltView    `json:"rebuilt_views"`
	ReconciledNonterminalJobCount int                     `json:"reconciled_nonterminal_job_count"`
	ReconciledLeaseCount          int                     `json:"reconciled_lease_count"`
	PostconditionSHA256           *string                 `json:"postcondition_sha256"`
	Warnings                      []RestoreSafeMessage    `json:"warnings"`
	Errors                        []RestoreSafeMessage    `json:"errors"`
}

func (result RestoreRebuildResult) ReadinessSatisfied() bool {
	return result.Validate() == nil && result.Status == RestoreStatusSucceeded
}

func (result RestoreRebuildResult) Validate() error {
	if result.SchemaID != RestoreRebuildResultSchemaID || result.AlgorithmID != RestoreAlgorithmID ||
		strings.TrimSpace(result.RestoreOperationID) == "" || uuid.Validate(result.TargetGenerationID) != nil ||
		!validSHA256String(result.ImplementationBindingSHA256) || !validSHA256String(result.SourceRegistrySHA256) ||
		result.ClearedTableIDs == nil || result.RebuiltViews == nil || result.Warnings == nil || result.Errors == nil ||
		result.ReconciledNonterminalJobCount < 0 || result.ReconciledLeaseCount < 0 {
		return NewRestoreError(RestoreErrorPostconditionFailed)
	}
	switch {
	case result.Status == RestoreStatusSucceeded && result.ReadinessOutcome == RestoreReadinessReady:
		if !equalStrings(result.ClearedTableIDs, restoreGraphTableIDs) || result.PostconditionSHA256 == nil ||
			!validSHA256String(*result.PostconditionSHA256) || len(result.Errors) != 0 {
			return NewRestoreError(RestoreErrorPostconditionFailed)
		}
	case result.Status == RestoreStatusFailed && result.ReadinessOutcome == RestoreReadinessIncomplete:
		if len(result.ClearedTableIDs) != 0 || len(result.RebuiltViews) != 0 || result.PostconditionSHA256 != nil || len(result.Errors) == 0 {
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
	if !sort.SliceIsSorted(result.RebuiltViews, func(i, j int) bool {
		if result.RebuiltViews[i].SourceRegistrationID != result.RebuiltViews[j].SourceRegistrationID {
			return result.RebuiltViews[i].SourceRegistrationID < result.RebuiltViews[j].SourceRegistrationID
		}
		return result.RebuiltViews[i].CandidateID < result.RebuiltViews[j].CandidateID
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

func mustRestoreCanonicalJSON(value any) []byte {
	body, _ := restoreCanonicalJSON(value)
	return body
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
