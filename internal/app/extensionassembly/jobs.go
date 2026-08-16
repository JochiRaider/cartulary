package extensionassembly

import (
	"fmt"

	"github.com/JochiRaider/cartulary/internal/modules/extensions"
	"github.com/JochiRaider/cartulary/internal/platform/jobs"
)

func JobDefinitions(catalog PublicationCatalog) ([]jobs.Definition, error) {
	workers := make([]extensions.WorkerPublication, 0, len(catalog.WorkerKinds()))
	for _, workerKind := range catalog.WorkerKinds() {
		worker, present := catalog.Worker(workerKind)
		if !present {
			return nil, fmt.Errorf("extension worker catalog lost %q", workerKind)
		}
		workers = append(workers, worker)
	}
	contracts := make([]extensions.JobKindContract, 0, len(catalog.JobKinds()))
	for _, jobKind := range catalog.JobKinds() {
		contract, present := catalog.Job(jobKind)
		if !present {
			return nil, fmt.Errorf("extension job catalog lost %q", jobKind)
		}
		contracts = append(contracts, contract)
	}
	return jobDefinitions(contracts, workers)
}

// RecognizedJobDefinitions projects every packaged current contract into the
// one durable Jobs catalog. Claim state is applied later by RuntimeSelection.
func RecognizedJobDefinitions(contracts []extensions.JobKindContract, workers []extensions.WorkerPublication) ([]jobs.Definition, error) {
	return jobDefinitions(contracts, workers)
}

func jobDefinitions(contracts []extensions.JobKindContract, workers []extensions.WorkerPublication) ([]jobs.Definition, error) {
	contractsByKind := make(map[string]extensions.JobKindContract, len(contracts))
	for _, contract := range contracts {
		if _, duplicate := contractsByKind[contract.JobKind]; duplicate {
			return nil, fmt.Errorf("duplicate extension job contract %q", contract.JobKind)
		}
		contractsByKind[contract.JobKind] = contract
	}
	result := make([]jobs.Definition, 0, len(contracts))
	assigned := make(map[string]struct{}, len(contracts))
	for _, worker := range workers {
		for _, jobKind := range worker.JobKinds {
			contract, present := contractsByKind[jobKind]
			if !present || contract.ProfileID != worker.ProfileID {
				return nil, fmt.Errorf("extension job %q has no exact generated worker", jobKind)
			}
			if _, duplicate := assigned[jobKind]; duplicate {
				return nil, fmt.Errorf("extension job %q has duplicate generated workers", jobKind)
			}
			assigned[jobKind] = struct{}{}
			result = append(result, jobDefinition(contract, worker.WorkerKind))
		}
	}
	if len(assigned) != len(contractsByKind) {
		return nil, fmt.Errorf("generated worker runtime contracts omit extension jobs")
	}
	return result, nil
}

func WorkerRuntimeContracts(catalog PublicationCatalog) ([]jobs.WorkerRuntimeContract, error) {
	workers := make([]extensions.WorkerPublication, 0, len(catalog.WorkerKinds()))
	for _, workerKind := range catalog.WorkerKinds() {
		worker, present := catalog.Worker(workerKind)
		if !present {
			return nil, fmt.Errorf("extension worker catalog lost %q", workerKind)
		}
		workers = append(workers, worker)
	}
	return workerRuntimeContracts(workers), nil
}

func workerRuntimeContracts(workers []extensions.WorkerPublication) []jobs.WorkerRuntimeContract {
	result := make([]jobs.WorkerRuntimeContract, 0, len(workers))
	for _, worker := range workers {
		result = append(result, jobs.WorkerRuntimeContract{
			ProfileID: worker.ProfileID, WorkerKind: worker.WorkerKind,
			JobKinds:                    append([]string(nil), worker.JobKinds...),
			MaxActiveAttemptsPerProcess: worker.MaxActiveAttemptsPerProcess,
		})
	}
	return result
}

func JobKinds(definitions []jobs.Definition) []string {
	result := make([]string, 0, len(definitions))
	for _, definition := range definitions {
		result = append(result, definition.JobKind)
	}
	return result
}

func jobDefinition(contract extensions.JobKindContract, workerKind string) jobs.Definition {
	resourceRefs := make([]jobs.ExtensionResourceRefContract, 0, len(contract.ResourceRefContracts))
	for _, resourceRef := range contract.ResourceRefContracts {
		resourceRefs = append(resourceRefs, jobs.ExtensionResourceRefContract{
			Kind:    resourceRef.ResourceRefKind,
			MaxRefs: resourceRef.MaxRefs,
		})
	}
	return jobs.Definition{
		JobKind:        contract.JobKind,
		ProgressUnitID: contract.ProgressUnitID,
		HandlerName:    workerKind,
		Extension: &jobs.ExtensionPolicy{
			OwnerProfileID: contract.ProfileID,
			OperationKind:  contract.OperationKind,
			ContractSHA256: contract.SHA256(),
			ProofRequired:  contract.ProofPolicy == "required_on_terminal_success",
			MaxProofBytes:  contract.MaxProofBytes,
			ResourceRefs:   resourceRefs,
		},
	}
}
