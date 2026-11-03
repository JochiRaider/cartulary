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

	contractsgen "github.com/JochiRaider/cartulary/internal/gen/contractextensions"
	"github.com/JochiRaider/cartulary/internal/platform/canonicaljson"
	"github.com/JochiRaider/cartulary/internal/platform/jobs/jobcontract"
)

const generatedExtensionsRoot = "contracts/extensions/generated/"

// PackagedArtifact is an immutable view of one build-time contract artifact.
// The byte slice returned to or received from callers is always copied.
type PackagedArtifact struct {
	JSON   []byte
	SHA256 string
}

// ArtifactSource is the coordinator's only build-package port. It deliberately
// exposes an explicit build inventory and immutable root, with no filesystem,
// discovery, transport, or profile callbacks.
type ArtifactSource interface {
	Artifact(path string) (PackagedArtifact, bool)
	ArtifactPaths() []string
	IntegritySHA256() string
}

type generatedArtifactSource struct{}

func (generatedArtifactSource) Artifact(path string) (PackagedArtifact, bool) {
	artifact, ok := contractsgen.Index[path]
	if !ok {
		return PackagedArtifact{}, false
	}
	return PackagedArtifact{JSON: []byte(artifact.JSON), SHA256: artifact.SHA256}, true
}

func (generatedArtifactSource) ArtifactPaths() []string {
	return append([]string(nil), contractsgen.RuntimeArtifactPaths...)
}

func (generatedArtifactSource) IntegritySHA256() string { return contractsgen.RegistryIntegritySHA256 }

// Finding is a safe coordination failure. It contains identities and contract
// tokens only; profile configuration values and implementation details never enter
// the coordinator.
type Finding struct {
	Code                  string
	Phase                 string
	ProfileID             string
	DependencyProfileID   string
	CollisionClass        string
	ConflictingTokens     []string
	ConflictingProfileIDs []string
	DuplicateCount        int
	Expected              string
	Actual                string
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
		result[index].ConflictingProfileIDs = append([]string(nil), result[index].ConflictingProfileIDs...)
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
	ProfileID string
	Claimable bool
	// ContractMajor is zero only for the unclaimable variant. Its canonical
	// descriptor and discovery representations use JSON null for that variant.
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
	configurationContract       map[string]any
	configurationContractSHA256 string
	descriptor                  Descriptor
	descriptorObject            map[string]any
	descriptorSHA256            string
	bindingObject               map[string]any
	bindingSHA256               string
	jobContracts                []JobKindContract
	workerRuntimeContracts      []WorkerPublication
}

