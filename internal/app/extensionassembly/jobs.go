package extensionassembly

import (
	"fmt"

	"github.com/JochiRaider/cartulary/internal/modules/extensions"
	"github.com/JochiRaider/cartulary/internal/platform/jobs"
)

var canonicalWorkerByJobKind = map[string]string{
	"import.discovery_v1":                             "import.discovery_worker_v1",
	"import.apply_v1":                                 "import.apply_worker_v1",
	"incident_portability.export_v1":                  "incident_portability.bundle_worker_v1",
	"incident_portability.import_v1":                  "incident_portability.bundle_worker_v1",
	"reference_pack.import_v1":                        "reference_pack.lifecycle_worker_v1",
	"reference_pack.reverify_v1":                      "reference_pack.lifecycle_worker_v1",
	"reference_pack.refresh_v1":                       "reference_pack.lifecycle_worker_v1",
	"snapshot_reporting.snapshot_create_v1":           "snapshot_reporting.job_worker_v1",
	"snapshot_reporting.release_create_v1":            "snapshot_reporting.job_worker_v1",
	"snapshot_reporting.composition_preview_v1":       "snapshot_reporting.job_worker_v1",
	"network_flow_activity.graph_view_materialize_v1": "network_flow_activity.graph_view_worker_v1",
}

func JobDefinitions(catalog PublicationCatalog) ([]jobs.Definition, error) {
	jobKinds := catalog.JobKinds()
	result := make([]jobs.Definition, 0, len(jobKinds))
	for _, jobKind := range jobKinds {
		contract, present := catalog.Job(jobKind)
		if !present {
			return nil, fmt.Errorf("extension job catalog lost %q", jobKind)
		}
		workerKind, present := canonicalWorkerByJobKind[jobKind]
		if !present {
			return nil, fmt.Errorf("extension job %q has no canonical worker", jobKind)
		}
		worker, present := catalog.Worker(workerKind)
		if !present || worker.ProfileID != contract.ProfileID {
			return nil, fmt.Errorf("extension job %q has no exact claimed worker", jobKind)
		}
		result = append(result, jobDefinition(contract, workerKind))
	}
	return result, nil
}

// RecognizedJobDefinitions projects every packaged current contract into the
// one durable Jobs catalog. Claim state is applied later by RuntimeSelection.
func RecognizedJobDefinitions(contracts []extensions.JobKindContract) ([]jobs.Definition, error) {
	result := make([]jobs.Definition, 0, len(contracts))
	for _, contract := range contracts {
		workerKind, present := canonicalWorkerByJobKind[contract.JobKind]
		if !present {
			return nil, fmt.Errorf("extension job %q has no canonical worker", contract.JobKind)
		}
		result = append(result, jobDefinition(contract, workerKind))
	}
	return result, nil
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
