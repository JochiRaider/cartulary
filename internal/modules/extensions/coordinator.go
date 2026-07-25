package extensions

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strings"

	contractsgen "github.com/JochiRaider/cartulary/internal/gen/contracts"
)

const generatedExtensionsRoot = "contracts/extensions/generated/"

// PackagedArtifact is an immutable view of one build-time contract artifact.
// The byte slice returned to or received from callers is always copied.
type PackagedArtifact struct {
	JSON   []byte
	SHA256 string
}

// ArtifactSource is the coordinator's only build-package port. It deliberately
// exposes no filesystem, package-discovery, transport, or profile callbacks.
type ArtifactSource interface {
	Artifact(path string) (PackagedArtifact, bool)
}

type generatedArtifactSource struct{}

func (generatedArtifactSource) Artifact(path string) (PackagedArtifact, bool) {
	artifact, ok := contractsgen.ExtensionArtifactsIndex[path]
	if !ok {
		return PackagedArtifact{}, false
	}
	return PackagedArtifact{JSON: []byte(artifact.JSON), SHA256: artifact.SHA256}, true
}

// Finding is a safe coordination failure. It contains identities and contract
// tokens only; profile configuration values and implementation details never enter
// the coordinator.
type Finding struct {
	Code                string
	Phase               string
	ProfileID           string
	DependencyProfileID string
	CollisionClass      string
	ConflictingTokens   []string
	Expected            string
	Actual              string
}

// ValidationError closes a coordinator stage without exposing a partial result.
type ValidationError struct {
	findings []Finding
}

func (e *ValidationError) Error() string {
	if e == nil || len(e.findings) == 0 {
		return "extension coordination failed"
	}
	return e.findings[0].Code
}

func (e *ValidationError) Findings() []Finding {
	if e == nil {
		return nil
	}
	result := append([]Finding(nil), e.findings...)
	for index := range result {
		result[index].ConflictingTokens = append([]string(nil), result[index].ConflictingTokens...)
	}
	return result
}

// Dependency is the only dependency shape visible to coordinator callers.
type Dependency struct {
	ProfileID             string
	RequiredContractMajor int
}

// Descriptor is a read-only transport-neutral projection of the canonical
// descriptor. Executable algorithms and profile-owned contract bodies stay in the
// immutable packaged artifacts.
type Descriptor struct {
	ProfileID      string
	Claimable      bool
	ContractMajor  int
	OwnerID        string
	ClaimConfigKey string
	RouteFamilies  []string
	WorkspaceKeys  []string
	CapabilityIDs  []string
	Dependencies   []Dependency
}

func cloneDescriptor(descriptor Descriptor) Descriptor {
	descriptor.RouteFamilies = append([]string(nil), descriptor.RouteFamilies...)
	descriptor.WorkspaceKeys = append([]string(nil), descriptor.WorkspaceKeys...)
	descriptor.CapabilityIDs = append([]string(nil), descriptor.CapabilityIDs...)
	descriptor.Dependencies = append([]Dependency(nil), descriptor.Dependencies...)
	return descriptor
}

type profileRecord struct {
	descriptor       Descriptor
	descriptorObject map[string]any
	descriptorSHA256 string
	bindingObject    map[string]any
	bindingSHA256    string
	jobContracts     []JobKindContract
}

// Coordinator is an immutable coordination facade over generated build inputs.
// Construction performs registry, collision, and binding admission once.
type Coordinator struct {
	profiles                      map[string]profileRecord
	orderedProfileIDs             []string
	registrySHA256                string
	clientSupportSHA256           string
	validationConditions          map[string]ValidationCondition
	statePlans                    map[string]StatePlan
	backupPlans                   map[string]BackupPlan
	participantContracts          map[string]ParticipantContract
	portabilityPolicies           []PortabilityPolicy
	inactiveConfigurationPolicies []InactiveConfigurationPolicy
}

// NewGeneratedCoordinator admits the repository-generated extension package.
func NewGeneratedCoordinator() (*Coordinator, error) {
	return NewCoordinator(generatedArtifactSource{})
}

