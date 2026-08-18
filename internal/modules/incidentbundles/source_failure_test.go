package incidentbundles

import (
	"errors"
	"testing"

	"github.com/JochiRaider/cartulary/internal/modules/incidentbundles/sourceport"
)

func TestSourcePortFailuresRequireDeclaredFamilyAndInvariant(t *testing.T) {
	t.Parallel()
	descriptor := sourceport.Descriptor{
		FamilyID: "fixture", ContractMajor: sourceport.ContractMajor,
		OwnerID: "module.fixture",
		Paths: []sourceport.Path{{
			LogicalPath: "data/fixture.ndjson", ContentRole: "source_rows",
			Versions: []int{2}, StableIdentity: []string{"id"},
			StableIdentityInvariantID: "fixture.declared",
		}},
		InvariantIDs: []string{"fixture.declared"},
	}
	port := sourceport.NewAdapter(sourceport.AdapterOptions{Descriptor: descriptor})
	attackerDescriptor := sourceport.Descriptor{
		FamilyID: "attacker", InvariantIDs: []string{"attacker.declared"},
	}
	undeclaredDescriptor := descriptor
	undeclaredDescriptor.InvariantIDs = append(undeclaredDescriptor.InvariantIDs, "fixture.attacker_selected")
	crossFamily := attackerDescriptor.DeclaredFailure("attacker.declared")
	undeclared := undeclaredDescriptor.DeclaredFailure("fixture.attacker_selected")

	for _, failure := range []error{
		crossFamily,
		undeclared,
	} {
		err := verificationErrorFromDeclaredPort(port, failure)
		var public *verificationError
		if errors.As(err, &public) {
			t.Fatalf("undeclared source failure became public: %#v", public)
		}
	}

	err := verificationErrorFromDeclaredPort(port, descriptor.DeclaredFailure("fixture.declared"))
	var public *verificationError
	if !errors.As(err, &public) || public.ReasonCode != "source_family_invalid" ||
		public.SourceFamilyID != "fixture" || public.InvariantID != "fixture.declared" {
		t.Fatalf("declared source failure = %#v, %v", public, err)
	}
}