// Coordinator is an immutable coordination facade over generated build inputs.
// Construction performs registry, collision, and binding admission once.
type Coordinator struct {
	transactionLimits             TransactionLimits
	validateFinding               func(any) error
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

// NewPackagedCoordinator admits the final browser package together with the
// generated Extensions catalog. Serving composition must use this constructor.
func NewPackagedCoordinator(browser BrowserPackage) (*Coordinator, error) {
	source, err := bindBrowserPackage(generatedArtifactSource{}, browser)
	if err != nil {
		return nil, err
	}
	return NewCoordinator(source)
}

// NewCoordinator admits one exact immutable artifact source. It is exported so
// application composition and contract tests can inject package adapters without
// granting the coordinator access to their internals.
func NewCoordinator(source ArtifactSource) (*Coordinator, error) {
	if source == nil {
		return nil, validationFailure(Finding{Code: "extension_registry_invalid", Phase: "registry_generation", Actual: "missing_artifact_source"})
	}
	sealed, err := admitRegistryPackage(source)
	if err != nil {
		return nil, err
	}
	source = sealed
	transactionLimits, err := readTransactionLimits(source)
	if err != nil {
		return nil, err
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
	if integrity["schema_id"] != "cartulary.extension_registry_integrity.v2" || stringValue(integrity["registry_sha256"]) != registryArtifact.SHA256 {
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
	if !ok || len(profileObjects) > 256 {
		return nil, invalidArtifact("profile_registry", errors.New("profiles must contain 0..256 objects"))
	}

	coordinator := &Coordinator{
		transactionLimits:    transactionLimits,
		validateFinding:      sealed.validateFinding,
		profiles:             make(map[string]profileRecord, len(profileObjects)),
		registrySHA256:       registryArtifact.SHA256,
		orderedProfileIDs:    make([]string, 0, len(profileObjects)),
		validationConditions: sealed.validationConditions,
	}
	for _, profileObject := range profileObjects {
		descriptor, parseErr := parseDescriptor(profileObject)
		if parseErr != nil {
			return nil, invalidArtifact("profile_descriptor", parseErr)
		}
		if _, duplicate := coordinator.profiles[descriptor.ProfileID]; duplicate {
			return nil, registryInvalid("duplicate_profile_projection")
		}
		descriptorObject, descriptorArtifact, readErr := readArtifactObject(source, generatedExtensionsRoot+"descriptors/"+descriptor.ProfileID+".json")
		if readErr != nil {
			return nil, readErr
		}
		if !equalCanonicalObjects(profileObject, descriptorObject) || descriptorDigests[descriptor.ProfileID] != descriptorArtifact.SHA256 {
			return nil, unavailableBinding(descriptor.ProfileID, descriptor.ContractMajor)
		}
		bindingPath := generatedExtensionsRoot + "implementation-bindings/" + descriptor.ProfileID + ".json"
		bindingObject := sealed.objects[bindingPath]
		bindingArtifact := sealed.artifacts[bindingPath]
		if bindingDigests[descriptor.ProfileID] != bindingArtifact.SHA256 {
			return nil, unavailableBinding(descriptor.ProfileID, descriptor.ContractMajor)
		}
		var contracts bindingContracts
		var bindingErr error
		if bindingObject != nil {
			contracts, bindingErr = admitBinding(descriptor, descriptorArtifact.SHA256, bindingObject, sealed)
		} else {
			contracts, bindingErr = admitDeclaredProfileContracts(descriptor.ProfileID, sealed)
		}
		if bindingErr != nil {
			return nil, unavailableBinding(descriptor.ProfileID, descriptor.ContractMajor)
		}
		admission, ok := descriptorObject["admission_validation"].(map[string]any)
		if !ok || len(admission) != 4 || admission["schema_id"] != "cartulary.extension_admission_validation.v1" {
			return nil, unavailableBinding(descriptor.ProfileID, descriptor.ContractMajor)
		}
		for _, phase := range []string{"preflight", "post_migration"} {
			value, present := admission[phase+"_algorithmid"]
			if !present || bindingObject != nil && value != bindingObject[phase+"_algorithm_id"] {
				return nil, unavailableBinding(descriptor.ProfileID, descriptor.ContractMajor)
			}
			if value != nil {
				id, ok := value.(string)
				if !ok || !strings.HasPrefix(id, descriptor.ProfileID+".") {
					return nil, unavailableBinding(descriptor.ProfileID, descriptor.ContractMajor)
				}
			}
		}
		coordinator.profiles[descriptor.ProfileID] = profileRecord{
			descriptor:             descriptor,
			descriptorObject:       cloneObject(profileObject),
			descriptorSHA256:       descriptorArtifact.SHA256,
			bindingObject:          cloneObject(bindingObject),
			bindingSHA256:          bindingArtifact.SHA256,
			jobContracts:           contracts.jobs,
			workerRuntimeContracts: contracts.workers,
		}
		coordinator.orderedProfileIDs = append(coordinator.orderedProfileIDs, descriptor.ProfileID)
	}
	sort.Strings(coordinator.orderedProfileIDs)
	if len(descriptorDigests) != len(coordinator.profiles) {
		return nil, invalidArtifact("registry_integrity", errors.New("descriptor digest set is incomplete or extra"))
	}
	for profileID := range bindingDigests {
		if _, recognized := coordinator.profiles[profileID]; !recognized {
			return nil, registryInvalid("binding_profile_unrecognized")
		}
	}
	inactivePolicies, err := admitInactiveConfigurationPolicies(source, coordinator.profiles, coordinator.orderedProfileIDs)
	if err != nil {
		return nil, err
	}
	coordinator.inactiveConfigurationPolicies = inactivePolicies

	_, clientSupportArtifact, err := readArtifactObject(source, generatedExtensionsRoot+"client-support-registry.json")
	if err != nil {
		return nil, err
	}
	if supportDigests["client-support-registry"] != clientSupportArtifact.SHA256 {
		return nil, invalidArtifact("client_support_registry", errors.New("integrity digest mismatch"))
	}
	coordinator.clientSupportSHA256 = clientSupportArtifact.SHA256

	if err := sealed.validateCompanionIdentities(); err != nil {
		return nil, err
	}
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
		} else if record.bindingSHA256 == "" {
			findings = append(findings, Finding{Code: "extension_implementation_unavailable", Phase: "implementation_binding", ProfileID: profileID, Actual: "binding_missing"})
		}
	}
	if len(findings) > 0 {
		return ClaimResolution{}, c.dependencyDiagnostics(validationFailure(findings...))
	}

	order, err := c.dependencyOrder(requested)
	if err != nil {
		return ClaimResolution{}, c.dependencyDiagnostics(err)
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
	ProfileID                   string
	WorkerKind                  string
	JobKinds                    []string
	MaxActiveAttemptsPerProcess int
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
	MinRefs            int
	MaxRefs            int
}

type JobKindContract struct {
	ProfileID                   string
	JobKind                     string
	ProgressUnitID              string
	OperationKind               string
	ProofPolicy                 string
	IdempotencyPolicy           string
	IdempotencyIdentitySchemaID string
	TerminalResultSchemaID      string
	TerminalSuccessCodes        []string
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
			"min_refs":              resourceRef.MinRefs,
			"max_refs":              resourceRef.MaxRefs,
		}
	}
	var identitySchema any
	if contract.IdempotencyIdentitySchemaID != "" {
		identitySchema = contract.IdempotencyIdentitySchemaID
	}
	return map[string]any{
		"schema_id":                      "cartulary.extension_job_kind_contract.v3",
		"profile_id":                     contract.ProfileID,
		"job_kind":                       contract.JobKind,
		"progress_unit_id":               contract.ProgressUnitID,
		"operation_kind":                 contract.OperationKind,
		"proof_policy":                   contract.ProofPolicy,
		"idempotency_policy":             contract.IdempotencyPolicy,
		"idempotency_identity_schema_id": identitySchema,
		"terminal_result_schema_id":      contract.TerminalResultSchemaID,
		"terminal_success_codes":         contract.TerminalSuccessCodes,
		"resource_ref_contracts":         resourceRefs,
		"cancellation_policy":            contract.CancellationPolicy,
		"max_proof_bytes":                contract.MaxProofBytes,
	}
}