// NewCoordinator admits one exact immutable artifact source. It is exported so
// application composition and contract tests can inject package adapters without
// granting the coordinator access to their internals.
func NewCoordinator(source ArtifactSource) (*Coordinator, error) {
	if source == nil {
		return nil, validationFailure(Finding{Code: "extension_registry_invalid", Phase: "registry_generation", Actual: "missing_artifact_source"})
	}
	registry, registryArtifact, err := readArtifactObject(source, generatedExtensionsRoot+"profile-registry.json")
	if err != nil {
		return nil, err
	}
	if err := requireExactKeys(registry, "schema_id", "profiles"); err != nil || registry["schema_id"] != "cartulary.extension_profile_registry.v1" {
		return nil, invalidArtifact("profile_registry", err)
	}
	integrity, _, err := readArtifactObject(source, generatedExtensionsRoot+"registry-integrity.json")
	if err != nil {
		return nil, err
	}
	if integrity["schema_id"] != "cartulary.extension_registry_integrity.v1" || stringValue(integrity["registry_sha256"]) != registryArtifact.SHA256 {
		return nil, validationFailure(Finding{Code: "extension_registry_invalid", Phase: "registry_generation", Expected: stringValue(integrity["registry_sha256"]), Actual: registryArtifact.SHA256})
	}
	descriptorDigests, err := digestRows(integrity["descriptor_digests"], "profile_id", "descriptor_sha256")
	if err != nil {
		return nil, invalidArtifact("registry_integrity", err)
	}
	bindingDigests, err := digestRows(integrity["implementation_binding_digests"], "profile_id", "binding_sha256")
	if err != nil {
		return nil, invalidArtifact("registry_integrity", err)
	}
	supportDigests, err := digestRows(integrity["supporting_contract_artifact_digests"], "artifact_id", "artifact_sha256")
	if err != nil {
		return nil, invalidArtifact("registry_integrity", err)
	}
	profileObjects, ok := objectSlice(registry["profiles"])
	if !ok || len(profileObjects) == 0 || len(profileObjects) > 256 {
		return nil, invalidArtifact("profile_registry", errors.New("profiles must contain 1..256 objects"))
	}

	coordinator := &Coordinator{
		profiles:             make(map[string]profileRecord, len(profileObjects)),
		registrySHA256:       registryArtifact.SHA256,
		orderedProfileIDs:    make([]string, 0, len(profileObjects)),
		validationConditions: map[string]ValidationCondition{},
	}
	for _, profileObject := range profileObjects {
		descriptor, parseErr := parseDescriptor(profileObject)
		if parseErr != nil {
			return nil, invalidArtifact("profile_descriptor", parseErr)
		}
		if _, duplicate := coordinator.profiles[descriptor.ProfileID]; duplicate {
			return nil, collisionFailure("profile_id", descriptor.ProfileID)
		}
		descriptorObject, descriptorArtifact, readErr := readArtifactObject(source, generatedExtensionsRoot+"descriptors/"+descriptor.ProfileID+".json")
		if readErr != nil {
			return nil, readErr
		}
		if !equalCanonicalObjects(profileObject, descriptorObject) || descriptorDigests[descriptor.ProfileID] != descriptorArtifact.SHA256 {
			return nil, unavailableBinding(descriptor.ProfileID, "descriptor_digest_mismatch")
		}
		bindingObject, bindingArtifact, readErr := readArtifactObject(source, generatedExtensionsRoot+"implementation-bindings/"+descriptor.ProfileID+".json")
		if readErr != nil {
			return nil, unavailableBinding(descriptor.ProfileID, "binding_missing")
		}
		if bindingDigests[descriptor.ProfileID] != bindingArtifact.SHA256 {
			return nil, unavailableBinding(descriptor.ProfileID, "binding_digest_mismatch")
		}
		if bindingErr := validateBinding(descriptor, descriptorArtifact.SHA256, bindingObject); bindingErr != nil {
			return nil, unavailableBinding(descriptor.ProfileID, bindingErr.Error())
		}
		jobContracts, parseErr := parseJobKindContracts(descriptor.ProfileID, bindingObject["job_kind_contracts"])
		if parseErr != nil {
			return nil, unavailableBinding(descriptor.ProfileID, parseErr.Error())
		}
		for _, jobContract := range jobContracts {
			artifactID := "job-contracts/" + descriptor.ProfileID + "/" + jobContract.JobKind
			jobObject, jobArtifact, readErr := readArtifactObject(source, generatedExtensionsRoot+artifactID+".json")
			if readErr != nil {
				return nil, unavailableBinding(descriptor.ProfileID, "job_contract_missing")
			}
			if supportDigests[artifactID] != jobArtifact.SHA256 ||
				!equalCanonicalObjects(jobObject, jobContract.object()) {
				return nil, unavailableBinding(descriptor.ProfileID, "job_contract_digest_mismatch")
			}
		}
		coordinator.profiles[descriptor.ProfileID] = profileRecord{
			descriptor:       descriptor,
			descriptorObject: cloneObject(profileObject),
			descriptorSHA256: descriptorArtifact.SHA256,
			bindingObject:    cloneObject(bindingObject),
			bindingSHA256:    bindingArtifact.SHA256,
			jobContracts:     cloneJobKindContracts(jobContracts),
		}
		coordinator.orderedProfileIDs = append(coordinator.orderedProfileIDs, descriptor.ProfileID)
	}
	sort.Strings(coordinator.orderedProfileIDs)
	if len(descriptorDigests) != len(coordinator.profiles) || len(bindingDigests) != len(coordinator.profiles) {
		return nil, invalidArtifact("registry_integrity", errors.New("descriptor or binding digest set is incomplete or extra"))
	}
	inactivePolicies, err := admitInactiveConfigurationPolicies(source, coordinator.profiles, coordinator.orderedProfileIDs)
	if err != nil {
		return nil, err
	}
	coordinator.inactiveConfigurationPolicies = inactivePolicies
	if err := coordinator.validateCollisions(source, supportDigests["build.base-route-reservations"]); err != nil {
		return nil, err
	}
	_, clientSupportArtifact, err := readArtifactObject(source, generatedExtensionsRoot+"client-support-registry.json")
	if err != nil {
		return nil, err
	}
	if supportDigests["client-support-registry"] != clientSupportArtifact.SHA256 {
		return nil, invalidArtifact("client_support_registry", errors.New("integrity digest mismatch"))
	}
	coordinator.clientSupportSHA256 = clientSupportArtifact.SHA256
	validationRegistry, validationArtifact, err := readArtifactObject(source, generatedExtensionsRoot+"validation-condition-registry.json")
	if err != nil {
		return nil, err
	}
	if supportDigests["validation-condition-registry"] != validationArtifact.SHA256 {
		return nil, invalidArtifact("validation_condition_registry", errors.New("integrity digest mismatch"))
	}
	conditions, err := parseValidationConditions(validationRegistry)
	if err != nil {
		return nil, invalidArtifact("validation_condition_registry", err)
	}
	coordinator.validationConditions = conditions
	if err := coordinator.admitRuntimeCatalogs(source, supportDigests); err != nil {
		return nil, err
	}
	return coordinator, nil
}

func (c *Coordinator) RegistrySHA256() string {
	if c == nil {
		return ""
	}
	return c.registrySHA256
}

// WithClientSupportRegistrySHA256 returns an immutable coordinator view bound
// to the final packaged browser build rather than the source-level semantic
// projection used during contract generation.
func (c *Coordinator) WithClientSupportRegistrySHA256(digest string) (*Coordinator, error) {
	if c == nil || len(digest) != 64 {
		return nil, validationFailure(Finding{Code: "extension_registry_invalid", Phase: "registry_generation", Actual: "invalid_client_support_registry_digest"})
	}
	decoded, err := hex.DecodeString(digest)
	if err != nil || len(decoded) != sha256.Size || strings.ToLower(digest) != digest {
		return nil, validationFailure(Finding{Code: "extension_registry_invalid", Phase: "registry_generation", Actual: "invalid_client_support_registry_digest"})
	}
	bound := *c
	bound.clientSupportSHA256 = digest
	return &bound, nil
}

// ValidationCondition returns the exact generated condition row. Unknown local
// conditions are never converted into implementation-selected diagnostics.
func (c *Coordinator) ValidationCondition(conditionID string) (ValidationCondition, bool) {
	if c == nil {
		return ValidationCondition{}, false
	}
	condition, ok := c.validationConditions[conditionID]
	return condition, ok
}

func (c *Coordinator) Descriptors() []Descriptor {
	if c == nil {
		return nil
	}
	result := make([]Descriptor, 0, len(c.orderedProfileIDs))
	for _, profileID := range c.orderedProfileIDs {
		result = append(result, cloneDescriptor(c.profiles[profileID].descriptor))
	}
	return result
}

func (c *Coordinator) Descriptor(profileID string) (Descriptor, bool) {
	if c == nil {
		return Descriptor{}, false
	}
	record, ok := c.profiles[profileID]
	return cloneDescriptor(record.descriptor), ok
}

// ClaimResolution binds one explicit claim request to the admitted registry and
// dependency order. The unexported registry identity prevents callers from
// constructing a resolution for a different coordinator.
type ClaimResolution struct {
	claims         ResolvedClaimSet
	admissionOrder []string
	registrySHA256 string
}

