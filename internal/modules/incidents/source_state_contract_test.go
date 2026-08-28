package incidents

import (
	"slices"
	"testing"

	"github.com/JochiRaider/cartulary/internal/modules/incidents/internal/sourcestate"
)

func TestSourcePortAndRecoveryContributionShareOneCatalog_Unit(t *testing.T) {
	port, err := NewIncidentBundleSourcePort()
	if err != nil {
		t.Fatalf("construct source port: %v", err)
	}
	contribution, err := RecoveryStateContribution()
	if err != nil {
		t.Fatalf("construct Recovery contribution: %v", err)
	}
	columns, err := sourcestate.IncidentColumns()
	if err != nil {
		t.Fatalf("load incident columns: %v", err)
	}

	descriptor := port.Descriptor()
	if descriptor.OwnerID != contribution.OwnerID || descriptor.Paths[0].SchemaID == "" {
		t.Fatalf("source and Recovery owner projections disagree: descriptor=%#v contribution=%#v", descriptor, contribution)
	}
	if len(columns) != 16 || columns[0] != "id" || columns[len(columns)-1] != "closed_at" {
		t.Fatalf("incident column catalog = %#v", columns)
	}
	gotRelations := make([]string, 0, len(contribution.Tables))
	for _, table := range contribution.Tables {
		gotRelations = append(gotRelations, table.TableName)
	}
	if !slices.Equal(gotRelations, []string{"incidents", "incident_memberships"}) {
		t.Fatalf("Recovery relations = %#v", gotRelations)
	}
}
