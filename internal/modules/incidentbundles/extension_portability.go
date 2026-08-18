package incidentbundles

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"path"
	"sort"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/modules/crossownertransaction"
	"github.com/JochiRaider/cartulary/internal/modules/stagedobjects"
)

const (
	PortabilityModeNoAuthoritativeState = "no_authoritative_incident_state"
	PortabilityModeParticipant          = "participant"
	PortabilityModeBlockedWhenPresent   = "blocked_when_present"

	ClaimStateClaimed               = "claimed"
	ClaimStateUnclaimed             = "unclaimed"
	ClaimStateRecognizedUnclaimable = "recognized_unclaimable"

	ExtensionExportResultSchema = "cartulary.extension_portability_export_result.v1"
	ExtensionImportResultSchema = "cartulary.extension_portability_import_preparation_result.v1"
	ExtensionPayloadSchema      = "cartulary.incident_bundle_extension_payload.v1"

	PortabilityParticipantByteLimit = 64 * 1024 * 1024
)

var (
	ErrPortabilityUnavailable = errors.New("extension_state_unavailable_for_portability")
	ErrPortabilityBlocked     = errors.New("extension_portability_blocked")
	ErrPortabilityPayload     = errors.New("extension_portability_payload_invalid")
	ErrPortabilityLimit       = errors.New("extension_portability_input_limit")
	ErrPortabilityResult      = errors.New("extension_portability_result_invalid")
)

// ExtensionPolicy is the Incident Bundles owner's immutable view of one
// Extensions declaration. Application composition translates the registry
// projection and resolved serving claim into this shape.
type ExtensionPolicy struct {
	ProfileID              string
	ClaimState             string
	ContractMajor          int
	Mode                   string
	ParticipantID          string
	ParticipantSHA256      string
	ParticipantSchemaID    string
	MaximumInputBytes      int64
	MaximumOutputBytes     int64
	AuthoritativeFamilyIDs []string
	BlockingFamilyIDs      []string
}

// StatePresence is the declarative storage-binding port. Implementations count
// retained authoritative members only and never invoke profile workers,
// migrations, probes, or participant code.
type StatePresence interface {
	AuthoritativeStatePresent(context.Context, StatePresenceQuery, uuid.UUID, string, []string) (bool, error)
}

// StatePresenceQuery is the transaction-bound read capability supplied by the
// Incident Bundles owner to declarative extension state bindings.
type StatePresenceQuery interface {
	QueryRow(context.Context, string, ...any) pgx.Row
}

type ExportInvocation struct {
	ProfileID     string
	ContractMajor int
	ClaimState    string
	IncidentID    uuid.UUID
	StatePresent  bool
}

type ExportResult struct {
	SchemaID             string
	Kind                 string
	PayloadSchemaID      string
	PayloadContractMajor int
	StateVersion         int
	Payload              []byte
}

type ImportInvocation struct {
	OperationID   string
	ProfileID     string
	ContractMajor int
	ClaimState    string
	IncidentID    uuid.UUID
	StateVersion  int
	PayloadSchema string
	Payload       []byte
}

type ImportPreparation struct {
	SchemaID               string
	Status                 string
	ParticipantInput       []byte
	ParticipantInputSHA256 string
	StagedOutputRefs       []string
	TransactionParticipant crossownertransaction.Participant
}

// ExtensionParticipant is the closed, owner-injected portability specialization.
// Export and import preparation are separate methods and result shapes.
type ExtensionParticipant interface {
	ID() string
	ContractSHA256() string
	SpecializationSchemaID() string
	Export(context.Context, ExportInvocation) (ExportResult, error)
	PrepareImport(context.Context, ImportInvocation, *PortabilityStagedOutputScope) (ImportPreparation, error)
}

// PortabilityStagedOutputScope exposes only non-authorizing logical references
// to participant code. The underlying staging identity remains private to the
// shared staged-object owner and is retained solely for abandon/transfer.
type PortabilityStagedOutputScope struct {
	scope      *stagedobjects.Scope
	references []stagedobjects.Reference
}

