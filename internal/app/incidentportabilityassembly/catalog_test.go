package incidentportabilityassembly

import (
	"errors"
	"strings"
	"testing"

	"github.com/JochiRaider/cartulary/internal/modules/incidentbundles/sourceport"
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
