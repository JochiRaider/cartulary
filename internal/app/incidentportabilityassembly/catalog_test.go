package incidentportabilityassembly

import (
	"errors"
	"strings"
	"testing"

	"github.com/JochiRaider/cartulary/internal/modules/incidentbundles/sourceport"
	referencefixture "github.com/JochiRaider/cartulary/internal/modules/reference_data/testsupport"
	"github.com/JochiRaider/cartulary/internal/testutil/revisionsupport"
)

func TestIncidentsSourcePortConstructionFailureHasOwnerContext_Unit(t *testing.T) {
	sentinel := errors.New("invalid manifest")
	port, err := constructIncidentsSourcePort(func() (sourceport.Port, error) {
		return nil, sentinel
	})
	if port != nil || !errors.Is(err, sentinel) ||
		!strings.Contains(err.Error(), "incident portability assembly: Incidents source port") {
		t.Fatalf("contextual source-port failure = port %#v, error %v", port, err)
	}
}

func TestCatalogRequiresTheComposedRevisionsSourcePort_Unit(t *testing.T) {
	if catalog, err := NewCatalog(nil, referencefixture.IndicatorRegistry{}, referencefixture.IndicatorRegistry{}); catalog != nil || !errors.Is(err, sourceport.ErrInvalidCatalog) {
		t.Fatalf("nil Revisions source port = catalog %#v, error %v", catalog, err)
	}
	port := revisionsupport.MustRuntime(t).IncidentBundleSourcePort()
	validator, ok := port.(sourceport.ContractValidator)
	if !ok {
		t.Fatal("composed Revisions port lacks contract validation")
	}
	complete := sourceOwnerOverride{Port: port, ContractValidator: validator, ownerID: "module.revisions"}
	if catalog, err := NewCatalog(complete, referencefixture.IndicatorRegistry{}, referencefixture.IndicatorRegistry{}); catalog == nil || err != nil {
		t.Fatalf("complete Revisions source port = catalog %#v, error %v", catalog, err)
	}
	complete.ownerID = "module.fixture"
	if catalog, err := NewCatalog(complete, referencefixture.IndicatorRegistry{}, referencefixture.IndicatorRegistry{}); catalog != nil || !errors.Is(err, sourceport.ErrInvalidCatalog) {
		t.Fatalf("complete wrong-owner Revisions source port = catalog %#v, error %v", catalog, err)
	}
}

type sourceOwnerOverride struct {
	sourceport.Port
	sourceport.ContractValidator
	ownerID string
}

func (p sourceOwnerOverride) Descriptor() sourceport.Descriptor {
	descriptor := p.Port.Descriptor()
	descriptor.OwnerID = p.ownerID
	return descriptor
}