func (r ClaimResolution) Claims() ResolvedClaimSet {
	claims, _ := NewResolvedClaimSet(r.claims.ProfileIDs())
	return claims
}

func (r ClaimResolution) AdmissionOrder() []string {
	return append([]string(nil), r.admissionOrder...)
}

// ResolveClaims validates only the explicit claim set. Required dependencies are
// never auto-claimed or substituted.
func (c *Coordinator) ResolveClaims(requestedProfileIDs []string) (ClaimResolution, error) {
	if c == nil {
		return ClaimResolution{}, validationFailure(Finding{Code: "extension_registry_invalid", Phase: "dependency_validation", Actual: "nil_coordinator"})
	}
	requested := make(map[string]struct{}, len(requestedProfileIDs))
	findings := []Finding{}
	for _, profileID := range requestedProfileIDs {
		if _, duplicate := requested[profileID]; duplicate {
			findings = append(findings, Finding{Code: "extension_registry_conflict", Phase: "claim_configuration", ProfileID: profileID, CollisionClass: "profile_id", ConflictingTokens: []string{profileID}})
			continue
		}
		requested[profileID] = struct{}{}
		record, recognized := c.profiles[profileID]
		if !recognized {
			findings = append(findings, Finding{Code: "extension_profile_unrecognized", Phase: "claim_configuration", ProfileID: profileID})
			continue
		}
		if !record.descriptor.Claimable {
			findings = append(findings, Finding{Code: "extension_profile_not_claimable", Phase: "claim_configuration", ProfileID: profileID})
		}
	}
	if len(findings) > 0 {
		return ClaimResolution{}, validationFailure(findings...)
	}

	for profileID := range requested {
		record := c.profiles[profileID]
		for _, dependency := range record.descriptor.Dependencies {
			dependencyRecord, recognized := c.profiles[dependency.ProfileID]
			if _, claimed := requested[dependency.ProfileID]; !claimed {
				findings = append(findings, Finding{Code: "extension_dependency_not_claimed", Phase: "dependency_validation", ProfileID: profileID, DependencyProfileID: dependency.ProfileID})
				continue
			}
			if !recognized || !dependencyRecord.descriptor.Claimable || dependencyRecord.descriptor.ContractMajor != dependency.RequiredContractMajor {
				findings = append(findings, Finding{Code: "extension_dependency_incompatible", Phase: "dependency_validation", ProfileID: profileID, DependencyProfileID: dependency.ProfileID, Expected: fmt.Sprint(dependency.RequiredContractMajor), Actual: fmt.Sprint(dependencyRecord.descriptor.ContractMajor)})
			}
		}
	}
	if len(findings) > 0 {
		sortFindings(findings)
		return ClaimResolution{}, validationFailure(findings...)
	}
	order, err := c.dependencyOrder(requested)
	if err != nil {
		return ClaimResolution{}, err
	}
	claims, err := NewResolvedClaimSet(requestedProfileIDs)
	if err != nil {
		return ClaimResolution{}, validationFailure(Finding{Code: "extension_profile_unrecognized", Phase: "claim_configuration"})
	}
	return ClaimResolution{claims: claims, admissionOrder: order, registrySHA256: c.registrySHA256}, nil
}

// RouteDispatch is the transport-neutral reservation/dispatch projection used by
// Core 01 during the later atomic publication switch.
type RouteDispatch struct {
	ProfileID      string
	RouteFamily    string
	DispatchState  string
	ContributionID *string
}

type WorkspacePublication struct {
	ProfileID      string
	WorkspaceKey   string
	ContributionID string
}

type WorkerPublication struct {
	ProfileID  string
	WorkerKind string
}

type ContributionPublication struct {
	ProfileID                   string
	ContributionID              string
	Kind                        string
	ContributionSHA256          string
	ImplementationBindingSHA256 string
	RouteFamily                 string
	WorkspaceKey                string
	ParticipantID               string
}

type JobResourceRefContract struct {
	ResourceRefKind    string
	ResourceIDSchemaID string
	MaxRefs            int
}

type JobKindContract struct {
	ProfileID                   string
	JobKind                     string
	OperationKind               string
	ProofPolicy                 string
	IdempotencyPolicy           string
	IdempotencyIdentitySchemaID string
	TerminalResultSchemaID      string
	ResourceRefContracts        []JobResourceRefContract
	CancellationPolicy          string
	MaxProofBytes               int
}

func (contract JobKindContract) SHA256() string {
	return canonicalDigest(contract.object())
}

func (contract JobKindContract) object() map[string]any {
	resourceRefs := make([]any, len(contract.ResourceRefContracts))
	for index, resourceRef := range contract.ResourceRefContracts {
		resourceRefs[index] = map[string]any{
			"resource_ref_kind":     resourceRef.ResourceRefKind,
			"resource_id_schema_id": resourceRef.ResourceIDSchemaID,
			"max_refs":              resourceRef.MaxRefs,
		}
	}
	return map[string]any{
		"schema_id":                      "cartulary.extension_job_kind_contract.v1",
		"profile_id":                     contract.ProfileID,
		"job_kind":                       contract.JobKind,
		"operation_kind":                 contract.OperationKind,
		"proof_policy":                   contract.ProofPolicy,
		"idempotency_policy":             contract.IdempotencyPolicy,
		"idempotency_identity_schema_id": contract.IdempotencyIdentitySchemaID,
		"terminal_result_schema_id":      contract.TerminalResultSchemaID,
		"resource_ref_contracts":         resourceRefs,
		"cancellation_policy":            contract.CancellationPolicy,
		"max_proof_bytes":                contract.MaxProofBytes,
	}
}

func cloneJobKindContracts(source []JobKindContract) []JobKindContract {
	result := make([]JobKindContract, len(source))
	for index, contract := range source {
		contract.ResourceRefContracts = append([]JobResourceRefContract(nil), contract.ResourceRefContracts...)
		result[index] = contract
	}
	return result
}

// JobKindContracts returns the exact generated job catalog in profile and job
// identity order. Internal ownership metadata is not part of the public job
// resource and remains confined to this application-facing contract.
func (c *Coordinator) JobKindContracts() []JobKindContract {
	if c == nil {
		return nil
	}
	result := []JobKindContract{}
	for _, profileID := range c.orderedProfileIDs {
		result = append(result, cloneJobKindContracts(c.profiles[profileID].jobContracts)...)
	}
	return result
}

type DiscoveryWorkspace struct {
	WorkspaceKey string
	MinimumRole  string
}

