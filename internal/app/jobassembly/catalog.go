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
)

type baseContract struct {
	SchemaID string `json:"schema_id"`
	Jobs     []struct {
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
	artifact, ok := contractjobs.Index["contracts/jobs/base.v1.json"]
	if !ok {
		return nil, nil, errors.New("base Jobs contract is missing")
	}
	return decodeBaseDefinitions(artifact.JSON)
}

func decodeBaseDefinitions(raw string) ([]jobs.Definition, []jobs.WorkerRuntimeContract, error) {
	if _, err := canonicaljson.DecodeStrict([]byte(raw)); err != nil {
		return nil, nil, err
	}
	decoder := json.NewDecoder(strings.NewReader(raw))
	decoder.DisallowUnknownFields()
	var contract baseContract
	if err := decoder.Decode(&contract); err != nil {
		return nil, nil, err
	}
	var trailing any
	if decoder.Decode(&trailing) != io.EOF || contract.SchemaID != "cartulary.base_job_catalog.v1" {
		return nil, nil, errors.New("base Jobs contract is invalid")
	}
	definitions := make([]jobs.Definition, 0, len(contract.Jobs))
	for _, item := range contract.Jobs {
		definitions = append(definitions, jobs.Definition{JobKind: item.JobKind, ProgressUnitID: item.ProgressUnitID, HandlerName: item.HandlerName})
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
