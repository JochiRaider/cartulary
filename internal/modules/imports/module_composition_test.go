package imports

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"

	"github.com/JochiRaider/cartulary/internal/modules/imports/ownerfacade"
	"github.com/JochiRaider/cartulary/internal/modules/revisions"
	"github.com/JochiRaider/cartulary/internal/platform/authn"
	"github.com/JochiRaider/cartulary/internal/platform/httpapi"
	"github.com/JochiRaider/cartulary/internal/platform/jobs"
)

type moduleTestDB struct{}

func (*moduleTestDB) Exec(context.Context, string, ...any) (pgconn.CommandTag, error) {
	return pgconn.CommandTag{}, nil
}

func (*moduleTestDB) Query(context.Context, string, ...any) (pgx.Rows, error) {
	return nil, errors.New("unexpected query")
}

func (*moduleTestDB) QueryRow(context.Context, string, ...any) pgx.Row { return nil }

func (*moduleTestDB) BeginTx(context.Context, pgx.TxOptions) (pgx.Tx, error) {
	return nil, errors.New("unexpected transaction")
}

type moduleTestJobs struct{}

func (*moduleTestJobs) CreateQueuedTx(context.Context, pgx.Tx, jobs.EnqueueParams, time.Time) (jobs.Resource, error) {
	return jobs.Resource{}, nil
}

func (*moduleTestJobs) ValidateExecutionTx(context.Context, pgx.Tx, jobs.Execution) error {
	return nil
}

func (*moduleTestJobs) ValidateCancellationExecutionTx(context.Context, pgx.Tx, jobs.Execution) error {
	return nil
}

func (*moduleTestJobs) HandlerPayload(context.Context, jobs.Execution) (json.RawMessage, error) {
	return nil, nil
}

func (*moduleTestJobs) ObserveExecution(context.Context, jobs.Execution) (jobs.Resource, error) {
	return jobs.Resource{}, nil
}

func (*moduleTestJobs) UpdateProgress(context.Context, jobs.Execution, jobs.Progress, *string) (jobs.Resource, error) {
	return jobs.Resource{}, nil
}

func (*moduleTestJobs) CompleteCanceled(context.Context, jobs.Execution, jobs.CancellationCompletion) (jobs.Resource, error) {
	return jobs.Resource{}, nil
}

type moduleTestRunner struct {
	registered []string
	failOn     string
}

func (r *moduleTestRunner) RegisterHandler(kind string, _ jobs.HandlerFunc) error {
	if kind == r.failOn {
		return errors.New("registration rejected")
	}
	r.registered = append(r.registered, kind)
	return nil
}

func (*moduleTestRunner) Notify(uuid.UUID) {}

type moduleTestFinalizer struct{}

func (*moduleTestFinalizer) FinalizeImportJobSuccess(context.Context, JobSuccessFinalization) (jobs.Resource, error) {
	return jobs.Resource{}, nil
}

func (*moduleTestFinalizer) FinalizeImportJobFailure(context.Context, JobFailureFinalization) (jobs.Resource, error) {
	return jobs.Resource{}, nil
}

func (*moduleTestFinalizer) FinalizeImportJobCancellation(context.Context, JobCancellationFinalization) (jobs.Resource, error) {
	return jobs.Resource{}, nil
}

type moduleTestFacade struct {
	binding ExtensionImportFacadeBinding
}

func (f *moduleTestFacade) Binding() ExtensionImportFacadeBinding { return f.binding }

func (*moduleTestFacade) PrepareImportUnitMapping(context.Context, ExtensionImportMappingRequest) (ExtensionImportMappingResult, error) {
	return ExtensionImportMappingResult{}, nil
}

func (*moduleTestFacade) ValidateImportUnitMappingResult(ExtensionImportMappingResult) error {
	return nil
}

func (*moduleTestFacade) ApplyImportUnitTx(context.Context, pgx.Tx, ExtensionImportApplyRequest) (ExtensionImportApplyResult, error) {
	return ExtensionImportApplyResult{}, nil
}

func (*moduleTestFacade) TranslateImportUnitError(error) (ExtensionImportErrorTranslation, bool) {
	return ExtensionImportErrorTranslation{}, false
}

func (*moduleTestFacade) ValidateImportUnitError(ExtensionImportOwnerError) error { return nil }

