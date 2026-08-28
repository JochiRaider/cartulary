package sourcestate

import (
	"errors"
	"fmt"
	"regexp"
	"slices"
	"strings"
	"sync"
)

const currentContractMajor = 2

var errInvalidCatalog = errors.New("incidents: invalid source-state catalog")

var (
	safeSQLIdentifier = regexp.MustCompile(`^[a-z][a-z0-9_]*$`)
	safeSchemaID      = regexp.MustCompile(`^cartulary\.[A-Za-z0-9_.-]+\.v[1-9][0-9]*$`)
)

var expectedIncidentColumns = []string{
	"id",
	"incident_key",
	"incident_key_canonical",
	"title",
	"description",
	"status",
	"severity",
	"tlp",
	"current_phase",
	"primary_external_case_ref",
	"created_by_user_id",
	"created_at",
	"updated_at",
	"updated_by_user_id",
	"incident_version",
	"closed_at",
}

var expectedInvariants = []string{
	"incident.source_identity_admitted",
	"incident.exact_shape",
	"incident.identity_key_lifecycle",
	"incident.attribution_version",
}

var expectedRecoveryRelations = []string{"incidents", "incident_memberships"}

type pathSpec struct {
	logicalPath               string
	contentRole               string
	schemaID                  string
	versions                  []int
	stableIdentity            []string
	stableIdentityInvariantID string
	columns                   []string
}

type manifestInput struct {
	familyID          string
	contractMajor     int
	ownerID           string
	ownerRelationIDs  []string
	path              pathSpec
	invariants        []string
	recoveryRelations []string
}

type manifest struct {
	familyID          string
	contractMajor     int
	ownerID           string
	ownerRelationIDs  []string
	path              pathSpec
	invariants        []string
	recoveryRelations []string
}

type SourcePath struct {
	LogicalPath               string
	ContentRole               string
	SchemaID                  string
	Versions                  []int
	StableIdentity            []string
	StableIdentityInvariantID string
}

type SourceDescriptor struct {
	FamilyID         string
	ContractMajor    int
	OwnerID          string
	OwnerRelationIDs []string
	Path             SourcePath
	InvariantIDs     []string
}

type RecoveryDescriptor struct {
	OwnerID   string
	Relations []string
}

var loadManifestOnce = sync.OnceValues(func() (manifest, error) {
	return validateManifest(authoredManifestInput())
})

func authoredManifestInput() manifestInput {
	return manifestInput{
		familyID:         "incident",
		contractMajor:    currentContractMajor,
		ownerID:          "module.incidents",
		ownerRelationIDs: []string{"incident-core"},
		path: pathSpec{
			logicalPath:               "data/incident.json",
			contentRole:               "singleton_json",
			schemaID:                  "cartulary.incident_bundle.incident.row.v1",
			versions:                  []int{3},
			stableIdentity:            []string{"id"},
			stableIdentityInvariantID: "incident.source_identity_admitted",
			columns:                   slices.Clone(expectedIncidentColumns),
		},
		invariants:        slices.Clone(expectedInvariants),
		recoveryRelations: slices.Clone(expectedRecoveryRelations),
	}
}

func loadManifest() (manifest, error) {
	return loadManifestOnce()
}

