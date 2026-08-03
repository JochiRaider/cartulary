package incidentbundles

import (
	"errors"
	"testing"

	"github.com/JochiRaider/cartulary/internal/modules/incidentbundles/sourceport"
)

func TestSourcePortFailuresRequireDeclaredFamilyAndInvariant(t *testing.T) {
	t.Parallel()
	port := sourceport.NewAdapter(sourceport.AdapterOptions{Descriptor: sourceport.Descriptor{
		FamilyID: "fixture", ContractMajor: sourceport.ContractMajor,
		OwnerID: "module.fixture",
		Paths: []sourceport.Path{{
			LogicalPath: "data/fixture.ndjson", ContentRole: "source_rows",
			Versions: []int{2}, StableIdentity: []string{"id"},
		}},
		InvariantIDs: []string{"fixture.declared"},
	}})

	for _, failure := range []*sourceport.Failure{
		{FamilyID: "attacker", InvariantID: "fixture.declared"},
		{FamilyID: "fixture", InvariantID: "fixture.attacker_selected"},
	} {
		err := verificationErrorFromDeclaredPort(port, failure)
		var public *VerificationError
		if errors.As(err, &public) {
			t.Fatalf("undeclared source failure became public: %#v", public)
		}
	}

	err := verificationErrorFromDeclaredPort(port, &sourceport.Failure{
		FamilyID: "fixture", InvariantID: "fixture.declared",
	})
	var public *VerificationError
	if !errors.As(err, &public) || public.ReasonCode != "source_family_invalid" ||
		public.SourceFamilyID != "fixture" || public.InvariantID != "fixture.declared" {
		t.Fatalf("declared source failure = %#v, %v", public, err)
	}
}
