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
	"sort"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/modules/crossownertransaction"
	"github.com/JochiRaider/cartulary/internal/modules/incidentbundles/sourceport"
	"github.com/JochiRaider/cartulary/internal/modules/incidentportability"
	"github.com/JochiRaider/cartulary/internal/modules/stagedobjects"
	"github.com/JochiRaider/cartulary/internal/platform/canonicaljson"
)

const (
	PortabilityModeNoAuthoritativeState = "no_authoritative_incident_state"
	PortabilityModeParticipant          = "participant"
	PortabilityModeBlockedWhenPresent   = "blocked_when_present"

	ClaimStateClaimed               = "claimed"
	ClaimStateUnclaimed             = "unclaimed"
	ClaimStateRecognizedUnclaimable = "recognized_unclaimable"

	ExtensionExportResultSchema  = "cartulary.extension_portability_export_result.v2"
	ExtensionImportResultSchema  = "cartulary.extension_portability_import_preparation_result.v2"
	ExtensionMemberCatalogSchema = "cartulary.incident_bundle_extension_members.v2"

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
	ProfileID  string
	ClaimState string
	// ContractMajor is absent (zero) only for recognized-unclaimable profiles.
	// Such a policy never emits a payload or invokes a profile participant.
	ContractMajor          int
	Mode                   string
	ParticipantID          string
	ParticipantSHA256      string
	ParticipantSchemaID    string
	Export                 PortabilityOperationPolicy
	Import                 PortabilityOperationPolicy
	AuthoritativeFamilyIDs []string
	BlockingFamilyIDs      []string
}

type PortabilityOperationPolicy struct {
	OutputSchemaID     string
	StateFamilyIDs     []string
	MaximumInputBytes  int64
	MaximumOutputBytes int64
	MaximumItems       int
}