type DiscoveryProfile struct {
	ProfileID     string
	Claimable     bool
	Claimed       bool
	ContractMajor *int
	RouteFamilies []string
	WorkspaceKeys []string
	Capabilities  []string
	Workspaces    []DiscoveryWorkspace
}

type ClaimPublication struct {
	ProfileID string
	Claimed   bool
}

type ListenerPublication struct {
	ComponentID string
}

type ImplementationBindingPublication struct {
	ProfileID     string
	BindingSHA256 string
}

// PublicationPlanSummary contains only canonical component identities. Component
// rows are available through copy-returning methods on PublicationPlan.
type PublicationPlanSummary struct {
	SchemaID                       string
	RegistrySHA256                 string
	ResolvedClaimSetSHA256         string
	ContributionRegistrySHA256     string
	RouteDispatchPlanSHA256        string
	WorkspaceRegistrySHA256        string
	WorkerPlanSHA256               string
	ListenerPlanSHA256             string
	ClientSupportRegistrySHA256    string
	ImplementationBindingSetSHA256 string
}

type PublicationPlan struct {
	summary                PublicationPlanSummary
	contributions          []ContributionPublication
	discovery              []DiscoveryProfile
	claims                 []ClaimPublication
	routes                 []RouteDispatch
	workspaces             []WorkspacePublication
	workers                []WorkerPublication
	jobContracts           []JobKindContract
	listeners              []ListenerPublication
	implementationBindings []ImplementationBindingPublication
	resolvedClaims         ResolvedClaimSet
}

func (p PublicationPlan) Summary() PublicationPlanSummary { return p.summary }

func (p PublicationPlan) Contributions() []ContributionPublication {
	return append([]ContributionPublication(nil), p.contributions...)
}

func (p PublicationPlan) Routes() []RouteDispatch {
	result := append([]RouteDispatch(nil), p.routes...)
	for index := range result {
		if result[index].ContributionID != nil {
			value := *result[index].ContributionID
			result[index].ContributionID = &value
		}
	}
	return result
}

func (p PublicationPlan) Workspaces() []WorkspacePublication {
	return append([]WorkspacePublication(nil), p.workspaces...)
}

func (p PublicationPlan) Workers() []WorkerPublication {
	return append([]WorkerPublication(nil), p.workers...)
}

func (p PublicationPlan) JobKindContracts() []JobKindContract {
	return cloneJobKindContracts(p.jobContracts)
}

func (p PublicationPlan) Discovery() []DiscoveryProfile {
	result := make([]DiscoveryProfile, len(p.discovery))
	for index, profile := range p.discovery {
		result[index] = DiscoveryProfile{
			ProfileID:     profile.ProfileID,
			Claimable:     profile.Claimable,
			Claimed:       profile.Claimed,
			ContractMajor: cloneIntPointer(profile.ContractMajor),
			RouteFamilies: append([]string(nil), profile.RouteFamilies...),
			WorkspaceKeys: append([]string(nil), profile.WorkspaceKeys...),
			Capabilities:  append([]string(nil), profile.Capabilities...),
			Workspaces:    append([]DiscoveryWorkspace(nil), profile.Workspaces...),
		}
	}
	return result
}

func (p PublicationPlan) Claims() []ClaimPublication {
	return append([]ClaimPublication(nil), p.claims...)
}

func (p PublicationPlan) Listeners() []ListenerPublication {
	return append([]ListenerPublication(nil), p.listeners...)
}

func (p PublicationPlan) ImplementationBindings() []ImplementationBindingPublication {
	return append([]ImplementationBindingPublication(nil), p.implementationBindings...)
}

func (p PublicationPlan) ResolvedClaims() ResolvedClaimSet {
	claims, _ := NewResolvedClaimSet(p.resolvedClaims.ProfileIDs())
	return claims
}

func cloneIntPointer(value *int) *int {
	if value == nil {
		return nil
	}
	cloned := *value
	return &cloned
}

