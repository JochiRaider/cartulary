package sourcestate

import (
	"errors"
	"fmt"
	"regexp"
	"slices"
	"strings"

	"github.com/JochiRaider/cartulary/internal/modules/incidentbundles/sourceport"
)

var errInvalidManifest = errors.New("links: invalid source-state manifest")

type relationKind uint8

const (
	relationInvalid relationKind = iota
	relationRecordLinks
	relationRecordTags
)

type pathKind uint8

const (
	pathInvalid pathKind = iota
	pathRecordLinks
	pathTagCatalog
	pathRecordTags
)

type relationSpec struct {
	kind    relationKind
	table   string
	columns []string
}

type pathSpec struct {
	kind            pathKind
	logicalPath     string
	contentRole     string
	versions        []int
	stableIdentity  []string
	relation        relationKind
	requiredColumns []string
	allowedColumns  []string
}

type manifestInput struct {
	relations  []relationSpec
	paths      []pathSpec
	invariants []string
}

type manifest struct {
	relations  []relationSpec
	paths      []pathSpec
	invariants []string
}

var safeSQLIdentifier = regexp.MustCompile(`^[a-z][a-z0-9_]*$`)

var expectedInvariants = []string{
	"links_tags.endpoints_same_incident",
	"links_tags.link_tuple_legal",
	"links_tags.link_unique",
	"links_tags.deletion_tuple_legal",
	"links_tags.tag_normalized",
	"links_tags.tag_catalog_exact",
	"links_tags.source_identity_admitted",
}

func loadManifest() (manifest, error) {
	return validateManifest(authoredManifestInput())
}

func authoredManifestInput() manifestInput {
	return manifestInput{
		relations: []relationSpec{
			{
				kind:  relationRecordLinks,
				table: "record_links",
				columns: []string{
					"record_link_id", "incident_id", "src_record_id", "dst_record_id",
					"link_type", "field_key", "provenance", "confidence", "owner_user_id",
					"created_by_user_id", "decided_at", "created_at", "deleted_at", "deleted_by_user_id",
				},
			},
			{
				kind:  relationRecordTags,
				table: "record_tags",
				columns: []string{
					"record_tag_id", "incident_id", "record_id", "tag_name", "normalized_tag_name",
					"created_by_user_id", "created_at", "updated_at", "deleted_at", "deleted_by_user_id",
				},
			},
		},
		paths: []pathSpec{
			{
				kind: pathRecordLinks, logicalPath: "data/record_links.ndjson", contentRole: "source_rows",
				versions: []int{3}, stableIdentity: []string{"record_link_id"}, relation: relationRecordLinks,
				requiredColumns: []string{"record_link_id", "incident_id"},
			},
			{
				kind: pathTagCatalog, logicalPath: "data/tags.ndjson", contentRole: "validation_rows",
				versions: []int{3}, stableIdentity: []string{"normalized_tag_name", "tag_name"},
				allowedColumns: []string{"tag_name", "normalized_tag_name"},
			},
			{
				kind: pathRecordTags, logicalPath: "data/record_tags.ndjson", contentRole: "source_rows",
				versions: []int{3}, stableIdentity: []string{"record_tag_id"}, relation: relationRecordTags,
				requiredColumns: []string{"record_tag_id", "record_id", "incident_id"},
			},
		},
		invariants: slices.Clone(expectedInvariants),
	}
}

