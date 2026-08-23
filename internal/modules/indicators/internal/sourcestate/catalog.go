package sourcestate

import (
	"fmt"
	"path"
	"regexp"
	"slices"
	"sort"
	"strings"
	"sync"
)

const portableBundleVersion = 2

var identifierPattern = regexp.MustCompile(`^[a-z][a-z0-9_]*$`)

type AuthoritativeRelation struct {
	TableName string
}

type RebuildableRelation struct {
	TableName          string
	RebuildInvariantID string
}

type PortabilityDescriptor struct {
	Order                     int
	LogicalPath               string
	ContentRole               string
	SchemaID                  string
	Versions                  []int
	StableIdentity            []string
	StableIdentityInvariantID string
}

type definition struct {
	authoritative []AuthoritativeRelation
	rebuildable   []RebuildableRelation
	portability   []PortabilityDescriptor
}

type Catalog struct {
	authoritative []AuthoritativeRelation
	rebuildable   []RebuildableRelation
	portability   []PortabilityDescriptor
}

var (
	loadOnce    sync.Once
	loadCatalog *Catalog
	loadErr     error
)

// Load validates the Indicator source-state facts once and returns an
// immutable catalog. Construction failures are propagated through application
// composition instead of becoming initialization panics.
func Load() (*Catalog, error) {
	loadOnce.Do(func() {
		loadCatalog, loadErr = build(canonicalDefinition())
	})
	return loadCatalog, loadErr
}

func (c *Catalog) AuthoritativeRelations() []AuthoritativeRelation {
	if c == nil {
		return nil
	}
	return slices.Clone(c.authoritative)
}

func (c *Catalog) RebuildableRelations() []RebuildableRelation {
	if c == nil {
		return nil
	}
	return slices.Clone(c.rebuildable)
}

func (c *Catalog) PortabilityDescriptors() []PortabilityDescriptor {
	if c == nil {
		return nil
	}
	return clonePortability(c.portability)
}

func canonicalDefinition() definition {
	return definition{
		authoritative: []AuthoritativeRelation{
			{TableName: "indicator_observations"},
			{TableName: "indicator_state_intervals"},
			{TableName: "indicators"},
		},
		rebuildable: []RebuildableRelation{
			{
				TableName:          "indicator_active_identities",
				RebuildInvariantID: "indicators.restore_active_identities.v1",
			},
		},
		portability: []PortabilityDescriptor{
			{
				Order:                     0,
				LogicalPath:               "data/indicators.ndjson",
				ContentRole:               "source_rows",
				SchemaID:                  "cartulary.incident_bundle.indicators.row.v1",
				Versions:                  []int{portableBundleVersion},
				StableIdentity:            []string{"record_id"},
				StableIdentityInvariantID: "indicators.source_identity_admitted",
			},
			{
				Order:                     1,
				LogicalPath:               "data/indicator_observations.ndjson",
				ContentRole:               "source_rows",
				SchemaID:                  "cartulary.incident_bundle.indicator_observations.row.v1",
				Versions:                  []int{portableBundleVersion},
				StableIdentity:            []string{"indicator_observation_id"},
				StableIdentityInvariantID: "indicators.source_identity_admitted",
			},
			{
				Order:                     2,
				LogicalPath:               "data/indicator_state_intervals.ndjson",
				ContentRole:               "source_rows",
				SchemaID:                  "cartulary.incident_bundle.indicator_state_intervals.row.v1",
				Versions:                  []int{portableBundleVersion},
				StableIdentity:            []string{"indicator_state_interval_id"},
				StableIdentityInvariantID: "indicators.source_identity_admitted",
			},
		},
	}
}

func build(input definition) (*Catalog, error) {
	if len(input.authoritative) == 0 || len(input.rebuildable) == 0 || len(input.portability) == 0 {
		return nil, fmt.Errorf("indicators: source-state catalog contains an empty inventory")
	}

	catalog := &Catalog{
		authoritative: slices.Clone(input.authoritative),
		rebuildable:   slices.Clone(input.rebuildable),
		portability:   clonePortability(input.portability),
	}
	relations := make(map[string]struct{}, len(catalog.authoritative)+len(catalog.rebuildable))
	for _, relation := range catalog.authoritative {
		if err := admitRelation(relations, relation.TableName); err != nil {
			return nil, err
		}
	}
	for _, relation := range catalog.rebuildable {
		if err := admitRelation(relations, relation.TableName); err != nil {
			return nil, err
		}
		if strings.TrimSpace(relation.RebuildInvariantID) == "" || relation.RebuildInvariantID != strings.TrimSpace(relation.RebuildInvariantID) {
			return nil, fmt.Errorf("indicators: rebuildable relation %s has an invalid rebuild invariant", relation.TableName)
		}
	}

	paths := make(map[string]struct{}, len(catalog.portability))
	schemas := make(map[string]struct{}, len(catalog.portability))
	orders := make(map[int]struct{}, len(catalog.portability))
	for _, descriptor := range catalog.portability {
		if err := validatePortabilityDescriptor(descriptor, paths, schemas, orders); err != nil {
			return nil, err
		}
	}
	for order := range catalog.portability {
		if _, present := orders[order]; !present {
			return nil, fmt.Errorf("indicators: source-state catalog is missing portable order %d", order)
		}
	}

	sort.Slice(catalog.authoritative, func(left, right int) bool {
		return catalog.authoritative[left].TableName < catalog.authoritative[right].TableName
	})
	sort.Slice(catalog.rebuildable, func(left, right int) bool {
		return catalog.rebuildable[left].TableName < catalog.rebuildable[right].TableName
	})
	sort.Slice(catalog.portability, func(left, right int) bool {
		return catalog.portability[left].Order < catalog.portability[right].Order
	})
	return catalog, nil
}