func (s *PortabilityStagedOutputScope) Allocate(ctx context.Context, operationID string, payload []byte) (string, error) {
	if s == nil || s.scope == nil {
		return "", stagedobjects.ErrScopeDenied
	}
	stagingID, err := s.scope.Allocate(ctx, operationID, payload)
	if err != nil {
		return "", err
	}
	s.references = append(s.references, stagedobjects.Reference{StagingID: stagingID})
	return stagedOutputLogicalRef(stagingID), nil
}

func (s *PortabilityStagedOutputScope) Refs() []string {
	if s == nil {
		return nil
	}
	result := make([]string, len(s.references))
	for index, reference := range s.references {
		result[index] = stagedOutputLogicalRef(reference.StagingID)
	}
	sort.Strings(result)
	return result
}

type ExtensionPayload struct {
	ProfileID       string
	ContractMajor   int
	StateVersion    int
	PayloadSchemaID string
	PayloadSHA256   string
	Payload         []byte
}

type PreparedPortability struct {
	Participants []crossownertransaction.Participant
	Transfers    []stagedobjects.Transfer
	scopes       []*stagedobjects.Scope
	allocator    stagedobjects.Allocator
	references   []stagedobjects.Reference
}

func (p *PreparedPortability) Abandon(ctx context.Context) error {
	if p == nil {
		return nil
	}
	var joined error
	for _, scope := range p.scopes {
		joined = errors.Join(joined, scope.Abandon(context.WithoutCancel(ctx)))
	}
	if len(p.scopes) == 0 && p.allocator != nil {
		for _, reference := range p.references {
			joined = errors.Join(joined, p.allocator.Abandon(context.WithoutCancel(ctx), reference))
		}
	}
	p.scopes = nil
	p.Transfers = nil
	p.Participants = nil
	p.references = nil
	return joined
}

func (p *PreparedPortability) Committed() {
	if p == nil {
		return
	}
	p.scopes = nil
	p.references = nil
	p.Transfers = nil
	p.Participants = nil
}

type PortabilityOrchestrator struct {
	policies     []ExtensionPolicy
	presence     StatePresence
	participants map[string]ExtensionParticipant
	allocator    stagedobjects.Allocator
}

func NewPortabilityOrchestrator(policies []ExtensionPolicy, presence StatePresence, participants []ExtensionParticipant, allocator stagedobjects.Allocator) (*PortabilityOrchestrator, error) {
	resolved := append([]ExtensionPolicy(nil), policies...)
	sort.Slice(resolved, func(i, j int) bool {
		return bytes.Compare([]byte(resolved[i].ProfileID), []byte(resolved[j].ProfileID)) < 0
	})
	participantMap := make(map[string]ExtensionParticipant, len(participants))
	for _, participant := range participants {
		if participant == nil || participant.ID() == "" ||
			participant.ContractSHA256() == "" || participant.SpecializationSchemaID() == "" {
			return nil, ErrPortabilityResult
		}
		if _, duplicate := participantMap[participant.ID()]; duplicate {
			return nil, ErrPortabilityResult
		}
		participantMap[participant.ID()] = participant
	}
	previous := ""
	admittedParticipants := make(map[string]struct{}, len(participantMap))
	for index := range resolved {
		policy := &resolved[index]
		policy.AuthoritativeFamilyIDs = canonicalStrings(policy.AuthoritativeFamilyIDs)
		policy.BlockingFamilyIDs = canonicalStrings(policy.BlockingFamilyIDs)
		if policy.ProfileID == "" || policy.ProfileID == previous || policy.ContractMajor < 1 ||
			(policy.ClaimState != ClaimStateClaimed && policy.ClaimState != ClaimStateUnclaimed && policy.ClaimState != ClaimStateRecognizedUnclaimable) {
			return nil, ErrPortabilityResult
		}
		switch policy.Mode {
		case PortabilityModeNoAuthoritativeState:
			if policy.ParticipantID != "" || len(policy.BlockingFamilyIDs) != 0 {
				return nil, ErrPortabilityResult
			}
		case PortabilityModeBlockedWhenPresent:
			if policy.ParticipantID != "" || len(policy.BlockingFamilyIDs) == 0 || presence == nil {
				return nil, ErrPortabilityResult
			}
		case PortabilityModeParticipant:
			if policy.ParticipantID == "" || policy.MaximumInputBytes < 1 ||
				policy.MaximumInputBytes > PortabilityParticipantByteLimit ||
				policy.MaximumOutputBytes < 0 || policy.MaximumOutputBytes > PortabilityParticipantByteLimit {
				return nil, ErrPortabilityResult
			}
			if policy.ClaimState == ClaimStateClaimed {
				participant, installed := participantMap[policy.ParticipantID]
				if !installed {
					return nil, ErrPortabilityUnavailable
				}
				if participant.ContractSHA256() != policy.ParticipantSHA256 ||
					participant.SpecializationSchemaID() != policy.ParticipantSchemaID {
					return nil, ErrPortabilityResult
				}
				admittedParticipants[policy.ParticipantID] = struct{}{}
			}
		default:
			return nil, ErrPortabilityResult
		}
		if len(policy.AuthoritativeFamilyIDs) > 0 && presence == nil {
			return nil, ErrPortabilityResult
		}
		previous = policy.ProfileID
	}
	if len(admittedParticipants) != len(participantMap) {
		return nil, ErrPortabilityResult
	}
	return &PortabilityOrchestrator{
		policies: resolved, presence: presence, participants: participantMap, allocator: allocator,
	}, nil
}

