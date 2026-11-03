// Package jobassembly composes the one durable Jobs catalog from explicit base
// and extension owner contracts. Optional claims affect execution, not recognition.
package jobassembly

import (
	"encoding/json"
	"errors"
	"io"
	"strings"

	"github.com/JochiRaider/cartulary/internal/app/extensionassembly"
	"github.com/JochiRaider/cartulary/internal/gen/contractjobs"
	"github.com/JochiRaider/cartulary/internal/modules/extensions"
	"github.com/JochiRaider/cartulary/internal/platform/canonicaljson"
	"github.com/JochiRaider/cartulary/internal/platform/jobs"
	"github.com/JochiRaider/cartulary/internal/platform/jobs/jobcontract"
)

type baseContract struct {
	SchemaID string `json:"schema_id"`
	Jobs     []struct {
		jobcontract.TerminalContract
		JobKind        string `json:"job_kind"`
		ProgressUnitID string `json:"progress_unit_id"`
		HandlerName    string `json:"handler_name"`
	} `json:"jobs"`
	Workers []struct {
		ProfileID  string   `json:"profile_id"`
		WorkerKind string   `json:"worker_kind"`
		JobKinds   []string `json:"job_kinds"`
		Maximum    int      `json:"max_active_attempts_per_process"`
	} `json:"workers"`
}

func baseDefinitions() ([]jobs.Definition, []jobs.WorkerRuntimeContract, error) {
	artifact, ok := contractjobs.Index["contracts/jobs/base.v2.json"]
	if !ok {
		return nil, nil, errors.New("base Jobs contract is missing")
	}
	return decodeBaseDefinitions(artifact.JSON)
}

func decodeBaseDefinitions(raw string) ([]jobs.Definition, []jobs.WorkerRuntimeContract, error) {
	value, err := canonicaljson.DecodeStrict([]byte(raw))
	if err != nil {
		return nil, nil, err
	}
	if !validBaseContractShape(value) {
		return nil, nil, errors.New("base Jobs contract has invalid members")
	}
	decoder := json.NewDecoder(strings.NewReader(raw))
	decoder.DisallowUnknownFields()
	var contract baseContract
	if err := decoder.Decode(&contract); err != nil {
		return nil, nil, err
	}
	var trailing any
	if decoder.Decode(&trailing) != io.EOF || contract.SchemaID != "cartulary.base_job_catalog.v2" {
		return nil, nil, errors.New("base Jobs contract is invalid")
	}
	definitions := make([]jobs.Definition, 0, len(contract.Jobs))
	for _, item := range contract.Jobs {
		definitions = append(definitions, jobs.Definition{JobKind: item.JobKind, ProgressUnitID: item.ProgressUnitID, HandlerName: item.HandlerName, Terminal: item.TerminalContract})
	}
	workers := make([]jobs.WorkerRuntimeContract, 0, len(contract.Workers))
	for _, item := range contract.Workers {
		if item.ProfileID != "base" {
			return nil, nil, errors.New("base Jobs contract has an extension worker")
		}
		workers = append(workers, jobs.WorkerRuntimeContract{ProfileID: item.ProfileID, WorkerKind: item.WorkerKind, JobKinds: item.JobKinds, MaxActiveAttemptsPerProcess: item.Maximum})
	}
	catalog, err := jobs.NewCatalog(definitions)
	if err != nil {
		return nil, nil, err
	}
	if _, err = jobs.FullRuntimeSelection(catalog, workers); err != nil {
		return nil, nil, err
	}
	return definitions, workers, nil
}

func validBaseContractShape(value any) bool {
	root, ok := exactBaseContractObject(value, "schema_id", "jobs", "workers")
	if !ok {
		return false
	}
	definitions, ok := root["jobs"].([]any)
	if !ok || len(definitions) == 0 {
		return false
	}
	for _, value := range definitions {
		row, ok := exactBaseContractObject(value, "job_kind", "progress_unit_id", "handler_name", "terminal_result_schema_id", "terminal_success_codes", "resource_ref_contracts")
		if !ok {
			return false
		}
		refs, ok := row["resource_ref_contracts"].([]any)
		if !ok {
			return false
		}
		for _, ref := range refs {
			if _, ok := exactBaseContractObject(ref, "resource_ref_kind", "resource_id_schema_id", "min_refs", "max_refs"); !ok {
				return false
			}
		}
	}
	workers, ok := root["workers"].([]any)
	if !ok {
		return false
	}
	for _, worker := range workers {
		if _, ok := exactBaseContractObject(worker, "profile_id", "worker_kind", "job_kinds", "max_active_attempts_per_process"); !ok {
			return false
		}
	}
	return true
}

func exactBaseContractObject(value any, keys ...string) (map[string]any, bool) {
	object, ok := value.(map[string]any)
	if !ok || len(object) != len(keys) {
		return nil, false
	}
	for _, key := range keys {
		if _, present := object[key]; !present {
			return nil, false
		}
	}
	return object, true
}

func RecognizedDefinitions(contracts []extensions.JobKindContract, workers []extensions.WorkerPublication) ([]jobs.Definition, error) {
	base, _, err := baseDefinitions()
	if err != nil {
		return nil, err
	}
	extensions, err := extensionassembly.RecognizedJobDefinitions(contracts, workers)
	if err != nil {
		return nil, err
	}
	return append(base, extensions...), nil
}

func SelectedDefinitions(catalog extensionassembly.PublicationCatalog) ([]jobs.Definition, error) {
	base, _, err := baseDefinitions()
	if err != nil {
		return nil, err
	}
	extensions, err := extensionassembly.JobDefinitions(catalog)
	if err != nil {
		return nil, err
	}
	return append(base, extensions...), nil
}

func WorkerContracts(catalog extensionassembly.PublicationCatalog) ([]jobs.WorkerRuntimeContract, error) {
	_, base, err := baseDefinitions()
	if err != nil {
		return nil, err
	}
	extensions, err := extensionassembly.WorkerRuntimeContracts(catalog)
	if err != nil {
		return nil, err
	}
	return append(base, extensions...), nil
}