// BuildPublicationPlan derives all Stage 6 component identities from one admitted
// resolution. It starts no listener or worker and mutates no external state.
func (c *Coordinator) BuildPublicationPlan(resolution ClaimResolution) (PublicationPlan, error) {
	if c == nil || resolution.registrySHA256 == "" || resolution.registrySHA256 != c.registrySHA256 {
		return PublicationPlan{}, validationFailure(Finding{Code: "extension_publication_failed", Phase: "publication", Actual: "claim_resolution_registry_mismatch"})
	}
	claimed := make(map[string]struct{}, len(resolution.claims.profileIDs))
	for _, profileID := range resolution.claims.profileIDs {
		if _, exists := c.profiles[profileID]; !exists {
			return PublicationPlan{}, validationFailure(Finding{Code: "extension_publication_failed", Phase: "publication", ProfileID: profileID, Actual: "unrecognized_claim"})
		}
		claimed[profileID] = struct{}{}
	}

	contributionItems := []any{}
	contributionPublications := []ContributionPublication{}
	routeItems := []any{}
	workspaceItems := []any{}
	workerItems := []any{}
	bindingItems := []any{}
	routes := []RouteDispatch{}
	workspaces := []WorkspacePublication{}
	workers := []WorkerPublication{}
	jobContracts := []JobKindContract{}
	discovery := []DiscoveryProfile{}
	claims := []ClaimPublication{}
	implementationBindings := []ImplementationBindingPublication{}
	for _, profileID := range c.orderedProfileIDs {
		record := c.profiles[profileID]
		_, isClaimed := claimed[profileID]
		contractMajor := record.descriptor.ContractMajor
		discoveryWorkspaces := make([]DiscoveryWorkspace, 0, len(record.descriptor.WorkspaceKeys))
		for _, workspaceKey := range record.descriptor.WorkspaceKeys {
			discoveryWorkspaces = append(discoveryWorkspaces, DiscoveryWorkspace{WorkspaceKey: workspaceKey, MinimumRole: "viewer"})
		}
		discovery = append(discovery, DiscoveryProfile{
			ProfileID: profileID, Claimable: record.descriptor.Claimable, Claimed: isClaimed,
			ContractMajor: &contractMajor, RouteFamilies: append([]string(nil), record.descriptor.RouteFamilies...),
			WorkspaceKeys: append([]string(nil), record.descriptor.WorkspaceKeys...),
			Capabilities:  append([]string(nil), record.descriptor.CapabilityIDs...), Workspaces: discoveryWorkspaces,
		})
		claims = append(claims, ClaimPublication{ProfileID: profileID, Claimed: isClaimed})
		contributions, _ := objectSlice(record.descriptorObject["contributions"])
		contributionByRoute := map[string]string{}
		contributionByWorkspace := map[string]string{}
		for _, contribution := range contributions {
			kind := stringValue(contribution["kind"])
			contributionID := stringValue(contribution["contribution_id"])
			if kind == "http_route_family" {
				contributionByRoute[stringValue(contribution["route_family"])] = contributionID
			}
			if kind == "incident_workspace" {
				contributionByWorkspace[stringValue(contribution["workspace_key"])] = contributionID
			}
			if isClaimed {
				contributionDigest := canonicalDigest(contribution)
				contributionItems = append(contributionItems, map[string]any{
					"profile_id": profileID, "contribution_id": contributionID, "kind": kind,
					"contribution_sha256": contributionDigest, "implementation_binding_sha256": record.bindingSHA256,
				})
				contributionPublications = append(contributionPublications, ContributionPublication{
					ProfileID:                   profileID,
					ContributionID:              contributionID,
					Kind:                        kind,
					ContributionSHA256:          contributionDigest,
					ImplementationBindingSHA256: record.bindingSHA256,
					RouteFamily:                 stringValue(contribution["route_family"]),
					WorkspaceKey:                stringValue(contribution["workspace_key"]),
					ParticipantID:               stringValue(contribution["participant_id"]),
				})
			}
		}
		for _, routeFamily := range record.descriptor.RouteFamilies {
			state := "inactive"
			var contributionID any
			var contributionPointer *string
			if isClaimed {
				state = "claimed"
				value := contributionByRoute[routeFamily]
				contributionID = value
				contributionPointer = &value
			}
			routeItems = append(routeItems, map[string]any{"profile_id": profileID, "route_family": routeFamily, "dispatch_state": state, "contribution_id": contributionID})
			routes = append(routes, RouteDispatch{ProfileID: profileID, RouteFamily: routeFamily, DispatchState: state, ContributionID: contributionPointer})
		}
		if !isClaimed {
			continue
		}
		for _, workspaceKey := range record.descriptor.WorkspaceKeys {
			contributionID := contributionByWorkspace[workspaceKey]
			workspaceItems = append(workspaceItems, map[string]any{"profile_id": profileID, "workspace_key": workspaceKey, "contribution_id": contributionID})
			workspaces = append(workspaces, WorkspacePublication{ProfileID: profileID, WorkspaceKey: workspaceKey, ContributionID: contributionID})
		}
		workerKinds, _ := stringSlice(record.bindingObject["worker_kinds"])
		for _, workerKind := range workerKinds {
			workerItems = append(workerItems, map[string]any{"profile_id": profileID, "worker_kind": workerKind})
			workers = append(workers, WorkerPublication{ProfileID: profileID, WorkerKind: workerKind})
		}
		jobContracts = append(jobContracts, cloneJobKindContracts(record.jobContracts)...)
		bindingItems = append(bindingItems, map[string]any{"profile_id": profileID, "binding_sha256": record.bindingSHA256})
		implementationBindings = append(implementationBindings, ImplementationBindingPublication{ProfileID: profileID, BindingSHA256: record.bindingSHA256})
	}
	sort.Slice(routeItems, func(i, j int) bool {
		left, right := routeItems[i].(map[string]any), routeItems[j].(map[string]any)
		return stringValue(left["route_family"])+"\x00"+stringValue(left["profile_id"]) < stringValue(right["route_family"])+"\x00"+stringValue(right["profile_id"])
	})
	sort.Slice(routes, func(i, j int) bool {
		return routes[i].RouteFamily+"\x00"+routes[i].ProfileID < routes[j].RouteFamily+"\x00"+routes[j].ProfileID
	})
	sort.Slice(contributionPublications, func(i, j int) bool {
		return contributionPublications[i].ProfileID+"\x00"+contributionPublications[i].ContributionID <
			contributionPublications[j].ProfileID+"\x00"+contributionPublications[j].ContributionID
	})
	sort.Slice(jobContracts, func(i, j int) bool {
		return jobContracts[i].ProfileID+"\x00"+jobContracts[i].JobKind <
			jobContracts[j].ProfileID+"\x00"+jobContracts[j].JobKind
	})

	componentDigests := map[string]string{
		"contribution_registry_sha256":      canonicalDigest(map[string]any{"schema_id": "cartulary.extension_contribution_publication_set.v1", "items": contributionItems}),
		"route_dispatch_plan_sha256":        canonicalDigest(map[string]any{"schema_id": "cartulary.extension_route_dispatch_plan.v1", "routes": routeItems}),
		"workspace_registry_sha256":         canonicalDigest(map[string]any{"schema_id": "cartulary.extension_workspace_registry.v1", "workspaces": workspaceItems}),
		"worker_plan_sha256":                canonicalDigest(map[string]any{"schema_id": "cartulary.extension_worker_plan.v1", "workers": workerItems}),
		"listener_plan_sha256":              canonicalDigest(map[string]any{"schema_id": "cartulary.extension_listener_activation_plan.v1", "http": true, "websocket": true, "job_dequeue": true}),
		"implementation_binding_set_sha256": canonicalDigest(map[string]any{"schema_id": "cartulary.extension_implementation_binding_set.v1", "bindings": bindingItems}),
	}
	document := map[string]any{
		"schema_id": "cartulary.extension_publication_plan.v1", "registry_sha256": c.registrySHA256,
		"resolved_claim_set_sha256": resolution.claims.SHA256(), "client_support_registry_sha256": c.clientSupportSHA256,
	}
	for key, value := range componentDigests {
		document[key] = value
	}
	canonical, err := canonicalJSON(document, true)
	if err != nil || len(canonical) > 16777216 {
		return PublicationPlan{}, validationFailure(Finding{Code: "extension_registry_limit_exceeded", Phase: "publication", Actual: fmt.Sprint(len(canonical))})
	}
	return PublicationPlan{
		summary: PublicationPlanSummary{
			SchemaID: "cartulary.extension_publication_plan.v1", RegistrySHA256: c.registrySHA256,
			ResolvedClaimSetSHA256: resolution.claims.SHA256(), ContributionRegistrySHA256: componentDigests["contribution_registry_sha256"],
			RouteDispatchPlanSHA256: componentDigests["route_dispatch_plan_sha256"], WorkspaceRegistrySHA256: componentDigests["workspace_registry_sha256"],
			WorkerPlanSHA256: componentDigests["worker_plan_sha256"], ListenerPlanSHA256: componentDigests["listener_plan_sha256"],
			ClientSupportRegistrySHA256: c.clientSupportSHA256, ImplementationBindingSetSHA256: componentDigests["implementation_binding_set_sha256"],
		},
		contributions: contributionPublications, discovery: discovery, claims: claims, routes: routes,
		workspaces: workspaces, workers: workers, jobContracts: jobContracts,
		listeners:              []ListenerPublication{{ComponentID: "http"}, {ComponentID: "job_dequeue"}, {ComponentID: "websocket"}},
		implementationBindings: implementationBindings, resolvedClaims: resolution.Claims(),
	}, nil
}