func (o *PortabilityOrchestrator) Export(ctx context.Context, query StatePresenceQuery, incidentID uuid.UUID) ([]ExtensionPayload, error) {
	if o == nil || query == nil || incidentID == uuid.Nil {
		return nil, ErrPortabilityUnavailable
	}
	results := make([]ExtensionPayload, 0)
	for _, policy := range o.policies {
		if err := contextError(ctx); err != nil {
			return nil, err
		}
		families := policy.AuthoritativeFamilyIDs
		if policy.Mode == PortabilityModeBlockedWhenPresent {
			families = policy.BlockingFamilyIDs
		}
		present := false
		if len(families) > 0 {
			var err error
			present, err = o.presence.AuthoritativeStatePresent(ctx, query, incidentID, policy.ProfileID, families)
			if err != nil {
				return nil, err
			}
		}
		switch policy.Mode {
		case PortabilityModeNoAuthoritativeState:
			if present {
				return nil, newPortabilityFailure(ErrPortabilityResult, policy.ProfileID)
			}
		case PortabilityModeBlockedWhenPresent:
			if present {
				return nil, newPortabilityFailure(ErrPortabilityBlocked, policy.ProfileID)
			}
		case PortabilityModeParticipant:
			if !present {
				continue
			}
			if policy.ClaimState != ClaimStateClaimed {
				return nil, newPortabilityFailure(ErrPortabilityUnavailable, policy.ProfileID)
			}
			participant := o.participants[policy.ParticipantID]
			result, err := participant.Export(ctx, ExportInvocation{
				ProfileID: policy.ProfileID, ContractMajor: policy.ContractMajor,
				ClaimState: policy.ClaimState, IncidentID: incidentID, StatePresent: true,
			})
			if err != nil {
				return nil, newPortabilityFailure(ErrPortabilityResult, policy.ProfileID)
			}
			payload, include, err := validateExportResult(policy, result)
			if err != nil {
				return nil, newPortabilityFailure(err, policy.ProfileID)
			}
			if include {
				results = append(results, payload)
			}
		}
	}
	return results, nil
}

// ValidatePublication is the authoritative last-moment guard. It evaluates
// only declarative blockers and never invokes extension participant code.
func (o *PortabilityOrchestrator) ValidatePublication(ctx context.Context, query StatePresenceQuery, incidentID uuid.UUID) error {
	if o == nil || query == nil || incidentID == uuid.Nil {
		return ErrPortabilityUnavailable
	}
	for _, policy := range o.policies {
		if policy.Mode != PortabilityModeBlockedWhenPresent {
			continue
		}
		if err := contextError(ctx); err != nil {
			return err
		}
		present, err := o.presence.AuthoritativeStatePresent(ctx, query, incidentID, policy.ProfileID, policy.BlockingFamilyIDs)
		if err != nil {
			return err
		}
		if present {
			return newPortabilityFailure(ErrPortabilityBlocked, policy.ProfileID)
		}
	}
	return nil
}