func admitRelation(relations map[string]struct{}, tableName string) error {
	if !identifierPattern.MatchString(tableName) {
		return fmt.Errorf("indicators: source-state catalog contains invalid relation %q", tableName)
	}
	if _, duplicate := relations[tableName]; duplicate {
		return fmt.Errorf("indicators: source-state catalog contains duplicate relation %s", tableName)
	}
	relations[tableName] = struct{}{}
	return nil
}

func validatePortabilityDescriptor(
	descriptor PortabilityDescriptor,
	paths map[string]struct{},
	schemas map[string]struct{},
	orders map[int]struct{},
) error {
	if descriptor.Order < 0 {
		return fmt.Errorf("indicators: portable path %s has invalid order %d", descriptor.LogicalPath, descriptor.Order)
	}
	if _, duplicate := orders[descriptor.Order]; duplicate {
		return fmt.Errorf("indicators: source-state catalog contains duplicate portable order %d", descriptor.Order)
	}
	if descriptor.LogicalPath != path.Clean(descriptor.LogicalPath) ||
		!strings.HasPrefix(descriptor.LogicalPath, "data/") ||
		strings.Count(descriptor.LogicalPath, "/") != 1 ||
		!strings.HasSuffix(descriptor.LogicalPath, ".ndjson") {
		return fmt.Errorf("indicators: source-state catalog contains invalid portable path %q", descriptor.LogicalPath)
	}
	base := strings.TrimSuffix(strings.TrimPrefix(descriptor.LogicalPath, "data/"), ".ndjson")
	if !identifierPattern.MatchString(base) {
		return fmt.Errorf("indicators: source-state catalog contains invalid portable path %q", descriptor.LogicalPath)
	}
	wantSchema := "cartulary.incident_bundle." + base + ".row.v1"
	if descriptor.SchemaID != wantSchema {
		return fmt.Errorf("indicators: portable path %s has schema %s, want %s", descriptor.LogicalPath, descriptor.SchemaID, wantSchema)
	}
	if descriptor.ContentRole != "source_rows" {
		return fmt.Errorf("indicators: portable path %s has invalid content role %q", descriptor.LogicalPath, descriptor.ContentRole)
	}
	if len(descriptor.Versions) != 1 || descriptor.Versions[0] != portableBundleVersion {
		return fmt.Errorf("indicators: portable path %s has invalid bundle versions %v", descriptor.LogicalPath, descriptor.Versions)
	}
	if len(descriptor.StableIdentity) == 0 {
		return fmt.Errorf("indicators: portable path %s has empty stable identity", descriptor.LogicalPath)
	}
	identities := make(map[string]struct{}, len(descriptor.StableIdentity))
	for _, identity := range descriptor.StableIdentity {
		if !identifierPattern.MatchString(identity) {
			return fmt.Errorf("indicators: portable path %s has invalid stable identity %q", descriptor.LogicalPath, identity)
		}
		if _, duplicate := identities[identity]; duplicate {
			return fmt.Errorf("indicators: portable path %s has duplicate stable identity %s", descriptor.LogicalPath, identity)
		}
		identities[identity] = struct{}{}
	}
	if descriptor.StableIdentityInvariantID != "indicators.source_identity_admitted" {
		return fmt.Errorf("indicators: portable path %s has invalid stable identity invariant %q", descriptor.LogicalPath, descriptor.StableIdentityInvariantID)
	}
	if _, duplicate := paths[descriptor.LogicalPath]; duplicate {
		return fmt.Errorf("indicators: source-state catalog contains duplicate portable path %s", descriptor.LogicalPath)
	}
	if _, duplicate := schemas[descriptor.SchemaID]; duplicate {
		return fmt.Errorf("indicators: source-state catalog contains duplicate portable schema %s", descriptor.SchemaID)
	}
	paths[descriptor.LogicalPath] = struct{}{}
	schemas[descriptor.SchemaID] = struct{}{}
	orders[descriptor.Order] = struct{}{}
	return nil
}

func clonePortability(input []PortabilityDescriptor) []PortabilityDescriptor {
	result := make([]PortabilityDescriptor, len(input))
	for index, descriptor := range input {
		result[index] = descriptor
		result[index].Versions = slices.Clone(descriptor.Versions)
		result[index].StableIdentity = slices.Clone(descriptor.StableIdentity)
	}
	return result
}