func validateManifest(input manifestInput) (manifest, error) {
	if len(input.relations) != 2 || len(input.paths) != 3 || !slices.Equal(input.invariants, expectedInvariants) {
		return manifest{}, errInvalidManifest
	}
	expectedRelations := []relationKind{relationRecordLinks, relationRecordTags}
	seenRelations := map[relationKind]struct{}{}
	for index, relation := range input.relations {
		if relation.kind != expectedRelations[index] || !safeSQLIdentifier.MatchString(relation.table) || len(relation.columns) == 0 {
			return manifest{}, errInvalidManifest
		}
		if _, duplicate := seenRelations[relation.kind]; duplicate {
			return manifest{}, errInvalidManifest
		}
		seenRelations[relation.kind] = struct{}{}
		if !validUniqueIdentifiers(relation.columns) {
			return manifest{}, errInvalidManifest
		}
	}
	expectedPaths := []pathKind{pathRecordLinks, pathTagCatalog, pathRecordTags}
	seenPaths := map[string]struct{}{}
	for index, path := range input.paths {
		if path.kind != expectedPaths[index] || !safeLogicalPath(path.logicalPath) ||
			(path.contentRole != "source_rows" && path.contentRole != "validation_rows") ||
			!slices.Equal(path.versions, []int{3}) || !validUniqueIdentifiers(path.stableIdentity) {
			return manifest{}, errInvalidManifest
		}
		if _, duplicate := seenPaths[path.logicalPath]; duplicate {
			return manifest{}, errInvalidManifest
		}
		seenPaths[path.logicalPath] = struct{}{}
		if path.kind == pathTagCatalog {
			if path.relation != relationInvalid || path.contentRole != "validation_rows" ||
				!slices.Equal(path.allowedColumns, []string{"tag_name", "normalized_tag_name"}) {
				return manifest{}, errInvalidManifest
			}
			continue
		}
		relation, ok := relationByKind(input.relations, path.relation)
		if !ok || path.contentRole != "source_rows" || !slices.Equal(path.allowedColumns, nil) ||
			!columnsContained(path.requiredColumns, relation.columns) {
			return manifest{}, errInvalidManifest
		}
	}
	return manifest{
		relations:  cloneRelations(input.relations),
		paths:      clonePaths(input.paths, input.relations),
		invariants: slices.Clone(input.invariants),
	}, nil
}

func (value manifest) descriptor() sourceport.Descriptor {
	paths := make([]sourceport.Path, 0, len(value.paths))
	for _, path := range value.paths {
		paths = append(paths, sourceport.Path{
			LogicalPath: path.logicalPath, ContentRole: path.contentRole,
			Versions: slices.Clone(path.versions), StableIdentity: slices.Clone(path.stableIdentity),
			StableIdentityInvariantID: "links_tags.source_identity_admitted",
		})
	}
	return sourceport.Descriptor{
		FamilyID: "links_tags", ContractMajor: sourceport.ContractMajor,
		OwnerID: "module.links", OwnerRelationIDs: []string{"links-and-tags"},
		Dependencies: []string{"assessments"}, Paths: paths,
		InvariantIDs: slices.Clone(value.invariants),
	}
}

func (value manifest) tableNames() []string {
	result := make([]string, 0, len(value.relations))
	for _, relation := range value.relations {
		result = append(result, relation.table)
	}
	return result
}

func (value manifest) pathSpecs() []pathSpec {
	return clonePaths(value.paths, value.relations)
}

func relationByKind(relations []relationSpec, kind relationKind) (relationSpec, bool) {
	for _, relation := range relations {
		if relation.kind == kind {
			return relation, true
		}
	}
	return relationSpec{}, false
}

func cloneRelations(source []relationSpec) []relationSpec {
	result := make([]relationSpec, len(source))
	for index, relation := range source {
		result[index] = relation
		result[index].columns = slices.Clone(relation.columns)
	}
	return result
}

func clonePaths(source []pathSpec, relations []relationSpec) []pathSpec {
	result := make([]pathSpec, len(source))
	for index, path := range source {
		result[index] = path
		result[index].versions = slices.Clone(path.versions)
		result[index].stableIdentity = slices.Clone(path.stableIdentity)
		result[index].requiredColumns = slices.Clone(path.requiredColumns)
		result[index].allowedColumns = slices.Clone(path.allowedColumns)
		if result[index].relation != relationInvalid && result[index].allowedColumns == nil {
			if relation, ok := relationByKind(relations, result[index].relation); ok {
				result[index].allowedColumns = slices.Clone(relation.columns)
			}
		}
	}
	return result
}

func validUniqueIdentifiers(values []string) bool {
	if len(values) == 0 {
		return false
	}
	seen := map[string]struct{}{}
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

func columnsContained(required []string, allowed []string) bool {
	if !validUniqueIdentifiers(required) {
		return false
	}
	for _, column := range required {
		if !slices.Contains(allowed, column) {
			return false
		}
	}
	return true
}

func safeLogicalPath(value string) bool {
	return strings.HasPrefix(value, "data/") && strings.HasSuffix(value, ".ndjson") &&
		!strings.Contains(value, "..") && !strings.ContainsAny(value, "\\\x00") &&
		strings.TrimSpace(value) == value
}

func manifestError(detail string) error {
	return fmt.Errorf("%w: %s", errInvalidManifest, detail)
}