func validModuleTestBinding() ExtensionImportFacadeBinding {
	return ExtensionImportFacadeBinding{
		SchemaID:               "cartulary.imports.analytical_facade_binding.v1",
		TargetKind:             "network_flow_table",
		ExtensionProfileID:     "network_flow_activity",
		OwnerContractRef:       "network_flow_activity@5",
		FacadeID:               "network_flow_import_facade_v1",
		ContractMajor:          5,
		MappingSchemaID:        "cartulary.network_flow.approved_mapping.v1",
		PreviewRequestSchemaID: "cartulary.network_flow.import_preview_request.v1",
		PreviewResultSchemaID:  "cartulary.network_flow.import_preview_result.v1",
		ApplyRequestSchemaID:   "cartulary.network_flow.import_apply_request.v1",
		ApplyResultSchemaID:    "cartulary.network_flow.import_unit_result.v1",
		ErrorSchemaID:          "cartulary.network_flow.import_owner_error.v1",
		ErrorTranslationID:     "network_flow_activity.import_error_translation.v1",
		CommitProtocolID:       "cartulary.imports.unit_commit.v1",
	}
}

func validModuleTestDependencies() ModuleDependencies {
	jobService := &moduleTestJobs{}
	return ModuleDependencies{
		Postgres:            &moduleTestDB{},
		JobTransactions:     jobService,
		JobOperations:       jobService,
		JobRunner:           &moduleTestRunner{},
		Limits:              Limits{MaxCSVSourceBytes: 1, MaxXLSXSourceBytes: 1, MaxRows: 1, MaxColumns: 1, MaxCells: 1},
		ArchiveLimits:       ArchiveLimits{DefaultMaxExtractedBytes: 1, MaxCompressionRatio: 1, MaxMembers: 1},
		OwnerCreateRegistry: &ownerfacade.ImportOwnerCreateRegistry{},
		RevisionAppender:    &revisions.Appender{},
		ExtensionProfileAdmission: func(profileID string) bool {
			return profileID == networkFlowExtensionProfileID
		},
		JobSuccessFinalizer:    &moduleTestFinalizer{},
		ExtensionImportFacades: []ExtensionImportFacade{&moduleTestFacade{binding: validModuleTestBinding()}},
	}
}

func TestImportsModuleRejectsEveryMissingDependency(t *testing.T) {
	tests := []struct {
		name   string
		remove func(*ModuleDependencies)
	}{
		{name: "Postgres", remove: func(deps *ModuleDependencies) { deps.Postgres = nil }},
		{name: "Jobs transactions", remove: func(deps *ModuleDependencies) { deps.JobTransactions = nil }},
		{name: "Jobs operations", remove: func(deps *ModuleDependencies) { deps.JobOperations = nil }},
		{name: "Jobs runner", remove: func(deps *ModuleDependencies) { deps.JobRunner = nil }},
		{name: "owner registry", remove: func(deps *ModuleDependencies) { deps.OwnerCreateRegistry = nil }},
		{name: "Revisions", remove: func(deps *ModuleDependencies) { deps.RevisionAppender = nil }},
		{name: "profile admission", remove: func(deps *ModuleDependencies) { deps.ExtensionProfileAdmission = nil }},
		{name: "job finalizer", remove: func(deps *ModuleDependencies) { deps.JobSuccessFinalizer = nil }},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			dependencies := validModuleTestDependencies()
			test.remove(&dependencies)
			if _, err := NewModule(dependencies); err == nil {
				t.Fatal("missing dependency was accepted")
			}
		})
	}
}

func TestImportsModuleRejectsEveryNonPositiveLimit(t *testing.T) {
	tests := []struct {
		name string
		set  func(*ModuleDependencies, int64)
	}{
		{name: "MaxCSVSourceBytes", set: func(deps *ModuleDependencies, value int64) { deps.Limits.MaxCSVSourceBytes = value }},
		{name: "MaxXLSXSourceBytes", set: func(deps *ModuleDependencies, value int64) { deps.Limits.MaxXLSXSourceBytes = value }},
		{name: "MaxRows", set: func(deps *ModuleDependencies, value int64) { deps.Limits.MaxRows = value }},
		{name: "MaxColumns", set: func(deps *ModuleDependencies, value int64) { deps.Limits.MaxColumns = value }},
		{name: "MaxCells", set: func(deps *ModuleDependencies, value int64) { deps.Limits.MaxCells = value }},
		{name: "DefaultMaxExtractedBytes", set: func(deps *ModuleDependencies, value int64) { deps.ArchiveLimits.DefaultMaxExtractedBytes = value }},
		{name: "MaxCompressionRatio", set: func(deps *ModuleDependencies, value int64) { deps.ArchiveLimits.MaxCompressionRatio = value }},
		{name: "MaxMembers", set: func(deps *ModuleDependencies, value int64) { deps.ArchiveLimits.MaxMembers = value }},
	}
	for _, test := range tests {
		for _, value := range []int64{0, -1} {
			t.Run(test.name, func(t *testing.T) {
				dependencies := validModuleTestDependencies()
				test.set(&dependencies, value)
				if _, err := NewModule(dependencies); err == nil {
					t.Fatalf("limit %s=%d was accepted", test.name, value)
				}
			})
		}
	}
}

