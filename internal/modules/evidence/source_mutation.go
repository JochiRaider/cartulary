package evidence

import (
	"errors"

	"github.com/JochiRaider/cartulary/internal/modules/collaboration"
	evidenceprojection "github.com/JochiRaider/cartulary/internal/modules/evidence/projectionports"
	"github.com/JochiRaider/cartulary/internal/modules/revisions"
	"github.com/JochiRaider/cartulary/internal/platform/postgres"
)

type sourceMutationService struct {
	source    evidenceSourceKernel
	mutations evidenceSourceMutationKernel
}

func newSourceMutationService(
	pool postgres.DB,
	projectionRows evidenceprojection.MutationRows,
	appender *revisions.Appender,
	intents collaboration.RecordChangedAppender,
	incidentState evidenceIncidentAdmissionPort,
	recordEnvelopes evidenceRecordEnvelopePort,
) (*sourceMutationService, error) {
	if pool == nil {
		return nil, errors.New("compose Evidence source mutations: Postgres is required")
	}
	if projectionRows == nil {
		return nil, errors.New("compose Evidence source mutations: Projections is required")
	}
	if appender == nil {
		return nil, errors.New("compose Evidence source mutations: Revisions is required")
	}
	if intents == nil {
		return nil, errors.New("compose Evidence source mutations: Collaboration is required")
	}
	if isNilMutationCapability(incidentState) {
		return nil, errors.New("compose Evidence source mutations: incident state is required")
	}
	if isNilMutationCapability(recordEnvelopes) {
		return nil, errors.New("compose Evidence source mutations: record envelopes are required")
	}
	service := &sourceMutationService{}
	service.source = evidenceSourceKernel{
		records:     recordEnvelopes,
		rows:        service,
		projections: projectionRows,
	}
	service.mutations = evidenceSourceMutationKernel{
		incidents:     incidentState,
		source:        service.source,
		revisions:     newRevisionAppendAdapter(appender),
		collaboration: intents,
	}
	return service, nil
}
