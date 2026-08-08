package extensionassembly

import (
	"fmt"

	"github.com/JochiRaider/cartulary/internal/platform/jobs"
)

var canonicalWorkerByJobKind = map[string]string{
	"import.discovery_v1":                       "import.discovery_worker_v1",
	"import.apply_v1":                           "import.apply_worker_v1",
	"incident_portability.export_v1":            "incident_portability.bundle_worker_v1",
	"incident_portability.import_v1":            "incident_portability.bundle_worker_v1",
	"reference_pack.import_v1":                  "reference_pack.lifecycle_worker_v1",
	"reference_pack.reverify_v1":                "reference_pack.lifecycle_worker_v1",
	"reference_pack.refresh_v1":                 "reference_pack.lifecycle_worker_v1",
	"snapshot_reporting.snapshot_create_v1":     "snapshot_reporting.job_worker_v1",
	"snapshot_reporting.release_create_v1":      "snapshot_reporting.job_worker_v1",
	"snapshot_reporting.composition_preview_v1": "snapshot_reporting.job_worker_v1",
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
		resourceRefs := make([]jobs.ExtensionResourceRefContract, 0, len(contract.ResourceRefContracts))
		for _, resourceRef := range contract.ResourceRefContracts {
			resourceRefs = append(resourceRefs, jobs.ExtensionResourceRefContract{
				Kind:    resourceRef.ResourceRefKind,
				MaxRefs: resourceRef.MaxRefs,
			})
		}
		result = append(result, jobs.Definition{
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
		})
	}
	return result, nil
}