func (o *PortabilityOrchestrator) PrepareImport(ctx context.Context, operationID string, incidentID uuid.UUID, files map[string][]byte) (PreparedPortability, error) {
	if o == nil || operationID == "" || incidentID == uuid.Nil {
		return PreparedPortability{}, ErrPortabilityPayload
	}
	payloads, err := decodeExtensionPayloads(files)
	if err != nil {
		return PreparedPortability{}, err
	}
	if len(payloads) == 0 {
		return PreparedPortability{Participants: []crossownertransaction.Participant{}, Transfers: []stagedobjects.Transfer{}}, nil
	}
	policies := make(map[string]ExtensionPolicy, len(o.policies))
	for _, policy := range o.policies {
		policies[policy.ProfileID] = policy
	}
	prepared := PreparedPortability{allocator: o.allocator}
	var aggregate int64
	for _, payload := range payloads {
		if err := contextError(ctx); err != nil {
			_ = prepared.Abandon(ctx)
			return PreparedPortability{}, err
		}
		policy, recognized := policies[payload.ProfileID]
		if !recognized || policy.Mode != PortabilityModeParticipant ||
			policy.ClaimState != ClaimStateClaimed ||
			payload.ContractMajor != policy.ContractMajor ||
			payload.StateVersion < 1 {
			_ = prepared.Abandon(ctx)
			return PreparedPortability{}, newPortabilityFailure(ErrPortabilityPayload, payload.ProfileID)
		}
		size := int64(len(payload.Payload))
		if size > policy.MaximumInputBytes || aggregate > PortabilityParticipantByteLimit-size {
			_ = prepared.Abandon(ctx)
			return PreparedPortability{}, ErrPortabilityLimit
		}
		aggregate += size
		participant := o.participants[policy.ParticipantID]
		if participant == nil || o.allocator == nil {
			_ = prepared.Abandon(ctx)
			return PreparedPortability{}, newPortabilityFailure(ErrPortabilityUnavailable, payload.ProfileID)
		}
		scope, err := stagedobjects.NewScope(operationID, payload.ProfileID, o.allocator)
		if err != nil {
			_ = prepared.Abandon(ctx)
			return PreparedPortability{}, err
		}
		prepared.scopes = append(prepared.scopes, scope)
		participantScope := &PortabilityStagedOutputScope{scope: scope}
		result, err := participant.PrepareImport(ctx, ImportInvocation{
			OperationID: operationID, ProfileID: payload.ProfileID, ContractMajor: policy.ContractMajor,
			ClaimState: policy.ClaimState, IncidentID: incidentID, StateVersion: payload.StateVersion,
			PayloadSchema: payload.PayloadSchemaID, Payload: bytes.Clone(payload.Payload),
		}, participantScope)
		if err != nil || validateImportPreparation(result, participantScope) != nil {
			_ = prepared.Abandon(ctx)
			return PreparedPortability{}, newPortabilityFailure(ErrPortabilityResult, payload.ProfileID)
		}
		prepared.references = append(prepared.references, participantScope.references...)
		prepared.Participants = append(prepared.Participants, &preparedPortabilityParticipant{
			delegate: result.TransactionParticipant,
			input:    bytes.Clone(result.ParticipantInput),
			sha256:   result.ParticipantInputSHA256,
		})
	}
	for _, scope := range prepared.scopes {
		transfer, err := scope.Transfer(operationID)
		if err != nil {
			_ = prepared.Abandon(ctx)
			return PreparedPortability{}, err
		}
		prepared.Transfers = append(prepared.Transfers, transfer)
	}
	prepared.scopes = nil
	return prepared, nil
}

