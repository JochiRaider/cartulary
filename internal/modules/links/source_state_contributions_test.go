package links_test

import (
	"reflect"
	"testing"

	"github.com/JochiRaider/cartulary/internal/modules/links"
)

func TestSourceStateContributionsRemainExact_Unit(t *testing.T) {
	port, err := links.NewIncidentBundleSourcePort()
	if err != nil {
		t.Fatalf("construct Links source port: %v", err)
	}
	descriptor := port.Descriptor()
	gotPaths := make([]string, 0, len(descriptor.Paths))
	gotRoles := make([]string, 0, len(descriptor.Paths))
	gotIdentities := make([][]string, 0, len(descriptor.Paths))
	for _, path := range descriptor.Paths {
		gotPaths = append(gotPaths, path.LogicalPath)
		gotRoles = append(gotRoles, path.ContentRole)
		gotIdentities = append(gotIdentities, path.StableIdentity)
		if !reflect.DeepEqual(path.Versions, []int{3}) {
			t.Fatalf("path %q versions = %v, want [3]", path.LogicalPath, path.Versions)
		}
	}
	if want := []string{"data/record_links.ndjson", "data/tags.ndjson", "data/record_tags.ndjson"}; !reflect.DeepEqual(gotPaths, want) {
		t.Fatalf("descriptor paths = %v, want %v", gotPaths, want)
	}
	if want := []string{"source_rows", "validation_rows", "source_rows"}; !reflect.DeepEqual(gotRoles, want) {
		t.Fatalf("descriptor roles = %v, want %v", gotRoles, want)
	}
	if want := [][]string{{"record_link_id"}, {"normalized_tag_name", "tag_name"}, {"record_tag_id"}}; !reflect.DeepEqual(gotIdentities, want) {
		t.Fatalf("descriptor identities = %v, want %v", gotIdentities, want)
	}
	if want := []string{
		"links_tags.endpoints_same_incident",
		"links_tags.link_tuple_legal",
		"links_tags.link_unique",
		"links_tags.deletion_tuple_legal",
		"links_tags.tag_normalized",
		"links_tags.tag_catalog_exact",
		"links_tags.source_identity_admitted",
	}; !reflect.DeepEqual(descriptor.InvariantIDs, want) {
		t.Fatalf("descriptor invariants = %v, want %v", descriptor.InvariantIDs, want)
	}

	contribution, err := links.RecoveryStateContribution()
	if err != nil {
		t.Fatalf("construct Links Recovery contribution: %v", err)
	}
	if contribution.OwnerID != "module.links" || len(contribution.Tables) != 2 {
		t.Fatalf("Recovery contribution = %#v, want module.links with two tables", contribution)
	}
	gotTables := []string{contribution.Tables[0].TableName, contribution.Tables[1].TableName}
	if want := []string{"record_links", "record_tags"}; !reflect.DeepEqual(gotTables, want) {
		t.Fatalf("Recovery tables = %v, want %v", gotTables, want)
	}
}