func cloneJobKindContracts(source []JobKindContract) []JobKindContract {
	result := make([]JobKindContract, len(source))
	for index, contract := range source {
		contract.TerminalSuccessCodes = append([]string(nil), contract.TerminalSuccessCodes...)
		contract.ResourceRefContracts = append([]JobResourceRefContract(nil), contract.ResourceRefContracts...)
		result[index] = contract
	}
	return result
}

func cloneWorkerPublication(source WorkerPublication) WorkerPublication {
	source.JobKinds = append([]string(nil), source.JobKinds...)
	return source
}

func cloneWorkerPublications(source []WorkerPublication) []WorkerPublication {
	result := make([]WorkerPublication, len(source))
	for index, worker := range source {
		result[index] = cloneWorkerPublication(worker)
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

// WorkerRuntimeContracts returns every packaged worker contract in profile and
// worker identity order, independent of claim state. Jobs uses this set to
// construct the recognized durable catalog without a handwritten mapping.
func (c *Coordinator) WorkerRuntimeContracts() []WorkerPublication {
	if c == nil {
		return nil
	}
	result := []WorkerPublication{}
	for _, profileID := range c.orderedProfileIDs {
		result = append(result, cloneWorkerPublications(c.profiles[profileID].workerRuntimeContracts)...)
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
	return cloneWorkerPublications(p.workers)
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
	if _, err := c.validateClaimResolution(resolution); err != nil {
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
		var contractMajor *int
		if record.descriptor.Claimable {
			value := record.descriptor.ContractMajor
			contractMajor = &value
		}
		discoveryWorkspaces := make([]DiscoveryWorkspace, 0, len(record.descriptor.WorkspaceKeys))
		for _, workspaceKey := range record.descriptor.WorkspaceKeys {
			discoveryWorkspaces = append(discoveryWorkspaces, DiscoveryWorkspace{WorkspaceKey: workspaceKey, MinimumRole: "viewer"})
		}
		discovery = append(discovery, DiscoveryProfile{
			ProfileID: profileID, Claimable: record.descriptor.Claimable, Claimed: isClaimed,
			ContractMajor: contractMajor, RouteFamilies: append([]string(nil), record.descriptor.RouteFamilies...),
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
		for _, worker := range record.workerRuntimeContracts {
			workerItems = append(workerItems, map[string]any{
				"profile_id": profileID, "worker_kind": worker.WorkerKind,
				"job_kinds":                       append([]string(nil), worker.JobKinds...),
				"max_active_attempts_per_process": worker.MaxActiveAttemptsPerProcess,
			})
			workers = append(workers, cloneWorkerPublication(worker))
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

func parseDescriptor(object map[string]any) (Descriptor, error) {
	if err := requireExactKeys(object, "schema_id", "profile_id", "claimable", "contract_major", "owner_id", "claim_config_key", "route_families", "workspace_keys", "capability_ids", "runtime_dependencies", "contributions", "public_schema_ids", "prestage_config_keys", "state_ownership", "admission_validation", "egress_mode", "incident_portability_mode", "snapshot_reporting_mode"); err != nil {
		return Descriptor{}, err
	}
	profileID := stringValue(object["profile_id"])
	major, ok := integerValue(object["contract_major"])
	claimable, boolOK := object["claimable"].(bool)
	majorValid := claimable && ok && major >= 1 || !claimable && object["contract_major"] == nil
	routes, routesOK := stringSlice(object["route_families"])
	workspaces, workspacesOK := stringSlice(object["workspace_keys"])
	capabilities, capabilitiesOK := stringSlice(object["capability_ids"])
	dependenciesObjects, dependenciesOK := objectSlice(object["runtime_dependencies"])
	if object["schema_id"] != "cartulary.extension_profile_descriptor.v3" || !extensionProfileIDPattern.MatchString(profileID) || !majorValid || !boolOK || !routesOK || !workspacesOK || !capabilitiesOK || !dependenciesOK || len(capabilities) != 0 {
		return Descriptor{}, errors.New("descriptor identity, scalar, collection, or capability contract is invalid")
	}
	dependencies := make([]Dependency, 0, len(dependenciesObjects))
	for _, dependencyObject := range dependenciesObjects {
		if err := requireExactKeys(dependencyObject, "profile_id", "required_contract_major"); err != nil {
			return Descriptor{}, err
		}
		dependencyMajor, integerOK := integerValue(dependencyObject["required_contract_major"])
		dependencyID := stringValue(dependencyObject["profile_id"])
		if !integerOK || dependencyMajor < 1 || !extensionProfileIDPattern.MatchString(dependencyID) || dependencyID == "base" {
			return Descriptor{}, errors.New("dependency is malformed")
		}
		dependencies = append(dependencies, Dependency{ProfileID: dependencyID, RequiredContractMajor: dependencyMajor})
	}
	return Descriptor{
		ProfileID: profileID, Claimable: claimable, ContractMajor: major,
		OwnerID: stringValue(object["owner_id"]), ClaimConfigKey: stringValue(object["claim_config_key"]),
		RouteFamilies: routes, WorkspaceKeys: workspaces, CapabilityIDs: capabilities, Dependencies: dependencies,
	}, nil
}

type bindingContracts struct {
	jobs    []JobKindContract
	workers []WorkerPublication
}

func admitBinding(descriptor Descriptor, descriptorDigest string, binding map[string]any, source *admittedPackage) (bindingContracts, error) {
	requiredKeys := []string{"schema_id", "profile_id", "contract_major", "descriptor_sha256", "implemented_contribution_ids", "supported_capability_ids", "state_ownership_kind", "preflight_algorithm_id", "post_migration_algorithm_id", "initialization_definition_sha256", "initialization_algorithm_id", "final_state_validation_algorithm_id", "dependency_probe_ids", "migration_definitions", "physical_state_binding_sha256", "backup_codec_bindings", "rebuild_algorithm_ids", "transaction_participant_limits", "supporting_schema_ids", "worker_kinds", "worker_runtime_contracts", "job_kind_contracts", "participant_contracts"}
	if err := requireExactKeys(binding, requiredKeys...); err != nil {
		return bindingContracts{}, err
	}
	major, ok := integerValue(binding["contract_major"])
	majorMatches := descriptor.Claimable && ok && major == descriptor.ContractMajor || !descriptor.Claimable && binding["contract_major"] == nil
	capabilities, capabilitiesOK := stringSlice(binding["supported_capability_ids"])
	if binding["schema_id"] != "cartulary.extension_implementation_binding.v2" || stringValue(binding["profile_id"]) != descriptor.ProfileID || !majorMatches || stringValue(binding["descriptor_sha256"]) != descriptorDigest || !capabilitiesOK || len(capabilities) != 0 {
		return bindingContracts{}, errors.New("binding identity, major, descriptor digest, or capability parity is invalid")
	}
	for _, key := range []string{"implemented_contribution_ids", "dependency_probe_ids", "rebuild_algorithm_ids", "supporting_schema_ids", "worker_kinds"} {
		values, valid := stringSlice(binding[key])
		if !valid || !strictlySortedUnique(values) {
			return bindingContracts{}, fmt.Errorf("binding %s is malformed", key)
		}
	}
	declaredSchemas, _ := stringSlice(binding["supporting_schema_ids"])
	requiredSchemas := source.requiredSchemas[descriptor.ProfileID]
	if len(declaredSchemas) != len(requiredSchemas) {
		return bindingContracts{}, errors.New("binding supporting schema set is incomplete or extra")
	}
	for i, id := range requiredSchemas {
		if declaredSchemas[i] != id {
			return bindingContracts{}, errors.New("binding supporting schema set differs from its operational contracts")
		}
	}
	if err := validateParticipantBindingReferences(descriptor.ProfileID, binding, source); err != nil {
		return bindingContracts{}, err
	}
	jobs, err := admitJobKindContracts(descriptor.ProfileID, binding["job_kind_contracts"], source)
	if err != nil {
		return bindingContracts{}, err
	}
	workerKinds, _ := stringSlice(binding["worker_kinds"])
	workers, err := parseWorkerRuntimeContracts(descriptor.ProfileID, binding["worker_runtime_contracts"], workerKinds, jobs)
	if err != nil {
		return bindingContracts{}, err
	}
	return bindingContracts{jobs: jobs, workers: workers}, nil
}

func parseWorkerRuntimeContracts(profileID string, value any, workerKinds []string, jobContracts []JobKindContract) ([]WorkerPublication, error) {
	rows, ok := objectSlice(value)
	if !ok || len(rows) != len(workerKinds) {
		return nil, errors.New("binding worker runtime contracts are incomplete")
	}
	knownJobs := make(map[string]bool, len(jobContracts))
	for _, contract := range jobContracts {
		knownJobs[contract.JobKind] = true
	}
	assignedJobs := make(map[string]bool, len(jobContracts))
	result := make([]WorkerPublication, 0, len(rows))
	for index, row := range rows {
		if err := requireExactKeys(row, "schema_id", "profile_id", "worker_kind", "job_kinds", "max_active_attempts_per_process"); err != nil {
			return nil, err
		}
		maximum, maximumOK := integerValue(row["max_active_attempts_per_process"])
		jobs, jobsOK := stringSlice(row["job_kinds"])
		if row["schema_id"] != "cartulary.extension_worker_runtime_contract.v1" ||
			stringValue(row["profile_id"]) != profileID ||
			stringValue(row["worker_kind"]) != workerKinds[index] ||
			!maximumOK || maximum < 1 || maximum > 8 ||
			!jobsOK || len(jobs) == 0 || !strictlySortedUnique(jobs) {
			return nil, errors.New("binding worker runtime contract is malformed")
		}
		for _, jobKind := range jobs {
			if !knownJobs[jobKind] || assignedJobs[jobKind] {
				return nil, errors.New("binding worker runtime job assignment is invalid")
			}
			assignedJobs[jobKind] = true
		}
		result = append(result, WorkerPublication{
			ProfileID: profileID, WorkerKind: workerKinds[index],
			JobKinds: jobs, MaxActiveAttemptsPerProcess: maximum,
		})
	}
	if len(assignedJobs) != len(knownJobs) {
		return nil, errors.New("binding worker runtime job assignment is incomplete")
	}
	return result, nil
}

func admitJobKindContracts(profileID string, value any, source *admittedPackage) ([]JobKindContract, error) {
	rows, ok := objectSlice(value)
	if !ok || len(rows) > 64 {
		return nil, errors.New("binding job_kind_contracts is malformed")
	}
	result := make([]JobKindContract, len(rows))
	previous := ""
	for index, reference := range rows {
		if err := requireExactKeys(reference, "job_kind", "job_kind_contract_sha256"); err != nil {
			return nil, err
		}
		kind := stringValue(reference["job_kind"])
		digest := stringValue(reference["job_kind_contract_sha256"])
		if !strings.HasPrefix(kind, profileID+".") || !extensionProfileIDPattern.MatchString(strings.TrimPrefix(kind, profileID+".")) || kind <= previous || !lowerDigest(digest) {
			return nil, errors.New("binding job reference is invalid")
		}
		path := generatedExtensionsRoot + "job-contracts/" + profileID + "/" + kind + ".json"
		artifact, exists := source.artifacts[path]
		if !exists {
			return nil, errors.New("job_contract_missing")
		}
		if artifact.SHA256 != digest {
			return nil, errors.New("job_contract_digest_mismatch")
		}
		row := source.objects[path]

		if err := requireExactKeys(row,
			"schema_id", "profile_id", "job_kind", "progress_unit_id", "operation_kind", "proof_policy",
			"idempotency_policy", "idempotency_identity_schema_id", "terminal_result_schema_id", "terminal_success_codes",
			"resource_ref_contracts", "cancellation_policy", "max_proof_bytes",
		); err != nil {
			return nil, err
		}
		contract := JobKindContract{
			ProfileID:                   stringValue(row["profile_id"]),
			JobKind:                     stringValue(row["job_kind"]),
			ProgressUnitID:              stringValue(row["progress_unit_id"]),
			OperationKind:               stringValue(row["operation_kind"]),
			ProofPolicy:                 stringValue(row["proof_policy"]),
			IdempotencyPolicy:           stringValue(row["idempotency_policy"]),
			IdempotencyIdentitySchemaID: stringValue(row["idempotency_identity_schema_id"]),
			TerminalResultSchemaID:      stringValue(row["terminal_result_schema_id"]),
			CancellationPolicy:          stringValue(row["cancellation_policy"]),
			MaxProofBytes:               intValue(row["max_proof_bytes"]),
		}
		codes, codesOK := stringSlice(row["terminal_success_codes"])
		if !codesOK {
			return nil, errors.New("binding job success codes are malformed")
		}
		contract.TerminalSuccessCodes = codes
		if row["schema_id"] != "cartulary.extension_job_kind_contract.v3" ||
			contract.ProfileID != profileID || contract.JobKind != kind ||
			!strings.HasPrefix(contract.JobKind, profileID+".") ||
			strings.Contains(contract.JobKind, "/") ||
			strings.Contains(contract.JobKind, `\`) ||
			(previous != "" && previous >= contract.JobKind) ||
			!validProgressUnitID(contract.ProgressUnitID) ||
			!strings.HasPrefix(contract.OperationKind, profileID+".") ||
			(contract.ProofPolicy != "required_on_terminal_success" && contract.ProofPolicy != "forbidden") ||
			!validJobIdentityPolicy(contract) ||
			(contract.CancellationPolicy != "precommit_observable" && contract.CancellationPolicy != "not_cancelable") ||
			contract.MaxProofBytes < 1 || contract.MaxProofBytes > 1048576 {
			return nil, errors.New("binding job kind contract is invalid")
		}
		if _, ok := jobcontract.ResolveTerminal(contract.TerminalResultSchemaID); !ok {
			return nil, errors.New("unresolved terminal result schema")
		}
		previous = contract.JobKind
		resourceRefs, ok := objectSlice(row["resource_ref_contracts"])
		if !ok || len(resourceRefs) > 64 {
			return nil, errors.New("binding job resource references are malformed")
		}
		previousRef := ""
		for _, resourceRef := range resourceRefs {
			if err := requireExactKeys(resourceRef, "resource_ref_kind", "resource_id_schema_id", "min_refs", "max_refs"); err != nil {
				return nil, err
			}
			parsed := JobResourceRefContract{
				ResourceRefKind:    stringValue(resourceRef["resource_ref_kind"]),
				ResourceIDSchemaID: stringValue(resourceRef["resource_id_schema_id"]),
				MinRefs:            intValue(resourceRef["min_refs"]),
				MaxRefs:            intValue(resourceRef["max_refs"]),
			}
			if parsed.ResourceRefKind == "" ||
				(previousRef != "" && previousRef >= parsed.ResourceRefKind) ||
				parsed.MaxRefs < 1 || parsed.MaxRefs > 1024 {
				return nil, errors.New("binding job resource reference is invalid")
			}
			previousRef = parsed.ResourceRefKind
			contract.ResourceRefContracts = append(contract.ResourceRefContracts, parsed)
		}
		if err := jobcontract.ValidateTerminalContract(contract.terminalContract()); err != nil {
			return nil, err
		}
		if contract.ProofPolicy == "forbidden" && len(contract.ResourceRefContracts) != 0 {
			return nil, errors.New("proof-forbidden job declares resources")
		}
		result[index] = contract
	}
	return result, nil
}

func validProgressUnitID(value string) bool {
	if len(value) == 0 || len(value) > 191 {
		return false
	}
	segments := strings.Split(value, ".")
	if len(segments) < 3 {
		return false
	}
	for _, segment := range segments[:len(segments)-1] {
		if len(segment) == 0 || len(segment) > 63 || segment[0] < 'a' || segment[0] > 'z' {
			return false
		}
		for _, character := range segment[1:] {
			if !((character >= 'a' && character <= 'z') ||
				(character >= '0' && character <= '9') || character == '_') {
				return false
			}
		}
	}
	version := segments[len(segments)-1]
	if len(version) < 2 || version[0] != 'v' || version[1] < '1' || version[1] > '9' {
		return false
	}
	for _, character := range version[2:] {
		if character < '0' || character > '9' {
			return false
		}
	}
	return true
}

func readArtifactObject(source ArtifactSource, path string) (map[string]any, PackagedArtifact, error) {
	artifact, ok := source.Artifact(path)
	if !ok {
		return nil, PackagedArtifact{}, invalidArtifact(path, nil)
	}
	if len(artifact.JSON) > packagedArtifactByteLimit(path) {
		return nil, PackagedArtifact{}, registryLimit("artifact_bytes", packagedArtifactByteLimit(path), len(artifact.JSON))
	}
	artifact.JSON = append([]byte(nil), artifact.JSON...)
	digest := sha256.Sum256(artifact.JSON)
	actualDigest := hex.EncodeToString(digest[:])
	if artifact.SHA256 != actualDigest {
		return nil, PackagedArtifact{}, invalidArtifact(path, nil)
	}
	value, err := canonicaljson.DecodeExtension(artifact.JSON)
	if err != nil {
		return nil, PackagedArtifact{}, invalidArtifact(path, err)
	}
	object, ok := value.(map[string]any)
	if !ok || object == nil {
		return nil, PackagedArtifact{}, invalidArtifact(path, errors.New("artifact must be an object"))
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
	encoded, err := canonicaljson.MarshalExtension(value)
	if err != nil {
		return nil, err
	}
	if !finalLF {
		encoded = encoded[:len(encoded)-1]
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

func strictlySortedUnique(values []string) bool {
	for index := 1; index < len(values); index++ {
		if values[index-1] >= values[index] {
			return false
		}
	}
	return true
}

func sortFindings(findings []Finding) {
	sort.Slice(findings, func(i, j int) bool {
		left := findingSortKey(findings[i])
		right := findingSortKey(findings[j])
		return left < right
	})
}

func findingSortKey(f Finding) string {
	return strings.Join([]string{f.Code, f.ProfileID, f.DependencyProfileID, f.CollisionClass, strings.Join(f.ConflictingProfileIDs, ","), strings.Join(f.ConflictingTokens, ","), f.Expected, f.Actual, fmt.Sprint(f.DuplicateCount)}, "\x00")
}

func validationFailure(findings ...Finding) error {
	for index := range findings {
		findings[index].ConflictingTokens = append([]string(nil), findings[index].ConflictingTokens...)
		findings[index].ConflictingProfileIDs = append([]string(nil), findings[index].ConflictingProfileIDs...)
	}
	cause := &ValidationError{findings: findings}
	if len(findings) == 1 && findings[0].Code == "extension_registry_invalid" {
		return startupFailure(cause, registryArtifactFinding("profile_registry"))
	}
	return cause
}

func invalidArtifact(slot string, _ error) error {
	cause := &ValidationError{findings: []Finding{{Code: "extension_registry_invalid", Phase: "registry_generation", Expected: slot, Actual: "redacted"}}}
	return startupFailure(cause, registryArtifactFinding(slot))
}

func unavailableBinding(profileID string, major int) error {
	cause := validationFailure(Finding{Code: "extension_implementation_unavailable", Phase: "implementation_binding", ProfileID: profileID, Actual: "redacted"})
	if major < 1 {
		return startupFailure(cause, registryArtifactFinding("implementation_binding"))
	}
	return implementationUnavailable(profileID, major, nil, cause)
}
