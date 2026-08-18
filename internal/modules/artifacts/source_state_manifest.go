package artifacts

import (
	"errors"
	"fmt"
	"regexp"
	"slices"
	"strings"
	"sync"

	"github.com/JochiRaider/cartulary/internal/modules/incidentbundles/sourceport"
	"github.com/JochiRaider/cartulary/internal/modules/incidentportability"
	recoverystate "github.com/JochiRaider/cartulary/internal/platform/recoverystate"
)

var (
	errInvalidSourceStateManifest = errors.New("artifacts source-state manifest is invalid")
	sourceStateIdentifierPattern  = regexp.MustCompile(`^[a-z][a-z0-9_]*$`)
	sourceStatePathPattern        = regexp.MustCompile(`^data/[a-z][a-z0-9_]*\.ndjson$`)

	sourceStateManifestOnce      sync.Once
	cachedSourceStateManifest    *sourceStateManifest
	cachedSourceStateManifestErr error
)

type sourceStateRelation struct {
	tableName             string
	logicalBundlePath     string
	supportedVersions     []int
	stableIdentity        []string
	requiredImportColumns []string
}

type sourceStateManifest struct {
	relations []sourceStateRelation
}

type sourceStateExportSpecification struct {
	logicalBundlePath string
	query             string
}

var authoredSourceStateRelations = []sourceStateRelation{
	{
		tableName:             "artifacts",
		logicalBundlePath:     "data/artifacts.ndjson",
		supportedVersions:     []int{2},
		stableIdentity:        []string{"record_id"},
		requiredImportColumns: []string{"record_id", "incident_id"},
	},
	{
		tableName:             "artifact_findings",
		logicalBundlePath:     "data/artifact_findings.ndjson",
		supportedVersions:     []int{2},
		stableIdentity:        []string{"record_id"},
		requiredImportColumns: []string{"record_id", "incident_id"},
	},
	{
		tableName:             "artifact_investigative_queries",
		logicalBundlePath:     "data/artifact_investigative_queries.ndjson",
		supportedVersions:     []int{2},
		stableIdentity:        []string{"record_id"},
		requiredImportColumns: []string{"record_id", "incident_id"},
	},
	{
		tableName:             "artifact_forensic_keywords",
		logicalBundlePath:     "data/artifact_forensic_keywords.ndjson",
		supportedVersions:     []int{2},
		stableIdentity:        []string{"record_id"},
		requiredImportColumns: []string{"record_id", "incident_id"},
	},
	{
		tableName:             "handoff_risk_refs",
		logicalBundlePath:     "data/handoff_risk_refs.ndjson",
		supportedVersions:     []int{2},
		stableIdentity:        []string{"risk_ref_id"},
		requiredImportColumns: []string{"risk_ref_id", "handoff_record_id"},
	},
}

func loadSourceStateManifest() (*sourceStateManifest, error) {
	sourceStateManifestOnce.Do(func() {
		cachedSourceStateManifest, cachedSourceStateManifestErr = newSourceStateManifest(authoredSourceStateRelations)
	})
	return cachedSourceStateManifest, cachedSourceStateManifestErr
}

func newSourceStateManifest(relations []sourceStateRelation) (*sourceStateManifest, error) {
	if len(relations) == 0 {
		return nil, fmt.Errorf("%w: no relations", errInvalidSourceStateManifest)
	}
	tableNames := make(map[string]struct{}, len(relations))
	logicalPaths := make(map[string]struct{}, len(relations))
	normalized := make([]sourceStateRelation, len(relations))
	for index, relation := range relations {
		if err := validateSourceStateRelation(relation); err != nil {
			return nil, fmt.Errorf("%w: relation %d: %v", errInvalidSourceStateManifest, index, err)
		}
		if _, duplicate := tableNames[relation.tableName]; duplicate {
			return nil, fmt.Errorf("%w: duplicate table %q", errInvalidSourceStateManifest, relation.tableName)
		}
		tableNames[relation.tableName] = struct{}{}
		if _, duplicate := logicalPaths[relation.logicalBundlePath]; duplicate {
			return nil, fmt.Errorf("%w: duplicate logical path %q", errInvalidSourceStateManifest, relation.logicalBundlePath)
		}
		logicalPaths[relation.logicalBundlePath] = struct{}{}
		normalized[index] = cloneSourceStateRelation(relation)
	}
	return &sourceStateManifest{relations: normalized}, nil
}

