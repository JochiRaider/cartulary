package extensionassembly

import (
	"errors"
	"slices"

	"github.com/JochiRaider/cartulary/internal/modules/extensions"
	"github.com/JochiRaider/cartulary/internal/modules/graphprojection"
	"github.com/JochiRaider/cartulary/internal/modules/networkflow"
	"github.com/JochiRaider/cartulary/internal/modules/reporting"
	"github.com/JochiRaider/cartulary/internal/modules/reporting/graphsourcecontract"
)

// NewAdmittedNetworkFlowReportingSource binds the admitted publication contract
// to the concrete owner's transactional reader without invoking it.
func NewAdmittedNetworkFlowReportingSource(catalog PublicationCatalog, module *networkflow.Module) (reporting.GraphSourceProvider, error) {
	contribution, present := catalog.Contribution(networkflow.ReportingGraphSourceContributionID)
	if !present || contribution.ProfileID != networkflow.ProfileID || contribution.Kind != "snapshot_reporting_participant" || contribution.ParticipantID != networkflow.ReportingGraphSourceParticipantID {
		return nil, errors.New("network flow reporting source contribution is not admitted")
	}
	contract, present := catalog.Participant(networkflow.ReportingGraphSourceParticipantID)
	source := module.ReportingGraphSource()
	if !present || source == nil {
		return nil, errors.New("network flow reporting source is unavailable")
	}
	if err := validateNetworkFlowReportingSourceContract(contract); err != nil {
		return nil, err
	}
	return source, nil
}

func validateNetworkFlowReportingSourceContract(contract extensions.ParticipantContract) error {
	if contract.ContractKind != "cartulary.extension_reporting_source_contract.v1" || contract.OwnerProfileID != networkflow.ProfileID || contract.ParticipantID != networkflow.ReportingGraphSourceParticipantID || contract.ParticipantKind != "snapshot_reporting" || len(contract.Operations) != 0 || contract.InputSchemaID != "" || !slices.Equal(contract.OwnedStateFamilyIDs, []string{networkflow.ReportingGraphSourceStateFamilyID}) || !slices.Equal(contract.AlgorithmIDs, []string{networkflow.ReportingGraphSourceOrderingAlgorithmID, networkflow.ReportingGraphSourceReadAlgorithmID}) {
		return errors.New("network flow reporting source contract is not implemented")
	}
	definition := contract.ReportingSource
	if definition.ProjectionSchemaID != graphprojection.ProjectionSchemaIDV2 || definition.LabelSchemaID != graphsourcecontract.SchemaID || definition.ReadAlgorithmID != networkflow.ReportingGraphSourceReadAlgorithmID || definition.OrderingAlgorithmID != networkflow.ReportingGraphSourceOrderingAlgorithmID || definition.MaxVertices != graphprojection.MaximumResultVerticesV2 || definition.MaxEdges != graphprojection.MaximumResultEdgesV2 || definition.MaxLabelBytes != networkflow.ReportingGraphSourceMaxLabelBytes {
		return errors.New("network flow reporting source definition is not implemented")
	}
	return nil
}