func validateManifest(input manifestInput) (manifest, error) {
	if input.familyID != "incident" || input.contractMajor != currentContractMajor ||
		input.ownerID != "module.incidents" || !slices.Equal(input.ownerRelationIDs, []string{"incident-core"}) {
		return manifest{}, catalogError("owner or contract identity drift")
	}
	path := input.path
	if path.logicalPath != "data/incident.json" || !safeLogicalPath(path.logicalPath) ||
		path.contentRole != "singleton_json" || !safeSchemaID.MatchString(path.schemaID) ||
		!slices.Equal(path.versions, []int{3}) || !slices.Equal(path.stableIdentity, []string{"id"}) ||
		path.stableIdentityInvariantID != expectedInvariants[0] {
		return manifest{}, catalogError("source path, version, or schema generation drift")
	}
	if !slices.Equal(path.columns, expectedIncidentColumns) || !validUniqueIdentifiers(path.columns) {
		return manifest{}, catalogError("incident column catalog drift")
	}
	if !slices.Equal(input.invariants, expectedInvariants) || !validUniqueDottedIDs(input.invariants) {
		return manifest{}, catalogError("invariant catalog drift")
	}
	if !slices.Equal(input.recoveryRelations, expectedRecoveryRelations) ||
		!validUniqueIdentifiers(input.recoveryRelations) {
		return manifest{}, catalogError("Recovery relation catalog drift")
	}
	return manifest{
		familyID:          input.familyID,
		contractMajor:     input.contractMajor,
		ownerID:           input.ownerID,
		ownerRelationIDs:  slices.Clone(input.ownerRelationIDs),
		path:              clonePath(input.path),
		invariants:        slices.Clone(input.invariants),
		recoveryRelations: slices.Clone(input.recoveryRelations),
	}, nil
}

func Source() (SourceDescriptor, error) {
	value, err := loadManifest()
	if err != nil {
		return SourceDescriptor{}, err
	}
	return value.descriptor(), nil
}

func IncidentColumns() ([]string, error) {
	value, err := loadManifest()
	if err != nil {
		return nil, err
	}
	return slices.Clone(value.path.columns), nil
}

func Recovery() (RecoveryDescriptor, error) {
	value, err := loadManifest()
	if err != nil {
		return RecoveryDescriptor{}, err
	}
	return RecoveryDescriptor{
		OwnerID:   value.ownerID,
		Relations: slices.Clone(value.recoveryRelations),
	}, nil
}

func Validate() error {
	_, err := loadManifest()
	return err
}

func (value manifest) descriptor() SourceDescriptor {
	return SourceDescriptor{
		FamilyID:         value.familyID,
		ContractMajor:    value.contractMajor,
		OwnerID:          value.ownerID,
		OwnerRelationIDs: slices.Clone(value.ownerRelationIDs),
		Path: SourcePath{
			LogicalPath:               value.path.logicalPath,
			ContentRole:               value.path.contentRole,
			SchemaID:                  value.path.schemaID,
			Versions:                  slices.Clone(value.path.versions),
			StableIdentity:            slices.Clone(value.path.stableIdentity),
			StableIdentityInvariantID: value.path.stableIdentityInvariantID,
		},
		InvariantIDs: slices.Clone(value.invariants),
	}
}

func clonePath(value pathSpec) pathSpec {
	value.versions = slices.Clone(value.versions)
	value.stableIdentity = slices.Clone(value.stableIdentity)
	value.columns = slices.Clone(value.columns)
	return value
}

func validUniqueIdentifiers(values []string) bool {
	if len(values) == 0 {
		return false
	}
	seen := make(map[string]struct{}, len(values))
	for _, value := range values {
		if !safeSQLIdentifier.MatchString(value) {
			return false
		}
		if _, duplicate := seen[value]; duplicate {
			return false
		}
		seen[value] = struct{}{}
	}
	return true
}

func validUniqueDottedIDs(values []string) bool {
	if len(values) == 0 {
		return false
	}
	seen := make(map[string]struct{}, len(values))
	for _, value := range values {
		parts := strings.Split(value, ".")
		if len(parts) < 2 || !validUniqueIdentifierParts(parts) {
			return false
		}
		if _, duplicate := seen[value]; duplicate {
			return false
		}
		seen[value] = struct{}{}
	}
	return true
}

func validUniqueIdentifierParts(values []string) bool {
	for _, value := range values {
		if !safeSQLIdentifier.MatchString(value) {
			return false
		}
	}
	return true
}

func safeLogicalPath(value string) bool {
	return value == "data/incident.json" && !strings.Contains(value, "..") &&
		!strings.ContainsAny(value, "\\\x00") && strings.TrimSpace(value) == value
}

func catalogError(detail string) error {
	return fmt.Errorf("%w: %s", errInvalidCatalog, detail)
}