func (c *Coordinator) dependencyOrder(requested map[string]struct{}) ([]string, error) {
	unresolved := make(map[string]map[string]struct{}, len(requested))
	for profileID := range requested {
		unresolved[profileID] = map[string]struct{}{}
		for _, dependency := range c.profiles[profileID].descriptor.Dependencies {
			unresolved[profileID][dependency.ProfileID] = struct{}{}
		}
	}
	order := make([]string, 0, len(requested))
	for len(order) < len(requested) {
		eligible := []string{}
		for profileID, dependencies := range unresolved {
			if dependencies != nil && len(dependencies) == 0 {
				eligible = append(eligible, profileID)
			}
		}
		if len(eligible) == 0 {
			cycle := []string{}
			for profileID, dependencies := range unresolved {
				for dependencyID := range dependencies {
					cycle = append(cycle, profileID+"->"+dependencyID)
				}
			}
			sort.Strings(cycle)
			return nil, collisionFailure("dependency_cycle", cycle...)
		}
		sort.Strings(eligible)
		next := eligible[0]
		order = append(order, next)
		unresolved[next] = nil
		for _, dependencies := range unresolved {
			delete(dependencies, next)
		}
	}
	return order, nil
}

func (c *Coordinator) validateCollisions(source ArtifactSource, expectedBaseReservationDigest string) error {
	claimKeys := map[string]string{}
	contributionIDs := map[string]string{}
	publicSchemas := map[string]string{}
	routeOwners := map[string]string{}
	for _, profileID := range c.orderedProfileIDs {
		record := c.profiles[profileID]
		if prior := claimKeys[record.descriptor.ClaimConfigKey]; prior != "" {
			return collisionFailure("claim_key", record.descriptor.ClaimConfigKey)
		}
		claimKeys[record.descriptor.ClaimConfigKey] = profileID
		for _, raw := range record.descriptorObject["public_schema_ids"].([]any) {
			schemaID := stringValue(raw)
			if publicSchemas[schemaID] != "" {
				return collisionFailure("public_schema_id", schemaID)
			}
			publicSchemas[schemaID] = profileID
		}
		contributions, _ := objectSlice(record.descriptorObject["contributions"])
		for _, contribution := range contributions {
			contributionID := stringValue(contribution["contribution_id"])
			if contributionIDs[contributionID] != "" {
				return collisionFailure("contribution_id", contributionID)
			}
			contributionIDs[contributionID] = profileID
		}
		for _, route := range record.descriptor.RouteFamilies {
			for existing, owner := range routeOwners {
				if routeFamiliesOverlap(route, existing) {
					return validationFailure(Finding{Code: "extension_registry_conflict", Phase: "registry_collision", CollisionClass: "route_family_overlap", ProfileID: profileID, ConflictingTokens: sortedStrings(route, existing), Actual: owner})
				}
			}
			routeOwners[route] = profileID
		}
	}
	base, baseArtifact, err := readArtifactObject(source, "contracts/extensions/build/base-route-reservations.json")
	if err != nil {
		return err
	}
	if expectedBaseReservationDigest == "" || expectedBaseReservationDigest != baseArtifact.SHA256 {
		return invalidArtifact("base_route_reservation_registry", errors.New("integrity digest mismatch"))
	}
	reservations, ok := objectSlice(base["reservations"])
	if !ok {
		return invalidArtifact("base_route_reservation_registry", errors.New("reservations must be an array"))
	}
	for route := range routeOwners {
		for _, reservation := range reservations {
			baseRoute := stringValue(reservation["path_template"])
			if baseReservationOverlaps(route, baseRoute, stringValue(reservation["match_scope"])) {
				return collisionFailure("base_route_capture", sortedStrings(route, baseRoute)...)
			}
		}
	}
	return nil
}

func parseDescriptor(object map[string]any) (Descriptor, error) {
	if err := requireExactKeys(object, "schema_id", "profile_id", "claimable", "contract_major", "owner_id", "claim_config_key", "route_families", "workspace_keys", "capability_ids", "runtime_dependencies", "contributions", "public_schema_ids", "prestage_config_keys", "state_ownership", "admission_validation", "egress_mode", "incident_portability_mode", "snapshot_reporting_mode", "conformance_manifest_id"); err != nil {
		return Descriptor{}, err
	}
	profileID := stringValue(object["profile_id"])
	major, ok := integerValue(object["contract_major"])
	claimable, boolOK := object["claimable"].(bool)
	routes, routesOK := stringSlice(object["route_families"])
	workspaces, workspacesOK := stringSlice(object["workspace_keys"])
	capabilities, capabilitiesOK := stringSlice(object["capability_ids"])
	dependenciesObjects, dependenciesOK := objectSlice(object["runtime_dependencies"])
	if object["schema_id"] != "cartulary.extension_profile_descriptor.v2" || !extensionProfileIDPattern.MatchString(profileID) || !ok || major < 1 || !boolOK || !routesOK || !workspacesOK || !capabilitiesOK || !dependenciesOK || len(capabilities) != 0 {
		return Descriptor{}, errors.New("descriptor identity, scalar, collection, or capability contract is invalid")
	}
	dependencies := make([]Dependency, 0, len(dependenciesObjects))
	seen := map[string]struct{}{}
	for _, dependencyObject := range dependenciesObjects {
		if err := requireExactKeys(dependencyObject, "profile_id", "required_contract_major"); err != nil {
			return Descriptor{}, err
		}
		dependencyMajor, integerOK := integerValue(dependencyObject["required_contract_major"])
		dependencyID := stringValue(dependencyObject["profile_id"])
		if !integerOK || dependencyMajor < 1 || dependencyID == profileID {
			return Descriptor{}, errors.New("dependency is malformed or self-referential")
		}
		if _, duplicate := seen[dependencyID]; duplicate {
			return Descriptor{}, errors.New("dependency is duplicated")
		}
		seen[dependencyID] = struct{}{}
		dependencies = append(dependencies, Dependency{ProfileID: dependencyID, RequiredContractMajor: dependencyMajor})
	}
	return Descriptor{
		ProfileID: profileID, Claimable: claimable, ContractMajor: major,
		OwnerID: stringValue(object["owner_id"]), ClaimConfigKey: stringValue(object["claim_config_key"]),
		RouteFamilies: routes, WorkspaceKeys: workspaces, CapabilityIDs: capabilities, Dependencies: dependencies,
	}, nil
}

