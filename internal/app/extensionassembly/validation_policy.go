package extensionassembly

import (
	"github.com/JochiRaider/cartulary/internal/modules/crossownertransaction"
	"github.com/JochiRaider/cartulary/internal/modules/extensions"
	"github.com/JochiRaider/cartulary/internal/modules/incidentbundles"
	"github.com/JochiRaider/cartulary/internal/modules/networkflow"
	"github.com/JochiRaider/cartulary/internal/platform/canonicaljson"
)

type CrossOwnerValidationPolicy struct {
	Coordinator *extensions.Coordinator
	NetworkFlow *networkflow.Module
}

func (p CrossOwnerValidationPolicy) AcceptDescriptor(d crossownertransaction.Descriptor) bool {
	if p.Coordinator == nil {
		return false
	}
	if d.ParticipantID == incidentbundles.ImportTransactionParticipantID {
		return crossownertransaction.SameDescriptor(d, incidentbundles.ImportTransactionDescriptor())
	}
	implementation, ok := networkflow.TransactionContract(d.ParticipantID)
	if !ok || !crossownertransaction.SameDescriptor(d, implementation) {
		return false
	}
	for _, contract := range CrossOwnerDescriptors(p.Coordinator.ParticipantContracts()) {
		if contract.ParticipantID == d.ParticipantID {
			return crossownertransaction.SameDescriptor(d, contract)
		}
	}
	return false
}

func (p CrossOwnerValidationPolicy) AcceptInput(d crossownertransaction.Descriptor, value any) bool {
	if !p.AcceptDescriptor(d) {
		return false
	}
	if d.ParticipantID == incidentbundles.ImportTransactionParticipantID {
		return incidentbundles.ValidateImportTransactionInput(value)
	}
	if p.NetworkFlow == nil {
		return false
	}
	return p.NetworkFlow.ValidateTransactionInput(d.InputSchemaID, value)
}

func (p CrossOwnerValidationPolicy) FailureFinding(id string, owner crossownertransaction.Descriptor, timeoutSeconds int64, timedOut bool) (crossownertransaction.Finding, bool) {
	if p.Coordinator == nil {
		return crossownertransaction.Finding{}, false
	}
	finding, ok := p.Coordinator.TransactionValidationFailure(id, owner.OwnerProfileID, owner.ValidationAlgorithmID, timeoutSeconds, timedOut)
	if !ok {
		return crossownertransaction.Finding{}, false
	}
	details, err := canonicaljson.MarshalExtension(finding.Details)
	if err != nil {
		return crossownertransaction.Finding{}, false
	}
	return crossownertransaction.Finding{Path: finding.Path, ReasonCode: finding.ReasonCode, Message: finding.Message, Details: append([]byte(nil), details[:len(details)-1]...)}, true
}

func (p CrossOwnerValidationPolicy) AcceptFinding(owner crossownertransaction.Descriptor, finding crossownertransaction.Finding) bool {
	details, err := canonicaljson.DecodeStrict(finding.Details)
	if err != nil {
		return false
	}
	object, ok := details.(map[string]any)
	return ok && p.Coordinator.ValidateTransactionFinding(owner.OwnerProfileID, owner.ValidationAlgorithmID, extensions.StartupFinding{Path: finding.Path, ReasonCode: finding.ReasonCode, Message: finding.Message, Details: object})
}