func validateExportResult(policy ExtensionPolicy, result ExportResult) (ExtensionPayload, bool, error) {
	if result.SchemaID != ExtensionExportResultSchema {
		return ExtensionPayload{}, false, ErrPortabilityResult
	}
	if result.Kind == "omit" {
		if result.PayloadSchemaID != "" || result.PayloadContractMajor != 0 || result.StateVersion != 0 || result.Payload != nil {
			return ExtensionPayload{}, false, ErrPortabilityResult
		}
		return ExtensionPayload{}, false, nil
	}
	if result.Kind != "payload" || result.PayloadSchemaID == "" || result.PayloadContractMajor < 1 ||
		result.StateVersion < 1 || result.Payload == nil || int64(len(result.Payload)) > policy.MaximumOutputBytes {
		return ExtensionPayload{}, false, ErrPortabilityResult
	}
	sum := sha256.Sum256(result.Payload)
	return ExtensionPayload{
		ProfileID: policy.ProfileID, ContractMajor: result.PayloadContractMajor,
		StateVersion: result.StateVersion, PayloadSchemaID: result.PayloadSchemaID,
		PayloadSHA256: hex.EncodeToString(sum[:]), Payload: bytes.Clone(result.Payload),
	}, true, nil
}

func validateImportPreparation(result ImportPreparation, scope *PortabilityStagedOutputScope) error {
	if result.SchemaID != ExtensionImportResultSchema || result.Status != "prepared" ||
		result.TransactionParticipant == nil || result.ParticipantInput == nil ||
		int64(len(result.ParticipantInput)) > PortabilityParticipantByteLimit ||
		result.ParticipantInputSHA256 != digestBytes(result.ParticipantInput) ||
		!equalStrings(canonicalStrings(result.StagedOutputRefs), result.StagedOutputRefs) ||
		!equalStrings(result.StagedOutputRefs, scope.Refs()) {
		return ErrPortabilityResult
	}
	return nil
}

// preparedPortabilityParticipant binds the side-effect-free preparation bytes
// to the later admitted transaction participant. A participant cannot replace
// or reinterpret those bytes after the bundle owner has accepted them.
type preparedPortabilityParticipant struct {
	delegate crossownertransaction.Participant
	input    []byte
	sha256   string
}

func (p *preparedPortabilityParticipant) ID() string {
	if p == nil || p.delegate == nil {
		return ""
	}
	return p.delegate.ID()
}

func (p *preparedPortabilityParticipant) BuildInput(ctx context.Context, operation crossownertransaction.OperationContext) (crossownertransaction.Input, error) {
	if p == nil || p.delegate == nil || p.sha256 != digestBytes(p.input) {
		return crossownertransaction.Input{}, ErrPortabilityResult
	}
	input, err := p.delegate.BuildInput(ctx, operation)
	if err != nil || !bytes.Equal(input.CanonicalBytes, p.input) {
		return crossownertransaction.Input{}, ErrPortabilityResult
	}
	input.CanonicalBytes = bytes.Clone(p.input)
	return input, nil
}

func (p *preparedPortabilityParticipant) Prepare(ctx context.Context, invocation crossownertransaction.Invocation) (crossownertransaction.PrepareResult, error) {
	return p.delegate.Prepare(ctx, invocation)
}

func (p *preparedPortabilityParticipant) Validate(ctx context.Context, invocation crossownertransaction.Invocation) (crossownertransaction.ValidationResult, error) {
	return p.delegate.Validate(ctx, invocation)
}

func (p *preparedPortabilityParticipant) Write(ctx context.Context, invocation crossownertransaction.Invocation) (crossownertransaction.WriteResult, error) {
	return p.delegate.Write(ctx, invocation)
}

func stagedOutputLogicalRef(stagingID string) string {
	return "cartulary:staged_output:" + stagingID
}

type extensionPayloadEnvelope struct {
	SchemaID        string          `json:"schema_id"`
	ProfileID       string          `json:"profile_id"`
	ContractMajor   int             `json:"contract_major"`
	StateVersion    int             `json:"state_version"`
	PayloadSchemaID string          `json:"payload_schema_id"`
	PayloadSHA256   string          `json:"payload_sha256"`
	Payload         json.RawMessage `json:"payload"`
}