func validateBinding(descriptor Descriptor, descriptorDigest string, binding map[string]any) error {
	requiredKeys := []string{"schema_id", "profile_id", "contract_major", "descriptor_sha256", "implemented_contribution_ids", "supported_capability_ids", "state_ownership_kind", "preflight_algorithm_id", "post_migration_algorithm_id", "initialization_definition_sha256", "initialization_algorithm_id", "final_state_validation_algorithm_id", "dependency_probe_ids", "migration_definitions", "physical_state_binding_sha256", "backup_codec_bindings", "rebuild_algorithm_ids", "transaction_participant_limits", "supporting_schema_ids", "worker_kinds", "job_kind_contracts", "participant_contracts"}
	if err := requireExactKeys(binding, requiredKeys...); err != nil {
		return err
	}
	major, ok := integerValue(binding["contract_major"])
	capabilities, capabilitiesOK := stringSlice(binding["supported_capability_ids"])
	if binding["schema_id"] != "cartulary.extension_implementation_binding.v1" || stringValue(binding["profile_id"]) != descriptor.ProfileID || !ok || major != descriptor.ContractMajor || stringValue(binding["descriptor_sha256"]) != descriptorDigest || !capabilitiesOK || len(capabilities) != 0 {
		return errors.New("binding identity, major, descriptor digest, or capability parity is invalid")
	}
	for _, key := range []string{"implemented_contribution_ids", "dependency_probe_ids", "rebuild_algorithm_ids", "supporting_schema_ids", "worker_kinds"} {
		values, valid := stringSlice(binding[key])
		if !valid || !strictlySortedUnique(values) {
			return fmt.Errorf("binding %s is malformed", key)
		}
	}
	if _, err := parseJobKindContracts(descriptor.ProfileID, binding["job_kind_contracts"]); err != nil {
		return err
	}
	return nil
}