func validateSourceStateRelation(relation sourceStateRelation) error {
	if !sourceStateIdentifierPattern.MatchString(relation.tableName) {
		return fmt.Errorf("unsafe or empty table identifier %q", relation.tableName)
	}
	if !sourceStatePathPattern.MatchString(relation.logicalBundlePath) {
		return fmt.Errorf("unsafe or empty logical path %q", relation.logicalBundlePath)
	}
	if len(relation.supportedVersions) == 0 {
		return errors.New("supported versions are empty")
	}
	for index, version := range relation.supportedVersions {
		if version < 1 || (index > 0 && version <= relation.supportedVersions[index-1]) {
			return fmt.Errorf("supported versions are not strictly increasing: %v", relation.supportedVersions)
		}
	}
	if err := validateSourceStateIdentifiers("stable identity", relation.stableIdentity); err != nil {
		return err
	}
	if err := validateSourceStateIdentifiers("required import column", relation.requiredImportColumns); err != nil {
		return err
	}
	for _, identity := range relation.stableIdentity {
		if !slices.Contains(relation.requiredImportColumns, identity) {
			return fmt.Errorf("stable identity %q is not a required import column", identity)
		}
	}
	return nil
}

func validateSourceStateIdentifiers(kind string, identifiers []string) error {
	if len(identifiers) == 0 {
		return fmt.Errorf("%s list is empty", kind)
	}
	seen := make(map[string]struct{}, len(identifiers))
	for _, identifier := range identifiers {
		if !sourceStateIdentifierPattern.MatchString(identifier) {
			return fmt.Errorf("unsafe or empty %s %q", kind, identifier)
		}
		if _, duplicate := seen[identifier]; duplicate {
			return fmt.Errorf("duplicate %s %q", kind, identifier)
		}
		seen[identifier] = struct{}{}
	}
	return nil
}

func (manifest *sourceStateManifest) recoveryTables() []recoverystate.Table {
	tableNames := make([]string, 0, len(manifest.relations))
	for _, relation := range manifest.relations {
		tableNames = append(tableNames, relation.tableName)
	}
	return recoverystate.AuthoritativeTables(tableNames...)
}

func (manifest *sourceStateManifest) sourcePortPaths() []sourceport.Path {
	paths := make([]sourceport.Path, 0, len(manifest.relations))
	for _, relation := range manifest.relations {
		paths = append(paths, sourceport.Path{
			LogicalPath:               relation.logicalBundlePath,
			ContentRole:               "source_rows",
			Versions:                  append([]int(nil), relation.supportedVersions...),
			StableIdentity:            append([]string(nil), relation.stableIdentity...),
			StableIdentityInvariantID: "artifacts.source_identity_admitted",
		})
	}
	return paths
}

func (manifest *sourceStateManifest) exportSpecifications() []sourceStateExportSpecification {
	specifications := make([]sourceStateExportSpecification, 0, len(manifest.relations))
	for _, relation := range manifest.relations {
		orderBy := make([]string, 0, len(relation.stableIdentity))
		for _, identifier := range relation.stableIdentity {
			orderBy = append(orderBy, quoteSourceStateIdentifier(identifier))
		}
		specifications = append(specifications, sourceStateExportSpecification{
			logicalBundlePath: relation.logicalBundlePath,
			query: fmt.Sprintf(
				"SELECT to_jsonb(t) FROM %s AS t WHERE %s = $1 ORDER BY %s",
				quoteSourceStateIdentifier(relation.tableName),
				quoteSourceStateIdentifier("incident_id"),
				strings.Join(orderBy, ", "),
			),
		})
	}
	return specifications
}

func (manifest *sourceStateManifest) importSpecifications() []incidentportability.FixedImportSpec {
	specifications := make([]incidentportability.FixedImportSpec, 0, len(manifest.relations))
	for _, relation := range manifest.relations {
		quotedTable := quoteSourceStateIdentifier(relation.tableName)
		specifications = append(specifications, incidentportability.FixedImportSpec{
			LogicalBundlePath: relation.logicalBundlePath,
			AttributionTable:  relation.tableName,
			StableIdentity:    append([]string(nil), relation.stableIdentity...),
			RequiredColumns:   append([]string(nil), relation.requiredImportColumns...),
			InsertSQL: fmt.Sprintf(
				"INSERT INTO %s SELECT * FROM jsonb_populate_record(NULL::%s, $1::jsonb)",
				quotedTable,
				quotedTable,
			),
		})
	}
	return specifications
}

func quoteSourceStateIdentifier(identifier string) string {
	return `"` + identifier + `"`
}

func cloneSourceStateRelation(relation sourceStateRelation) sourceStateRelation {
	relation.supportedVersions = append([]int(nil), relation.supportedVersions...)
	relation.stableIdentity = append([]string(nil), relation.stableIdentity...)
	relation.requiredImportColumns = append([]string(nil), relation.requiredImportColumns...)
	return relation
}
