package producer

import (
	"github.com/JochiRaider/cartulary/internal/gen/contractreferencepacks"
	"strings"
	"testing"
)

// This projection assertion is separate from acquisition's scratch mutations.
func assertATTACKProducerExcludedFromRuntime(t *testing.T) {
	t.Helper()
	for _, artifact := range contractreferencepacks.Artifacts {
		if strings.Contains(artifact.Path, "reference-pack-producers") || strings.Contains(artifact.JSON, "cartulary.attack.native.candidate") {
			t.Fatal("producer candidate leaked into runtime projection")
		}
	}
}