func TestImportsModuleValidatesEveryAnalyticalBindingMember(t *testing.T) {
	tests := []struct {
		name   string
		mutate func(*ExtensionImportFacadeBinding)
	}{
		{name: "schema_id", mutate: func(binding *ExtensionImportFacadeBinding) { binding.SchemaID += ".wrong" }},
		{name: "target_kind", mutate: func(binding *ExtensionImportFacadeBinding) { binding.TargetKind += ".wrong" }},
		{name: "extension_profile_id", mutate: func(binding *ExtensionImportFacadeBinding) { binding.ExtensionProfileID += ".wrong" }},
		{name: "owner_contract_ref", mutate: func(binding *ExtensionImportFacadeBinding) { binding.OwnerContractRef += ".wrong" }},
		{name: "facade_id", mutate: func(binding *ExtensionImportFacadeBinding) { binding.FacadeID += ".wrong" }},
		{name: "contract_major", mutate: func(binding *ExtensionImportFacadeBinding) { binding.ContractMajor++ }},
		{name: "mapping_schema_id", mutate: func(binding *ExtensionImportFacadeBinding) { binding.MappingSchemaID += ".wrong" }},
		{name: "preview_request_schema_id", mutate: func(binding *ExtensionImportFacadeBinding) { binding.PreviewRequestSchemaID += ".wrong" }},
		{name: "preview_result_schema_id", mutate: func(binding *ExtensionImportFacadeBinding) { binding.PreviewResultSchemaID += ".wrong" }},
		{name: "apply_request_schema_id", mutate: func(binding *ExtensionImportFacadeBinding) { binding.ApplyRequestSchemaID += ".wrong" }},
		{name: "apply_result_schema_id", mutate: func(binding *ExtensionImportFacadeBinding) { binding.ApplyResultSchemaID += ".wrong" }},
		{name: "error_schema_id", mutate: func(binding *ExtensionImportFacadeBinding) { binding.ErrorSchemaID += ".wrong" }},
		{name: "error_translation_id", mutate: func(binding *ExtensionImportFacadeBinding) { binding.ErrorTranslationID += ".wrong" }},
		{name: "commit_protocol_id", mutate: func(binding *ExtensionImportFacadeBinding) { binding.CommitProtocolID += ".wrong" }},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			dependencies := validModuleTestDependencies()
			binding := validModuleTestBinding()
			test.mutate(&binding)
			dependencies.ExtensionImportFacades = []ExtensionImportFacade{&moduleTestFacade{binding: binding}}
			if _, err := NewModule(dependencies); err == nil {
				t.Fatal("mismatched binding was accepted")
			}
		})
	}
}

