package sourceport

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"sort"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/modules/incidentportability"
)

const ContractMajor = 1

var (
	ErrInvalidCatalog  = errors.New("incident bundle source-port catalog is invalid")
	ErrPreparedBinding = errors.New("incident bundle prepared source belongs to another port")
	schemaIDPattern    = regexp.MustCompile(`^cartulary\.[A-Za-z0-9_.-]+\.v[1-9][0-9]*$`)
)

type Path struct {
	LogicalPath               string   `json:"logical_path"`
	ContentRole               string   `json:"content_role"`
	SchemaID                  string   `json:"schema_id,omitempty"`
	Versions                  []int    `json:"versions"`
	StableIdentity            []string `json:"stable_identity"`
	StableIdentityInvariantID string   `json:"stable_identity_invariant_id"`
}

type Descriptor struct {
	FamilyID         string
	ContractMajor    int
	OwnerID          string
	OwnerRelationIDs []string
	Dependencies     []string
	Paths            []Path
	InvariantIDs     []string
}

type ImportContext struct {
	IncidentID           uuid.UUID
	ActorUserID          uuid.UUID
	BundleVersion        int
	OperationID          string
	Attributions         incidentportability.AttributionRecorder
	Actors               ActorCatalog
	RewrittenObjectBlobs []byte
}

type PortableAttributionResolver interface {
	ResolvePortableSourceActors(
		context.Context,
		incidentportability.Queryer,
		uuid.UUID,
		string,
		string,
		[]string,
	) (map[string]string, error)
}

type ExportContext struct {
	Query                incidentportability.Queryer
	IncidentID           uuid.UUID
	PortableAttributions PortableAttributionResolver
}

type Bundle interface {
	File(string) ([]byte, bool)
	Paths() []string
}

type MapBundle map[string][]byte

func (b MapBundle) File(path string) ([]byte, bool) {
	payload, ok := b[path]
	return payload, ok
}

func (b MapBundle) Paths() []string {
	paths := make([]string, 0, len(b))
	for path := range b {
		paths = append(paths, path)
	}
	sort.Strings(paths)
	return paths
}

type Prepared struct {
	bindingKey string
	value      any
}

func NewPrepared(portKey string, operationID string, value any) Prepared {
	return Prepared{bindingKey: portKey + "\x00" + operationID, value: value}
}

func (p Prepared) ValueFor(portKey string, operationID string) (any, error) {
	if portKey == "" || operationID == "" || p.bindingKey != portKey+"\x00"+operationID {
		return nil, ErrPreparedBinding
	}
	return p.value, nil
}

type Port interface {
	Descriptor() Descriptor
	Export(context.Context, ExportContext) ([]incidentportability.File, error)
	PrepareImport(context.Context, Bundle, ImportContext) (Prepared, error)
	ApplyImportTx(context.Context, pgx.Tx, Prepared, ImportContext) error
	ValidateImportTx(context.Context, pgx.Tx, Prepared, ImportContext) error
}

type ContractValidator interface {
	ValidateSourcePortContract() error
}

type Failure struct {
	familyID    string
	invariantID string
}

func (e *Failure) Error() string {
	return fmt.Sprintf("incident bundle source family %s failed invariant %s", e.familyID, e.invariantID)
}

func (e *Failure) FamilyID() string {
	if e == nil {
		return ""
	}
	return e.familyID
}

func (e *Failure) InvariantID() string {
	if e == nil {
		return ""
	}
	return e.invariantID
}

// DeclaredFailure constructs a read-only failure bound to this descriptor's
// family and closed invariant registry.
func (d Descriptor) DeclaredFailure(invariantID string) error {
	if strings.TrimSpace(d.FamilyID) == "" || !strings.HasPrefix(invariantID, d.FamilyID+".") {
		return fmt.Errorf("%w: invariant %s is outside family %s", ErrInvalidCatalog, invariantID, d.FamilyID)
	}
	for _, declared := range d.InvariantIDs {
		if invariantID == declared {
			return &Failure{familyID: d.FamilyID, invariantID: invariantID}
		}
	}
	return fmt.Errorf("%w: invariant %s is undeclared for family %s", ErrInvalidCatalog, invariantID, d.FamilyID)
}

type Catalog struct {
	ports            []Port
	descriptors      []Descriptor
	byFamily         map[string]Port
	specialConsumers map[int]map[string]string
}

type CatalogOptions struct {
	Ports                  []Port
	RequiredPathsByVersion map[int][]string
	AllowedRelationIDs     map[string]struct{}
	SpecialConsumers       map[int]map[string]string
}

