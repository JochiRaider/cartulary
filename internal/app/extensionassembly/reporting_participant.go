package extensionassembly

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/JochiRaider/cartulary/internal/modules/extensions"
	"github.com/JochiRaider/cartulary/internal/modules/reporting"
)

type admittedRenderExportInvoker struct {
	participant reporting.RenderExportParticipant
	timeout     time.Duration
}

func NewAdmittedRenderExportInvoker(
	catalog PublicationCatalog,
	participant reporting.RenderExportParticipant,
	timeout time.Duration,
) (reporting.RenderExportInvoker, error) {
	contribution, present := catalog.Contribution(reporting.RenderExportContributionID)
	if !present || contribution.ProfileID != reporting.ProfileID ||
		contribution.Kind != "snapshot_reporting_participant" ||
		contribution.ParticipantID != reporting.RenderExportParticipantID {
		return nil, errors.New("snapshot/reporting render participant contribution is not admitted")
	}
	contract, present := catalog.Participant(reporting.RenderExportParticipantID)
	if !present {
		return nil, errors.New("snapshot/reporting render participant contract is not admitted")
	}
	if err := validateRenderExportContract(contract); err != nil {
		return nil, err
	}
	if participant == nil {
		return nil, errors.New("snapshot/reporting render participant implementation is missing")
	}
	if timeout <= 0 {
		return nil, errors.New("snapshot/reporting render participant timeout must be positive")
	}
	return admittedRenderExportInvoker{participant: participant, timeout: timeout}, nil
}

func (invoker admittedRenderExportInvoker) Invoke(
	ctx context.Context,
	invocation reporting.RenderExportInvocation,
) (reporting.RenderExportResult, error) {
	if invoker.participant == nil || invoker.timeout <= 0 {
		return reporting.RenderExportResult{}, errors.New("snapshot/reporting render participant is unavailable")
	}
	callCtx, cancel := context.WithTimeout(ctx, invoker.timeout)
	defer cancel()
	return invoker.participant.Emit(callCtx, invocation)
}

func validateRenderExportContract(contract extensions.ParticipantContract) error {
	if contract.ParticipantID != reporting.RenderExportParticipantID ||
		contract.OwnerProfileID != reporting.ProfileID ||
		contract.ParticipantKind != reporting.RenderExportParticipantKind ||
		contract.InputSchemaID != reporting.RenderExportContextSchemaID ||
		len(contract.Operations) != 1 {
		return fmt.Errorf("snapshot/reporting render participant contract identity is invalid")
	}
	operation := contract.Operations[0]
	if operation.OperationKind != reporting.RenderExportOperationKind ||
		operation.ResultSchemaID != reporting.RenderExportResultSchemaID ||
		operation.AlgorithmID != reporting.RenderExportAlgorithmID ||
		operation.OutputSchemaID != reporting.ExportModelSchemaID ||
		operation.OrderingAlgorithmID != reporting.RenderExportOrderingAlgorithmID ||
		operation.MaxInputBytes != reporting.RenderExportMaxInputBytes ||
		operation.MaxOutputBytes != reporting.RenderExportMaxOutputBytes ||
		operation.MaxItems != reporting.RenderExportMaxItems {
		return fmt.Errorf("snapshot/reporting render participant operation contract is invalid")
	}
	return nil
}