func parseJobKindContracts(profileID string, value any) ([]JobKindContract, error) {
	rows, ok := objectSlice(value)
	if !ok || len(rows) > 256 {
		return nil, errors.New("binding job_kind_contracts is malformed")
	}
	result := make([]JobKindContract, len(rows))
	previous := ""
	for index, row := range rows {
		if err := requireExactKeys(row,
			"schema_id", "profile_id", "job_kind", "operation_kind", "proof_policy",
			"idempotency_policy", "idempotency_identity_schema_id", "terminal_result_schema_id",
			"resource_ref_contracts", "cancellation_policy", "max_proof_bytes",
		); err != nil {
			return nil, err
		}
		contract := JobKindContract{
			ProfileID:                   stringValue(row["profile_id"]),
			JobKind:                     stringValue(row["job_kind"]),
			OperationKind:               stringValue(row["operation_kind"]),
			ProofPolicy:                 stringValue(row["proof_policy"]),
			IdempotencyPolicy:           stringValue(row["idempotency_policy"]),
			IdempotencyIdentitySchemaID: stringValue(row["idempotency_identity_schema_id"]),
			TerminalResultSchemaID:      stringValue(row["terminal_result_schema_id"]),
			CancellationPolicy:          stringValue(row["cancellation_policy"]),
			MaxProofBytes:               intValue(row["max_proof_bytes"]),
		}
		if row["schema_id"] != "cartulary.extension_job_kind_contract.v1" ||
			contract.ProfileID != profileID ||
			!strings.HasPrefix(contract.JobKind, profileID+".") ||
			strings.Contains(contract.JobKind, "/") ||
			strings.Contains(contract.JobKind, `\`) ||
			(previous != "" && previous >= contract.JobKind) ||
			!strings.HasPrefix(contract.OperationKind, profileID+".") ||
			contract.ProofPolicy != "required_on_terminal_success" ||
			contract.IdempotencyPolicy != "required" ||
			contract.IdempotencyIdentitySchemaID != "cartulary.route_scoped_idempotency_identity.v1" ||
			contract.TerminalResultSchemaID != "cartulary.common_job_terminal_success.v1" ||
			contract.CancellationPolicy != "precommit_observable" ||
			contract.MaxProofBytes < 1 || contract.MaxProofBytes > 1048576 {
			return nil, errors.New("binding job kind contract is invalid")
		}
		previous = contract.JobKind
		resourceRefs, ok := objectSlice(row["resource_ref_contracts"])
		if !ok || len(resourceRefs) > 64 {
			return nil, errors.New("binding job resource references are malformed")
		}
		previousRef := ""
		for _, resourceRef := range resourceRefs {
			if err := requireExactKeys(resourceRef, "resource_ref_kind", "resource_id_schema_id", "max_refs"); err != nil {
				return nil, err
			}
			parsed := JobResourceRefContract{
				ResourceRefKind:    stringValue(resourceRef["resource_ref_kind"]),
				ResourceIDSchemaID: stringValue(resourceRef["resource_id_schema_id"]),
				MaxRefs:            intValue(resourceRef["max_refs"]),
			}
			if parsed.ResourceRefKind == "" ||
				(previousRef != "" && previousRef >= parsed.ResourceRefKind) ||
				parsed.ResourceIDSchemaID != "cartulary.common_job_resource_ref_id.v1" ||
				parsed.MaxRefs < 1 || parsed.MaxRefs > 1024 {
				return nil, errors.New("binding job resource reference is invalid")
			}
			previousRef = parsed.ResourceRefKind
			contract.ResourceRefContracts = append(contract.ResourceRefContracts, parsed)
		}
		result[index] = contract
	}
	return result, nil
}

func readArtifactObject(source ArtifactSource, path string) (map[string]any, PackagedArtifact, error) {
	artifact, ok := source.Artifact(path)
	if !ok {
		return nil, PackagedArtifact{}, validationFailure(Finding{Code: "extension_registry_invalid", Phase: "registry_generation", Actual: "missing:" + path})
	}
	artifact.JSON = append([]byte(nil), artifact.JSON...)
	digest := sha256.Sum256(artifact.JSON)
	actualDigest := hex.EncodeToString(digest[:])
	if artifact.SHA256 != actualDigest {
		return nil, PackagedArtifact{}, validationFailure(Finding{Code: "extension_registry_invalid", Phase: "registry_generation", Expected: artifact.SHA256, Actual: actualDigest})
	}
	decoder := json.NewDecoder(bytes.NewReader(artifact.JSON))
	decoder.UseNumber()
	var object map[string]any
	if err := decoder.Decode(&object); err != nil || object == nil {
		return nil, PackagedArtifact{}, invalidArtifact(path, err)
	}
	var extra any
	if err := decoder.Decode(&extra); err == nil {
		return nil, PackagedArtifact{}, invalidArtifact(path, errors.New("multiple JSON values"))
	}
	return object, artifact, nil
}

func digestRows(value any, identityKey, digestKey string) (map[string]string, error) {
	rows, ok := objectSlice(value)
	if !ok {
		return nil, errors.New("digest rows must be an array")
	}
	result := make(map[string]string, len(rows))
	for _, row := range rows {
		identity, digest := stringValue(row[identityKey]), stringValue(row[digestKey])
		if identity == "" || len(digest) != 64 || result[identity] != "" {
			return nil, errors.New("digest row is malformed or duplicated")
		}
		result[identity] = digest
	}
	return result, nil
}

func requireExactKeys(object map[string]any, keys ...string) error {
	if len(object) != len(keys) {
		return fmt.Errorf("object has %d members; want %d", len(object), len(keys))
	}
	for _, key := range keys {
		if _, ok := object[key]; !ok {
			return fmt.Errorf("object omits %s", key)
		}
	}
	return nil
}

func objectSlice(value any) ([]map[string]any, bool) {
	values, ok := value.([]any)
	if !ok {
		return nil, false
	}
	result := make([]map[string]any, 0, len(values))
	for _, value := range values {
		object, objectOK := value.(map[string]any)
		if !objectOK {
			return nil, false
		}
		result = append(result, object)
	}
	return result, true
}

func stringSlice(value any) ([]string, bool) {
	values, ok := value.([]any)
	if !ok {
		return nil, false
	}
	result := make([]string, 0, len(values))
	for _, value := range values {
		text, textOK := value.(string)
		if !textOK {
			return nil, false
		}
		result = append(result, text)
	}
	return result, true
}

func integerValue(value any) (int, bool) {
	number, ok := value.(json.Number)
	if !ok {
		return 0, false
	}
	integer, err := number.Int64()
	return int(integer), err == nil && int64(int(integer)) == integer
}

func stringValue(value any) string {
	text, _ := value.(string)
	return text
}

func canonicalJSON(value any, finalLF bool) ([]byte, error) {
	encoded, err := json.Marshal(value)
	if err != nil {
		return nil, err
	}
	if finalLF {
		encoded = append(encoded, '\n')
	}
	return encoded, nil
}

func canonicalDigest(value any) string {
	encoded, _ := canonicalJSON(value, true)
	digest := sha256.Sum256(encoded)
	return hex.EncodeToString(digest[:])
}

func equalCanonicalObjects(left, right map[string]any) bool {
	leftJSON, leftErr := canonicalJSON(left, false)
	rightJSON, rightErr := canonicalJSON(right, false)
	return leftErr == nil && rightErr == nil && bytes.Equal(leftJSON, rightJSON)
}

func cloneObject(source map[string]any) map[string]any {
	encoded, _ := json.Marshal(source)
	decoder := json.NewDecoder(bytes.NewReader(encoded))
	decoder.UseNumber()
	var clone map[string]any
	_ = decoder.Decode(&clone)
	return clone
}

func routeFamiliesOverlap(left, right string) bool {
	leftSegments, leftOK := routeSegments(left)
	rightSegments, rightOK := routeSegments(right)
	if !leftOK || !rightOK {
		return false
	}
	limit := len(leftSegments)
	if len(rightSegments) < limit {
		limit = len(rightSegments)
	}
	for index := 0; index < limit; index++ {
		leftParameter := strings.HasPrefix(leftSegments[index], "{")
		rightParameter := strings.HasPrefix(rightSegments[index], "{")
		if !leftParameter && !rightParameter && leftSegments[index] != rightSegments[index] {
			return false
		}
	}
	return true
}

func baseReservationOverlaps(extensionRoute, baseRoute, baseScope string) bool {
	extensionSegments, extensionOK := routeSegments(extensionRoute)
	baseSegments, baseOK := routeSegments(baseRoute)
	if !extensionOK || !baseOK || (baseScope != "exact" && baseScope != "descendants") {
		return false
	}
	if baseScope == "exact" && len(extensionSegments) != len(baseSegments) {
		return false
	}
	limit := len(baseSegments)
	if len(extensionSegments) < limit {
		if baseScope == "exact" {
			return false
		}
		limit = len(extensionSegments)
	}
	for index := 0; index < limit; index++ {
		extensionParameter := strings.HasPrefix(extensionSegments[index], "{")
		baseParameter := strings.HasPrefix(baseSegments[index], "{")
		if !extensionParameter && !baseParameter && extensionSegments[index] != baseSegments[index] {
			return false
		}
	}
	return baseScope == "descendants" || len(extensionSegments) == len(baseSegments)
}

func routeSegments(route string) ([]string, bool) {
	if !strings.HasPrefix(route, "/api/v1/") || strings.HasSuffix(route, "/") || strings.ContainsAny(route, "?#%\\\x00") {
		return nil, false
	}
	segments := strings.Split(strings.TrimPrefix(route, "/"), "/")
	if len(segments) > 32 {
		return nil, false
	}
	for _, segment := range segments {
		if segment == "" || segment == "." || segment == ".." {
			return nil, false
		}
	}
	return segments, true
}

func strictlySortedUnique(values []string) bool {
	for index := 1; index < len(values); index++ {
		if values[index-1] >= values[index] {
			return false
		}
	}
	return true
}

func sortedStrings(values ...string) []string {
	result := append([]string(nil), values...)
	sort.Strings(result)
	return result
}

func sortFindings(findings []Finding) {
	sort.Slice(findings, func(i, j int) bool {
		left := findings[i].Code + "\x00" + findings[i].ProfileID + "\x00" + findings[i].DependencyProfileID
		right := findings[j].Code + "\x00" + findings[j].ProfileID + "\x00" + findings[j].DependencyProfileID
		return left < right
	})
}

func validationFailure(findings ...Finding) error {
	for index := range findings {
		findings[index].ConflictingTokens = append([]string(nil), findings[index].ConflictingTokens...)
	}
	return &ValidationError{findings: findings}
}

func invalidArtifact(kind string, cause error) error {
	actual := "invalid"
	if cause != nil {
		actual = cause.Error()
	}
	return validationFailure(Finding{Code: "extension_registry_invalid", Phase: "registry_generation", Expected: kind, Actual: actual})
}

func collisionFailure(class string, tokens ...string) error {
	return validationFailure(Finding{Code: "extension_registry_conflict", Phase: "registry_collision", CollisionClass: class, ConflictingTokens: sortedStrings(tokens...)})
}

func unavailableBinding(profileID, actual string) error {
	return validationFailure(Finding{Code: "extension_implementation_unavailable", Phase: "implementation_binding", ProfileID: profileID, Actual: actual})
}