func TestImportsModuleRejectsInvalidAnalyticalContributions(t *testing.T) {
	t.Run("nil", func(t *testing.T) {
		dependencies := validModuleTestDependencies()
		dependencies.ExtensionImportFacades = []ExtensionImportFacade{nil}
		if _, err := NewModule(dependencies); err == nil {
			t.Fatal("nil facade was accepted")
		}
	})

	t.Run("typed nil", func(t *testing.T) {
		dependencies := validModuleTestDependencies()
		var facade *moduleTestFacade
		dependencies.ExtensionImportFacades = []ExtensionImportFacade{facade}
		if _, err := NewModule(dependencies); err == nil {
			t.Fatal("typed-nil facade was accepted")
		}
	})

	t.Run("unknown selector", func(t *testing.T) {
		dependencies := validModuleTestDependencies()
		binding := validModuleTestBinding()
		binding.TargetKind = "unknown_target"
		dependencies.ExtensionImportFacades = []ExtensionImportFacade{&moduleTestFacade{binding: binding}}
		if _, err := NewModule(dependencies); err == nil {
			t.Fatal("unknown selector was accepted")
		}
	})

	t.Run("duplicate selector", func(t *testing.T) {
		dependencies := validModuleTestDependencies()
		facade := &moduleTestFacade{binding: validModuleTestBinding()}
		dependencies.ExtensionImportFacades = []ExtensionImportFacade{facade, facade}
		if _, err := NewModule(dependencies); err == nil {
			t.Fatal("duplicate selector was accepted")
		}
	})

	t.Run("duplicate facade id", func(t *testing.T) {
		binding := validModuleTestBinding()
		fixtureSelector := analyticalImportTargetKey{TargetKind: "fixture_target", ExtensionProfileID: "fixture_profile"}
		fixtureTarget := analyticalImportTargets[analyticalImportTargetKey{
			TargetKind:         binding.TargetKind,
			ExtensionProfileID: binding.ExtensionProfileID,
		}]
		fixtureTarget.TargetKind = fixtureSelector.TargetKind
		fixtureTarget.ExtensionProfileID = fixtureSelector.ExtensionProfileID
		analyticalImportTargets[fixtureSelector] = fixtureTarget
		defer delete(analyticalImportTargets, fixtureSelector)

		fixtureBinding := binding
		fixtureBinding.TargetKind = fixtureSelector.TargetKind
		fixtureBinding.ExtensionProfileID = fixtureSelector.ExtensionProfileID
		dependencies := validModuleTestDependencies()
		dependencies.ExtensionImportFacades = []ExtensionImportFacade{
			&moduleTestFacade{binding: binding},
			&moduleTestFacade{binding: fixtureBinding},
		}
		if _, err := NewModule(dependencies); err == nil {
			t.Fatal("duplicate facade id was accepted")
		}
	})
}

func TestImportsModuleRequiresClaimedFacadeAndPermitsUnclaimedOmission(t *testing.T) {
	claimed := validModuleTestDependencies()
	claimed.ExtensionImportFacades = nil
	if _, err := NewModule(claimed); err == nil {
		t.Fatal("claimed analytical target omitted its facade")
	}

	unclaimed := validModuleTestDependencies()
	unclaimed.ExtensionProfileAdmission = func(string) bool { return false }
	unclaimed.ExtensionImportFacades = nil
	if _, err := NewModule(unclaimed); err != nil {
		t.Fatalf("known unclaimed analytical target required a facade: %v", err)
	}
}

func TestImportsModuleConstructsWithoutSideEffectsAndRegistersExactSurface(t *testing.T) {
	dependencies := validModuleTestDependencies()
	runner := dependencies.JobRunner.(*moduleTestRunner)
	now := time.Date(2026, 8, 30, 12, 0, 0, 0, time.UTC)
	dependencies.Now = func() time.Time { return now }
	module, err := NewModule(dependencies)
	if err != nil {
		t.Fatalf("construct Imports module: %v", err)
	}
	if len(runner.registered) != 0 {
		t.Fatalf("construction registered workers: %v", runner.registered)
	}
	if module.service.cursorCodec == nil || !module.service.now().Equal(now) {
		t.Fatal("cursor fallback or injected clock was not installed")
	}
	registry, err := httpapi.NewRouteRegistry(nil)
	if err != nil {
		t.Fatalf("construct public route registry: %v", err)
	}
	if err := module.RegisterRoutes()(http.NewServeMux(), httpapi.DependencySet{PublicRoutes: registry}); err != nil {
		t.Fatalf("register Imports routes: %v", err)
	}
	if operations := registry.Snapshot(); len(operations) != 11 {
		t.Fatalf("bound Imports operations = %d, want 11", len(operations))
	}
	if err := module.RegisterWorkers(); err != nil {
		t.Fatalf("register Imports workers: %v", err)
	}
	if len(runner.registered) != 2 || runner.registered[0] != importDiscoveryJobHandlerName || runner.registered[1] != importApplyJobHandlerName {
		t.Fatalf("registered Imports workers = %v", runner.registered)
	}
}

func TestImportsModuleRejectsInvalidAuthMaterialAndWorkerRegistrationFailure(t *testing.T) {
	invalidAuth := validModuleTestDependencies()
	invalidAuth.Env = map[string]string{authn.AuthMasterKeyEnv: "not-base64"}
	if _, err := NewModule(invalidAuth); err == nil {
		t.Fatal("invalid authentication key material was accepted")
	}

	registrationFailure := validModuleTestDependencies()
	runner := registrationFailure.JobRunner.(*moduleTestRunner)
	runner.failOn = importApplyJobHandlerName
	module, err := NewModule(registrationFailure)
	if err != nil {
		t.Fatalf("construct module for registration failure: %v", err)
	}
	if err := module.RegisterWorkers(); err == nil {
		t.Fatal("worker registration failure was ignored")
	}
}