func extensionPayloadPath(profileID string) string {
	return "ext/extensions/" + profileID + "/payload.json"
}

func EncodeExtensionPayload(payload ExtensionPayload) (string, []byte, error) {
	if payload.ProfileID == "" || payload.ContractMajor < 1 || payload.StateVersion < 1 ||
		payload.PayloadSchemaID == "" || payload.Payload == nil || payload.PayloadSHA256 != digestBytes(payload.Payload) {
		return "", nil, ErrPortabilityPayload
	}
	encoded, err := json.Marshal(extensionPayloadEnvelope{
		SchemaID: ExtensionPayloadSchema, ProfileID: payload.ProfileID,
		ContractMajor: payload.ContractMajor, StateVersion: payload.StateVersion,
		PayloadSchemaID: payload.PayloadSchemaID, PayloadSHA256: payload.PayloadSHA256,
		Payload: bytes.Clone(payload.Payload),
	})
	if err != nil {
		return "", nil, err
	}
	return extensionPayloadPath(payload.ProfileID), append(encoded, '\n'), nil
}

func decodeExtensionPayloads(files map[string][]byte) ([]ExtensionPayload, error) {
	paths := make([]string, 0)
	for filePath := range files {
		if strings.HasPrefix(filePath, "ext/extensions/") {
			paths = append(paths, filePath)
		}
	}
	sort.Strings(paths)
	results := make([]ExtensionPayload, 0, len(paths))
	previous := ""
	for _, filePath := range paths {
		clean := path.Clean(filePath)
		parts := strings.Split(clean, "/")
		if clean != filePath || len(parts) != 4 || parts[0] != "ext" || parts[1] != "extensions" || parts[3] != "payload.json" {
			return nil, ErrPortabilityPayload
		}
		var envelope extensionPayloadEnvelope
		decoder := json.NewDecoder(bytes.NewReader(files[filePath]))
		decoder.DisallowUnknownFields()
		if err := decoder.Decode(&envelope); err != nil || envelope.SchemaID != ExtensionPayloadSchema ||
			envelope.ProfileID != parts[2] || envelope.ProfileID == previous ||
			envelope.ContractMajor < 1 || envelope.StateVersion < 1 ||
			envelope.PayloadSchemaID == "" || envelope.Payload == nil ||
			envelope.PayloadSHA256 != digestBytes(envelope.Payload) {
			return nil, ErrPortabilityPayload
		}
		var extra any
		if err := decoder.Decode(&extra); !errors.Is(err, io.EOF) {
			return nil, ErrPortabilityPayload
		}
		results = append(results, ExtensionPayload{
			ProfileID: envelope.ProfileID, ContractMajor: envelope.ContractMajor,
			StateVersion: envelope.StateVersion, PayloadSchemaID: envelope.PayloadSchemaID,
			PayloadSHA256: envelope.PayloadSHA256, Payload: bytes.Clone(envelope.Payload),
		})
		previous = envelope.ProfileID
	}
	return results, nil
}

type PortabilityFailure struct {
	Kind      error
	ProfileID string
}

func (e *PortabilityFailure) Error() string {
	return fmt.Sprintf("%s: %s", e.Kind, e.ProfileID)
}

func (e *PortabilityFailure) Unwrap() error { return e.Kind }

func newPortabilityFailure(kind error, profileID string) error {
	return &PortabilityFailure{Kind: kind, ProfileID: profileID}
}

func canonicalStrings(source []string) []string {
	result := append([]string(nil), source...)
	sort.Strings(result)
	write := 0
	for _, value := range result {
		if value == "" || (write > 0 && result[write-1] == value) {
			continue
		}
		result[write] = value
		write++
	}
	return result[:write]
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

func digestBytes(payload []byte) string {
	sum := sha256.Sum256(payload)
	return hex.EncodeToString(sum[:])
}

func contextError(ctx context.Context) error {
	if ctx == nil {
		return context.Canceled
	}
	return ctx.Err()
}