func NewCatalog(options CatalogOptions) (*Catalog, error) {
	if len(options.Ports) == 0 || len(options.RequiredPathsByVersion) == 0 {
		return nil, ErrInvalidCatalog
	}
	byFamily := make(map[string]Port, len(options.Ports))
	descriptors := make([]Descriptor, 0, len(options.Ports))
	pathClaims := map[int]map[string]string{}
	for version := range options.RequiredPathsByVersion {
		pathClaims[version] = map[string]string{}
		for logicalPath, consumerID := range options.SpecialConsumers[version] {
			if strings.TrimSpace(logicalPath) == "" || strings.TrimSpace(consumerID) == "" {
				return nil, fmt.Errorf("%w: incomplete special consumer", ErrInvalidCatalog)
			}
			pathClaims[version][logicalPath] = consumerID
		}
	}
	for _, port := range options.Ports {
		if port == nil {
			return nil, ErrInvalidCatalog
		}
		descriptor := canonicalDescriptor(port.Descriptor())
		if err := validateDescriptor(descriptor, options.AllowedRelationIDs); err != nil {
			return nil, err
		}
		if _, duplicate := byFamily[descriptor.FamilyID]; duplicate {
			return nil, fmt.Errorf("%w: duplicate family %s", ErrInvalidCatalog, descriptor.FamilyID)
		}
		byFamily[descriptor.FamilyID] = port
		descriptors = append(descriptors, descriptor)
		for _, path := range descriptor.Paths {
			for _, version := range path.Versions {
				claims, admitted := pathClaims[version]
				if !admitted {
					return nil, fmt.Errorf("%w: path %s uses unsupported version %d", ErrInvalidCatalog, path.LogicalPath, version)
				}
				if prior := claims[path.LogicalPath]; prior != "" {
					return nil, fmt.Errorf("%w: path %s claimed by %s and %s", ErrInvalidCatalog, path.LogicalPath, prior, descriptor.FamilyID)
				}
				claims[path.LogicalPath] = descriptor.FamilyID
			}
		}
	}
	for _, descriptor := range descriptors {
		for _, dependency := range descriptor.Dependencies {
			if _, ok := byFamily[dependency]; !ok {
				return nil, fmt.Errorf("%w: family %s has unknown dependency %s", ErrInvalidCatalog, descriptor.FamilyID, dependency)
			}
		}
	}
	for _, port := range options.Ports {
		validator, ok := port.(ContractValidator)
		if !ok {
			return nil, fmt.Errorf("%w: source port does not expose contract validation", ErrInvalidCatalog)
		}
		if err := validator.ValidateSourcePortContract(); err != nil {
			return nil, err
		}
	}
	for version, requiredPaths := range options.RequiredPathsByVersion {
		claims := pathClaims[version]
		required := map[string]struct{}{}
		for _, path := range requiredPaths {
			if _, duplicate := required[path]; duplicate {
				return nil, fmt.Errorf("%w: duplicate required path %s", ErrInvalidCatalog, path)
			}
			required[path] = struct{}{}
			if claims[path] == "" {
				return nil, fmt.Errorf("%w: uncovered required path %s for version %d", ErrInvalidCatalog, path, version)
			}
		}
		for path := range claims {
			if _, ok := required[path]; !ok {
				return nil, fmt.Errorf("%w: unrequired path claim %s for version %d", ErrInvalidCatalog, path, version)
			}
		}
	}
	ordered, err := topologicalOrder(descriptors)
	if err != nil {
		return nil, err
	}
	ports := make([]Port, 0, len(ordered))
	for _, descriptor := range ordered {
		ports = append(ports, byFamily[descriptor.FamilyID])
	}
	specialConsumers := map[int]map[string]string{}
	for version, claims := range options.SpecialConsumers {
		specialConsumers[version] = map[string]string{}
		for logicalPath, consumerID := range claims {
			specialConsumers[version][logicalPath] = consumerID
		}
	}
	return &Catalog{ports: ports, descriptors: ordered, byFamily: byFamily, specialConsumers: specialConsumers}, nil
}

func (c *Catalog) Ports() []Port {
	if c == nil {
		return nil
	}
	return append([]Port(nil), c.ports...)
}

func (c *Catalog) Descriptors() []Descriptor {
	if c == nil {
		return nil
	}
	result := make([]Descriptor, 0, len(c.descriptors))
	for _, descriptor := range c.descriptors {
		result = append(result, cloneDescriptor(descriptor))
	}
	return result
}

func (c *Catalog) ConsumerFor(version int, path string) (string, bool) {
	if c == nil {
		return "", false
	}
	for _, descriptor := range c.descriptors {
		for _, candidate := range descriptor.Paths {
			if candidate.LogicalPath == path && containsVersion(candidate.Versions, version) {
				return descriptor.FamilyID, true
			}
		}
	}
	if consumerID := c.specialConsumers[version][path]; consumerID != "" {
		return consumerID, true
	}
	return "", false
}