func validPortabilityOperation(policy PortabilityOperationPolicy) bool {
	return policy.OutputSchemaID != "" && policy.MaximumInputBytes > 0 &&
		policy.MaximumInputBytes <= PortabilityParticipantByteLimit && policy.MaximumOutputBytes >= 0 &&
		policy.MaximumOutputBytes <= PortabilityParticipantByteLimit && policy.MaximumItems >= 0 && policy.MaximumItems <= 1048576
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

// ExportSnapshotQuery is available only to application assembly and the owner's
// physical capability factory. Participant code receives the refined logical
// capability, never this query interface or a transaction handle.
type ExportSnapshotQuery interface {
	StatePresenceQuery
	incidentportability.Queryer
}

type ExportReadCapability interface {
	ProfileScope() string
	IncidentScope() uuid.UUID
}

type ExportReadFactory interface {
	ForSnapshot(ExportSnapshotQuery, uuid.UUID, string, []string) (ExportReadCapability, error)
}

type ExportInvocation struct {
	ProfileID     string
	ContractMajor int
	ClaimState    string
	IncidentID    uuid.UUID
	ReadAccess    ExportReadCapability
	Members       incidentportability.MemberWriter
	Actors        incidentportability.ActorReferences
}

type ExportResult struct {
	SchemaID        string
	Kind            string
	ContentSchemaID string
	ContractMajor   int
	StateVersion    int
}

type PortabilityImportContext struct {
	OperationID string
	IncidentID  uuid.UUID
	ActorUserID uuid.UUID
	Actors      sourceport.ActorCatalog
}

// ImportDependencyAccess is refined by each owner to the admitted logical source
// dependencies it needs. The participant never receives the whole bundle.
type ImportDependencyAccess interface {
	ProfileScope() string
	IncidentScope() uuid.UUID
}
type ImportDependencyFactory interface {
	ForImport(context.Context, PortabilityImportContext, incidentportability.MemberReader, string) (ImportDependencyAccess, error)
}

// PreparationMembers holds derived, operation-owned scratch inputs. It is
// sealed after preparation and abandoned after either commit or rollback; it
// never becomes an incident artifact or an exported bundle member.
type PreparationMembers interface {
	incidentportability.MemberReader
	incidentportability.MemberWriter
}

type ImportInvocation struct {
	PreparationMembers PreparationMembers
	Dependencies       ImportDependencyAccess
	PortabilityImportContext
	ProfileID       string
	ContractMajor   int
	ClaimState      string
	StateVersion    int
	ContentSchemaID string
	CatalogSHA256   string
	Members         incidentportability.MemberReader
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
	scope          *stagedobjects.Scope
	references     []stagedobjects.Reference
	budget         *memberBudget
	remainingItems *int
	publications   []stagedobjects.Publication
}

func (s *PortabilityStagedOutputScope) Allocate(ctx context.Context, operationID string, resource stagedobjects.ResourceBinding, payload []byte) (string, error) {
	if s == nil || s.scope == nil || s.budget == nil || !resource.Valid() {
		return "", stagedobjects.ErrScopeDenied
	}
	if s.remainingItems != nil && *s.remainingItems <= 0 {
		return "", ErrPortabilityLimit
	}
	if s.budget.members <= 0 || int64(len(payload)) > s.budget.bytes || len(payload) > PortabilityParticipantByteLimit {
		return "", ErrPortabilityLimit
	}
	stagingID, err := s.scope.Allocate(ctx, operationID, payload)
	if err != nil {
		return "", err
	}
	s.budget.bytes -= int64(len(payload))
	s.budget.members--
	if s.remainingItems != nil {
		(*s.remainingItems)--
	}
	s.references = append(s.references, stagedobjects.Reference{StagingID: stagingID})
	s.publications = append(s.publications, stagedobjects.Publication{StagingID: stagingID, ResourceKind: resource.Kind, ResourceID: resource.ID, ByteSize: int64(len(payload)), SHA256: digestBytes(payload)})
	return stagedOutputLogicalRef(stagingID), nil
}

// AllocateStream retains bytes under the operation lease and returns a logical
// content binding. Physical identity and publication remain with shared owners.
func (s *PortabilityStagedOutputScope) AllocateStream(ctx context.Context, operationID string, resource stagedobjects.ResourceBinding, source io.Reader, maximumBytes int64) (string, stagedobjects.ContentBinding, error) {
	if s == nil || s.scope == nil || s.budget == nil || !resource.Valid() {
		return "", stagedobjects.ContentBinding{}, stagedobjects.ErrScopeDenied
	}
	if s.remainingItems != nil && *s.remainingItems <= 0 {
		return "", stagedobjects.ContentBinding{}, ErrPortabilityLimit
	}
	if s.budget.members <= 0 || maximumBytes < 0 || maximumBytes > s.budget.bytes {
		return "", stagedobjects.ContentBinding{}, ErrPortabilityLimit
	}
	sealed, err := s.scope.AllocateStream(ctx, operationID, source, maximumBytes)
	if err != nil {
		return "", stagedobjects.ContentBinding{}, err
	}
	s.budget.bytes -= sealed.ByteSize
	s.budget.members--
	if s.remainingItems != nil {
		(*s.remainingItems)--
	}
	s.references = append(s.references, sealed.Reference)
	s.publications = append(s.publications, stagedobjects.Publication{StagingID: sealed.StagingID, ResourceKind: resource.Kind, ResourceID: resource.ID, ByteSize: sealed.ByteSize, SHA256: sealed.SHA256})
	return stagedOutputLogicalRef(sealed.StagingID), sealed.ContentBinding, nil
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

type PreparedPortability struct {
	Participants      []crossownertransaction.Participant
	Publications      []stagedobjects.PublicationPlan
	scopes            []*stagedobjects.Scope
	preparationScopes []*stagedobjects.Scope
	allocator         stagedobjects.Allocator
	references        []stagedobjects.Reference
}

func (p *PreparedPortability) Abandon(ctx context.Context) error {
	if p == nil {
		return nil
	}
	var joined error
	for _, scope := range p.preparationScopes {
		joined = errors.Join(joined, scope.Abandon(context.WithoutCancel(ctx)))
	}
	p.preparationScopes = nil
	for _, scope := range p.scopes {
		joined = errors.Join(joined, scope.Abandon(context.WithoutCancel(ctx)))
	}
	if p.allocator != nil {
		for _, reference := range p.references {
			joined = errors.Join(joined, p.allocator.Abandon(context.WithoutCancel(ctx), reference))
		}
	}
	p.scopes = nil
	p.Publications = nil
	p.Participants = nil
	p.references = nil
	return joined
}

func (p *PreparedPortability) Committed(ctx context.Context) error {
	if p == nil {
		return nil
	}
	var joined error
	for _, scope := range p.preparationScopes {
		joined = errors.Join(joined, scope.Abandon(context.WithoutCancel(ctx)))
	}
	p.preparationScopes = nil
	p.scopes = nil
	p.references = nil
	p.Publications = nil
	p.Participants = nil
	return joined
}

type PortabilityOrchestrator struct {
	policies     []ExtensionPolicy
	presence     StatePresence
	participants map[string]ExtensionParticipant
	allocator    stagedobjects.Allocator
	reads        ExportReadFactory
	dependencies ImportDependencyFactory
	limits       Limits
}

type PortabilityOptions struct {
	Policies           []ExtensionPolicy
	Presence           StatePresence
	Participants       []ExtensionParticipant
	Allocator          stagedobjects.Allocator
	SnapshotReads      ExportReadFactory
	ImportDependencies ImportDependencyFactory
	Limits             Limits
}

func NewPortabilityOrchestrator(options PortabilityOptions) (*PortabilityOrchestrator, error) {
	policies, presence, participants, allocator := options.Policies, options.Presence, options.Participants, options.Allocator
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
		policy.Export.StateFamilyIDs = append([]string(nil), policy.Export.StateFamilyIDs...)
		policy.Import.StateFamilyIDs = append([]string(nil), policy.Import.StateFamilyIDs...)
		validMajor := policy.ClaimState == ClaimStateRecognizedUnclaimable && policy.ContractMajor == 0 ||
			policy.ClaimState != ClaimStateRecognizedUnclaimable && policy.ContractMajor >= 1 && policy.ContractMajor <= 2147483647
		if policy.ProfileID == "" || policy.ProfileID == previous || !validMajor ||
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
			if policy.ParticipantID == "" || !validPortabilityOperation(policy.Export) || !validPortabilityOperation(policy.Import) {
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
	if len(admittedParticipants) != len(participantMap) || len(participantMap) > 0 && options.SnapshotReads == nil {
		return nil, ErrPortabilityResult
	}
	return &PortabilityOrchestrator{
		policies: resolved, presence: presence, participants: participantMap, allocator: allocator, reads: options.SnapshotReads, dependencies: options.ImportDependencies, limits: options.Limits,
	}, nil
}

func (o *PortabilityOrchestrator) Export(ctx context.Context, query ExportSnapshotQuery, incidentID uuid.UUID, members incidentportability.MemberWriter, actors incidentportability.ActorReferences) error {
	if o == nil || query == nil || incidentID == uuid.Nil || members == nil || actors == nil {
		return ErrPortabilityUnavailable
	}
	var inputBytes, metadataBytes int64
	for _, policy := range o.policies {
		if err := contextError(ctx); err != nil {
			return err
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
				return err
			}
		}
		switch policy.Mode {
		case PortabilityModeNoAuthoritativeState:
			if present {
				return newPortabilityFailure(ErrPortabilityResult, policy.ProfileID)
			}
		case PortabilityModeBlockedWhenPresent:
			if present {
				return newPortabilityFailure(ErrPortabilityBlocked, policy.ProfileID)
			}
		case PortabilityModeParticipant:
			if !present {
				continue
			}
			if policy.ClaimState != ClaimStateClaimed {
				return newPortabilityFailure(ErrPortabilityUnavailable, policy.ProfileID)
			}
			contextBytes, err := canonicaljson.MarshalExtension(portabilityContextMetadata(policy, incidentID, "export"))
			if err != nil {
				return ErrPortabilityResult
			}
			contextSize := int64(len(contextBytes))
			if len(policy.Export.StateFamilyIDs) > policy.Export.MaximumItems || contextSize > policy.Export.MaximumInputBytes || contextSize > PortabilityParticipantByteLimit-inputBytes {
				return ErrPortabilityLimit
			}
			inputBytes += contextSize
			access, err := o.reads.ForSnapshot(snapshotReadOnly{query}, incidentID, policy.ProfileID, append([]string(nil), policy.Export.StateFamilyIDs...))
			if err != nil || access == nil || access.ProfileScope() != policy.ProfileID || access.IncidentScope() != incidentID {
				return newPortabilityFailure(ErrPortabilityUnavailable, policy.ProfileID)
			}
			emitter := &extensionMemberWriter{destination: members, profileID: policy.ProfileID, paths: map[string]incidentportability.Member{}, maximumItems: policy.Export.MaximumItems}
			result, err := o.participants[policy.ParticipantID].Export(ctx, ExportInvocation{
				ProfileID: policy.ProfileID, ContractMajor: policy.ContractMajor, ClaimState: policy.ClaimState,
				IncidentID: incidentID, ReadAccess: access, Members: emitter, Actors: actors,
			})
			emitter.closed = true
			if errors.Is(err, ErrPortabilityLimit) {
				return ErrPortabilityLimit
			}
			if err != nil {
				return newPortabilityFailure(ErrPortabilityResult, policy.ProfileID)
			}
			if result.SchemaID != ExtensionExportResultSchema || result.Kind != "members" ||
				result.ContentSchemaID != policy.Export.OutputSchemaID || result.ContractMajor != policy.ContractMajor || result.StateVersion < 1 || len(emitter.paths) == 0 {
				return newPortabilityFailure(ErrPortabilityResult, policy.ProfileID)
			}
			catalog := extensionMemberCatalog{SchemaID: ExtensionMemberCatalogSchema, ProfileID: policy.ProfileID,
				ContractMajor: result.ContractMajor, StateVersion: result.StateVersion, ContentSchemaID: result.ContentSchemaID, Members: emitter.members()}
			encoded, err := encodeExtensionCatalog(catalog)
			if err != nil {
				return err
			}
			resultBytes, err := canonicaljson.MarshalExtension(map[string]any{
				"schema_id": result.SchemaID, "kind": result.Kind, "content_schema_id": result.ContentSchemaID,
				"contract_major": result.ContractMajor, "state_version": result.StateVersion,
			})
			if err != nil {
				return ErrPortabilityResult
			}
			outputSize := int64(len(encoded)) + int64(len(resultBytes))
			if outputSize > policy.Export.MaximumOutputBytes || outputSize > PortabilityParticipantByteLimit-metadataBytes {
				return ErrPortabilityLimit
			}
			metadataBytes += outputSize
			if _, err := members.WriteMember(ctx, extensionCatalogPath(policy.ProfileID), bytes.NewReader(encoded)); err != nil {
				return err
			}
		}
	}
	return nil
}

func portabilityContextMetadata(policy ExtensionPolicy, incidentID uuid.UUID, operation string) map[string]any {
	return map[string]any{
		"schema_id": "cartulary.extension_portability_participant_context.v2", "operation_kind": operation,
		"profile_id": policy.ProfileID, "contract_major": policy.ContractMajor, "claim_state": policy.ClaimState,
		"incident_id": incidentID.String(), "content_schema_id": policy.Export.OutputSchemaID,
	}
}

// This wrapper deliberately hides any Exec, Commit or Rollback methods on the
// caller's concrete transaction before invoking the trusted capability factory.
type snapshotReadOnly struct{ ExportSnapshotQuery }

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

func (o *PortabilityOrchestrator) PrepareImport(ctx context.Context, input PortabilityImportContext, files incidentportability.MemberReader) (PreparedPortability, error) {
	operationID, incidentID := input.OperationID, input.IncidentID
	if o == nil || operationID == "" || incidentID == uuid.Nil || input.ActorUserID == uuid.Nil {
		return PreparedPortability{}, ErrPortabilityPayload
	}
	payloads, err := decodeExtensionCatalogs(ctx, files)
	if err != nil {
		return PreparedPortability{}, err
	}
	if len(payloads) == 0 {
		return PreparedPortability{Participants: []crossownertransaction.Participant{}, Publications: []stagedobjects.PublicationPlan{}}, nil
	}
	policies := make(map[string]ExtensionPolicy, len(o.policies))
	for _, policy := range o.policies {
		policies[policy.ProfileID] = policy
	}
	prepared := PreparedPortability{allocator: o.allocator}
	var aggregate, preparedBytes int64
	budget := memberBudget{bytes: o.limits.IncidentBundles.MaxExtractedBytes, members: o.limits.Archives.MaxMembers}
	if budget.bytes <= 0 {
		budget.bytes = defaultIncidentBundleMaxExtractedBytes
	}
	if budget.members <= 0 {
		budget.members = defaultArchiveMaxMembers
	}
	for _, payload := range payloads {
		if err := contextError(ctx); err != nil {
			_ = prepared.Abandon(ctx)
			return PreparedPortability{}, err
		}
		policy, recognized := policies[payload.ProfileID]
		if !recognized || policy.Mode != PortabilityModeParticipant || policy.ClaimState != ClaimStateClaimed {
			_ = prepared.Abandon(ctx)
			return PreparedPortability{}, newPortabilityFailure(ErrPortabilityUnavailable, payload.ProfileID)
		}
		if payload.ContractMajor != policy.ContractMajor {
			_ = prepared.Abandon(ctx)
			return PreparedPortability{}, newPortabilityFailure(ErrPortabilityResult, payload.ProfileID)
		}
		if payload.ContentSchemaID != policy.Export.OutputSchemaID || payload.StateVersion < 1 {
			_ = prepared.Abandon(ctx)
			return PreparedPortability{}, newPortabilityFailure(ErrPortabilityPayload, payload.ProfileID)
		}
		metadata := portabilityContextMetadata(policy, incidentID, "import")
		metadata["operation_id"], metadata["actor_user_id"] = input.OperationID, input.ActorUserID.String()
		metadata["state_version"], metadata["catalog_sha256"] = payload.StateVersion, payload.sha256
		contextBytes, err := canonicaljson.MarshalExtension(metadata)
		if err != nil {
			_ = prepared.Abandon(ctx)
			return PreparedPortability{}, ErrPortabilityResult
		}
		size := payload.byteSize + int64(len(contextBytes))
		if len(payload.Members) > policy.Import.MaximumItems || size > policy.Import.MaximumInputBytes || aggregate > PortabilityParticipantByteLimit-size {
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
		remainingItems := policy.Import.MaximumItems
		participantScope := &PortabilityStagedOutputScope{scope: scope, budget: &budget, remainingItems: &remainingItems}
		var preparation *memberStore
		var preparationMembers PreparationMembers
		if _, ok := o.allocator.(stagedobjects.StreamAllocator); ok {
			preparationScope, scopeErr := stagedobjects.NewScope(operationID, payload.ProfileID, o.allocator)
			if scopeErr != nil {
				_ = prepared.Abandon(ctx)
				return PreparedPortability{}, scopeErr
			}
			prepared.preparationScopes = append(prepared.preparationScopes, preparationScope)
			preparation, err = newMemberStore(operationID, preparationScope, o.limits)
			if err != nil {
				_ = prepared.Abandon(ctx)
				return PreparedPortability{}, err
			}
			preparation.pathPrefix = "prepared/" + payload.ProfileID + "/"
			preparation.sharedBudget = &budget
			preparation.remainingItems = &remainingItems
			preparationMembers = preparation
		}
		var dependencies ImportDependencyAccess
		if o.dependencies != nil {
			dependencies, err = o.dependencies.ForImport(ctx, input, files, payload.ProfileID)
			if err != nil || dependencies == nil || dependencies.ProfileScope() != payload.ProfileID || dependencies.IncidentScope() != incidentID {
				_ = prepared.Abandon(ctx)
				return PreparedPortability{}, newPortabilityFailure(ErrPortabilityPayload, payload.ProfileID)
			}
		}
		result, err := participant.PrepareImport(ctx, ImportInvocation{
			PreparationMembers: preparationMembers, Dependencies: dependencies, PortabilityImportContext: input, ProfileID: payload.ProfileID, ContractMajor: policy.ContractMajor,
			ClaimState: policy.ClaimState, StateVersion: payload.StateVersion,
			ContentSchemaID: payload.ContentSchemaID, CatalogSHA256: payload.sha256, Members: payload.bound,
		}, participantScope)
		if preparation != nil {
			preparation.Seal()
		}
		if errors.Is(err, ErrPortabilityLimit) {
			_ = prepared.Abandon(ctx)
			return PreparedPortability{}, ErrPortabilityLimit
		}
		if err != nil || validateImportPreparation(result, participantScope) != nil {
			_ = prepared.Abandon(ctx)
			return PreparedPortability{}, newPortabilityFailure(ErrPortabilityResult, payload.ProfileID)
		}
		resultMetadata, err := canonicaljson.MarshalExtension(map[string]any{
			"schema_id": result.SchemaID, "status": result.Status,
			"participant_input":        json.RawMessage(result.ParticipantInput),
			"participant_input_sha256": result.ParticipantInputSHA256,
			"staged_output_refs":       result.StagedOutputRefs,
		})
		if err != nil {
			_ = prepared.Abandon(ctx)
			return PreparedPortability{}, ErrPortabilityResult
		}
		publicationMetadata, err := canonicaljson.MarshalExtension(append([]stagedobjects.Publication{}, participantScope.publications...))
		if err != nil {
			_ = prepared.Abandon(ctx)
			return PreparedPortability{}, ErrPortabilityResult
		}
		resultSize := int64(len(resultMetadata)) + int64(len(publicationMetadata))
		if resultSize > policy.Import.MaximumOutputBytes || resultSize > PortabilityParticipantByteLimit-preparedBytes {
			_ = prepared.Abandon(ctx)
			return PreparedPortability{}, ErrPortabilityLimit
		}
		preparedBytes += resultSize
		prepared.references = append(prepared.references, participantScope.references...)
		transfer, transferErr := scope.Transfer(operationID)
		if transferErr != nil {
			_ = prepared.Abandon(ctx)
			return PreparedPortability{}, transferErr
		}
		prepared.Publications = append(prepared.Publications, stagedobjects.PublicationPlan{Transfer: transfer, Publications: participantScope.publications})
		prepared.Participants = append(prepared.Participants, &preparedPortabilityParticipant{
			delegate:    result.TransactionParticipant,
			operationID: operationID,
			input:       bytes.Clone(result.ParticipantInput),
			sha256:      result.ParticipantInputSHA256,
		})
	}
	prepared.scopes = nil
	return prepared, nil
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
	delegate    crossownertransaction.Participant
	input       []byte
	sha256      string
	operationID string
}

func (p *preparedPortabilityParticipant) ID() string {
	if p == nil || p.delegate == nil {
		return ""
	}
	return p.delegate.ID()
}

func (p *preparedPortabilityParticipant) BuildInput(ctx context.Context, operation crossownertransaction.OperationContext) (crossownertransaction.Input, error) {
	if p == nil || p.delegate == nil || p.sha256 != digestBytes(p.input) || operation.OperationID != p.operationID {
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
