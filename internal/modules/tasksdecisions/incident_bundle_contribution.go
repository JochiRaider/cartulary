package tasksdecisions

import (
	"github.com/JochiRaider/cartulary/internal/modules/incidentbundles/sourceport"
	"github.com/JochiRaider/cartulary/internal/modules/records/subtypepresence"
	"github.com/JochiRaider/cartulary/internal/modules/tasksdecisions/internal/envelopefacts"
	"github.com/JochiRaider/cartulary/internal/modules/tasksdecisions/internal/providers/incidentbundle"
)

type EnvelopeFact = envelopefacts.Fact
type EnvelopeFactReader = envelopefacts.Reader

func NewIncidentBundleSourcePort(linkFacts LinkFactsCapability, fieldReferences FieldReferenceReader, envelopes EnvelopeFactReader) (sourceport.Port, error) {
	return incidentbundle.NewSourcePort(linkFacts, fieldReferences, envelopes)
}

func IncidentBundleSubtypeContribution() subtypepresence.Contribution {
	return incidentbundle.SubtypeContribution()
}