func canonicalDescriptor(descriptor Descriptor) Descriptor {
	descriptor.Dependencies = canonicalStrings(descriptor.Dependencies)
	descriptor.OwnerRelationIDs = canonicalStrings(descriptor.OwnerRelationIDs)
	descriptor.InvariantIDs = canonicalStrings(descriptor.InvariantIDs)
	descriptor.Paths = append([]Path(nil), descriptor.Paths...)
	sort.Slice(descriptor.Paths, func(i, j int) bool {
		return descriptor.Paths[i].LogicalPath < descriptor.Paths[j].LogicalPath
	})
	for index := range descriptor.Paths {
		descriptor.Paths[index].Versions = canonicalInts(descriptor.Paths[index].Versions)
		descriptor.Paths[index].StableIdentity = canonicalStrings(descriptor.Paths[index].StableIdentity)
	}
	return descriptor
}

func cloneDescriptor(descriptor Descriptor) Descriptor {
	result := descriptor
	result.Dependencies = append([]string(nil), descriptor.Dependencies...)
	result.OwnerRelationIDs = append([]string(nil), descriptor.OwnerRelationIDs...)
	result.InvariantIDs = append([]string(nil), descriptor.InvariantIDs...)
	result.Paths = make([]Path, 0, len(descriptor.Paths))
	for _, path := range descriptor.Paths {
		path.Versions = append([]int(nil), path.Versions...)
		path.StableIdentity = append([]string(nil), path.StableIdentity...)
		result.Paths = append(result.Paths, path)
	}
	return result
}

func validateDescriptor(descriptor Descriptor, allowedRelations map[string]struct{}) error {
	if strings.TrimSpace(descriptor.FamilyID) == "" || descriptor.ContractMajor != ContractMajor ||
		strings.TrimSpace(descriptor.OwnerID) == "" || len(descriptor.Paths) == 0 ||
		len(descriptor.InvariantIDs) == 0 {
		return fmt.Errorf("%w: incomplete descriptor %s", ErrInvalidCatalog, descriptor.FamilyID)
	}
	for _, relationID := range descriptor.OwnerRelationIDs {
		if _, ok := allowedRelations[relationID]; !ok {
			return fmt.Errorf("%w: family %s has unknown owner relation %s", ErrInvalidCatalog, descriptor.FamilyID, relationID)
		}
	}
	seenPaths := map[string]struct{}{}
	for _, path := range descriptor.Paths {
		if strings.TrimSpace(path.LogicalPath) == "" || strings.TrimSpace(path.ContentRole) == "" ||
			len(path.Versions) == 0 || len(path.StableIdentity) == 0 ||
			strings.TrimSpace(path.StableIdentityInvariantID) == "" {
			return fmt.Errorf("%w: incomplete path in %s", ErrInvalidCatalog, descriptor.FamilyID)
		}
		if path.SchemaID != "" && !schemaIDPattern.MatchString(path.SchemaID) {
			return fmt.Errorf("%w: invalid path schema id in %s", ErrInvalidCatalog, descriptor.FamilyID)
		}
		if _, duplicate := seenPaths[path.LogicalPath]; duplicate {
			return fmt.Errorf("%w: duplicate path %s", ErrInvalidCatalog, path.LogicalPath)
		}
		if failure := descriptor.DeclaredFailure(path.StableIdentityInvariantID); errors.Is(failure, ErrInvalidCatalog) {
			return failure
		}
		seenPaths[path.LogicalPath] = struct{}{}
	}
	for _, invariantID := range descriptor.InvariantIDs {
		if !strings.HasPrefix(invariantID, descriptor.FamilyID+".") {
			return fmt.Errorf("%w: invariant %s is outside family %s", ErrInvalidCatalog, invariantID, descriptor.FamilyID)
		}
	}
	return nil
}

func topologicalOrder(descriptors []Descriptor) ([]Descriptor, error) {
	remaining := make(map[string]Descriptor, len(descriptors))
	for _, descriptor := range descriptors {
		remaining[descriptor.FamilyID] = descriptor
	}
	resolved := map[string]struct{}{}
	ordered := make([]Descriptor, 0, len(descriptors))
	for len(remaining) > 0 {
		ready := make([]string, 0, len(remaining))
		for familyID, descriptor := range remaining {
			allResolved := true
			for _, dependency := range descriptor.Dependencies {
				if _, ok := resolved[dependency]; !ok {
					allResolved = false
					break
				}
			}
			if allResolved {
				ready = append(ready, familyID)
			}
		}
		if len(ready) == 0 {
			return nil, fmt.Errorf("%w: dependency cycle", ErrInvalidCatalog)
		}
		sort.Strings(ready)
		for _, familyID := range ready {
			ordered = append(ordered, remaining[familyID])
			delete(remaining, familyID)
			resolved[familyID] = struct{}{}
		}
	}
	return ordered, nil
}

func canonicalStrings(values []string) []string {
	result := append([]string(nil), values...)
	sort.Strings(result)
	return result
}

func canonicalInts(values []int) []int {
	result := append([]int(nil), values...)
	sort.Ints(result)
	return result
}

func containsVersion(versions []int, version int) bool {
	for _, candidate := range versions {
		if candidate == version {
			return true
		}
	}
	return false
}
